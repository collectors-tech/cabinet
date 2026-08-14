//go:build windows

package ai

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCodexInstalledPathsForHomeFindsNewestVSCodeChatGPTRuntime(t *testing.T) {
	home := t.TempDir()
	older := filepath.Join(home, ".vscode", "extensions", "openai.chatgpt-26.1.0-win32-x64", "bin", "windows-x86_64", "codex.exe")
	newer := filepath.Join(home, ".vscode", "extensions", "openai.chatgpt-26.2.0-win32-x64", "bin", "windows-x86_64", "codex.exe")
	for _, path := range []string{older, newer} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create extension runtime directory: %v", err)
		}
		if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
			t.Fatalf("create extension runtime: %v", err)
		}
	}
	paths := codexInstalledPathsForHome(home)
	if len(paths) != 2 || paths[0] != newer || paths[1] != older {
		t.Fatalf("expected newest VS Code ChatGPT runtime first, got %#v", paths)
	}
}
