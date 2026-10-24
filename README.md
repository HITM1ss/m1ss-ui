# M1ss-UI
**An Advanced Web Panel • Built on SagerNet/Sing-Box**

M1ss-UI is a customized fork of [S-UI](https://github.com/alireza0/s-ui), with a lightweight responsive interface and a bundled frontend. See [自动构建与部署](docs/BUILD-AND-RELEASE.md) for build artifacts, release triggers and installation steps.

![](https://img.shields.io/github/v/release/HITM1ss/m1ss-ui.svg)
[![CI](https://github.com/HITM1ss/m1ss-ui/actions/workflows/ci.yml/badge.svg)](https://github.com/HITM1ss/m1ss-ui/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/alireza0/s-ui)](https://goreportcard.com/report/github.com/alireza0/s-ui)
[![Downloads](https://img.shields.io/github/downloads/HITM1ss/m1ss-ui/total.svg)](https://img.shields.io/github/downloads/HITM1ss/m1ss-ui/total.svg)
[![License](https://img.shields.io/badge/license-GPL%20V3-blue.svg?longCache=true)](https://www.gnu.org/licenses/gpl-3.0.en.html)

> **Disclaimer:** This project is only for personal learning and communication, please do not use it for illegal purposes, please do not use it in a production environment

**If you think this project is helpful to you, you may wish to give a**:star2:

**Want to contribute?** See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, coding conventions, testing, and the pull request process.

[!["Buy Me A Coffee"](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/alireza7)

<a href="https://nowpayments.io/donation/alireza7" target="_blank" rel="noreferrer noopener">
   <img src="https://nowpayments.io/images/embeds/donation-button-white.svg" alt="Crypto donation button by NOWPayments">
</a>

## Quick Overview
| Features                               |      Enable?       |
| -------------------------------------- | :----------------: |
| Multi-Protocol                         | :heavy_check_mark: |
| Multi-Language                         | :heavy_check_mark: |
| Multi-Client/Inbound                   | :heavy_check_mark: |
| Advanced Traffic Routing Interface     | :heavy_check_mark: |
| Client & Traffic & System Status       | :heavy_check_mark: |
| Subscription Link (link/json/clash + info)| :heavy_check_mark: |
| Dark/Light Theme                       | :heavy_check_mark: |
| API Interface                          | :heavy_check_mark: |

## Supported Platforms
| Platform | Architecture | Status |
|----------|--------------|---------|
| Linux    | amd64, arm64, armv7, armv6, armv5, 386, s390x | ✅ Supported |
| Windows  | amd64, 386, arm64 | ✅ Supported |
| macOS    | amd64, arm64 | 🚧 Experimental |

## Screenshots

!["Main"](https://github.com/alireza0/s-ui-frontend/raw/main/media/main.png)

[Other UI Screenshots](https://github.com/alireza0/s-ui-frontend/blob/main/screenshots.md)

## Documentation

Full documentation lives in the [Wiki](https://github.com/alireza0/s-ui/wiki):

| Page | Contents |
|------|----------|
| [Subscription Service](https://github.com/alireza0/s-ui/wiki/Subscription-Service) | Subscription URLs, the three formats, response headers |
| [Subscription JSON Template](https://github.com/alireza0/s-ui/wiki/Subscription-JSON-Template) | Structure and supported keys of the sing-box subscription template |
| [Subscription Clash Template](https://github.com/alireza0/s-ui/wiki/Subscription-Clash-Template) | The Clash.Meta template, proxy groups and filters |
| [API Documentation](https://github.com/alireza0/s-ui/wiki/API-Documentation) | The token-authenticated REST API (`/apiv2`) |
| [Configuration Objects](https://github.com/alireza0/s-ui/wiki/Configuration-Objects) | Shape of the objects read and written through the API |
| [Settings Reference](https://github.com/alireza0/s-ui/wiki/Settings-Reference) | Every panel setting and its default |

## Default Installation Information
- Panel Port: 2095
- Panel Path: /app/
- Subscription Port: 2096
- Subscription Path: /sub/
- User/Password: admin

## Install & Upgrade to Latest Version

### Linux
```sh
bash <(curl -Ls https://raw.githubusercontent.com/HITM1ss/m1ss-ui/main/install.sh)
```

#### Update from the panel

After installing a version with panel updates, open **Settings → Software update** and select **Check for updates**. A newer stable release can be installed with **Update now**; no scheduled checks or upgrades run in the background.

Panel updates support the standard `/usr/local/s-ui` Linux installation managed by systemd or Alpine/OpenRC. Downloads require HTTPS and `SHA256SUMS`. The updater downloads and verifies the archive before stopping the service, backs up the executable and SQLite database (including any WAL), migrates, and checks that the new panel, subscription listener, and core start. Failed migration or startup, and caught HUP/INT/TERM signals, trigger an attempt to restore the previous executable and database. The `rolledBack` result confirms that restoration and startup checks completed; `failed` requires inspection. Maintenance mode is preserved.

Updates briefly interrupt proxy connections. One rollback snapshot and update logs are retained under `/usr/local/s-ui/.m1ss-update/job-*/`; a failed or interrupted update retains its recovery files. Docker installations should pull the new image and recreate the container instead. The first installation of a release containing this feature still uses the installer above.

##### First upgrade from v1.6.8 or earlier

The update worker is embedded in the currently running executable. The first web upgrade to v1.6.9 therefore still uses the old worker; its OpenRC rollback fix takes effect after v1.6.9 has started.

For this first OpenRC upgrade, use SSH:

1. Confirm that no update worker or migration is running, then stop the service as described in the recovery steps below.
2. Back up the existing executable, the database at its original configured path, and any WAL after stopping the service. Keep them together in a manual snapshot folder, using the `backup/sui`, `backup/database` and optional `backup/database-wal` names shown below. Retain this snapshot throughout the upgrade.
3. Run the installer pinned to this release from a root Bash shell. It downloads the versioned archive and checks the published checksum:

   ```bash
   bash <(curl -fsSL https://raw.githubusercontent.com/HITM1ss/m1ss-ui/v1.6.9/install.sh) v1.6.9
   ```

4. Verify the running version, panel, subscription endpoint and proxy behavior. If installation, migration or startup fails, stop the service and restore your manual snapshot using the recovery block below, setting `update_job` to that snapshot folder and `update_db` to the original database path.

Once v1.6.9 is running, subsequent panel updates use the corrected worker.

##### Recover after an interrupted update

A power loss, or SIGKILL of the worker itself (including an OOM kill), bypasses its exit handler. Automatic rollback is not guaranteed in these cases. If the panel is reachable, a subsequent update-status request marks an active job as `interrupted` once its last state change is over 30 seconds old and its worker PID is no longer alive. The panel then asks you to inspect the service and recovery files. If the service is down, use SSH; the webpage cannot diagnose or restore a stopped panel.

The files on a standard installation are:

| File | Purpose |
| --- | --- |
| `/usr/local/s-ui/.m1ss-update/status.json` | Current job directory name, worker PID, phase, target version and previous version |
| `/usr/local/s-ui/.m1ss-update/job-*/worker.log` | That job's download, migration, startup and rollback output |
| `/usr/local/s-ui/.m1ss-update/job-*/backup/sui` | Previous executable |
| `/usr/local/s-ui/.m1ss-update/job-*/backup/database` | Previous SQLite database file |
| `/usr/local/s-ui/.m1ss-update/job-*/backup/database-wal` | Previous SQLite WAL, only if one existed when backed up |

Use a root shell on the affected server and follow these steps:

1. **Identify the job and confirm that it has exited.** Read `status.json` and the matching `worker.log`. Check the PID against the process command and ensure that neither that worker nor its migration child is still running; after a reboot the PID may have been reused. A systemd worker has a separate unit named `m1ss-ui-update-<job>.service`, for example `m1ss-ui-update-job-123.service`. If the worker is still running, allow its rollback to finish or stop that exact job and its children, then inspect the result again before restoring files.

2. **Stop the panel service and confirm it stopped.** For systemd use `systemctl stop s-ui` and inspect `systemctl status s-ui`. For OpenRC use `rc-service --ifstarted s-ui stop` and inspect `rc-service s-ui status`. An already stopped service is expected to report inactive; a stop error with a still-running service must be resolved before continuing. Ensure that no `sui migrate` process remains, since it can write the database independently of the panel service.

3. **Select a complete snapshot and the original database path.** Use the job named in `status.json` and restore the executable, database and optional WAL from that same job. Jobs that reached `installing` or `restarting` have completed the backup copies; an interruption during `backingUp` may leave partial backup files. The presence of a file alone does not prove that the snapshot is complete. During `downloading`, `verifying` or `backingUp`, installation and migration have not started: inspect the existing files and restart the unchanged installation rather than overwriting it with a partial snapshot. If snapshot completeness cannot be established, use an existing known-good backup.

   The default database is `/usr/local/s-ui/db/s-ui.db`. If the service sets `SUI_DB_FOLDER`, the database is `<SUI_DB_FOLDER>/s-ui.db`; relative folders resolve from the service's working directory, normally `/usr/local/s-ui`. Check the original service configuration, including systemd overrides (`systemctl cat s-ui`) or OpenRC's `/etc/init.d/s-ui` and `/etc/conf.d/s-ui`. The backup is always named `database`, so its name does not identify the original path. Keep the service configured to use that original folder.

4. **Restore the stopped installation.** Replace the job placeholder and, if necessary, `update_db` below before running this Bash block. This returns database contents to the snapshot time. The original backup files are kept. Do not run a new migration when restoring the previous executable together with its matching database.

   ```bash
   bash <<'RECOVER'
   set -e
   update_job='/usr/local/s-ui/.m1ss-update/job-REPLACE_WITH_JOB'
   update_binary='/usr/local/s-ui/sui'
   update_db='/usr/local/s-ui/db/s-ui.db'

   test -f "$update_job/backup/sui"
   test -f "$update_job/backup/database"
   cp -p -- "$update_job/backup/sui" "$update_binary.rollback"
   mv -f -- "$update_binary.rollback" "$update_binary"
   cp -p -- "$update_job/backup/database" "$update_db.rollback"
   mv -f -- "$update_db.rollback" "$update_db"
   rm -f -- "$update_db-wal" "$update_db-shm"
   if [[ -f "$update_job/backup/database-wal" ]]; then
       cp -p -- "$update_job/backup/database-wal" "$update_db-wal"
   fi
   rm -f -- /usr/local/s-ui/.m1ss-update/ready
   RECOVER
   ```

   Stop here if any command fails; keep the service stopped while resolving the error. Never combine a database from one snapshot with another snapshot's WAL, and remove a leftover WAL even when the selected backup has no WAL.

5. **Start and verify the previous version.** Use `systemctl start s-ui` and `systemctl status s-ui`, or `rc-service s-ui start` and `rc-service s-ui status`. Check service output with `journalctl -u s-ui -n 100 --no-pager` on systemd, or `/var/log/s-ui.log` on the standard OpenRC installation. Sign in at the configured panel URL and confirm the previous version, subscription endpoint and proxy connections; if maintenance mode was enabled, the core should remain stopped. Keep the job's backups and logs until these checks pass.

Manual restoration does not change the recorded result to `rolledBack`. An `interrupted` or `failed` result remains an incident record; confirm recovery from the running version and service behavior before checking for another update.

#### Installer language

The installer is available in the same six languages as the panel: `en` (default), `fa`, `ru`, `vi`, `zhcn`, `zhtw`. Choose one with the `SUI_LANG` environment variable (when unset, your system `$LANG` is used as a hint):

```sh
SUI_LANG=fa bash <(curl -Ls https://raw.githubusercontent.com/HITM1ss/m1ss-ui/main/install.sh)
```

### Alpine Linux
Alpine uses `apk` and OpenRC instead of `apt`/systemd. The install script detects Alpine automatically and sets up an OpenRC service. Since Alpine has no `bash` by default, install it first:

```sh
apk add bash
bash <(curl -Ls https://raw.githubusercontent.com/HITM1ss/m1ss-ui/main/install.sh)
```

Manage the service with OpenRC: `rc-service s-ui start|stop|restart` and `rc-update add s-ui default`.

### Windows
1. Download the latest Windows release from [GitHub Releases](https://github.com/HITM1ss/m1ss-ui/releases/latest)
2. Extract the ZIP file
3. Run `install-windows.bat` as Administrator
4. Follow the installation wizard

## Install legacy Version

**Step 1:** To install your desired legacy version, add the version to the end of the installation command. e.g., ver `v1.5.0`:

```sh
VERSION=v1.5.0 && bash <(curl -Ls https://raw.githubusercontent.com/HITM1ss/m1ss-ui/$VERSION/install.sh) $VERSION
```

## Manual installation

### Linux
1. Get the latest version of S-UI based on your OS/Architecture from GitHub: [https://github.com/HITM1ss/m1ss-ui/releases/latest](https://github.com/HITM1ss/m1ss-ui/releases/latest)
2. **OPTIONAL** Get the latest version of `s-ui.sh` [https://raw.githubusercontent.com/HITM1ss/m1ss-ui/main/s-ui.sh](https://raw.githubusercontent.com/HITM1ss/m1ss-ui/main/s-ui.sh)
3. **OPTIONAL** Copy `s-ui.sh` to /usr/bin/ and run `chmod +x /usr/bin/s-ui`.
4. Extract s-ui tar.gz file to a directory of your choice and navigate to the directory where you extracted the tar.gz file.
5. Copy *.service files to /etc/systemd/system/ and run `systemctl daemon-reload`.
6. Enable autostart and start S-UI service using `systemctl enable s-ui --now`
7. The panel manages its embedded Sing-Box core; a separate sing-box system service is not required.

### Windows
1. Get the latest Windows version from GitHub: [https://github.com/HITM1ss/m1ss-ui/releases/latest](https://github.com/HITM1ss/m1ss-ui/releases/latest)
2. Download the appropriate Windows package (e.g., `s-ui-windows-amd64.zip`)
3. Extract the ZIP file to a directory of your choice
4. Run `install-windows.bat` as Administrator
5. Follow the installation wizard
6. Access the panel at http://localhost:2095/app

## Uninstall S-UI

### systemd
```sh
sudo -i

systemctl disable s-ui  --now

rm -f /etc/systemd/system/sing-box.service
systemctl daemon-reload

rm -fr /usr/local/s-ui
rm /usr/bin/s-ui
```

### Alpine (OpenRC)
```sh
rc-service s-ui stop
rc-update del s-ui default
rm -f /etc/init.d/s-ui

rm -fr /usr/local/s-ui
rm /usr/bin/s-ui
```

## Install using Docker

<details>
   <summary>Click for details</summary>

### Usage

**Step 1:** Install Docker

```shell
curl -fsSL https://get.docker.com | sh
```

**Step 2:** Install S-UI

> Docker compose method

```shell
mkdir s-ui && cd s-ui
wget -q https://raw.githubusercontent.com/HITM1ss/m1ss-ui/main/docker-compose.yml
docker compose up -d
```

> Use docker

```shell
mkdir s-ui && cd s-ui
docker run -itd \
    -p 2095:2095 -p 2096:2096 -p 443:443 -p 80:80 \
    -v $PWD/db/:/app/db/ \
    -v $PWD/cert/:/app/cert/ \
    --name s-ui --restart=unless-stopped \
    ghcr.io/hitm1ss/m1ss-ui:latest
```

> Build your own image

```shell
git clone https://github.com/HITM1ss/m1ss-ui
cd m1ss-ui
docker build -t s-ui .
```

</details>

## Manual run ( contribution )

<details>
   <summary>Click for details</summary>

### Build and run whole project
```shell
./runSUI.sh
```

### Clone the repository
```shell
# clone repository
git clone https://github.com/HITM1ss/m1ss-ui
cd m1ss-ui
# frontend/ is already included in this repository
```


### - Frontend

Frontend code is included in `frontend/`, derived from [s-ui-frontend](https://github.com/alireza0/s-ui-frontend).

### - Backend
> Please build frontend once before!

To build backend:
```shell
# remove old frontend compiled files
rm -fr web/html/*
# apply new frontend compiled files
cp -R frontend/dist/ web/html/
# build
go build -o sui main.go
```

To run backend (from root folder of repository):
```shell
./sui
```

</details>

## Languages

- English
- Farsi
- Vietnamese
- Chinese (Simplified)
- Chinese (Traditional)
- Russian

## Features

- Supported protocols:
  - General:  Mixed, SOCKS, HTTP, HTTPS, Direct, Redirect, TProxy
  - V2Ray based: VLESS, VMess, Trojan, Shadowsocks
  - Other protocols: ShadowTLS, Hysteria, Hysteria2, Naive, TUIC
- Supports XTLS protocols
- An advanced interface for routing traffic, incorporating PROXY Protocol, External, and Transparent Proxy, SSL Certificate, and Port
- An advanced interface for inbound and outbound configuration
- Clients’ traffic cap and expiration date
- Displays online clients, inbounds and outbounds with traffic statistics, and system status monitoring
- Subscription service with ability to add external links and subscription
- HTTPS for secure access to the web panel and subscription service (self-provided domain + SSL certificate)
- Dark/Light theme

## Environment Variables

<details>
  <summary>Click for details</summary>

### Usage

| Variable       |                      Type                      | Default       |
| -------------- | :--------------------------------------------: | :------------ |
| SUI_LOG_LEVEL  | `"debug"` \| `"info"` \| `"warn"` \| `"error"` | `"info"`      |
| SUI_DEBUG      |                   `boolean`                    | `false`       |
| SUI_BIN_FOLDER |                    `string`                    | `"bin"`       |
| SUI_DB_FOLDER  |                    `string`                    | `"db"`        |
| SINGBOX_API    |                    `string`                    | -             |

</details>

## SSL Certificate

<details>
  <summary>Click for details</summary>

### Certbot

```bash
snap install core; snap refresh core
snap install --classic certbot
ln -s /snap/bin/certbot /usr/bin/certbot

certbot certonly --standalone --register-unsafely-without-email --non-interactive --agree-tos -d <Your Domain Name>
```

</details>

## Third-party Projects

Community-made projects built around S-UI. These are not affiliated with or maintained by S-UI — use them at your own discretion:

- [itning/reset-s-ui-traffic](https://github.com/itning/reset-s-ui-traffic) — periodic traffic reset for all users
- [zqh2333/s-ui-traffic-reset](https://github.com/zqh2333/s-ui-traffic-reset) — traffic reset tool
- [Sownix21/SUI-Bot](https://github.com/Sownix21/SUI-Bot) - telegram bot
- [Android App](https://github.com/Sownix21/SUI-Node) - multi panel controller

> Building something on top of S-UI (a Telegram bot, monitoring, automation, ...)? Open an issue/PR to get it listed here.

## Star History

<a href="https://www.star-history.com/?repos=alireza0%2Fs-ui&type=date&releases=&legend=bottom-right">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=alireza0/s-ui&type=date&theme=dark&legend=bottom-right" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=alireza0/s-ui&type=date&legend=bottom-right" />
   <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=alireza0/s-ui&type=date&legend=bottom-right" />
 </picture>
</a>
