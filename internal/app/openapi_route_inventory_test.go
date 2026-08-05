package app

import (
	"bufio"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type runtimeRouteContract struct {
	Path    string
	Methods []string
}

type openAPIOperationContract struct {
	OperationID    string
	HasClientError bool
}

func TestOpenAPIParitySuite(t *testing.T) {
	t.Parallel()

	runtimeRoutes := discoverRuntimeRouteInventory(t)
	openAPIRoutes := readOpenAPIRouteInventory(t)

	t.Run("runtime and documented routes agree in both directions", func(t *testing.T) {
		var mismatches []string
		for path, runtimeMethods := range runtimeRoutes {
			documentedMethods, ok := openAPIRoutes[path]
			if !ok {
				methods := make([]string, 0, len(runtimeMethods))
				for method := range runtimeMethods {
					methods = append(methods, strings.ToUpper(method))
				}
				sort.Strings(methods)
				mismatches = append(mismatches, "runtime-only path "+path+" ("+strings.Join(methods, ", ")+")")
				continue
			}
			mismatches = append(mismatches, methodMismatches(path, runtimeMethods, documentedMethods)...)
		}
		for path := range openAPIRoutes {
			if _, ok := runtimeRoutes[path]; !ok {
				mismatches = append(mismatches, "documented-only path "+path)
			}
		}
		sort.Strings(mismatches)
		if len(mismatches) > 0 {
			t.Fatalf("runtime/OpenAPI route parity failed:\n%s", strings.Join(mismatches, "\n"))
		}
	})

	t.Run("operations have unique identifiers and client errors", func(t *testing.T) {
		seenOperationIDs := make(map[string]string)
		var failures []string
		for path, methods := range openAPIRoutes {
			for method, operation := range methods {
				key := strings.ToUpper(method) + " " + path
				if operation.OperationID == "" {
					failures = append(failures, key+" has no operationId")
				} else if previous, exists := seenOperationIDs[operation.OperationID]; exists {
					failures = append(failures, key+" duplicates operationId "+operation.OperationID+" from "+previous)
				} else {
					seenOperationIDs[operation.OperationID] = key
				}
				if !operation.HasClientError {
					failures = append(failures, key+" has no 4XX response")
				}
			}
		}
		sort.Strings(failures)
		if len(failures) > 0 {
			t.Fatalf("OpenAPI operation contract failed:\n%s", strings.Join(failures, "\n"))
		}
	})

	t.Run("critical security and payload contracts are explicit", func(t *testing.T) {
		_, raw := readOpenAPISpec(t)
		for path, required := range map[string][]string{
			"/api/companion/pairing/requests": {
				"security: []", "CabinetLocalSession", "CompanionDeviceID", "CompanionPairingRequestInput", "CompanionPairingReceipt", "CompanionPairingSummary", `"413":`, `"429":`,
			},
			"/api/companion/pairing/approvals": {
				"CabinetLocalSession", "CabinetOIDCSession", "CompanionPairingApproval", "CompanionPairingSummary", `"401":`, `"409":`,
			},
			"/api/companion/pairing/exchanges": {
				"security: []", "CompanionPairingExchange", "CompanionCredentialResponse", "Cache-Control", `"409":`, `"429":`,
			},
			"/api/companion/session": {
				"CompanionProfileBearer", "CompanionDeviceID", "CompanionSession", "Calling credential revoked", `"401":`, `"429":`,
			},
			"/api/companion/session/rotate": {
				"CompanionProfileBearer", "CompanionCredentialResponse", "Cache-Control", `"401":`, `"429":`,
			},
			"/api/companion/sessions": {
				"CabinetLocalSession", "CabinetOIDCSession", "CompanionSession", "revoked_count", `"401":`,
			},
			"/api/companion/modules": {
				"CompanionProfileBearer", "profile-scoped registry", "application/json", "CompanionModuleRegistry", `"401":`, `"429":`,
			},
			"/api/companion/payloads": {
				"CompanionProfileBearer", "application/json", "CompanionPayloadSubmission", "CompanionAcceptedPayload", `"413":`, `"429":`,
			},
			"/api/companion/media-submissions": {
				"CompanionProfileBearer", "X-Cabinet-Profile", "X-Cabinet-Idempotency-Key", "X-Cabinet-Media-SHA256", "8388608", `"413":`, `"501":`,
			},
			"/api/profiles/{profileID}/integration-instances": {
				"CabinetLocalSession", "CabinetOIDCSession", "application/json", "IntegrationInstancePatch", "IntegrationInstanceList", "IntegrationInstanceResponse", "name: id", `"400":`,
			},
			"/api/providers/bonza/run": {
				"CabinetLocalSession", "CabinetOIDCSession", "application/json", "ProviderSearchRunRequest", "ProviderRunResponse", `"400":`,
			},
			"/api/providers/frontline/discovery": {
				"CabinetLocalSession", "CabinetOIDCSession", "application/json", "FrontlineDiscoveryRequest", "FrontlineDiscoveryResponse", `"400":`,
			},
			"/api/providers/frontline/run": {
				"CabinetLocalSession", "CabinetOIDCSession", "application/json", "BrowserAwareProviderRunRequest", "ProviderRunResponse", `"400":`,
			},
			"/api/providers/hobbytech/run": {
				"CabinetLocalSession", "CabinetOIDCSession", "application/json", "BrowserAwareProviderRunRequest", "ProviderRunResponse", `"400":`,
			},
		} {
			section, ok := openAPIPathSection(raw, path)
			if !ok {
				t.Errorf("OpenAPI security contract missing path %s", path)
				continue
			}
			for _, fragment := range required {
				if !strings.Contains(section, fragment) {
					t.Errorf("OpenAPI %s security contract missing %q", path, fragment)
				}
			}
		}
	})
}

func discoverRuntimeRouteInventory(t *testing.T) map[string]map[string]openAPIOperationContract {
	t.Helper()

	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	repoRoot := filepath.Clean(filepath.Join(workingDirectory, "..", ".."))
	appDir := filepath.Join(repoRoot, "internal", "app")
	parsed, err := parser.ParseDir(token.NewFileSet(), appDir, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parse runtime routes: %v", err)
	}

	result := make(map[string]map[string]openAPIOperationContract)
	for _, pkg := range parsed {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok || len(call.Args) < 2 {
					return true
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || selector.Sel.Name != "HandleFunc" {
					return true
				}
				pathLiteral, ok := call.Args[0].(*ast.BasicLit)
				if !ok || pathLiteral.Kind != token.STRING {
					return true
				}
				path, err := strconv.Unquote(pathLiteral.Value)
				if err != nil || (path != "/healthz" && !strings.HasPrefix(path, "/api/")) {
					return true
				}
				if reason, excluded := runtimeRouteExclusions[path]; excluded {
					if strings.TrimSpace(reason) == "" {
						t.Fatalf("runtime route exclusion %s has no reviewed reason", path)
					}
					return true
				}
				if family, ok := runtimeRouteFamilies[path]; ok {
					for _, route := range family {
						addRuntimeRoute(t, result, route.Path, route.Methods)
					}
					return true
				}
				methods := methodsUsedByHandler(call.Args[1])
				if override, ok := runtimeRouteMethodOverrides[path]; ok {
					methods = override
				}
				if len(methods) == 0 {
					t.Fatalf("runtime route %s has no discoverable HTTP method; add a reviewed override", path)
				}
				addRuntimeRoute(t, result, path, methods)
				return true
			})
		}
	}
	return result
}

func methodsUsedByHandler(handler ast.Expr) []string {
	found := make(map[string]struct{})
	ast.Inspect(handler, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		identifier, ok := selector.X.(*ast.Ident)
		if !ok || identifier.Name != "http" || !strings.HasPrefix(selector.Sel.Name, "Method") {
			return true
		}
		method := strings.ToLower(strings.TrimPrefix(selector.Sel.Name, "Method"))
		if method != "head" {
			found[method] = struct{}{}
		}
		return true
	})
	methods := make([]string, 0, len(found))
	for method := range found {
		methods = append(methods, method)
	}
	sort.Strings(methods)
	return methods
}

func addRuntimeRoute(t *testing.T, inventory map[string]map[string]openAPIOperationContract, path string, methods []string) {
	t.Helper()
	if _, exists := inventory[path]; exists {
		t.Fatalf("duplicate runtime route inventory path %s", path)
	}
	inventory[path] = make(map[string]openAPIOperationContract, len(methods))
	for _, method := range methods {
		inventory[path][strings.ToLower(method)] = openAPIOperationContract{}
	}
}

func readOpenAPIRouteInventory(t *testing.T) map[string]map[string]openAPIOperationContract {
	t.Helper()

	specPath, raw := readOpenAPISpec(t)
	pathPattern := regexp.MustCompile(`^  (/[^:]+):\s*$`)
	methodPattern := regexp.MustCompile(`^    (get|post|put|patch|delete|options):\s*$`)
	operationIDPattern := regexp.MustCompile(`^      operationId:\s*([A-Za-z][A-Za-z0-9]*)\s*$`)
	clientErrorPattern := regexp.MustCompile(`^        ["']?4[0-9]{2}["']?:`)

	inventory := make(map[string]map[string]openAPIOperationContract)
	currentPath := ""
	currentMethod := ""
	scanner := bufio.NewScanner(strings.NewReader(string(raw)))
	for scanner.Scan() {
		line := scanner.Text()
		if match := pathPattern.FindStringSubmatch(line); len(match) == 2 {
			currentPath = match[1]
			currentMethod = ""
			inventory[currentPath] = make(map[string]openAPIOperationContract)
			continue
		}
		if match := methodPattern.FindStringSubmatch(line); len(match) == 2 && currentPath != "" {
			currentMethod = match[1]
			inventory[currentPath][currentMethod] = openAPIOperationContract{}
			continue
		}
		if currentPath == "" || currentMethod == "" {
			continue
		}
		operation := inventory[currentPath][currentMethod]
		if match := operationIDPattern.FindStringSubmatch(line); len(match) == 2 {
			operation.OperationID = match[1]
		}
		if clientErrorPattern.MatchString(line) {
			operation.HasClientError = true
		}
		inventory[currentPath][currentMethod] = operation
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan %s: %v", specPath, err)
	}
	return inventory
}

func methodMismatches(path string, runtime, documented map[string]openAPIOperationContract) []string {
	var mismatches []string
	for method := range runtime {
		if _, ok := documented[method]; !ok {
			mismatches = append(mismatches, "runtime-only operation "+strings.ToUpper(method)+" "+path)
		}
	}
	for method := range documented {
		if _, ok := runtime[method]; !ok {
			mismatches = append(mismatches, "documented-only operation "+strings.ToUpper(method)+" "+path)
		}
	}
	return mismatches
}

var runtimeRouteMethodOverrides = map[string][]string{
	"/healthz":                   {"get"},
	"/api/runtime":               {"get"},
	"/api/auth/zitadel/login":    {"get"},
	"/api/auth/zitadel/callback": {"get"},
	"/api/auth/zitadel/session":  {"get"},
	"/api/auth/zitadel/refresh":  {"post"},
	"/api/auth/zitadel/logout":   {"post"},
}

var runtimeRouteExclusions = map[string]string{
	"/api/test/reset":                       "E2E-only reset hook registered only when CABINET_E2E_MODE is enabled.",
	"/api/test/bootstrap":                   "E2E-only fixture bootstrap hook registered only when CABINET_E2E_MODE is enabled.",
	"/api/test/scale/bootstrap":             "E2E-only scale fixture hook registered only when CABINET_E2E_MODE is enabled.",
	"/api/test/runtime/setup-status":        "E2E-only runtime setup fixture hook registered only when CABINET_E2E_MODE is enabled.",
	"/api/test/runtime/setup-config":        "E2E-only runtime setup fixture hook registered only when CABINET_E2E_MODE is enabled.",
	"/api/test/runtime/setup-import-source": "E2E-only setup-import fixture hook registered only when CABINET_E2E_MODE is enabled.",
	"/api/test/auth/provider-options":       "E2E-only authentication fixture hook registered only when CABINET_E2E_MODE is enabled.",
}

var runtimeRouteFamilies = map[string][]runtimeRouteContract{
	"/api/profiles/": {
		{Path: "/api/profiles/{profileID}/settings", Methods: []string{"get", "put"}},
		{Path: "/api/profiles/{profileID}/mcp-http-status", Methods: []string{"get"}},
		{Path: "/api/profiles/{profileID}/mcp-http-credential", Methods: []string{"post"}},
		{Path: "/api/profiles/{profileID}/mcp-http-config", Methods: []string{"put"}},
		{Path: "/api/profiles/{profileID}/integration-instances", Methods: []string{"delete", "get", "post", "put"}},
		{Path: "/api/profiles/{profileID}/saved-filters", Methods: []string{"delete", "get", "post", "put"}},
		{Path: "/api/profiles/{profileID}/storage", Methods: []string{"get"}},
		{Path: "/api/profiles/{profileID}/secrets", Methods: []string{"delete", "get", "put"}},
		{Path: "/api/profiles/{profileID}/license", Methods: []string{"get", "put"}},
	},
	"/api/users/": {
		{Path: "/api/users/{userID}", Methods: []string{"delete", "put"}},
	},
	"/api/barcodes/": {
		{Path: "/api/barcodes/{barcode}", Methods: []string{"get"}},
		{Path: "/api/barcodes/{barcode}/external-search", Methods: []string{"get"}},
	},
	"/api/scanner/query-sets/": {
		{Path: "/api/scanner/query-sets/{querySetID}", Methods: []string{"delete", "put"}},
	},
	"/api/scanner/candidates/": {
		{Path: "/api/scanner/candidates/{candidate_id}", Methods: []string{"patch"}},
	},
	"/api/wishlist/": {
		{Path: "/api/wishlist/{wishlistID}/restore", Methods: []string{"post"}},
	},
	"/api/chat/workflow-runs/": {
		{Path: "/api/chat/workflow-runs/{runID}", Methods: []string{"get", "patch"}},
	},
	"/api/chat/inbox/": {
		{Path: "/api/chat/inbox/{inboxID}", Methods: []string{"patch"}},
	},
	"/api/media/assets/": {
		{Path: "/api/media/assets/{assetID}/metadata", Methods: []string{"patch"}},
	},
	"/api/items/": {
		{Path: "/api/items/bulk-edit", Methods: []string{"post"}},
		{Path: "/api/items/{itemID}", Methods: []string{"delete", "put"}},
		{Path: "/api/items/{itemID}/restore", Methods: []string{"post"}},
		{Path: "/api/items/{itemID}/barcodes", Methods: []string{"get", "post"}},
		{Path: "/api/items/{itemID}/instances", Methods: []string{"get", "post"}},
		{Path: "/api/items/{itemID}/instances/{instanceID}", Methods: []string{"put"}},
		{Path: "/api/items/{itemID}/photos", Methods: []string{"get", "post"}},
		{Path: "/api/items/{itemID}/photos/reorder", Methods: []string{"post"}},
		{Path: "/api/items/{itemID}/photos/{photoID}", Methods: []string{"delete"}},
		{Path: "/api/items/{itemID}/photos/{photoID}/primary", Methods: []string{"put"}},
		{Path: "/api/items/{itemID}/photos/{photoID}/rotate", Methods: []string{"put"}},
		{Path: "/api/items/{itemID}/photos/{photoID}/file", Methods: []string{"get"}},
		{Path: "/api/items/{itemID}/photos-rebuild", Methods: []string{"post"}},
	},
}
