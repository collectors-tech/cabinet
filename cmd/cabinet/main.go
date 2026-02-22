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

func main() {
	cfg := config.Load()
	a, err := app.New(cfg)
	if err != nil {
		log.Fatalf("startup failed: %v", err)
	}
	if openBrowserEnabled(os.LookupEnv("CABINET_OPEN_BROWSER")) {
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
