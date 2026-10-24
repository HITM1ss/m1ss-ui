package updater

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
)

func fixtureFile(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
}

func TestWorkerInstallAndRollback(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the service updater runs on Linux; these fixtures also run on macOS")
	}
	for _, scenario := range []string{
		"success", "checksum", "missing-checksum", "version-mismatch", "migration", "health", "rollback-stop",
		"start-failure", "signal-hup", "signal-int", "signal-term", "signal-kill",
		"download-failure", "initial-stop-failure",
	} {
		t.Run(scenario, func(t *testing.T) {
			dir := t.TempDir()
			install := filepath.Join(dir, "install with spaces")
			root := filepath.Join(install, ".m1ss-update")
			job := filepath.Join(root, "job-test")
			binary := filepath.Join(install, "sui")
			db := filepath.Join(dir, "custom db", "s-ui.db")
			bin := filepath.Join(dir, "tools")
			const oldBinary = "#!/bin/bash\necho 'S-UI Panel 1.6.6'\n"
			fixtureFile(t, binary, oldBinary, 0o700)
			fixtureFile(t, db, "original database", 0o600)
			fixtureFile(t, db+"-wal", "original WAL", 0o600)
			fixtureFile(t, filepath.Join(root, "status.json"), `{"phase":"queued","job":"job-test"}`, 0o600)
			fixtureFile(t, filepath.Join(job, "worker.sh"), string(worker), 0o700)
			candidate := `#!/bin/bash
if [[ "$1" == -v ]]; then
  if [[ "$SCENARIO" == version-mismatch ]]; then echo 'S-UI Panel 1.6.8'
  else echo 'S-UI Panel 1.6.7'; fi
  exit 0
fi
if [[ "$1" == migrate ]]; then
  printf 'migrated database' >"$SUI_DB_FOLDER/s-ui.db"
  printf 'migrated WAL' >"$SUI_DB_FOLDER/s-ui.db-wal"
  # The timeout fixture execs this process, so PPID is the actual worker.
  case "$SCENARIO" in
    signal-hup) kill -HUP "$PPID";;
    signal-int) kill -INT "$PPID";;
    signal-term) kill -TERM "$PPID";;
    signal-kill) kill -KILL "$PPID";;
  esac
  [[ "$SCENARIO" != migration ]]
fi
`
			archive := filepath.Join(dir, "release.tar.gz")
			f, err := os.Create(archive)
			if err != nil {
				t.Fatal(err)
			}
			gz := gzip.NewWriter(f)
			tw := tar.NewWriter(gz)
			if err := tw.WriteHeader(&tar.Header{Name: "s-ui/sui", Mode: 0o755, Size: int64(len(candidate))}); err != nil {
				t.Fatal(err)
			}
			if _, err := tw.Write([]byte(candidate)); err != nil {
				t.Fatal(err)
			}
			if err := tw.Close(); err != nil {
				t.Fatal(err)
			}
			if err := gz.Close(); err != nil {
				t.Fatal(err)
			}
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(archive)
			if err != nil {
				t.Fatal(err)
			}
			sums := fmt.Sprintf("%x  s-ui-linux-amd64.tar.gz\n", sha256.Sum256(data))
			if scenario == "checksum" {
				sums = strings.Repeat("0", 64) + "  s-ui-linux-amd64.tar.gz\n"
			}
			if scenario == "missing-checksum" {
				sums = strings.Repeat("0", 64) + "  another-file\n"
			}
			fixtureFile(t, filepath.Join(dir, "sums"), sums, 0o600)
			fixtureFile(t, filepath.Join(bin, "curl"), `#!/bin/bash
while [[ $# -gt 0 ]]; do
  if [[ "$1" == -o ]]; then output=$2; shift 2; else url=$1; shift; fi
done
if [[ "$SCENARIO" == download-failure ]]; then
  printf 'partial download' >"$output"
  echo 'fixture download failed' >&2
  exit 22
fi
case "$url" in
  */SHA256SUMS) cp "$FIXTURE/sums" "$output";;
  */s-ui-linux-amd64.tar.gz) cp "$FIXTURE/release.tar.gz" "$output";;
  *) exit 1;;
esac
`, 0o700)
			fixtureFile(t, filepath.Join(bin, "sleep"), "#!/bin/bash\nexit 0\n", 0o700)
			fixtureFile(t, filepath.Join(bin, "timeout"), "#!/bin/bash\nshift\nexec \"$@\"\n", 0o700)
			hashTool, err := exec.LookPath("sha256sum")
			hashArgs := ""
			if err != nil {
				hashTool, err = exec.LookPath("shasum")
				hashArgs = "-a 256"
			}
			if err != nil {
				t.Fatal("a SHA256 tool is required for the worker integration fixtures")
			}
			fixtureFile(t, filepath.Join(bin, "sha256sum"), fmt.Sprintf("#!/bin/bash\nexec %q %s \"$@\"\n", hashTool, hashArgs), 0o700)
			fixtureFile(t, filepath.Join(bin, "rc-service"), `#!/bin/bash
root="$INSTALL/.m1ss-update"
if_started=false
if [[ "${1:-}" == --ifstarted ]]; then if_started=true; shift; fi
[[ $# == 2 && "$1" == s-ui ]] || exit 2
if [[ "$if_started" == true && ! -f "$FIXTURE/started" ]]; then
  echo skipped >>"$FIXTURE/service.log"
  exit 0
fi
case "$2" in
 stop)
  echo stop >>"$FIXTURE/service.log"
  # OpenRC reports an error for a repeated stop unless --ifstarted skipped it.
  [[ -f "$FIXTURE/started" ]] || exit 1
  if [[ "$SCENARIO" == initial-stop-failure ]]; then
    echo 'fixture initial stop failed' >&2
    exit 1
  fi
  if [[ "$SCENARIO" == rollback-stop && -f "$FIXTURE/new-started" ]]; then exit 1; fi
  kill "$(cat "$FIXTURE/pid")" || exit 1
  rm -f "$FIXTURE/pid" "$FIXTURE/started";;
 start)
  echo start >>"$FIXTURE/service.log"
  # OpenRC succeeds without launching a duplicate of an already started service.
  [[ ! -f "$FIXTURE/started" ]] || exit 0
  version=$("$INSTALL/sui" -v | awk '{print $3}')
  if [[ "$SCENARIO" == start-failure && "$version" == 1.6.7 ]]; then exit 1; fi
  /bin/sleep 120 &
  pid=$!
  echo "$pid" >"$FIXTURE/pid"
  echo "$pid" >>"$FIXTURE/pids"
  touch "$FIXTURE/started"
  if [[ "$version" == 1.6.7 ]]; then touch "$FIXTURE/new-started"; fi
  if [[ "$version" == 1.6.6 || ( "$SCENARIO" != health && "$SCENARIO" != rollback-stop ) ]]; then
    echo "$version $pid" >"$root/ready"
  fi;;
 status) [[ -f "$FIXTURE/started" && -f "$FIXTURE/pid" ]] && kill -0 "$(cat "$FIXTURE/pid")" 2>/dev/null;;
 *) exit 2;;
esac
`, 0o700)
			initialService := exec.Command("/bin/sleep", "120")
			if err := initialService.Start(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				_ = initialService.Process.Kill()
				_ = initialService.Wait()
				if pids, err := os.ReadFile(filepath.Join(dir, "pids")); err == nil {
					for _, pid := range strings.Fields(string(pids)) {
						if n, err := strconv.Atoi(pid); err == nil {
							if p, err := os.FindProcess(n); err == nil {
								_ = p.Kill()
							}
						}
					}
				}
			})
			fixtureFile(t, filepath.Join(dir, "pid"), strconv.Itoa(initialService.Process.Pid), 0o600)
			fixtureFile(t, filepath.Join(dir, "started"), "", 0o600)
			command := exec.Command("bash", "-x", filepath.Join(job, "worker.sh"), binary, db, "1.6.7", "1.6.6", "amd64", "openrc", "1")
			command.Env = append(os.Environ(), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"), "FIXTURE="+dir, "INSTALL="+install, "SCENARIO="+scenario)
			output, runErr := command.CombinedOutput()
			stateData, err := os.ReadFile(filepath.Join(root, "status.json"))
			if err != nil {
				t.Fatal(err)
			}
			var state State
			if err := json.Unmarshal(stateData, &state); err != nil {
				t.Fatal(err)
			}
			want := "succeeded"
			if scenario == "version-mismatch" || scenario == "migration" || scenario == "health" || scenario == "start-failure" ||
				scenario == "signal-hup" || scenario == "signal-int" || scenario == "signal-term" {
				want = "rolledBack"
			}
			if scenario == "checksum" || scenario == "missing-checksum" || scenario == "rollback-stop" ||
				scenario == "download-failure" || scenario == "initial-stop-failure" {
				want = "failed"
			}
			if scenario == "signal-kill" {
				want = "installing"
			}
			if state.Phase != want || (scenario == "success") != (runErr == nil) {
				log, _ := os.ReadFile(filepath.Join(job, "worker.log"))
				t.Fatalf("phase=%s want=%s err=%v\n%s\n%s", state.Phase, want, runErr, output, log)
			}
			expectedBinary, expectedDB, expectedWAL := oldBinary, "original database", "original WAL"
			if scenario == "success" || scenario == "rollback-stop" || scenario == "signal-kill" {
				expectedBinary, expectedDB, expectedWAL = candidate, "migrated database", "migrated WAL"
			}
			for path, expected := range map[string]string{binary: expectedBinary, db: expectedDB, db + "-wal": expectedWAL} {
				got, err := os.ReadFile(path)
				if err != nil || string(got) != expected {
					t.Errorf("%s: got %q, want %q, err=%v", path, got, expected, err)
				}
			}
			beforeBackupFailure := scenario == "checksum" || scenario == "missing-checksum" ||
				scenario == "download-failure" || scenario == "initial-stop-failure"
			if beforeBackupFailure {
				if _, err := os.Stat(filepath.Join(job, "backup")); !os.IsNotExist(err) {
					t.Errorf("created a backup before verification and stopping succeeded: %v", err)
				}
			} else {
				for name, expected := range map[string]string{"sui": oldBinary, "database": "original database", "database-wal": "original WAL"} {
					got, err := os.ReadFile(filepath.Join(job, "backup", name))
					if err != nil || string(got) != expected {
						t.Errorf("backup %s not retained: %q %v", name, got, err)
					}
				}
			}
			if scenario == "signal-kill" {
				result, ok := command.ProcessState.Sys().(syscall.WaitStatus)
				if !ok || !result.Signaled() || result.Signal() != syscall.SIGKILL {
					t.Errorf("worker was not killed by SIGKILL: %v", runErr)
				}
				for _, name := range []string{"pid", "started"} {
					if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
						t.Errorf("service unexpectedly recovered after SIGKILL: %s, err=%v", name, err)
					}
				}
				return // SIGKILL cannot execute the worker's recovery or cleanup trap.
			}
			pid, err := os.ReadFile(filepath.Join(dir, "pid"))
			if err != nil {
				t.Fatal("service was not recovered:", err)
			}
			n, err := strconv.Atoi(strings.TrimSpace(string(pid)))
			if err != nil {
				t.Fatal(err)
			}
			process, err := os.FindProcess(n)
			if err != nil || process.Signal(syscall.Signal(0)) != nil {
				t.Error("service process is not alive")
			}
			if scenario == "rollback-stop" && n == initialService.Process.Pid {
				t.Error("rollback-stop did not exercise the running new service")
			}
			if beforeBackupFailure && n != initialService.Process.Pid {
				t.Error("replaced the original service process before installation")
			}
			if scenario == "download-failure" || scenario == "initial-stop-failure" {
				log, err := os.ReadFile(filepath.Join(job, "worker.log"))
				failure := "fixture download failed"
				if scenario == "initial-stop-failure" {
					failure = "fixture initial stop failed"
				}
				if err != nil || !strings.Contains(string(log), failure) {
					t.Errorf("failure log not retained: %q %v", log, err)
				}
				if _, err := os.Stat(filepath.Join(dir, "pids")); !os.IsNotExist(err) {
					t.Errorf("launched an additional service after an early failure: %v", err)
				}
			}
			if scenario == "initial-stop-failure" {
				commands, err := os.ReadFile(filepath.Join(dir, "service.log"))
				if err != nil || string(commands) != "stop\nstart\n" {
					t.Errorf("unexpected early-stop recovery commands: %q %v", commands, err)
				}
			} else if beforeBackupFailure {
				if _, err := os.Stat(filepath.Join(dir, "service.log")); !os.IsNotExist(err) {
					t.Error("stopped service before download verification succeeded")
				}
			}
			if _, err := os.Stat(filepath.Join(job, "archive.tar.gz")); !os.IsNotExist(err) {
				t.Error("large archive was not cleaned up")
			}
		})
	}
}
