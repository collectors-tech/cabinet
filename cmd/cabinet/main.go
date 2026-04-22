package main

import (
	"context"
	"errors"
	"flag"
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
	overrides, err := parseStartupArgs(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		log.Fatalf("invalid startup args: %v", err)
	}
	if err := validateStartupOverrides(overrides); err != nil {
		log.Fatalf("invalid startup args: %v", err)
	}
	applyStartupOverrides(overrides)

	cfg := config.Load()
	openBrowserValue, openBrowserSet := os.LookupEnv("CABINET_OPEN_BROWSER")
	log.Printf("%s", buildEffectiveStartupConfigLine(cfg))
	browserLaunch := resolveBrowserLaunch(os.Args[1:], openBrowserValue, openBrowserSet)
	allowParallel := startupAllowsParallel()
	restartRequested := startupWantsRestart()
	requestedProbe := resolveRequestedRuntimeProbe(cfg, fetchRuntimeEndpointProbe, isRuntimeAddrInUse)
	log.Printf("%s", runtimeEndpointStatusLogLine(requestedProbe))
	if restartRequested {
		switch requestedProbe.Status {
		case "occupied":
			log.Fatalf("runtime restart failed: requested endpoint %s is occupied by a non-Cabinet listener", requestedProbe.URL)
		case "cabinet":
			restartResult, err := restartRequestedRuntime(
				requestedProbe,
				cfg.Addr,
				requestRuntimeShutdown,
				waitForRuntimeRestartReady,
				isRuntimeProcessAlive,
				terminateRuntimeProcess,
			)
			if err != nil {
				log.Fatalf("runtime restart failed: %v", err)
			}
			log.Printf("%s", runtimeRestartLogLine(restartResult, cfg.DataDir))
		}
	}
	attachDecision, attachErr := resolveRunningRuntimeAttach(cfg.DataDir, isRuntimeProcessAlive, isRuntimeEndpointHealthy)
	if attachErr != nil {
		log.Printf("runtime attach check skipped: %v", attachErr)
	} else if attachDecision.Attach {
		log.Printf("%s", runtimeAttachLogLine(attachDecision, cfg.DataDir))
		if !browserLaunch.Enabled {
			if strings.TrimSpace(browserLaunch.DisableNote) != "" {
				log.Printf("%s", browserLaunch.DisableNote)
			}
			return
		}
		if err := launcher.OpenBrowser(attachDecision.URL); err != nil {
			log.Printf("browser attach launch skipped: %v", err)
		}
		return
	}
	requestedAttachDecision := runtimeAttachDecision{}
	if !restartRequested && requestedProbe.Status == "cabinet" {
		requestedAttachDecision = resolveRequestedRuntimeAttach(cfg, allowParallel, isRuntimeEndpointHealthy)
	}
	if requestedAttachDecision.Attach {
		log.Printf("%s", runtimeAttachLogLine(requestedAttachDecision, cfg.DataDir))
		if !browserLaunch.Enabled {
			if strings.TrimSpace(browserLaunch.DisableNote) != "" {
				log.Printf("%s", browserLaunch.DisableNote)
			}
			return
		}
		if err := launcher.OpenBrowser(requestedAttachDecision.URL); err != nil {
			log.Printf("browser attach launch skipped: %v", err)
		}
		return
	}

	a, err := app.New(cfg)
	if err != nil {
		log.Fatalf("startup failed: %v", err)
	}
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

func startupAllowsParallel() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("CABINET_ALLOW_PARALLEL"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
