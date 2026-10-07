# PanSou · production-source release

[简体中文](README.md) · [Operations](docs/OPERATIONS.en.md) · [Contributing](CONTRIBUTING.md) · [Security](SECURITY.md)

PanSou searches Telegram channels and resource plugins, groups results by storage provider, and uses a two-level cache. The upstream project is [fish2018/pansou](https://github.com/fish2018/pansou). Its **MIT license and author attribution are retained**. This repository provides the API backend, not a web frontend.

## Source of v1.0.0

This release restores the deployed source revision `7f7991654adbf72a9652b9a4fd92a4a6c5d4df3c` and adds packaging and bilingual operations material. A read-only comparison on 2026-10-07 matched all 139 inspected source/reference files to that historical commit. [PRODUCTION_SOURCE.json](docs/PRODUCTION_SOURCE.json) records runtime file hashes. Later development from the previous main remains in Git history; no history was force-rewritten.

Publishing on GitHub does not deploy or restart production. The distributed executable is rebuilt from verified source; it is **not claimed to be byte-identical to the existing production binary**. Go 1.24.9 is the build toolchain for this production baseline. Dependency upgrades require separate validation and release work.

## Download and verify

Download `pansou-v1.0.0-linux-amd64.tar.gz`, `PROVENANCE.json`, and `SHA256SUMS` from [Releases](https://github.com/Tumblr-code/pansou/releases).

```bash
sha256sum -c SHA256SUMS
tar -xzf pansou-v1.0.0-linux-amd64.tar.gz
cd pansou-v1.0.0-linux-amd64
```

Follow the [operations guide](docs/OPERATIONS.en.md) for native installation, service users, systemd, port protection, and cache permissions. The binary listens on all interfaces and authentication defaults to disabled. Restrict access with the supplied port guard or a container loopback mapping before starting it.

## API and configuration

- `GET /api/health`: process health and configured channel/plugin information.
- `GET/POST /api/search`: existing search API; use `kw` in JSON requests.
- `POST /api/check/links`: link checks; authentication endpoints are under `/api/auth/*`.
- Configure `PORT`, `CHANNELS`, `ENABLED_PLUGINS`, `CACHE_PATH`, proxies, and `AUTH_*` through the process environment. [pansou.env.example](deploy/pansou.env.example) contains no real credentials.
- When authentication is enabled, obtain a Bearer token through the existing API. Store credentials in a private environment file, not command arguments or Git.

The historical upstream configuration reference is preserved in [UPSTREAM_README.zh-CN.md](UPSTREAM_README.zh-CN.md). Its upstream image examples are separate from this distribution. Use the fixed `v1.0.0` asset or `ghcr.io/tumblr-code/pansou:v1.0.0`; main pushes do not publish images or update `latest`.

## Development and packaging

```bash
go version                    # go1.24.9
python3 scripts/verify_production_source.py
go mod verify
go test ./...
go vet ./...
python3 scripts/package_release.py --version v1.0.0
```

The Linux amd64 package includes the executable, license, templates, bilingual guides, and provenance. CI checks tests, source hashes, Docker builds, and release packaging. Only a final GitHub Release publishes the fixed image tag. Offline CI does not certify live provider searches, authentication integration, or production rollout.
