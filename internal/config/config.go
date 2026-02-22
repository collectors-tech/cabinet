package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/collectors-tech/cabinet/internal/update"
)

type Config struct {
	Addr            string
	DataDir         string
	DBPath          string
	UpdateChannel   update.Channel
	UpdatePublicKey string
	WebAuthnRPID    string
	WebAuthnOrigin  string
	WebAuthnName    string
	BackupInterval  int
}

func Load() Config {
	addr := valueOrDefault("CABINET_ADDR", "127.0.0.1:8080")
	dataDir := valueOrDefault("CABINET_DATA_DIR", defaultDataDir())
	dbPath := valueOrDefault("CABINET_DB_PATH", filepath.Join(dataDir, "cabinet.db"))
	updateChannel := update.ParseChannel(valueOrDefault("CABINET_UPDATE_CHANNEL", "stable"))
	updatePublicKey := os.Getenv("CABINET_UPDATE_PUBLIC_KEY")
	waRPID := valueOrDefault("CABINET_WEBAUTHN_RP_ID", "127.0.0.1")
	waOrigin := valueOrDefault("CABINET_WEBAUTHN_ORIGIN", "http://127.0.0.1:8080")
	waName := valueOrDefault("CABINET_WEBAUTHN_RP_NAME", "Cabinet")
	backupInterval, _ := strconv.Atoi(valueOrDefault("CABINET_BACKUP_INTERVAL_MINUTES", "60"))
	if backupInterval <= 0 {
		backupInterval = 60
	}
	return Config{
		Addr:            addr,
		DataDir:         dataDir,
		DBPath:          dbPath,
		UpdateChannel:   updateChannel,
		UpdatePublicKey: updatePublicKey,
		WebAuthnRPID:    waRPID,
		WebAuthnOrigin:  waOrigin,
		WebAuthnName:    waName,
		BackupInterval:  backupInterval,
	}
}

func valueOrDefault(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
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
