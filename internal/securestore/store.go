package securestore

import (
	"fmt"
	"os"

	"github.com/zalando/go-keyring"
)

const serviceName = "cabinet"

func Set(profileID, key, value string) error {
	if os.Getenv("CABINET_FORCE_SECURESTORE_FAIL") == "1" {
		return fmt.Errorf("forced secure store failure")
	}
	return keyring.Set(serviceName, profileID+":"+key, value)
}

func Get(profileID, key string) (string, error) {
	if os.Getenv("CABINET_FORCE_SECURESTORE_FAIL") == "1" {
		return "", fmt.Errorf("forced secure store failure")
	}
	v, err := keyring.Get(serviceName, profileID+":"+key)
	if err != nil {
		return "", fmt.Errorf("keyring get: %w", err)
	}
	return v, nil
}
