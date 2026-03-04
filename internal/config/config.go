package config

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/collectors-tech/cabinet/internal/update"
)

type Config struct {
	Addr            string
	Host            string
	Port            int
	BindMode        string
	DataDir         string
	DBPath          string
	EnableE2EHooks  bool
	UpdateChannel   update.Channel
	UpdatePublicKey string
	WebAuthnRPID    string
	WebAuthnOrigin  string
	WebAuthnName    string
	BackupInterval  int
	ValidationError string
}

func Load() Config {
	loadDotEnv(".env")

	bindMode, bindModeErr := resolveBindMode()
	host, port, addr, addrValidationErr := resolveRuntimeAddress(bindMode)
	dataDir := valueOrDefault("CABINET_DATA_DIR", defaultDataDir())
	dbPath := valueOrDefault("CABINET_DB_PATH", filepath.Join(dataDir, "cabinet.db"))
	updateChannel := update.ParseChannel(valueOrDefault("CABINET_UPDATE_CHANNEL", "stable"))
	enableE2EHooks := parseBoolEnv(valueOrDefault("CABINET_E2E_MODE", "false"))
	updatePublicKey := os.Getenv("CABINET_UPDATE_PUBLIC_KEY")
	waRPID := valueOrDefault("CABINET_WEBAUTHN_RP_ID", "127.0.0.1")
	waOrigin := valueOrDefault("CABINET_WEBAUTHN_ORIGIN", "http://127.0.0.1:17880")
	waName := valueOrDefault("CABINET_WEBAUTHN_RP_NAME", "Cabinet")
	backupInterval, _ := strconv.Atoi(valueOrDefault("CABINET_BACKUP_INTERVAL_MINUTES", "60"))
	if backupInterval <= 0 {
		backupInterval = 60
	}
	validationErr := ""
	if strings.TrimSpace(bindModeErr) != "" {
		validationErr = strings.TrimSpace(bindModeErr)
	} else if strings.TrimSpace(addrValidationErr) != "" {
		validationErr = strings.TrimSpace(addrValidationErr)
	}

	return Config{
		Addr:            addr,
		Host:            host,
		Port:            port,
		BindMode:        bindMode,
		DataDir:         dataDir,
		DBPath:          dbPath,
		EnableE2EHooks:  enableE2EHooks,
		UpdateChannel:   updateChannel,
		UpdatePublicKey: updatePublicKey,
		WebAuthnRPID:    waRPID,
		WebAuthnOrigin:  waOrigin,
		WebAuthnName:    waName,
		BackupInterval:  backupInterval,
		ValidationError: validationErr,
	}
}

func resolveBindMode() (string, string) {
	raw := strings.ToLower(strings.TrimSpace(valueOrDefault("CABINET_BIND_MODE", "local")))
	switch raw {
	case "local", "lan":
		return raw, ""
	default:
		return "local", fmt.Sprintf("invalid CABINET_BIND_MODE: %q (expected local or lan)", raw)
	}
}

func resolveRuntimeAddress(bindMode string) (string, int, string, string) {
	const (
		defaultHost = "127.0.0.1"
		defaultPort = 17880
	)

	hostRaw := strings.TrimSpace(os.Getenv("CABINET_HOST"))
	portRaw := strings.TrimSpace(os.Getenv("CABINET_PORT"))
	addrRaw := strings.TrimSpace(os.Getenv("CABINET_ADDR"))

	if hostRaw != "" || portRaw != "" {
		host := hostRaw
		if host == "" {
			if strings.EqualFold(bindMode, "lan") {
				host = "0.0.0.0"
			} else {
				host = defaultHost
			}
		}
		port := defaultPort
		if portRaw != "" {
			parsedPort, err := strconv.Atoi(portRaw)
			if err != nil || parsedPort < 1 || parsedPort > 65535 {
				return host, defaultPort, net.JoinHostPort(host, strconv.Itoa(defaultPort)), fmt.Sprintf("invalid CABINET_PORT: %q (expected integer between 1 and 65535)", portRaw)
			}
			port = parsedPort
		}
		return host, port, net.JoinHostPort(host, strconv.Itoa(port)), ""
	}

	if addrRaw == "" {
		if strings.EqualFold(bindMode, "lan") {
			return "0.0.0.0", defaultPort, net.JoinHostPort("0.0.0.0", strconv.Itoa(defaultPort)), ""
		}
		return defaultHost, defaultPort, net.JoinHostPort(defaultHost, strconv.Itoa(defaultPort)), ""
	}

	host, portString, err := net.SplitHostPort(addrRaw)
	if err != nil {
		return defaultHost, defaultPort, net.JoinHostPort(defaultHost, strconv.Itoa(defaultPort)), fmt.Sprintf("invalid CABINET_ADDR: %q (expected host:port)", addrRaw)
	}
	port, err := strconv.Atoi(strings.TrimSpace(portString))
	if err != nil || port < 1 || port > 65535 {
		return host, defaultPort, net.JoinHostPort(host, strconv.Itoa(defaultPort)), fmt.Sprintf("invalid CABINET_ADDR port in %q", addrRaw)
	}
	if strings.TrimSpace(host) == "" {
		host = defaultHost
	}
	return host, port, addrRaw, ""
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}
		if existing, exists := os.LookupEnv(key); exists && strings.TrimSpace(existing) != "" {
			continue
		}
		value := strings.TrimSpace(parts[1])
		if len(value) >= 2 {
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) || (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
		}
		_ = os.Setenv(key, value)
	}
}

func valueOrDefault(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func parseBoolEnv(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func defaultDataDir() string {
	if resolved := resolveExecutableLocalDataDir(); strings.TrimSpace(resolved) != "" {
		return resolved
	}

	if runtime.GOOS == "windows" {
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			base = "."
		}
		return filepath.Join(base, "Cabinet")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ".cabinet"
	}
	return filepath.Join(home, ".cabinet")
}

func resolveExecutableLocalDataDir() string {
	if override := strings.TrimSpace(os.Getenv("CABINET_EXE_DIR")); override != "" {
		return filepath.Join(override, "data")
	}
	exePath, err := os.Executable()
	if err != nil || strings.TrimSpace(exePath) == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(exePath), "data")
}
