package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCabinetContainerImageContract(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join("..")
	dockerfilePath := filepath.Join(repoRoot, "Dockerfile")
	raw, err := os.ReadFile(dockerfilePath)
	if err != nil {
		t.Fatalf("read Dockerfile: %v", err)
	}
	content := string(raw)

	requiredFragments := []string{
		"FROM node:22-bookworm AS ui-build",
		"RUN npm ci",
		"COPY docs/help-center/ /src/docs/help-center/",
		"RUN npm run build",
		"FROM golang:1.24-bookworm AS app-build",
		"RUN go mod download",
		"COPY --from=ui-build /src/internal/ui/static ./internal/ui/static",
		"CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build",
		"FROM debian:bookworm-slim AS runtime",
		"ca-certificates curl tzdata",
		"mkdir -p /data",
		"chown cabinet:cabinet /data",
		"ENV CABINET_OPEN_BROWSER=0",
		"EXPOSE 17880",
		"VOLUME [\"/data\"]",
		"HEALTHCHECK --interval=5s --timeout=3s --start-period=10s --retries=12 CMD curl -fsS http://127.0.0.1:17880/healthz || exit 1",
		"USER cabinet",
		"ENTRYPOINT [\"/app/cabinet\"]",
		"--no-open-browser",
		"--listen",
		"0.0.0.0:17880",
		"--data-dir",
		"/data",
		"--profile",
		"e2e-cypress",
		"--instance-name",
		"cypress-container",
		"--allow-parallel",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(content, fragment) {
			t.Fatalf("Dockerfile missing required fragment %q", fragment)
		}
	}
}

func TestCabinetDockerIgnoreKeepsImageBuildContextBounded(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join("..", ".dockerignore"))
	if err != nil {
		t.Fatalf("read .dockerignore: %v", err)
	}
	content := string(raw)

	requiredEntries := []string{
		".git",
		".agentbus",
		".antfarm",
		".logs",
		".tmp",
		".work-agent",
		"bin",
		"data",
		"node_modules",
		"tmp",
		"ui.web/node_modules",
		"internal/ui/static",
		"cabinet.exe",
	}

	for _, entry := range requiredEntries {
		if !strings.Contains(content, entry) {
			t.Fatalf(".dockerignore missing required entry %q", entry)
		}
	}
}
