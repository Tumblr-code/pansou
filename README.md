# PanSou · 生产源码发行版

[English](README.en.md) · [部署与运维](docs/OPERATIONS.zh-CN.md) · [贡献](CONTRIBUTING.md) · [安全](SECURITY.md)

PanSou 提供 Telegram 频道与插件资源搜索、网盘分类和两级缓存。上游为 [fish2018/pansou](https://github.com/fish2018/pansou)，本仓保留原有 **MIT 许可证及作者声明**，没有 Web 前端。

## v1.0.0 的来源

本版恢复已在线运行的 `7f7991654adbf72a9652b9a4fd92a4a6c5d4df3c` 源码，并增加发布与双语运维材料。2026-10-07 只读核验的 139 个源码及参考文件全部匹配该历史提交；机器可校验的运行源码清单见 [PRODUCTION_SOURCE.json](docs/PRODUCTION_SOURCE.json)。恢复前 main 的后续开发仍保留在 Git 历史，没有强推或删除提交。

GitHub 发布不会部署或重启生产。发行二进制由清单对应源码重新构建，**不宣称与原生产二进制逐字节相同**。Go 1.24.9 为此生产基线的构建工具链；后续依赖升级应单独测试与发布。

## 下载与核验

从 [Releases](https://github.com/Tumblr-code/pansou/releases) 下载 `pansou-v1.0.0-linux-amd64.tar.gz`、`PROVENANCE.json` 与 `SHA256SUMS`。

```bash
sha256sum -c SHA256SUMS
tar -xzf pansou-v1.0.0-linux-amd64.tar.gz
cd pansou-v1.0.0-linux-amd64
```

原生安装、专用账号、systemd、端口保护和缓存权限见[部署教程](docs/OPERATIONS.zh-CN.md)。不要直接将默认无认证的 8888 端口暴露到公网；二进制监听所有接口，由模板中的端口保护或容器 loopback 映射限制入口。

## 接口和配置

- `GET /api/health`：进程健康及当前频道/插件信息。
- `GET/POST /api/search`：现有搜索接口，JSON 参数使用 `kw`。
- `POST /api/check/links`：链接检查；认证接口在 `/api/auth/*`。
- `PORT`、`CHANNELS`、`ENABLED_PLUGINS`、`CACHE_PATH`、代理及 `AUTH_*` 均由运行环境提供。示例见 [pansou.env.example](deploy/pansou.env.example)，无真实配置或凭据。
- 开启认证后按现有 API 获取 Bearer Token；凭据保存在私有环境文件中，不放命令行或 Git。

上游基线的详细配置与协议说明保存在 [UPSTREAM_README.zh-CN.md](UPSTREAM_README.zh-CN.md)。其中上游镜像、示例与当前维护者发行渠道不同；安装本版使用固定 `v1.0.0` 资产或 `ghcr.io/tumblr-code/pansou:v1.0.0`。本仓不在 main 推送时发布镜像，不更新 `latest`。

## 开发与构建

```bash
go version                    # go1.24.9
python3 scripts/verify_production_source.py
go mod verify
go test ./...
go vet ./...
python3 scripts/package_release.py --version v1.0.0
```

Linux amd64 附件包含程序、许可证、模板、双语文档及来源信息。CI 执行测试、来源校验、Docker 构建和发布包检查；仅正式 GitHub Release 才发布固定版本镜像。真实上游搜索、认证集成及生产切换不作为离线 CI 已验收事项。
