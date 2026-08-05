package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBetaReleaseTrackersUseOneAcyclicGateModel(t *testing.T) {
	t.Parallel()

	repoRoot := resolveRepoRoot(t)
	for relativePath, required := range map[string][]string{
		"TASK_LIST.md": {
			"#2037 [DONE]",
			"#2036 [IN REVIEW]",
			"#2033",
			"#2035",
			"#2032",
			"#1944",
			"#1945",
			"#2034",
			"private/internal candidate",
			"temporary release freeze",
		},
		filepath.Join("openspec", "migration", "beta-release-evidence-index.md"): {
			"Source/live ready for packaging",
			"does not require #1869 closure",
			"private/internal candidate",
			"#1869 packaged acceptance",
			"#1867 packaged data-safety",
			"explicit #1864 approval",
			"external prerelease publication",
			"`develop` to `main`",
			"temporary release freeze",
			"issue-1944-frontline-live-provider",
			"issue-1929-bonza-live-path-probe",
		},
		filepath.Join("openspec", "migration", "beta-packaged-core-workflow-acceptance.md"): {
			"private/internal candidate",
			"Cabinet package SHA-256",
			"Browser Companion package SHA-256",
			"Voglers",
			"Hobbytech",
			"Frontline",
			"Bonza",
			"final approval follows packaged acceptance",
		},
	} {
		raw, err := os.ReadFile(filepath.Join(repoRoot, relativePath))
		if err != nil {
			t.Fatalf("read %s: %v", relativePath, err)
		}
		content := string(raw)
		for _, fragment := range required {
			if !strings.Contains(content, fragment) {
				t.Errorf("%s missing release-governance contract %q", relativePath, fragment)
			}
		}
	}
}

func TestBetaReleaseApprovalCannotBlockCandidateAcceptance(t *testing.T) {
	t.Parallel()

	repoRoot := resolveRepoRoot(t)
	for _, relativePath := range []string{
		filepath.Join("openspec", "migration", "beta-release-evidence-index.md"),
		filepath.Join("openspec", "migration", "beta-packaged-core-workflow-acceptance.md"),
		filepath.Join("release", "windows-portable-artifact-validation.md"),
		filepath.Join("release", "windows-portable-upgrade-validation.md"),
	} {
		raw, err := os.ReadFile(filepath.Join(repoRoot, relativePath))
		if err != nil {
			t.Fatalf("read %s: %v", relativePath, err)
		}
		content := string(raw)
		for _, fragment := range []string{
			"Internal candidate creation does not require final #1864 approval.",
			"Final #1864 approval is required before external prerelease publication or `develop` to `main` promotion.",
		} {
			if !strings.Contains(content, fragment) {
				t.Errorf("%s missing acyclic approval wording %q", relativePath, fragment)
			}
		}
	}
}

func TestBetaReleaseFreezeAndBranchDispositionAreExplicit(t *testing.T) {
	t.Parallel()

	repoRoot := resolveRepoRoot(t)
	path := filepath.Join(repoRoot, "openspec", "migration", "beta-release-evidence-index.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read beta evidence index: %v", err)
	}
	content := string(raw)
	for _, fragment := range []string{
		"P0 release blocker",
		"new exact candidate commit",
		"e04ca27e",
		"2c0e51b2",
		"archived as read-only evidence",
		"must not be merged wholesale",
	} {
		if !strings.Contains(content, fragment) {
			t.Errorf("beta evidence index missing freeze/branch disposition %q", fragment)
		}
	}
}
