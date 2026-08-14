//go:build !windows

package ai

import "os/exec"

func codexExecutableCandidates() []string {
	return []string{"codex"}
}

func codexInstalledPaths() []string {
	return nil
}

func configureBackgroundCommand(_ *exec.Cmd) {}
