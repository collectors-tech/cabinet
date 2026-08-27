//go:build !windows

package tests

func watchdogProcessExists(int) bool {
	return false
}
