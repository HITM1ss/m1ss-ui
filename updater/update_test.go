package updater

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func testManager(t *testing.T, body string) *Manager {
	t.Helper()
	root := t.TempDir()
	return &Manager{
		binary: filepath.Join(root, "sui"), dbPath: filepath.Join(root, "db", "s-ui.db"),
		version: "1.6.6", arch: "amd64", service: "openrc", endpoint: "https://api.github.com/test",
		alive: func(int) bool { return false }, launch: func(string, []string, string) error { return nil },
		client: &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
			if r.Header.Get("User-Agent") != "M1ss-UI/1.6.6" {
				t.Error("missing release request user agent")
			}
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
		})},
	}
}

const stableRelease = `{"tag_name":"v1.6.7","assets":[{"name":"s-ui-linux-amd64.tar.gz"},{"name":"SHA256SUMS"}]}`

func TestVersionAndArchitecture(t *testing.T) {
	for _, tt := range []struct {
		latest, current string
		want            bool
	}{
		{"v1.6.7", "1.6.6", true}, {"v1.6.10", "1.6.9", true}, {"v2.0.0", "1.99.9", true},
		{"v1.6.6", "1.6.6", false}, {"v1.6.5", "1.6.6", false},
		{"v1.6.7-rc.1", "1.6.6", false}, {"v01.6.7", "1.6.6", false},
		{"v1.6.7;id", "1.6.6", false}, {"v1.6.7", "development", false},
	} {
		if got := newer(tt.latest, tt.current); got != tt.want {
			t.Errorf("newer(%q, %q) = %v", tt.latest, tt.current, got)
		}
	}
	for input, want := range map[string]string{"amd64": "amd64", "arm64": "arm64", "386": "386", "s390x": "s390x", "riscv64": "", "arm": ""} {
		if got := architecture(input, ""); got != want {
			t.Errorf("architecture(%s) = %s", input, got)
		}
	}
	for _, arm := range []string{"5", "6", "7"} {
		if architecture("arm", arm) != "armv"+arm {
			t.Fatal("incorrect ARM release archive")
		}
	}
}

func TestCheckRequiresStableCompleteRelease(t *testing.T) {
	for _, body := range []string{
		`{"tag_name":"v1.6.7-rc.1"}`, `{"tag_name":"v1.6.7","prerelease":true}`,
		`{"tag_name":"v1.6.7","draft":true}`, `{"tag_name":"v1.6.7"}`,
		`{"tag_name":"v1.6.7","assets":[{"name":"SHA256SUMS"}]}`,
		`{"tag_name":"v1.6.7","assets":[{"name":"s-ui-linux-amd64.tar.gz"}]}`, `not json`,
	} {
		if _, err := testManager(t, body).Check(context.Background()); err == nil {
			t.Errorf("accepted incomplete or unsafe release: %s", body)
		}
	}
	m := testManager(t, stableRelease)
	release, err := m.Check(context.Background())
	if err != nil || !release.Available || release.LatestVersion != "1.6.7" || release.URL != repository+"/releases/tag/v1.6.7" {
		t.Fatalf("unexpected release: %+v, %v", release, err)
	}
	m.reason = "container"
	if _, err := m.Start(context.Background(), "1.6.7"); err == nil {
		t.Fatal("container installation allowed self-update")
	}
}

func TestUpdateIsPinnedAndSingleFlight(t *testing.T) {
	m := testManager(t, stableRelease)
	if _, err := m.Start(context.Background(), "1.6.8"); err == nil {
		t.Fatal("accepted an unchecked version")
	}
	var launches atomic.Int32
	m.launch = func(script string, args []string, manager string) error {
		launches.Add(1)
		if !strings.HasPrefix(script, m.root()+string(os.PathSeparator)) || args[2] != "1.6.7" || manager != "openrc" {
			t.Errorf("unexpected worker arguments: %s %v %s", script, args, manager)
		}
		return nil
	}
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Go(func() { _, _ = m.Start(context.Background(), "1.6.7") })
	}
	group.Wait()
	if launches.Load() != 1 || m.Status().State.Phase != "queued" {
		t.Fatalf("launched %d workers", launches.Load())
	}
	info, err := os.Stat(m.root())
	if err != nil || info.Mode().Perm() != 0o700 {
		t.Fatalf("update directory is not private: %v", err)
	}
}

func TestLaunchFailureAndInterruptedWorker(t *testing.T) {
	m := testManager(t, stableRelease)
	m.launch = func(string, []string, string) error { return fmt.Errorf("service manager unavailable") }
	if _, err := m.Start(context.Background(), "1.6.7"); err == nil || m.Status().State.Phase != "failed" {
		t.Fatal("launch failure left update running")
	}
	state := State{Phase: "installing", PID: 123, UpdatedAt: time.Now().Add(-time.Minute).Unix()}
	if err := m.writeState(state); err != nil {
		t.Fatal(err)
	}
	m.alive = func(int) bool { return true }
	if !m.Status().State.Active() {
		t.Fatal("marked a live worker as interrupted")
	}
	m.alive = func(int) bool { return false }
	if m.Status().State.Phase != "interrupted" {
		t.Fatal("dead worker permanently blocks later updates")
	}
}
