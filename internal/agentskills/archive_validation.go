package agentskills

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const SkillManifestFile = "cabinet.skill.json"

type ImportResultState string

const (
	ImportValidReadyToInstall      ImportResultState = "valid-ready-to-install"
	ImportValidWithWarnings        ImportResultState = "valid-with-warnings"
	ImportBlockedMissingDependency ImportResultState = "blocked-missing-dependency"
	ImportBlockedInvalidManifest   ImportResultState = "blocked-invalid-manifest"
	ImportBlockedUnsafeArchive     ImportResultState = "blocked-unsafe-archive"
	ImportInstalledDisabled        ImportResultState = "installed-disabled"
	ImportInstalledEnabled         ImportResultState = "installed-enabled"
)

type ArchiveValidationOptions struct {
	MaxFiles int
	MaxBytes int64
}

type ArchiveValidationResult struct {
	State      ImportResultState `json:"state"`
	Skill      Skill             `json:"skill,omitempty"`
	Errors     []string          `json:"errors,omitempty"`
	Warnings   []string          `json:"warnings,omitempty"`
	Provenance string            `json:"provenance,omitempty"`
}

type SkillImporter struct {
	Registry Registry
	Store    *InstalledSkillStore
}

type SkillImportResult struct {
	State          ImportResultState   `json:"state"`
	Skill          Skill               `json:"skill,omitempty"`
	InstalledState InstalledSkillState `json:"installed_state,omitempty"`
	Errors         []string            `json:"errors,omitempty"`
	Warnings       []string            `json:"warnings,omitempty"`
	Provenance     string              `json:"provenance,omitempty"`
}

type skillManifest struct {
	Schema                  string              `json:"schema"`
	ID                      string              `json:"id"`
	Version                 string              `json:"version"`
	DisplayName             string              `json:"displayName"`
	Description             string              `json:"description"`
	Category                string              `json:"category"`
	Source                  manifestSource      `json:"source"`
	SafetyLevel             SafetyLevel         `json:"safetyLevel"`
	Status                  Status              `json:"status"`
	Modes                   []string            `json:"modes"`
	Capabilities            []string            `json:"capabilities"`
	GuidedWorkflows         []string            `json:"guidedWorkflows"`
	UITargets               []string            `json:"uiTargets"`
	IntegrationRequirements []string            `json:"integrationRequirements"`
	InputSchemaRef          string              `json:"inputSchemaRef"`
	OutputSchemaRef         string              `json:"outputSchemaRef"`
	Permissions             manifestPermissions `json:"permissions"`
	Compatibility           map[string]any      `json:"compatibility"`
	Audit                   manifestAudit       `json:"audit"`
	Checksums               map[string]string   `json:"checksums"`
}

type manifestSource struct {
	Type string `json:"type"`
}

type manifestPermissions struct {
	CabinetReads   []string `json:"cabinetReads"`
	CabinetWrites  []string `json:"cabinetWrites"`
	ExternalReads  []string `json:"externalReads"`
	ExternalWrites []string `json:"externalWrites"`
	SecretAccess   bool     `json:"secretAccess"`
	Destructive    bool     `json:"destructive"`
}

type manifestAudit struct {
	ActionTimeline       string `json:"actionTimeline"`
	RequiresConfirmation bool   `json:"requiresConfirmation"`
}

func (r Registry) ValidateSkillFolder(root string, opts ArchiveValidationOptions) ArchiveValidationResult {
	root = filepath.Clean(strings.TrimSpace(root))
	if root == "." || root == "" {
		return invalidManifestResult("folder path is required")
	}
	info, err := os.Stat(root)
	if err != nil {
		return invalidManifestResult(fmt.Sprintf("skill folder is not readable: %v", err))
	}
	if !info.IsDir() {
		return invalidManifestResult("skill source must be an unpacked folder")
	}
	if opts.MaxFiles <= 0 {
		opts.MaxFiles = 100
	}
	if opts.MaxBytes <= 0 {
		opts.MaxBytes = 10 << 20
	}

	fileCount := 0
	totalBytes := int64(0)
	var errorsOut []string
	warnings := []string{}
	seenManifest := false

	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			errorsOut = append(errorsOut, fmt.Sprintf("cannot read archive path: %v", walkErr))
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			errorsOut = append(errorsOut, fmt.Sprintf("cannot resolve archive path %q: %v", path, err))
			return nil
		}
		if rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if unsafeArchivePath(rel) {
			errorsOut = append(errorsOut, "unsafe archive path: "+rel)
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			errorsOut = append(errorsOut, "symlink paths are not supported: "+rel)
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		fileCount++
		if fileCount > opts.MaxFiles {
			errorsOut = append(errorsOut, "archive file count exceeds Cabinet limit")
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			errorsOut = append(errorsOut, fmt.Sprintf("cannot read file metadata for %s: %v", rel, err))
			return nil
		}
		totalBytes += info.Size()
		if totalBytes > opts.MaxBytes {
			errorsOut = append(errorsOut, "archive size exceeds Cabinet limit")
		}
		if unsupportedSkillArchiveFile(rel) {
			errorsOut = append(errorsOut, "unsupported executable or native file: "+rel)
		}
		if rel == SkillManifestFile {
			seenManifest = true
		}
		return nil
	})
	if err != nil {
		return unsafeArchiveResult(fmt.Sprintf("cannot walk skill folder: %v", err))
	}
	if len(errorsOut) > 0 {
		return ArchiveValidationResult{State: ImportBlockedUnsafeArchive, Errors: errorsOut, Warnings: warnings}
	}
	if !seenManifest {
		return invalidManifestResult("required manifest cabinet.skill.json is missing")
	}

	manifestBytes, err := os.ReadFile(filepath.Join(root, SkillManifestFile))
	if err != nil {
		return invalidManifestResult(fmt.Sprintf("manifest is not readable: %v", err))
	}
	var manifest skillManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return invalidManifestResult(fmt.Sprintf("manifest is not valid JSON: %v", err))
	}
	result := r.validateSkillManifest(root, manifest)
	result.Provenance = root
	return result
}

func (r Registry) ValidateSkillZipArchive(path string, opts ArchiveValidationOptions) ArchiveValidationResult {
	path = strings.TrimSpace(path)
	if path == "" {
		return invalidManifestResult("archive path is required")
	}
	reader, err := zip.OpenReader(path)
	if err != nil {
		return invalidManifestResult(fmt.Sprintf("skill archive is not readable: %v", err))
	}
	defer reader.Close()
	if opts.MaxFiles <= 0 {
		opts.MaxFiles = 100
	}
	if opts.MaxBytes <= 0 {
		opts.MaxBytes = 10 << 20
	}

	tmp, err := os.MkdirTemp("", "cabinet-skill-archive-*")
	if err != nil {
		return unsafeArchiveResult(fmt.Sprintf("cannot prepare archive validation folder: %v", err))
	}
	defer os.RemoveAll(tmp)

	fileCount := 0
	totalBytes := int64(0)
	var errorsOut []string
	for _, file := range reader.File {
		rel := filepath.ToSlash(file.Name)
		if strings.HasSuffix(rel, "/") {
			if unsafeArchivePath(strings.TrimSuffix(rel, "/")) {
				errorsOut = append(errorsOut, "unsafe archive path: "+rel)
			}
			continue
		}
		if unsafeArchivePath(rel) {
			errorsOut = append(errorsOut, "unsafe archive path: "+rel)
			continue
		}
		fileCount++
		if fileCount > opts.MaxFiles {
			errorsOut = append(errorsOut, "archive file count exceeds Cabinet limit")
			continue
		}
		totalBytes += int64(file.UncompressedSize64)
		if totalBytes > opts.MaxBytes {
			errorsOut = append(errorsOut, "archive size exceeds Cabinet limit")
			continue
		}
		if unsupportedSkillArchiveFile(rel) {
			errorsOut = append(errorsOut, "unsupported executable or native file: "+rel)
			continue
		}
		if file.FileInfo().Mode()&os.ModeSymlink != 0 {
			errorsOut = append(errorsOut, "symlink paths are not supported: "+rel)
			continue
		}
		if err := extractZipFile(tmp, file, rel); err != nil {
			errorsOut = append(errorsOut, err.Error())
		}
	}
	if len(errorsOut) > 0 {
		return ArchiveValidationResult{State: ImportBlockedUnsafeArchive, Errors: errorsOut, Provenance: path}
	}
	result := r.ValidateSkillFolder(tmp, opts)
	result.Provenance = path
	result.Skill.Provenance = path
	return result
}

func (i SkillImporter) ImportSkillFolder(profileID, root string, opts ArchiveValidationOptions) SkillImportResult {
	return i.importValidation(profileID, i.Registry.ValidateSkillFolder(root, opts))
}

func (i SkillImporter) ImportSkillZipArchive(profileID, path string, opts ArchiveValidationOptions) SkillImportResult {
	return i.importValidation(profileID, i.Registry.ValidateSkillZipArchive(path, opts))
}

func (i SkillImporter) importValidation(profileID string, validation ArchiveValidationResult) SkillImportResult {
	result := SkillImportResult{
		State:      validation.State,
		Skill:      validation.Skill,
		Errors:     append([]string{}, validation.Errors...),
		Warnings:   append([]string{}, validation.Warnings...),
		Provenance: validation.Provenance,
	}
	if validation.State != ImportValidReadyToInstall && validation.State != ImportValidWithWarnings {
		return result
	}
	if i.Store == nil {
		result.State = ImportBlockedInvalidManifest
		result.Errors = append(result.Errors, "installed skill store is required")
		return result
	}
	installed, state, err := i.Registry.InstalledStateFromValidation(profileID, validation)
	if err != nil {
		result.State = ImportBlockedInvalidManifest
		result.Errors = append(result.Errors, err.Error())
		return result
	}
	saved, err := i.Store.Save(installed)
	if err != nil {
		result.State = ImportBlockedInvalidManifest
		result.Errors = append(result.Errors, err.Error())
		return result
	}
	result.State = state
	result.InstalledState = saved
	result.Skill.Provenance = validation.Provenance
	return result
}

func (r Registry) validateSkillManifest(root string, manifest skillManifest) ArchiveValidationResult {
	var errorsOut []string
	var warnings []string
	required := map[string]string{
		"schema":               manifest.Schema,
		"id":                   manifest.ID,
		"version":              manifest.Version,
		"displayName":          manifest.DisplayName,
		"description":          manifest.Description,
		"category":             manifest.Category,
		"source.type":          manifest.Source.Type,
		"safetyLevel":          string(manifest.SafetyLevel),
		"status":               string(manifest.Status),
		"audit.actionTimeline": manifest.Audit.ActionTimeline,
	}
	for field, value := range required {
		if strings.TrimSpace(value) == "" {
			errorsOut = append(errorsOut, "manifest field is required: "+field)
		}
	}
	if len(manifest.Modes) == 0 {
		errorsOut = append(errorsOut, "manifest field is required: modes")
	}
	if len(manifest.Compatibility) == 0 {
		errorsOut = append(errorsOut, "manifest field is required: compatibility")
	}
	if !manifestPermissionsExplicit(manifest.Permissions) {
		errorsOut = append(errorsOut, "manifest field is required: permissions")
	}
	if manifest.Schema != "https://collectors.tech/cabinet/schemas/agent-skill.v1.json" {
		errorsOut = append(errorsOut, "manifest schema is not recognised")
	}
	if manifest.Source.Type != "archive" {
		errorsOut = append(errorsOut, "source.type must be archive for local imports")
	}
	if !validSkillID(manifest.ID) {
		errorsOut = append(errorsOut, "skill id must be lower-case dot-separated text")
	}
	if !validSemver(manifest.Version) {
		errorsOut = append(errorsOut, "skill version must be semantic version text")
	}
	if !knownSafetyLevel(manifest.SafetyLevel) {
		errorsOut = append(errorsOut, "safetyLevel is not recognised")
	}
	if !knownArchiveStatus(manifest.Status) {
		errorsOut = append(errorsOut, "status is not recognised")
	}
	if err := r.ValidateImportedSkill(Skill{ID: manifest.ID}); err != nil {
		errorsOut = append(errorsOut, err.Error())
	}
	for _, ref := range []string{manifest.InputSchemaRef, manifest.OutputSchemaRef} {
		if strings.TrimSpace(ref) != "" && unsafeArchivePath(ref) {
			errorsOut = append(errorsOut, "manifest path escapes archive root: "+ref)
		}
	}
	for path, declared := range manifest.Checksums {
		if unsafeArchivePath(path) {
			errorsOut = append(errorsOut, "checksum path escapes archive root: "+path)
			continue
		}
		if !strings.HasPrefix(declared, "sha256:") {
			errorsOut = append(errorsOut, "checksum must use sha256: "+path)
			continue
		}
		actual, err := sha256File(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			errorsOut = append(errorsOut, fmt.Sprintf("checksum target is not readable: %s", path))
			continue
		}
		if "sha256:"+actual != declared {
			errorsOut = append(errorsOut, "checksum mismatch: "+path)
		}
	}
	if len(errorsOut) > 0 {
		return ArchiveValidationResult{State: ImportBlockedInvalidManifest, Errors: errorsOut, Warnings: warnings}
	}

	missing := r.missingSkillDependencies(manifest)
	if len(missing) > 0 && manifest.Status != StatusInvalid && manifest.Status != StatusRequiresImplementation {
		return ArchiveValidationResult{State: ImportBlockedMissingDependency, Errors: missing, Warnings: warnings}
	}
	if len(missing) > 0 {
		warnings = append(warnings, missing...)
	}

	skill := deriveExecutionState(Skill{
		ID:                   strings.TrimSpace(manifest.ID),
		Version:              strings.TrimSpace(manifest.Version),
		DisplayName:          strings.TrimSpace(manifest.DisplayName),
		Description:          strings.TrimSpace(manifest.Description),
		Category:             strings.TrimSpace(manifest.Category),
		Source:               SourceArchive,
		Status:               manifest.Status,
		SafetyLevel:          manifest.SafetyLevel,
		Capabilities:         append([]string{}, manifest.Capabilities...),
		GuidedWorkflows:      append([]string{}, manifest.GuidedWorkflows...),
		UITargets:            append([]string{}, manifest.UITargets...),
		IntegrationWorkflows: append([]string{}, manifest.IntegrationRequirements...),
		InputSchemaRefs:      singleRef(manifest.InputSchemaRef),
		OutputSchemaRefs:     singleRef(manifest.OutputSchemaRef),
		Permissions: PermissionDeclaration{
			LocalRead:       len(manifest.Permissions.CabinetReads) > 0,
			LocalWrite:      len(manifest.Permissions.CabinetWrites) > 0,
			ExternalRead:    len(manifest.Permissions.ExternalReads) > 0,
			ExternalWrite:   len(manifest.Permissions.ExternalWrites) > 0,
			SecretAccess:    manifest.Permissions.SecretAccess,
			Destructive:     manifest.Permissions.Destructive,
			RequiresConfirm: manifest.Audit.RequiresConfirmation || manifest.SafetyLevel == SafetyConfirmRequired || manifest.SafetyLevel == SafetyExternalWrite || manifest.SafetyLevel == SafetyDestructive,
		},
		AuditBehavior: strings.TrimSpace(manifest.Audit.ActionTimeline),
		Provenance:    "local archive",
		Removable:     true,
		Enabled:       manifest.SafetyLevel == SafetyReadOnly,
	})
	if skill.SafetyLevel != SafetyReadOnly {
		skill.Status = StatusDisabled
		skill.Executable = false
	}
	state := ImportValidReadyToInstall
	if len(warnings) > 0 {
		state = ImportValidWithWarnings
	}
	return ArchiveValidationResult{State: state, Skill: skill, Warnings: warnings}
}

func (r Registry) InstalledStateFromValidation(profileID string, result ArchiveValidationResult) (InstalledSkillState, ImportResultState, error) {
	if result.State != ImportValidReadyToInstall && result.State != ImportValidWithWarnings {
		return InstalledSkillState{}, result.State, errors.New("skill validation result is not installable")
	}
	if strings.TrimSpace(profileID) == "" {
		return InstalledSkillState{}, result.State, errors.New("profile id is required")
	}
	if err := r.ValidateImportedSkill(result.Skill); err != nil {
		return InstalledSkillState{}, result.State, err
	}
	enabled := result.Skill.SafetyLevel == SafetyReadOnly
	status := StatusAvailable
	installState := ImportInstalledEnabled
	if !enabled {
		status = StatusDisabled
		installState = ImportInstalledDisabled
	}
	return InstalledSkillState{
		ProfileID:          strings.TrimSpace(profileID),
		SkillID:            result.Skill.ID,
		Enabled:            enabled,
		Status:             status,
		ValidationWarnings: append([]string{}, result.Warnings...),
		ValidationErrors:   append([]string{}, result.Errors...),
	}, installState, nil
}

func invalidManifestResult(message string) ArchiveValidationResult {
	return ArchiveValidationResult{State: ImportBlockedInvalidManifest, Errors: []string{message}}
}

func unsafeArchiveResult(message string) ArchiveValidationResult {
	return ArchiveValidationResult{State: ImportBlockedUnsafeArchive, Errors: []string{message}}
}

func unsafeArchivePath(path string) bool {
	path = strings.TrimSpace(strings.ReplaceAll(path, "\\", "/"))
	if path == "" || strings.HasPrefix(path, "/") || strings.Contains(path, ":") {
		return true
	}
	clean := filepath.ToSlash(filepath.Clean(path))
	return clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../")
}

func unsupportedSkillArchiveFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	unsupported := []string{".bat", ".cmd", ".dll", ".dylib", ".exe", ".go", ".js", ".ps1", ".py", ".sh", ".so", ".ts"}
	return slices.Contains(unsupported, ext)
}

func extractZipFile(root string, file *zip.File, rel string) error {
	target := filepath.Join(root, filepath.FromSlash(rel))
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("cannot resolve extraction root: %v", err)
	}
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("cannot resolve archive path %s: %v", rel, err)
	}
	if targetAbs != rootAbs && !strings.HasPrefix(targetAbs, rootAbs+string(os.PathSeparator)) {
		return fmt.Errorf("unsafe archive path: %s", rel)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("cannot create archive path %s: %v", rel, err)
	}
	source, err := file.Open()
	if err != nil {
		return fmt.Errorf("cannot read archive file %s: %v", rel, err)
	}
	defer source.Close()
	destination, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("cannot extract archive file %s: %v", rel, err)
	}
	defer destination.Close()
	if _, err := io.Copy(destination, source); err != nil {
		return fmt.Errorf("cannot extract archive file %s: %v", rel, err)
	}
	return nil
}

func manifestPermissionsExplicit(permissions manifestPermissions) bool {
	return permissions.CabinetReads != nil &&
		permissions.CabinetWrites != nil &&
		permissions.ExternalReads != nil &&
		permissions.ExternalWrites != nil
}

func validSkillID(id string) bool {
	id = strings.TrimSpace(id)
	if id == "" || strings.Contains(id, "..") || strings.HasPrefix(id, ".") || strings.HasSuffix(id, ".") {
		return false
	}
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			continue
		}
		return false
	}
	return strings.Contains(id, ".")
}

func validSemver(version string) bool {
	parts := strings.Split(strings.TrimSpace(version), ".")
	if len(parts) != 3 {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	return true
}

func knownSafetyLevel(level SafetyLevel) bool {
	switch level {
	case SafetyReadOnly, SafetyPreviewOnly, SafetyConfirmRequired, SafetyExternalWrite, SafetyDestructive:
		return true
	default:
		return false
	}
}

func knownArchiveStatus(status Status) bool {
	switch status {
	case StatusAvailable, StatusPreviewOnly, StatusDisabled, StatusInvalid, StatusRequiresImplementation:
		return true
	default:
		return false
	}
}

func (r Registry) missingSkillDependencies(manifest skillManifest) []string {
	capabilities := capabilityIDs()
	guided := guidedWorkflowIDs()
	uiTargets := uiTargetIDs(r)
	var missing []string
	for _, id := range manifest.Capabilities {
		if !slices.Contains(capabilities, id) {
			missing = append(missing, "missing capability: "+id)
		}
	}
	for _, id := range manifest.GuidedWorkflows {
		if !slices.Contains(guided, id) {
			missing = append(missing, "missing guided workflow: "+id)
		}
	}
	for _, id := range manifest.UITargets {
		if !slices.Contains(uiTargets, id) {
			missing = append(missing, "missing ui target: "+id)
		}
	}
	return missing
}

func uiTargetIDs(r Registry) []string {
	seen := map[string]struct{}{}
	for _, skill := range r.List() {
		for _, id := range skill.UITargets {
			seen[id] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	slices.Sort(out)
	return out
}

func singleRef(ref string) []string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil
	}
	return []string{ref}
}

func sha256File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
