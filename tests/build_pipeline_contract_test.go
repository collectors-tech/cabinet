package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBuildCabinetAlwaysBuildsUIFirst(t *testing.T) {
	t.Parallel()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file location: runtime.Caller failed")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), ".."))
	buildScript := filepath.Join(repoRoot, "scripts", "build-cabinet.ps1")

	data, err := os.ReadFile(buildScript)
	if err != nil {
		t.Fatalf("read build script: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "build-ui-static.ps1") {
		t.Fatalf("build script must invoke ui static build before go build")
	}
	if strings.Contains(content, "SkipUIBuild") {
		t.Fatalf("build script must not provide a path that skips ui build")
	}
}

func TestReadmeBuildEntrypointDoesNotDocumentSkipUIBuild(t *testing.T) {
	t.Parallel()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file location: runtime.Caller failed")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), ".."))
	readmePath := filepath.Join(repoRoot, "README.md")

	data, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	content := string(data)

	if strings.Contains(content, "-SkipUIBuild") {
		t.Fatalf("README must not document skipping ui build for canonical build entrypoint")
	}
}

func TestPackageInstallersBuildsUIBeforeCrossPlatformLoop(t *testing.T) {
	t.Parallel()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file location: runtime.Caller failed")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), ".."))
	packageScript := filepath.Join(repoRoot, "scripts", "package-installers.ps1")

	data, err := os.ReadFile(packageScript)
	if err != nil {
		t.Fatalf("read package script: %v", err)
	}
	content := string(data)

	uiBuildIndex := strings.Index(content, "build-ui-static.ps1")
	targetsIndex := strings.Index(content, "$targets = @(")
	if uiBuildIndex < 0 {
		t.Fatalf("package script must invoke ui build")
	}
	if targetsIndex < 0 {
		t.Fatalf("package script must define cross-platform target loop")
	}
	if uiBuildIndex > targetsIndex {
		t.Fatalf("package script must build ui before target build loop")
	}
	if !strings.Contains(content, "throw \"ui.web build failed") {
		t.Fatalf("package script must fail fast on ui build failure")
	}
}
