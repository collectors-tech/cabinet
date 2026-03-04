package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/collectors-tech/cabinet/internal/app"
	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/launcher"
)

type browserLaunchSettings struct {
	Enabled     bool
	DisableNote string
}

func main() {
	cfg := config.Load()
	a, err := app.New(cfg)
	if err != nil {
		log.Fatalf("startup failed: %v", err)
	}
	openBrowserValue, openBrowserSet := os.LookupEnv("CABINET_OPEN_BROWSER")
	browserLaunch := resolveBrowserLaunch(os.Args[1:], openBrowserValue, openBrowserSet)
	if !browserLaunch.Enabled && strings.TrimSpace(browserLaunch.DisableNote) != "" {
		log.Printf("%s", browserLaunch.DisableNote)
	}
	if browserLaunch.Enabled {
		startupURL := launcher.StartupURLFromAddr(cfg.Addr)
		go func() {
			time.Sleep(300 * time.Millisecond)
			if err := launcher.OpenBrowser(startupURL); err != nil {
				log.Printf("browser launch skipped: %v", err)
			}
		}()
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := a.Run(ctx); err != nil {
		log.Fatalf("runtime failed: %v", err)
	}
}

func resolveBrowserLaunch(args []string, envValue string, envOK bool) browserLaunchSettings {
	if hasNoOpenBrowserFlag(args) {
		return browserLaunchSettings{
			Enabled:     false,
			DisableNote: "browser auto-open disabled (--no-open-browser)",
		}
	}
	if !openBrowserEnabled(envValue, envOK) {
		return browserLaunchSettings{
			Enabled:     false,
			DisableNote: "browser auto-open disabled (CABINET_OPEN_BROWSER)",
		}
	}
	return browserLaunchSettings{Enabled: true}
}

func hasNoOpenBrowserFlag(args []string) bool {
	for _, arg := range args {
		if strings.EqualFold(strings.TrimSpace(arg), "--no-open-browser") {
			return true
		}
	}
	return false
}

func openBrowserEnabled(value string, ok bool) bool {
	if !ok || strings.TrimSpace(value) == "" {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}
