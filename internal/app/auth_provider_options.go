package app

import (
	"os"
	"strings"
	"sync"
)

type authProviderOption struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Enabled bool   `json:"enabled"`
}

type authProviderOptionsPayload struct {
	IdentityMode    string               `json:"identity_mode"`
	ClerkConfigured bool                 `json:"clerk_configured"`
	Providers       []authProviderOption `json:"providers"`
}

var (
	authProviderOptionsMu          sync.RWMutex
	authProviderOptionsE2EOverride *authProviderOptionsPayload
)

func resetAuthProviderOptionsOverride() {
	authProviderOptionsMu.Lock()
	defer authProviderOptionsMu.Unlock()
	authProviderOptionsE2EOverride = nil
}

func setAuthProviderOptionsOverride(payload authProviderOptionsPayload) {
	authProviderOptionsMu.Lock()
	defer authProviderOptionsMu.Unlock()
	copyPayload := payload
	authProviderOptionsE2EOverride = &copyPayload
}

func resolveAuthProviderOptions() authProviderOptionsPayload {
	authProviderOptionsMu.RLock()
	override := authProviderOptionsE2EOverride
	authProviderOptionsMu.RUnlock()
	if override != nil {
		return *override
	}
	return defaultAuthProviderOptionsFromEnv()
}

func defaultAuthProviderOptionsFromEnv() authProviderOptionsPayload {
	clerkKeyConfigured := strings.TrimSpace(os.Getenv("VITE_CLERK_PUBLISHABLE_KEY")) != ""

	identityMode := strings.TrimSpace(strings.ToLower(os.Getenv("CABINET_AUTH_IDENTITY_MODE")))
	if identityMode == "" {
		if clerkKeyConfigured {
			identityMode = "clerk"
		} else {
			identityMode = "local"
		}
	}
	if identityMode != "clerk" {
		identityMode = "local"
	}

	return authProviderOptionsPayload{
		IdentityMode:    identityMode,
		ClerkConfigured: clerkKeyConfigured,
		Providers: []authProviderOption{
			{
				ID:      "google",
				Label:   "Google",
				Enabled: providerEnabledFromEnv("CABINET_AUTH_PROVIDER_GOOGLE_ENABLED", true),
			},
			{
				ID:      "apple",
				Label:   "Apple",
				Enabled: providerEnabledFromEnv("CABINET_AUTH_PROVIDER_APPLE_ENABLED", true),
			},
			{
				ID:      "microsoft",
				Label:   "Microsoft",
				Enabled: providerEnabledFromEnv("CABINET_AUTH_PROVIDER_MICROSOFT_ENABLED", true),
			},
		},
	}
}

func providerEnabledFromEnv(envKey string, fallback bool) bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(envKey)))
	if raw == "" {
		return fallback
	}
	switch raw {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
