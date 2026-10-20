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
	for _, scenario := range []string{"success", "checksum", "missing-checksum", "migration", "health"} {
		t.Run(scenario, func(t *testing.T) {
			dir := t.TempDir()
			install := filepath.Join(dir, "install with spaces")
			root := filepath.Join(install, ".m1ss-update")
			job := filepath.Join(root, "job-test")
			binary := filepath.Join(install, "sui")
			db := filepath.Join(dir, "custom db", "s-ui.db")
			bin := filepath.Join(dir, "tools")
			fixtureFile(t, binary, "#!/bin/bash\necho 'S-UI Panel 1.6.6'\n", 0o700)
			fixtureFile(t, db, "original database", 0o600)
			fixtureFile(t, db+"-wal", "original WAL", 0o600)
			fixtureFile(t, filepath.Join(root, "status.json"), `{"phase":"queued","job":"job-test"}`, 0o600)
			fixtureFile(t, filepath.Join(job, "worker.sh"), string(worker), 0o700)
			candidate := `#!/bin/bash
if [[ "$1" == -v ]]; then echo 'S-UI Panel 1.6.7'; exit 0; fi
if [[ "$1" == migrate ]]; then
  printf 'migrated database' >"$SUI_DB_FOLDER/s-ui.db"
  printf 'migrated WAL' >"$SUI_DB_FOLDER/s-ui.db-wal"
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
case "$2" in
 stop)
  echo stop >>"$FIXTURE/service.log"
  [[ ! -f "$FIXTURE/pid" ]] || kill "$(cat "$FIXTURE/pid")" 2>/dev/null || true
  rm -f "$FIXTURE/pid";;
 start)
  echo start >>"$FIXTURE/service.log"
  /bin/sleep 120 &
  pid=$!
  echo "$pid" >"$FIXTURE/pid"
  version=$("$INSTALL/sui" -v | awk '{print $3}')
  if [[ "$SCENARIO" != health || "$version" == 1.6.6 ]]; then echo "$version $pid" >"$root/ready"; fi;;
 status) [[ -f "$FIXTURE/pid" ]] && kill -0 "$(cat "$FIXTURE/pid")" 2>/dev/null;;
esac
`, 0o700)
			t.Cleanup(func() {
				if pid, err := os.ReadFile(filepath.Join(dir, "pid")); err == nil {
					if n, err := strconv.Atoi(strings.TrimSpace(string(pid))); err == nil {
						if p, err := os.FindProcess(n); err == nil {
							_ = p.Kill()
						}
					}
				}
			})
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
			if scenario == "migration" || scenario == "health" {
				want = "rolledBack"
			}
			if scenario == "checksum" || scenario == "missing-checksum" {
				want = "failed"
			}
			if state.Phase != want || (scenario == "success") != (runErr == nil) {
				log, _ := os.ReadFile(filepath.Join(job, "worker.log"))
				t.Fatalf("phase=%s want=%s err=%v\n%s\n%s", state.Phase, want, runErr, output, log)
			}
			if scenario != "success" {
				for path, expected := range map[string]string{db: "original database", db + "-wal": "original WAL"} {
					got, err := os.ReadFile(path)
					if err != nil || string(got) != expected {
						t.Errorf("%s not restored: %s %v", path, got, err)
					}
				}
				old, _ := os.ReadFile(binary)
				if !strings.Contains(string(old), "1.6.6") {
					t.Error("old executable not restored")
				}
			}
			if want == "failed" {
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
