package agentskills

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestSkillArchiveValidationAcceptsValidFolderAndInstallsDisabledByDefault(t *testing.T) {
	t.Parallel()

	root := writeSkillArchiveFixture(t, validSkillManifest(`{
		"id": "cabinet.example.update_item_guided",
		"safetyLevel": "confirm-required",
		"status": "preview-only",
		"capabilities": ["inventory.item.update"],
		"guidedWorkflows": ["inventory.item.update"],
		"uiTargets": ["inventory.item.editor.title"],
		"permissions": {
			"cabinetReads": ["inventory.item"],
			"cabinetWrites": ["inventory.item"],
			"externalReads": [],
			"externalWrites": [],
			"secretAccess": false,
			"destructive": false
		},
		"audit": {
			"actionTimeline": "records preview id and selected inventory item",
			"requiresConfirmation": true
		}
	}`))

	registry := NewRegistry(nil)
	result := registry.ValidateSkillFolder(root, ArchiveValidationOptions{})
	if result.State != ImportValidReadyToInstall {
		t.Fatalf("expected valid-ready-to-install, got %+v", result)
	}
	if result.Skill.ID != "cabinet.example.update_item_guided" || result.Skill.Source != SourceArchive || result.Skill.Executable {
		t.Fatalf("expected imported non-executable disabled skill model, got %+v", result.Skill)
	}
	if result.Skill.Status != StatusDisabled || result.Skill.Enabled {
		t.Fatalf("confirm-required imports must validate but install disabled by default, got %+v", result.Skill)
	}

	state, installState, err := registry.InstalledStateFromValidation("profile-a", result)
	if err != nil {
		t.Fatalf("installed state from validation: %v", err)
	}
	if installState != ImportInstalledDisabled || state.Status != StatusDisabled || state.Enabled {
		t.Fatalf("expected disabled installed state, installState=%s state=%+v", installState, state)
	}
}

func TestSkillArchiveValidationRejectsInvalidManifestAndUnsafeArchive(t *testing.T) {
	t.Parallel()

	registry := NewRegistry(nil)

	missingManifest := t.TempDir()
	writeArchiveFile(t, missingManifest, "README.md", "# missing manifest\n")
	missing := registry.ValidateSkillFolder(missingManifest, ArchiveValidationOptions{})
	if missing.State != ImportBlockedInvalidManifest || !containsFragment(missing.Errors, "manifest") {
		t.Fatalf("expected missing manifest to block as invalid manifest, got %+v", missing)
	}

	unsafe := writeSkillArchiveFixture(t, validSkillManifest(`{
		"id": "cabinet.example.unsafe_archive",
		"safetyLevel": "read-only",
		"status": "available"
	}`))
	writeArchiveFile(t, unsafe, "scripts/run.ps1", "Write-Host unsafe")
	unsafeResult := registry.ValidateSkillFolder(unsafe, ArchiveValidationOptions{})
	if unsafeResult.State != ImportBlockedUnsafeArchive || !containsFragment(unsafeResult.Errors, "unsupported") {
		t.Fatalf("expected executable/native file to block as unsafe archive, got %+v", unsafeResult)
	}

	checksumPath := writeSkillArchiveFixture(t, validSkillManifest(`{
		"id": "cabinet.example.path_traversal",
		"safetyLevel": "read-only",
		"status": "available",
		"checksums": {
			"../escape.json": "sha256:abcd"
		}
	}`))
	traversal := registry.ValidateSkillFolder(checksumPath, ArchiveValidationOptions{})
	if traversal.State != ImportBlockedInvalidManifest || !containsFragment(traversal.Errors, "escapes archive root") {
		t.Fatalf("expected manifest path traversal to block as invalid manifest, got %+v", traversal)
	}
}

func TestSkillArchiveValidationRejectsUnknownDependencyAndBuiltInOverride(t *testing.T) {
	t.Parallel()

	registry := NewRegistry(nil)

	unknown := writeSkillArchiveFixture(t, validSkillManifest(`{
		"id": "cabinet.example.unknown_dependency",
		"safetyLevel": "read-only",
		"status": "available",
		"capabilities": ["cabinet.missing.capability"]
	}`))
	unknownResult := registry.ValidateSkillFolder(unknown, ArchiveValidationOptions{})
	if unknownResult.State != ImportBlockedMissingDependency || !containsFragment(unknownResult.Errors, "missing capability") {
		t.Fatalf("expected unknown dependency block, got %+v", unknownResult)
	}

	duplicate := writeSkillArchiveFixture(t, validSkillManifest(`{
		"id": "cabinet.navigate.open_surface",
		"safetyLevel": "read-only",
		"status": "available"
	}`))
	duplicateResult := registry.ValidateSkillFolder(duplicate, ArchiveValidationOptions{})
	if duplicateResult.State != ImportBlockedInvalidManifest || !containsFragment(duplicateResult.Errors, "cannot override built-in") {
		t.Fatalf("expected built-in override to block as invalid manifest, got %+v", duplicateResult)
	}
}

func TestSkillArchiveValidationVerifiesChecksums(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeArchiveFile(t, root, "schemas/input.schema.json", `{"type":"object"}`)
	sum := sha256.Sum256([]byte(`{"type":"object"}`))
	manifest := validSkillManifest(`{
		"id": "cabinet.example.checksummed_reader",
		"safetyLevel": "read-only",
		"status": "available",
		"inputSchemaRef": "schemas/input.schema.json",
		"checksums": {
			"schemas/input.schema.json": "sha256:` + hex.EncodeToString(sum[:]) + `"
		}
	}`)
	writeArchiveFile(t, root, SkillManifestFile, manifest)

	result := NewRegistry(nil).ValidateSkillFolder(root, ArchiveValidationOptions{})
	if result.State != ImportValidReadyToInstall || result.Skill.Status != StatusAvailable || !result.Skill.Enabled || !result.Skill.Executable {
		t.Fatalf("expected read-only checksummed archive to validate enabled, got %+v", result)
	}
}

func validSkillManifest(overrides string) string {
	base := `{
		"schema": "https://collectors.tech/cabinet/schemas/agent-skill.v1.json",
		"id": "cabinet.example.open_inventory_help",
		"version": "1.0.0",
		"displayName": "Open inventory help",
		"description": "Explains Inventory without changing data.",
		"category": "inventory",
		"source": {"type": "archive"},
		"safetyLevel": "read-only",
		"status": "available",
		"modes": ["in-app", "assistant"],
		"capabilities": ["navigate.open_surface"],
		"guidedWorkflows": [],
		"uiTargets": ["inventory.item.editor.title"],
		"integrationRequirements": [],
		"permissions": {
			"cabinetReads": ["inventory.help"],
			"cabinetWrites": [],
			"externalReads": [],
			"externalWrites": [],
			"secretAccess": false,
			"destructive": false
		},
		"compatibility": {"cabinetMinVersion": "0.1.0", "schemaVersion": "v1"},
		"audit": {
			"actionTimeline": "records non-secret help route",
			"requiresConfirmation": false
		}
	}`
	if strings.TrimSpace(overrides) == "" {
		return base
	}
	return mergeJSONObjectsForTest(base, overrides)
}

func writeSkillArchiveFixture(t *testing.T, manifest string) string {
	t.Helper()
	root := t.TempDir()
	writeArchiveFile(t, root, SkillManifestFile, manifest)
	writeArchiveFile(t, root, "README.md", "# Test skill\n")
	return root
}

func writeArchiveFile(t *testing.T, root, path, content string) {
	t.Helper()
	fullPath := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir fixture path: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture file %s: %v", path, err)
	}
}

func containsFragment(values []string, fragment string) bool {
	return slices.ContainsFunc(values, func(value string) bool {
		return strings.Contains(value, fragment)
	})
}

func mergeJSONObjectsForTest(base, overrides string) string {
	trimmedBase := strings.TrimSpace(base)
	trimmedOverrides := strings.TrimSpace(overrides)
	return strings.TrimSuffix(trimmedBase, "}") + "," + strings.TrimPrefix(trimmedOverrides, "{")
}
