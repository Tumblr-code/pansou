# PanSou installation and operations

[简体中文](OPERATIONS.zh-CN.md) · [Home](../README.en.md)

## Requirements and installation

Target: Linux amd64, Ubuntu 24.04/systemd, with `ca-certificates`, `iptables` including IPv6, `curl`, Python 3, and `tar`. Download and verify all three Release assets first. Run this fresh-install example inside the extracted package. Existing installations must follow the upgrade section instead of replacing private configuration with a template.

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

Release directories stay read-only to the service account. Cache writes go to `/var/lib/pansou/cache`. The template uses port 8888; the port guard rejects non-loopback IPv4/IPv6 input. Changing the port requires changing the guard too. This guard does not manage Docker forwarding: use the `127.0.0.1:8888:8888` mapping in `deploy/compose.example.yml` for containers.

## Configuration

`CHANNELS` and `ENABLED_PLUGINS` are explicit comma-separated lists; examples are not a copy of production configuration. Keep `CACHE_PATH` inside the permitted writable directory. Configure proxies with `PROXY` or `HTTP_PROXY/HTTPS_PROXY`. For authentication, privately configure `AUTH_ENABLED=true`, `AUTH_USERS`, and a stable `AUTH_JWT_SECRET`. Never put their real values in documentation, shell history, issues, or images. Preserve the existing API rather than inventing another proxy authentication protocol.

## Verification and troubleshooting

```bash
systemctl is-active pansou-native.service pansou-port-guard.service
systemctl show -p MainPID,NRestarts pansou-native.service
curl --fail --silent http://127.0.0.1:8888/api/health
journalctl -u pansou-native.service -n 50 --no-pager
```

- Connection failure: inspect the service, port, proxy and local health response. Do not remove the firewall boundary to expose the port.
- Healthy but empty results: inspect enabled sources, upstream connectivity and cache. Process health does not prove all providers work.
- Cache permission errors: align `CACHE_PATH`, directory ownership, and systemd `ReadWritePaths`.
- Authentication errors: check private token/configuration consistency without logging credentials.

Rollout acceptance also requires a separate host to confirm port 8888 is inaccessible and an explicitly authorized real query. CI only performs isolated health checks; it does not search third-party providers.

## Upgrade and rollback

The Bot may remain running if it handles the brief backend outage. Verify and unpack the candidate into a new root-owned release directory. Save the previous `current` target, private environment file and units. Stop only the backend, atomically replace `current` using a temporary symlink and `mv -Tf`, then start and verify health/logs. On failure, restore the previous release and matching configuration. Do not mix unknown cache formats, overwrite an executing binary, or follow `latest` automatically.

## Backup and restore

Cache is not a customer ledger, but it may contain query content. Stop the backend before archiving `/etc/pansou`, related systemd units, `current/previous` targets and `/var/lib/pansou/cache`. Use an external 0700 backup directory, 0600 files, encryption and off-host storage; never upload backups to GitHub. Record revision, time and checksums. Test recovery on an isolated machine with the same release and production egress disabled, restoring ownership and validating only local health. Schedule production recovery separately. Cache can be rebuilt under the retention policy; copying cache is not equivalent to a business-data backup.

## Release boundary

`v1.0.0` fixes the verified production source baseline. Main/PR CI only tests and builds. A final Release uploads packages and a fixed image tag; it never updates `latest`. Installation and service changes remain separate operator actions; this documentation/release task did not alter the VM.
