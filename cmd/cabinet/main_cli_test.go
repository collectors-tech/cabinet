package main

import (
	"strings"
	"testing"
)

func TestResolveBrowserLaunch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		args     []string
		env      string
		envOK    bool
		wantOpen bool
		wantNote bool
	}{
		{
			name:     "default open when no flag and no env override",
			args:     []string{},
			env:      "",
			envOK:    false,
			wantOpen: true,
			wantNote: false,
		},
		{
			name:     "no-open-browser flag disables open and emits note",
			args:     []string{"--no-open-browser"},
			env:      "",
			envOK:    false,
			wantOpen: false,
			wantNote: true,
		},
		{
			name:     "flag disables even when env enables",
			args:     []string{"--no-open-browser"},
			env:      "true",
			envOK:    true,
			wantOpen: false,
			wantNote: true,
		},
		{
			name:     "env disables open and emits note",
			args:     []string{},
			env:      "false",
			envOK:    true,
			wantOpen: false,
			wantNote: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := resolveBrowserLaunch(tc.args, tc.env, tc.envOK)
			if got.Enabled != tc.wantOpen {
				t.Fatalf("resolveBrowserLaunch(...).Enabled = %v, want %v", got.Enabled, tc.wantOpen)
			}
			if (got.DisableNote != "") != tc.wantNote {
				t.Fatalf("resolveBrowserLaunch(...).DisableNote empty=%v, want note=%v", got.DisableNote == "", tc.wantNote)
			}
		})
	}
}

func TestRuntimeAttachLogLineIncludesURLPIDAndResolvedPort(t *testing.T) {
	t.Parallel()

	line := runtimeAttachLogLine(runtimeAttachDecision{
		Attach: true,
		URL:    "http://127.0.0.1:19090",
		PID:    4242,
	}, "C:/cabinet/data")
	for _, token := range []string{
		"CABINET_RUNTIME_ATTACH",
		"url=http://127.0.0.1:19090",
		"pid=4242",
		"data_dir=C:/cabinet/data",
		"resolved_port=19090",
	} {
		if !strings.Contains(line, token) {
			t.Fatalf("expected token %q in line %q", token, line)
		}
	}
}

func TestRuntimeAttachUserMessageForRequestedEndpointReuse(t *testing.T) {
	t.Parallel()

	line := runtimeAttachUserMessage(runtimeAttachDecision{
		Attach: true,
		URL:    "http://127.0.0.1:19090/",
		Reason: "requested_endpoint_healthy",
	})
	for _, token := range []string{
		"Cabinet is already running",
		"http://127.0.0.1:19090/",
		"--restart",
		"--allow-parallel",
	} {
		if !strings.Contains(line, token) {
			t.Fatalf("expected token %q in line %q", token, line)
		}
	}
}
