// Package updater installs official releases for service-managed Linux installs.
// Network access and workers are started only by an authenticated panel action.
package updater

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alireza0/s-ui/config"
)

const repository = "https://github.com/HITM1ss/m1ss-ui"

//go:embed worker.sh
var worker []byte

var versionPattern = regexp.MustCompile(`^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

type State struct {
	Phase           string `json:"phase"`
	TargetVersion   string `json:"targetVersion,omitempty"`
	PreviousVersion string `json:"previousVersion,omitempty"`
	StartedAt       int64  `json:"startedAt,omitempty"`
	UpdatedAt       int64  `json:"updatedAt,omitempty"`
	PID             int    `json:"pid,omitempty"`
	Job             string `json:"job,omitempty"`
}

func (s State) Active() bool {
	switch s.Phase {
	case "queued", "downloading", "verifying", "backingUp", "installing", "restarting", "rollingBack":
		return true
	}
	return false
}

type Status struct {
	CurrentVersion string `json:"currentVersion"`
	Supported      bool   `json:"supported"`
	Reason         string `json:"reason,omitempty"`
	State          State  `json:"state"`
}

type Release struct {
	Status
	LatestVersion string `json:"latestVersion"`
	Available     bool   `json:"available"`
	URL           string `json:"url"`
	PublishedAt   string `json:"publishedAt"`
}

type Manager struct {
	mu       sync.Mutex
	binary   string
	dbPath   string
	version  string
	arch     string
	service  string
	reason   string
	client   *http.Client
	endpoint string
	launch   func(string, []string, string) error
	alive    func(int) bool
}

var defaultOnce sync.Once
var defaultManager *Manager

func Default() *Manager {
	defaultOnce.Do(func() {
		binary, service, reason := installation()
		arch := architecture(runtime.GOARCH, buildARM())
		if arch == "" && reason == "" {
			reason = "architecture"
		}
		dbPath, err := filepath.Abs(config.GetDBPath())
		if err != nil {
			reason = "installation"
		}
		defaultManager = &Manager{
			binary: binary, dbPath: dbPath, version: config.GetVersion(), arch: arch,
			service: service, reason: reason,
			client:   &http.Client{Timeout: 15 * time.Second},
			endpoint: "https://api.github.com/repos/HITM1ss/m1ss-ui/releases/latest",
			launch:   launchWorker, alive: processAlive,
		}
	})
	return defaultManager
}

func buildARM() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "GOARM" {
				return strings.Split(setting.Value, ",")[0]
			}
		}
	}
	return ""
}

func architecture(arch, arm string) string {
	switch arch {
	case "amd64", "arm64", "386", "s390x":
		return arch
	case "arm":
		if arm == "5" || arm == "6" || arm == "7" {
			return "armv" + arm
		}
	}
	return ""
}

func newer(latest, current string) bool {
	a, b := versionPattern.FindStringSubmatch(latest), versionPattern.FindStringSubmatch(current)
	if a == nil || b == nil {
		return false
	}
	for i := 1; i <= 3; i++ {
		// Compare decimal strings without integer overflow.
		if len(a[i]) != len(b[i]) {
			return len(a[i]) > len(b[i])
		}
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}

func (m *Manager) root() string {
	return filepath.Join(filepath.Dir(m.binary), ".m1ss-update")
}

func (m *Manager) readState() State {
	f, err := os.Open(filepath.Join(m.root(), "status.json"))
	if err != nil {
		return State{Phase: "idle"}
	}
	defer f.Close()
	var state State
	if json.NewDecoder(io.LimitReader(f, 4096)).Decode(&state) != nil {
		return State{Phase: "failed"}
	}
	return state
}

func (m *Manager) writeState(state State) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return writeAtomic(filepath.Join(m.root(), "status.json"), data)
}

func writeAtomic(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".state-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

func (m *Manager) Status() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	state := m.readState()
	// A worker that exited without its trap (e.g. a reboot) must not lock out
	// all later updates. Give the service manager time to launch a queued job.
	if state.Active() && time.Now().Unix()-state.UpdatedAt > 30 && !m.alive(state.PID) {
		state.Phase = "interrupted"
		state.UpdatedAt = time.Now().Unix()
		_ = m.writeState(state)
	}
	return Status{CurrentVersion: m.version, Supported: m.reason == "", Reason: m.reason, State: state}
}

func (m *Manager) Check(ctx context.Context) (*Release, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, m.endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "M1ss-UI/"+m.version)
	response, err := m.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("check GitHub release: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub release request returned HTTP %d", response.StatusCode)
	}
	var release struct {
		Tag         string `json:"tag_name"`
		Draft       bool   `json:"draft"`
		Prerelease  bool   `json:"prerelease"`
		PublishedAt string `json:"published_at"`
		Assets      []struct {
			Name string `json:"name"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&release); err != nil {
		return nil, fmt.Errorf("read GitHub release: %w", err)
	}
	if release.Draft || release.Prerelease || !strings.HasPrefix(release.Tag, "v") || !versionPattern.MatchString(release.Tag) {
		return nil, errors.New("the latest release is not a stable version")
	}
	archive, sums := false, false
	for _, asset := range release.Assets {
		archive = archive || asset.Name == "s-ui-linux-"+m.arch+".tar.gz"
		sums = sums || asset.Name == "SHA256SUMS"
	}
	status := m.Status()
	if status.Supported && (!archive || !sums) {
		return nil, errors.New("release archive or SHA256SUMS is missing; try again after publishing completes")
	}
	return &Release{
		Status: status, LatestVersion: strings.TrimPrefix(release.Tag, "v"),
		Available: newer(release.Tag, m.version), URL: repository + "/releases/tag/" + release.Tag,
		PublishedAt: release.PublishedAt,
	}, nil
}

func (m *Manager) Start(ctx context.Context, target string) (State, error) {
	if m.reason != "" {
		return State{}, fmt.Errorf("panel updates are unavailable: %s", m.reason)
	}
	release, err := m.Check(ctx)
	if err != nil {
		return State{}, err
	}
	if !release.Available || target != release.LatestVersion {
		return State{}, errors.New("release changed or is not newer; check for updates again")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.readState().Active() {
		return State{}, errors.New("an update is already running")
	}
	if err := os.MkdirAll(m.root(), 0o700); err != nil {
		return State{}, err
	}
	if err := os.Chmod(m.root(), 0o700); err != nil {
		return State{}, err
	}
	job, err := os.MkdirTemp(m.root(), "job-")
	if err != nil {
		return State{}, err
	}
	script := filepath.Join(job, "worker.sh")
	if err := os.WriteFile(script, worker, 0o700); err != nil {
		return State{}, err
	}
	now := time.Now().Unix()
	state := State{Phase: "queued", TargetVersion: target, PreviousVersion: m.version, StartedAt: now, UpdatedAt: now, Job: filepath.Base(job)}
	if err := m.writeState(state); err != nil {
		return State{}, err
	}
	args := []string{m.binary, m.dbPath, target, m.version, m.arch, m.service, strconv.FormatInt(now, 10)}
	if err := m.launch(script, args, m.service); err != nil {
		state.Phase = "failed"
		_ = m.writeState(state)
		return State{}, err
	}
	return state, nil
}

// MarkReady is called only after the panel, subscription listener and core
// have started. It avoids insecure loopback TLS probes and a second Go process
// for every health check on small VPSs.
func MarkReady() error {
	m := Default()
	state := m.readState()
	if state.Phase != "restarting" && state.Phase != "rollingBack" {
		return nil
	}
	return writeAtomic(filepath.Join(m.root(), "ready"), []byte(fmt.Sprintf("%s %d\n", m.version, os.Getpid())))
}
