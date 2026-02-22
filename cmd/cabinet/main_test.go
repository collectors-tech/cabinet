package main

import "testing"

func TestOpenBrowserEnabled(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		env  string
		ok   bool
		want bool
	}{
		{name: "missing env defaults true", ok: false, want: true},
		{name: "empty env defaults true", env: "", ok: true, want: true},
		{name: "false disables", env: "false", ok: true, want: false},
		{name: "0 disables", env: "0", ok: true, want: false},
		{name: "off disables", env: "off", ok: true, want: false},
		{name: "no disables", env: "no", ok: true, want: false},
		{name: "true enables", env: "true", ok: true, want: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := openBrowserEnabled(tc.env, tc.ok); got != tc.want {
				t.Fatalf("openBrowserEnabled(%q, %v) = %v, want %v", tc.env, tc.ok, got, tc.want)
			}
		})
	}
}
