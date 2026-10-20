//go:build !linux

package updater

import "errors"

func installation() (string, string, string) { return "/usr/local/s-ui/sui", "", "platform" }
func processAlive(int) bool                  { return false }
func launchWorker(string, []string, string) error {
	return errors.New("panel updates require a service-managed Linux installation")
}
