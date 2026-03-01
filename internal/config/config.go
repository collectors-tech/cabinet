package config

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/collectors-tech/cabinet/internal/update"
)

type Config struct {
	Addr            string
	DataDir         string
	DBPath          string
	EnableE2EHooks  bool
	UpdateChannel   update.Channel
	UpdatePublicKey string
	WebAuthnRPID    string
	WebAuthnOrigin  string
	WebAuthnName    string
	BackupInterval  int
}

func Load() Config {
	loadDotEnv(".env")

	addr := valueOrDefault("CABINET_ADDR", "127.0.0.1:17880")
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
	return Config{
		Addr:            addr,
		DataDir:         dataDir,
		DBPath:          dbPath,
		EnableE2EHooks:  enableE2EHooks,
		UpdateChannel:   updateChannel,
		UpdatePublicKey: updatePublicKey,
		WebAuthnRPID:    waRPID,
		WebAuthnOrigin:  waOrigin,
		WebAuthnName:    waName,
		BackupInterval:  backupInterval,
	}
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
