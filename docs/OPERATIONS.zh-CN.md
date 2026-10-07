# PanSou 安装与运维

[English](OPERATIONS.en.md) · [首页](../README.md)

## 前提与安装

适用 Linux amd64、Ubuntu 24.04/systemd；需 `ca-certificates`、`iptables`（含 IPv6）、`curl`、Python 3 和 `tar`。先下载并校验三个 Release 附件。在解压目录执行以下新装示例；已有安装先阅读升级章节，不能用模板覆盖现有配置。

```bash
sudo useradd --system --home /nonexistent --shell /usr/sbin/nologin pansou
sudo install -d -o root -g root -m 0755 /opt/pansou/releases
sudo install -d -o root -g pansou -m 0750 /etc/pansou
sudo install -d -o pansou -g pansou -m 0700 /var/lib/pansou/cache
sudo cp -a . /opt/pansou/releases/v1.0.0
sudo chown -R root:root /opt/pansou/releases/v1.0.0
sudo chmod -R go-w /opt/pansou/releases/v1.0.0
sudo install -o root -g pansou -m 0640 deploy/pansou.env.example /etc/pansou/pansou.env
sudoedit /etc/pansou/pansou.env
sudo install -m 0755 deploy/pansou-port-guard.sh /usr/local/sbin/pansou-port-guard
sudo install -m 0644 deploy/systemd/*.service /etc/systemd/system/
sudo ln -s /opt/pansou/releases/v1.0.0 /opt/pansou/current
sudo systemctl daemon-reload
sudo systemctl enable --now pansou-port-guard.service pansou-native.service
```

运行时源码不写入 release。缓存只写 `/var/lib/pansou/cache`。模板默认端口 8888，`pansou-port-guard` 同时阻止非 loopback 的 IPv4/IPv6 入站；更改端口须同步修改保护规则。该规则不管理 Docker 转发，容器请使用 `deploy/compose.example.yml` 的 `127.0.0.1:8888:8888` 映射。

## 配置

`CHANNELS` 和 `ENABLED_PLUGINS` 为显式逗号分隔列表；模板仅提供起点，不代表生产完整列表。`CACHE_PATH` 必须落在服务可写目录。代理由 `PROXY` 或 `HTTP_PROXY/HTTPS_PROXY` 设置。需要认证时在私有文件中配置 `AUTH_ENABLED=true`、`AUTH_USERS` 与稳定的 `AUTH_JWT_SECRET`；不要把真实值写入教程、命令历史、issue 或镜像。保持服务原接口，不增加自定义代理鉴权协议。

## 验证和故障定位

```bash
systemctl is-active pansou-native.service pansou-port-guard.service
systemctl show -p MainPID,NRestarts pansou-native.service
curl --fail --silent http://127.0.0.1:8888/api/health
journalctl -u pansou-native.service -n 50 --no-pager
```

- 连接失败：检查服务、端口、代理和本机健康响应，不能通过解除防火墙暴露端口解决。
- 健康正常但无结果：检查允许的频道/插件、上游可达性和缓存，健康接口不证明每个来源都有效。
- 缓存权限失败：确认 `CACHE_PATH`、目录属主和 systemd `ReadWritePaths` 一致。
- 认证失败：核对 token 与私有认证配置，不在日志中输出它们。

部署验收还应从独立主机验证 8888 不可达，并在另行授权后做一次实际查询。CI 只做隔离健康检查，不请求第三方搜索。

## 升级与回退

停止 Bot 不是后端升级的前提；先确认调用端会处理短暂不可用。校验新包并解压到新的 root 所有 release 目录，保存原 `current`、环境文件和单元副本。停止后端后，用同目录临时链接加 `mv -Tf` 原子切换 `current`，再启动并核验健康及日志；仅改此服务。失败时指回原 release、恢复对应配置并重启，不混用未知版本缓存结构。不要覆盖正在运行的二进制，不自动跟随 `latest`。

## 备份与恢复

缓存不是客户账本，但仍可能含查询内容。停后端后，将 `/etc/pansou`、相关 systemd 单元、`current/previous` 目标和 `/var/lib/pansou/cache` 打包到权限为 0700 的外部备份目录；备份文件 0600，异机加密保存，不上传 GitHub。记录版本、时间和校验和。恢复时先在隔离机器安装同版程序，保持生产出口关闭，恢复配置/缓存及属主，只验证本机健康；通过后再安排生产恢复。缓存可按保留策略重建，不能声称复制缓存等同于业务备份。

## 发布边界

`v1.0.0` 固定生产源码基线。main/PR CI 只测试与构建，正式 Release 才上传包和固定版本镜像；不发布 `latest`。任何安装、升级、回退都由维护者另行执行，本轮文档与发布工作没有修改 VM。
