package launcher

import (
	"fmt"
	"net"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
)

func StartupURLFromAddr(addr string) string {
	host, port, err := net.SplitHostPort(strings.TrimSpace(addr))
	if err != nil {
		return "http://127.0.0.1:17880/"
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	if port == "" {
		port = "17880"
	}
	return (&url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, port),
		Path:   "/",
	}).String()
}

func OpenBrowser(targetURL string) error {
	name, args, err := browserCommandForOS(runtime.GOOS, targetURL)
	if err != nil {
		return err
	}
	return exec.Command(name, args...).Start()
}

func browserCommandForOS(goos string, targetURL string) (string, []string, error) {
	switch goos {
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler", targetURL}, nil
	case "darwin":
		return "open", []string{targetURL}, nil
	case "linux":
		return "xdg-open", []string{targetURL}, nil
	default:
		return "", nil, fmt.Errorf("unsupported platform for browser launch: %s", goos)
	}
}
