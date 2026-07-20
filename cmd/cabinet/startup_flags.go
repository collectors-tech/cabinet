package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/collectors-tech/cabinet/internal/config"
)

type startupOverrides struct {
	Env map[string]string
}

func parseStartupArgs(args []string) (startupOverrides, error) {
	fs := flag.NewFlagSet("cabinet", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var port int
	var listen string
	var dataDir string
	var profile string
	var instanceName string
	var authMode string
	var baseURL string
	var restart bool
	var allowParallel bool
	var seedSampleData bool
	var logLevel string
	var noOpenBrowser bool

	fs.IntVar(&port, "port", 0, "Runtime port override.")
	fs.StringVar(&listen, "listen", "", "Runtime listen address override (host:port).")
	fs.StringVar(&dataDir, "data-dir", "", "Data directory override.")
	fs.StringVar(&profile, "profile", "", "Profile key override.")
	fs.StringVar(&instanceName, "instance-name", "", "Instance/profile alias override.")
	fs.StringVar(&authMode, "auth-mode", "", "Auth mode override (local|clerk|zitadel).")
	fs.StringVar(&baseURL, "base-url", "", "Base URL override for runtime callbacks/origin.")
	fs.BoolVar(&restart, "restart", false, "Restart an already-running Cabinet instance on the requested endpoint.")
	fs.BoolVar(&allowParallel, "allow-parallel", false, "Allow parallel runtime instances.")
	fs.BoolVar(&seedSampleData, "seed-sample-data", false, "Seed idempotent sample data for the active profile on startup.")
	fs.StringVar(&logLevel, "log-level", "", "Log level override (debug|info|warn|error).")
	fs.BoolVar(&noOpenBrowser, "no-open-browser", false, "Disable browser auto-open on startup.")

	if err := fs.Parse(args); err != nil {
		return startupOverrides{}, err
	}

	env := map[string]string{}
	if port != 0 {
		if port < 1 || port > 65535 {
			return startupOverrides{}, fmt.Errorf("invalid --port value %d (expected 1-65535)", port)
		}
		env["CABINET_PORT"] = strconv.Itoa(port)
	}

	listen = strings.TrimSpace(listen)
	if listen != "" {
		host, portPart, err := net.SplitHostPort(listen)
		if err != nil {
			return startupOverrides{}, fmt.Errorf("invalid --listen value %q (expected host:port)", listen)
		}
		listenPort, err := strconv.Atoi(strings.TrimSpace(portPart))
		if err != nil || listenPort < 1 || listenPort > 65535 {
			return startupOverrides{}, fmt.Errorf("invalid --listen port in %q (expected 1-65535)", listen)
		}
		if port != 0 && port != listenPort {
			return startupOverrides{}, fmt.Errorf("conflicting --port (%d) and --listen (%s)", port, listen)
		}
		if strings.TrimSpace(host) != "" {
			env["CABINET_HOST"] = host
		}
		env["CABINET_PORT"] = strconv.Itoa(listenPort)
	}

	if trimmed := strings.TrimSpace(dataDir); trimmed != "" {
		env["CABINET_DATA_DIR"] = trimmed
	}

	profile = strings.TrimSpace(profile)
	instanceName = strings.TrimSpace(instanceName)
	if profile != "" {
		env["CABINET_PROFILE"] = profile
	} else if instanceName != "" {
		env["CABINET_PROFILE"] = instanceName
	}

	authMode = strings.ToLower(strings.TrimSpace(authMode))
	if authMode != "" {
		switch authMode {
		case "local", "clerk", "zitadel":
			env["CABINET_AUTH_MODE"] = authMode
			env["CABINET_AUTH_IDENTITY_MODE"] = authMode
		default:
			return startupOverrides{}, fmt.Errorf("invalid --auth-mode value %q (expected local, clerk or zitadel)", authMode)
		}
	}

	baseURL = strings.TrimSpace(baseURL)
	if baseURL != "" {
		parsed, err := url.Parse(baseURL)
		if err != nil || strings.TrimSpace(parsed.Scheme) == "" || strings.TrimSpace(parsed.Host) == "" {
			return startupOverrides{}, fmt.Errorf("invalid --base-url value %q", baseURL)
		}
		env["CABINET_BASE_URL"] = baseURL
	}

	if restart {
		env["CABINET_RESTART"] = "true"
	}

	if allowParallel {
		env["CABINET_ALLOW_PARALLEL"] = "true"
	}

	if seedSampleData {
		env["CABINET_SEED_SAMPLE_DATA"] = "true"
	}

	logLevel = strings.ToLower(strings.TrimSpace(logLevel))
	if logLevel != "" {
		switch logLevel {
		case "debug", "info", "warn", "error":
			env["CABINET_LOG_LEVEL"] = logLevel
		default:
			return startupOverrides{}, fmt.Errorf("invalid --log-level value %q (expected debug|info|warn|error)", logLevel)
		}
	}
	_ = noOpenBrowser

	if extra := fs.Args(); len(extra) > 0 {
		return startupOverrides{}, fmt.Errorf("unexpected positional arguments: %s", strings.Join(extra, " "))
	}

	return startupOverrides{Env: env}, nil
}

func applyStartupOverrides(overrides startupOverrides) {
	for key, value := range overrides.Env {
		_ = os.Setenv(key, value)
	}
}

func buildEffectiveStartupConfigLine(cfg config.Config) string {
	return fmt.Sprintf(
		"CABINET_EFFECTIVE_CONFIG addr=%s host=%s port=%d data_dir=%s profile=%s auth_mode=%s base_url=%s allow_parallel=%s seed_sample_data=%s log_level=%s",
		cfg.Addr,
		cfg.Host,
		cfg.Port,
		cfg.DataDir,
		envOrDefault("CABINET_PROFILE", "default"),
		envOrDefault("CABINET_AUTH_MODE", "local"),
		envOrDefault("CABINET_BASE_URL", fmt.Sprintf("http://%s", cfg.Addr)),
		envOrDefault("CABINET_ALLOW_PARALLEL", "false"),
		envOrDefault("CABINET_SEED_SAMPLE_DATA", "false"),
		envOrDefault("CABINET_LOG_LEVEL", "info"),
	)
}

func envOrDefault(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func validateStartupOverrides(overrides startupOverrides) error {
	if _, ok := overrides.Env["CABINET_RESTART"]; ok {
		if _, parallel := overrides.Env["CABINET_ALLOW_PARALLEL"]; parallel {
			return errors.New("restart cannot be combined with allow-parallel")
		}
	}
	if _, ok := overrides.Env["CABINET_ALLOW_PARALLEL"]; ok {
		if strings.TrimSpace(overrides.Env["CABINET_PROFILE"]) == "" && strings.TrimSpace(os.Getenv("CABINET_PROFILE")) == "" {
			return errors.New("allow-parallel requires --profile or --instance-name")
		}
	}
	return nil
}
