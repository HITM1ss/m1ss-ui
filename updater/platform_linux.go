//go:build linux

package updater

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func installation() (string, string, string) {
	const binary = "/usr/local/s-ui/sui"
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return binary, "", "container"
	}
	if cgroup, err := os.ReadFile("/proc/1/cgroup"); err == nil && (strings.Contains(string(cgroup), "docker") || strings.Contains(string(cgroup), "kubepods")) {
		return binary, "", "container"
	}
	exe, err := os.Executable()
	if err != nil || filepath.Clean(exe) != binary {
		return binary, "", "installation"
	}
	if os.Geteuid() != 0 {
		return binary, "", "permissions"
	}
	for _, tool := range []string{"bash", "curl", "tar", "sha256sum", "timeout"} {
		if _, err := exec.LookPath(tool); err != nil {
			return binary, "", "tools"
		}
	}
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		if _, err := exec.LookPath("systemd-run"); err != nil {
			return binary, "", "service"
		}
		output, err := exec.Command("systemctl", "show", "s-ui", "--property=MainPID", "--value").Output()
		if err == nil && strings.TrimSpace(string(output)) == strconv.Itoa(os.Getpid()) {
			return binary, "systemd", ""
		}
	} else if _, err := exec.LookPath("rc-service"); err == nil {
		pid, err := os.ReadFile("/run/s-ui.pid")
		if err == nil {
			value := strings.TrimSpace(string(pid))
			if value == strconv.Itoa(os.Getpid()) {
				return binary, "openrc", ""
			}
			// supervise-daemon records the supervisor, not the child process.
			parent := os.Getppid()
			supervisor, linkErr := os.Readlink(fmt.Sprintf("/proc/%d/exe", parent))
			if linkErr == nil && value == strconv.Itoa(parent) && filepath.Base(supervisor) == "supervise-daemon" {
				return binary, "openrc", ""
			}
		}
	}
	return binary, "", "service"
}

func processAlive(pid int) bool {
	return pid > 0 && syscall.Kill(pid, 0) == nil
}

func launchWorker(script string, args []string, service string) error {
	if service == "systemd" {
		// A child inside s-ui.service would be killed by systemctl stop s-ui.
		// The transient unit gives the updater its own cgroup and lifetime.
		command := []string{"--unit=m1ss-ui-update-" + filepath.Base(filepath.Dir(script)), "--collect", "--property=Type=exec", "bash", script}
		command = append(command, args...)
		if output, err := exec.Command("systemd-run", command...).CombinedOutput(); err != nil {
			return fmt.Errorf("start update service: %w: %s", err, strings.TrimSpace(string(output)))
		}
		return nil
	}
	command := exec.Command("bash", append([]string{script}, args...)...)
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	log, err := os.OpenFile(filepath.Join(filepath.Dir(script), "worker.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer log.Close()
	command.Stdout, command.Stderr = log, log
	if err := command.Start(); err != nil {
		return err
	}
	go func() { _ = command.Wait() }()
	return nil
}
