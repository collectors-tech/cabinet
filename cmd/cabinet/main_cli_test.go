package main

import "testing"

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
