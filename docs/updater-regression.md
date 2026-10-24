# Panel updater regression verification

Verified on 2026-10-24 against the working tree based on `cce6f9e`, including the updater fixes and regression tests. These results precede a release or deployment.

## Automated checks

| Check | Result |
| --- | --- |
| `go test -race ./updater ./api` | Passed on macOS/arm64 with Go 1.27.2 |
| `go vet ./updater ./api ./app` | Passed |
| `gofmt -l updater api app` | No formatting changes required |
| `bash -n updater/worker.sh` | Passed |
| `npm test` | 7 files, 181 tests passed |
| `npm run build` | Type checking and production build passed |
| ESLint on the 12 changed frontend source/test files | Passed |
| Native Linux updater package tests | Passed on Linux/arm64, including all 14 worker scenarios |

The Linux test binary was built with `GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go test -c -o updater-linux.test ./updater` and executed inside the disposable VM. Race detection ran separately on macOS. See [worker tests](../updater/worker_test.go), [update store tests](../frontend/src/store/modules/update.test.ts), and [late response tests](../frontend/src/plugins/httputil.test.ts).

The worker suite covers normal installation; download failure; invalid or missing checksum; initial service-stop failure; candidate version mismatch; migration failure; startup failure; readiness failure; rollback service-stop failure; HUP, INT, TERM and SIGKILL. In particular, the new initial-stop and download tests check that the original process and binary/database/WAL stay intact, no incomplete backup is created, and the error log is retained.

## Real Linux service managers

Both runs used a temporary ARM64 Linux VM and the original `updater/worker.sh`:

- Debian 13.7, Linux `6.12.111+deb13-arm64`, systemd `257.13-1~deb13u1` running as PID 1. The fixture copied the project's exact `s-ui.service`, including `PrivateTmp`, `NoNewPrivileges` and `ProtectHome`. The worker ran in an independent transient `Type=exec` unit.
- Alpine 3.24.3 with OpenRC 0.63.2 and Bash 5.3.9, in a separate chroot within that VM. The fixture used real `rc-service` and `supervise-daemon`, with the installation's executable, working directory, PID file and supervision settings. The worker ran in a separate session using `setsid`.

| Scenario | systemd | OpenRC | Verified outcome |
| --- | --- | --- | --- |
| Successful installation | Passed | Passed | New version ready; matching rollback snapshot retained |
| Migration failure | Passed | Passed | Previous executable, database and WAL restored; old service ready |
| Candidate version mismatch | Passed | Passed | Previous files restored; old service ready |
| New service startup failure | Passed | Passed | Previous files restored; old service ready |
| New service readiness failure | Passed | Passed | Previous files restored; old service ready |
| Download failure | Passed | Passed | Original PID and files unchanged; no backup created |
| Invalid checksum | Passed | Passed | Original PID and files unchanged; no backup created |
| Missing checksum entry | Passed | Passed | Original PID and files unchanged; no backup created |

Each case checked the actual ready PID against systemd's MainPID or OpenRC's supervisor/child relationship, process liveness, service status, installed version, database/WAL bytes, retained backups and logs, and download cleanup. The original worker SHA256 was `89298a44d25958e2f79a57de26c32a854a609cdc76ad2740e253b0dcb4886a8c`; the systemd unit SHA256 was `c7504e06a662a76552d61392ba5353cc023fd1d62a966f0abbde8693e1838a62`.

Downloads used local archives and checksums. Application executables and database/WAL contents were controlled fixtures; readiness sleeps were shortened. This verifies worker/service integration and backup restoration, rather than upgrading a production panel with a published release or migrating a production SQLite database. The Alpine chroot emitted boot-dependency/cgroup warnings; full Alpine boot and cgroup configuration are outside this run's coverage.

## Browser flows

Real Chromium ran against the local Vue app, with every panel API request intercepted by fixtures. All 12 cases passed with no page exceptions:

- Four lifecycle cases: continue monitoring after SPA navigation, pause/resume global and dashboard data polling, restore a job after page reload and initial disconnection, and clear both pollers on logout. Successful updates refreshed the current route once without a reload loop.
- Five submission-recovery cases: lost response followed by reconnection, an already completed update, an unchanged historical task, login expiry followed by login/recovery, and the 25-minute monitoring deadline followed by read-only reconnection. Each case sent exactly one installation POST.
- Three interrupted-state cases: English, Simplified Chinese and Traditional Chinese recovery guidance at a 390px viewport. The warning displayed SSH/README guidance and no active update progress bar.

Only an existing Vuetify `dense` deprecation warning was observed. The browser flows did not trigger a real updater, service stop or database migration.

## Recovery and cleanup

The README recovery block was separately executed in temporary directories with a custom database path containing spaces: a backup with WAL, a backup without WAL, and a missing-backup case. Restoration preserved backup files and modes, removed stale WAL/SHM as appropriate, and stopped before changes when required backup files were absent.

After the Linux runs, both fixture services, their executable directories and temporary curl/sleep helpers were confirmed absent, with no remaining fixture daemon. The local frontend server was also stopped. Linux result tables, original-source hashes and per-case logs were exported before the disposable VM was shut down. The VM and its dedicated image cache were then deleted.
