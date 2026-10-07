package jupansou

import (
	"bufio"
	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"pansou/model"
	"pansou/plugin"
	"pansou/util/json"
)

const (
	jupansouPluginName      = "jupansou"
	jupansouAPIURL          = "https://pan.dyuzi.com/api/other/web_search?title=%s&is_type=all&is_show=1&skip_check=1&max=120"
	jupansouDefaultPriority = 3
	jupansouTimeout         = 20 * time.Second
	jupansouMaxRetries      = 3
)

type JuPansouPlugin struct {
	*plugin.BaseAsyncPlugin
	client *http.Client
}

type juPansouStreamItem struct {
	Title  string `json:"title"`
	URL    string `json:"url"`
	IsType int    `json:"is_type"`
}

func init() {
	plugin.RegisterGlobalPlugin(NewJuPansouPlugin())
}

func NewJuPansouPlugin() *JuPansouPlugin {
	return &JuPansouPlugin{
		BaseAsyncPlugin: plugin.NewBaseAsyncPlugin(jupansouPluginName, jupansouDefaultPriority),
		client: &http.Client{
			Timeout: jupansouTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        32,
				MaxIdleConnsPerHost: 8,
				MaxConnsPerHost:     16,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func (p *JuPansouPlugin) Search(keyword string, ext map[string]interface{}) ([]model.SearchResult, error) {
	result, err := p.SearchWithResult(keyword, ext)
	if err != nil {
		return nil, err
	}
	return result.Results, nil
}

func (p *JuPansouPlugin) SearchWithResult(keyword string, ext map[string]interface{}) (model.PluginSearchResult, error) {
	return p.AsyncSearchWithResult(keyword, p.searchImpl, p.MainCacheKey, ext)
}

func (p *JuPansouPlugin) searchImpl(client *http.Client, keyword string, ext map[string]interface{}) ([]model.SearchResult, error) {
	if p.client != nil {
		client = p.client
	}

	searchURL := fmt.Sprintf(jupansouAPIURL, url.QueryEscape(keyword))
	ctx, cancel := context.WithTimeout(context.Background(), jupansouTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("[%s] 创建请求失败: %w", p.Name(), err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/event-stream,application/json,text/plain,*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://pan.dyuzi.com/")

	resp, err := doJuPansouRequestWithRetry(req, client)
	if err != nil {
		return nil, fmt.Errorf("[%s] 请求失败: %w", p.Name(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[%s] 接口返回状态码: %d", p.Name(), resp.StatusCode)
	}

	results := make([]model.SearchResult, 0)
	seen := make(map[string]struct{})

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		payload := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
		if payload == "" || payload == "[DONE]" {
			continue
		}

		var item juPansouStreamItem
		if err := json.Unmarshal([]byte(payload), &item); err != nil {
			continue
		}
		item.Title = strings.TrimSpace(item.Title)
		item.URL = strings.TrimSpace(item.URL)
		if item.Title == "" || item.URL == "" {
			continue
		}
		if _, ok := seen[item.URL]; ok {
			continue
		}

		linkType := mapJuPansouLinkType(item.IsType, item.URL)
		seen[item.URL] = struct{}{}
		sum := md5.Sum([]byte(item.URL))

		results = append(results, model.SearchResult{
			UniqueID: fmt.Sprintf("%s-%x", p.Name(), sum),
			Title:    item.Title,
			Content:  "来源: 剧盘搜",
			Channel:  "",
			Datetime: time.Now(),
			Tags:     []string{linkType},
			Links: []model.Link{
				{
					Type:     linkType,
					URL:      item.URL,
					Password: extractJuPansouPassword(item.URL),
				},
			},
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("[%s] 读取流式结果失败: %w", p.Name(), err)
	}

	return plugin.FilterResultsByKeyword(results, keyword), nil
}

func mapJuPansouLinkType(isType int, rawURL string) string {
	switch isType {
	case 0:
		return "quark"
	case 1:
		return "aliyun"
	case 2:
		return "baidu"
	case 3:
		return "uc"
	case 4:
		return "xunlei"
	default:
		urlValue := strings.ToLower(rawURL)
		switch {
		case strings.Contains(urlValue, "pan.quark.cn"):
			return "quark"
		case strings.Contains(urlValue, "pan.baidu.com"):
			return "baidu"
		case strings.Contains(urlValue, "alipan.com"), strings.Contains(urlValue, "aliyundrive.com"):
			return "aliyun"
		case strings.Contains(urlValue, "drive.uc.cn"):
			return "uc"
		case strings.Contains(urlValue, "pan.xunlei.com"):
			return "xunlei"
		default:
			return "other"
		}
	}
}

func extractJuPansouPassword(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	for _, key := range []string{"pwd", "code", "passcode"} {
		if value := strings.TrimSpace(u.Query().Get(key)); value != "" {
			return value
		}
	}
	return ""
}

func doJuPansouRequestWithRetry(req *http.Request, client *http.Client) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt < jupansouMaxRetries; attempt++ {
		resp, err := client.Do(req.Clone(req.Context()))
		if err == nil && resp.StatusCode == http.StatusOK {
			return resp, nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		lastErr = err
		if attempt < jupansouMaxRetries-1 {
			time.Sleep(200 * time.Millisecond * time.Duration(1<<attempt))
		}
	}
	return nil, fmt.Errorf("重试 %d 次后失败: %w", jupansouMaxRetries, lastErr)
}
