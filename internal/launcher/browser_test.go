package launcher

import "testing"

func TestBrowserCommandForOS(t *testing.T) {
	t.Parallel()

	name, args, err := browserCommandForOS("windows", "http://127.0.0.1:8080/")
	if err != nil {
		t.Fatalf("windows command error = %v", err)
	}
	if name == "" || len(args) == 0 {
		t.Fatalf("invalid windows command: %q %v", name, args)
	}

	name, args, err = browserCommandForOS("darwin", "http://127.0.0.1:8080/")
	if err != nil {
		t.Fatalf("darwin command error = %v", err)
	}
	if name != "open" || len(args) != 1 {
		t.Fatalf("invalid darwin command: %q %v", name, args)
	}

	name, args, err = browserCommandForOS("linux", "http://127.0.0.1:8080/")
	if err != nil {
		t.Fatalf("linux command error = %v", err)
	}
	if name != "xdg-open" || len(args) != 1 {
		t.Fatalf("invalid linux command: %q %v", name, args)
	}
}

func TestStartupURLFromAddr(t *testing.T) {
	t.Parallel()

	if got := StartupURLFromAddr("127.0.0.1:8080"); got != "http://127.0.0.1:8080/" {
		t.Fatalf("unexpected url: %q", got)
	}
	if got := StartupURLFromAddr("0.0.0.0:8080"); got != "http://127.0.0.1:8080/" {
		t.Fatalf("unexpected wildcard url: %q", got)
	}
	if got := StartupURLFromAddr(":8080"); got != "http://127.0.0.1:8080/" {
		t.Fatalf("unexpected empty-host url: %q", got)
	}
}
