//go:build windows

package ai

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"syscall"
)

func codexExecutableCandidates() []string {
	return []string{"codex.exe"}
}

func codexInstalledPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return codexInstalledPathsForHome(home)
}

func codexInstalledPathsForHome(home string) []string {
	paths := make([]string, 0)
	for _, root := range []string{".vscode", ".vscode-insiders"} {
		matches, _ := filepath.Glob(filepath.Join(home, root, "extensions", "openai.chatgpt-*", "bin", "windows-x86_64", "codex.exe"))
		paths = append(paths, matches...)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(paths)))
	return paths
}

func configureBackgroundCommand(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}
