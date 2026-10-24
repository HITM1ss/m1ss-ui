#!/usr/bin/env bash
# Invoked by the panel with validated release identifiers, never by a web shell.
set -Eeuo pipefail
umask 077

binary=$1 db=$2 target=$3 previous=$4 arch=$5 manager=$6 started=$7
job=$(cd -- "$(dirname -- "$0")" && pwd)
root=$(dirname -- "$job")
job_name=$(basename -- "$job")
export SUI_DB_FOLDER="$(dirname -- "$db")"
cd -- "$(dirname -- "$binary")"
exec >>"$job/worker.log" 2>&1
# A delayed worker must not replace the state of a newer retry.
grep -Fq "\"job\":\"$job_name\"" "$root/status.json" || exit 1

state() {
    local phase=$1
    printf '{"phase":"%s","targetVersion":"%s","previousVersion":"%s","startedAt":%s,"updatedAt":%s,"pid":%s,"job":"%s"}\n' \
        "$phase" "$target" "$previous" "$started" "$(date +%s)" "$$" "$job_name" >"$root/status-$$.tmp"
    mv -f -- "$root/status-$$.tmp" "$root/status.json"
}

service() {
    if [[ "$manager" == systemd ]]; then systemctl "$1" s-ui
    # OpenRC otherwise reports an already-stopped service as a failure.
    elif [[ "$1" == stop ]]; then rc-service --ifstarted s-ui stop
    else rc-service s-ui "$1"; fi
}

wait_ready() {
    local expected=$1 version pid stable=0
    for ((i = 0; i < 45; i++)); do
        if [[ -f "$root/ready" ]] && read -r version pid <"$root/ready" \
            && [[ "$version" == "$expected" && "$pid" =~ ^[1-9][0-9]*$ ]] \
            && kill -0 "$pid" 2>/dev/null && service status >/dev/null 2>&1; then
            stable=$((stable + 1))
            [[ $stable -ge 3 ]] && return 0
        else stable=0; fi
        sleep 2
    done
    return 1
}

stopped=0
backed_up=0
on_exit() {
    local code=$?
    trap - EXIT HUP INT TERM
    # Release download space before restoring the old database and executable.
    rm -f -- "$job/archive.tar.gz" "$job/SHA256SUMS"
    rm -rf -- "$job/stage"
    if [[ $code -ne 0 ]]; then
        if [[ $stopped == 1 && $backed_up == 1 ]]; then
            state rollingBack
            # Never restore database pages under a still-running writer.
            if service stop; then
                if cp -p -- "$job/backup/sui" "$binary.rollback" \
                    && mv -f -- "$binary.rollback" "$binary" \
                    && cp -p -- "$job/backup/database" "$db.rollback" \
                    && mv -f -- "$db.rollback" "$db" \
                    && rm -f -- "$db-wal" "$db-shm"; then
                    if [[ -f "$job/backup/database-wal" ]]; then
                        cp -p -- "$job/backup/database-wal" "$db-wal" || { state failed; exit 1; }
                    fi
                    rm -f -- "$root/ready"
                    if service start && wait_ready "$previous"; then state rolledBack
                    else state failed; fi
                else state failed; fi
            else state failed; fi
        elif [[ $stopped == 1 ]]; then
            # Downloaded bytes have not been installed; recover the old service.
            service start || true
            state failed
        else state failed; fi
    fi
    # Failed jobs retain their logs and backups for recovery. Downloads can be
    # large; remove them in either outcome, without touching retained backups.
    exit "$code"
}
trap on_exit EXIT
trap 'exit 1' HUP INT TERM

state downloading
base="https://github.com/HITM1ss/m1ss-ui/releases/download/v$target"
asset="s-ui-linux-$arch.tar.gz"
curl --fail --location --silent --show-error --retry 2 --connect-timeout 15 --max-time 600 \
    --max-filesize 536870912 --proto '=https' --proto-redir '=https' -o "$job/archive.tar.gz" "$base/$asset"
curl --fail --location --silent --show-error --retry 2 --connect-timeout 15 --max-time 60 \
    --max-filesize 65536 --proto '=https' --proto-redir '=https' -o "$job/SHA256SUMS" "$base/SHA256SUMS"

state verifying
expected=$(awk -v f="$asset" '$2 == f || $2 == "*" f { print $1; exit }' "$job/SHA256SUMS")
[[ "$expected" =~ ^[a-fA-F0-9]{64}$ ]] || exit 1
actual=$(sha256sum "$job/archive.tar.gz")
[[ "${actual%% *}" == "$(printf '%s' "$expected" | tr '[:upper:]' '[:lower:]')" ]] || exit 1
mkdir -m 700 -- "$job/stage"
tar -xzof "$job/archive.tar.gz" -C "$job/stage" s-ui/sui
candidate="$job/stage/s-ui/sui"
[[ -f "$candidate" && ! -L "$candidate" ]] || exit 1
chmod 755 "$candidate"

# No concurrent Go process while the panel is serving: download and extract
# first, then stop before probing or migrating the new binary on a small VPS.
state backingUp
stopped=1
service stop
mkdir -m 700 -- "$job/backup"
cp -p -- "$binary" "$job/backup/sui"
cp -p -- "$db" "$job/backup/database"
if [[ -f "$db-wal" ]]; then cp -p -- "$db-wal" "$job/backup/database-wal"; fi
backed_up=1
version_output=$(timeout 30 "$candidate" -v)
version=$(printf '%s\n' "$version_output" | awk '$1 == "S-UI" && $2 == "Panel" { print $3; exit }')
[[ "$version" == "$target" ]] || exit 1

state installing
mv -f -- "$candidate" "$binary"
timeout 120 "$binary" migrate
state restarting
rm -f -- "$root/ready"
service start
wait_ready "$target"

# Keep this rollback snapshot, replacing older snapshots only after success.
for old in "$root"/job-*; do
    [[ "$old" == "$job" || ! -d "$old" ]] || rm -rf -- "$old"
done
state succeeded
