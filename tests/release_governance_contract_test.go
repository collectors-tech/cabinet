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
			"#2036 [DONE]",
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

func TestTaskListNamesTheCurrentPrivateBetaCriticalPathTruthfully(t *testing.T) {
	t.Parallel()

	repoRoot := resolveRepoRoot(t)
	raw, err := os.ReadFile(filepath.Join(repoRoot, "TASK_LIST.md"))
	if err != nil {
		t.Fatalf("read TASK_LIST.md: %v", err)
	}
	content := string(raw)

	for _, fragment := range []string{
		"#2050 [DONE]",
		"#2051 [DONE]",
		"#2052 [DONE]",
		"#2053 [DONE]",
		"#2054 [DONE]",
		"#2055 [DONE]",
		"#2056 [DONE]",
		"#2062 [DONE]",
		"#2064 [DONE]",
		"#2065 [DONE]",
		"#2057 -> #2048 -> refresh #2066",
		"repository settings were later applied",
		"verified compliant",
		"local unstaged product-documentation",
		"local acceptance recorder",
		"user-present live proof",
		"do not satisfy #1944 or #1945",
		"exact private/internal candidate",
		"exact packaged Windows acceptance",
		"same-candidate data safety",
		"owner/legal decisions",
		"explicit #1864 approval",
	} {
		if !strings.Contains(content, fragment) {
			t.Errorf("TASK_LIST.md missing current release-path contract %q", fragment)
		}
	}

	for _, stale := range []string{
		"#2050 [READY]",
		"#2053 [READY]",
		"#2054 [READY]",
		"#2055 [READY]",
		"#2051 [READY]",
		"#2056 [IN PROGRESS]",
		"#2051 -> #2052 -> #2056 -> #2057 -> #2062 -> #2064 -> #2065 -> #2048",
		"#2051 -> #2056 -> #2057 -> #2048 -> refresh #2066",
		"staged-only dependency/security patch",
		"local unstaged branch-protection",
		"| #2052 | In progress",
		"| #2056 | In progress",
		"| #2062 | In progress",
		"| #2064 | In progress",
		"| #2065 | In progress",
		"remaining source patches are locally prepared and unmerged",
		"1 critical and 6 high production package families",
		"develop` is 1,766 commits ahead",
	} {
		if strings.Contains(content, stale) {
			t.Errorf("TASK_LIST.md retains stale release-path claim %q", stale)
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
