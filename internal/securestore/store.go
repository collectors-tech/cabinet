package securestore

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

const serviceName = "cabinet"

func Set(profileID, key, value string) error {
	return keyring.Set(serviceName, profileID+":"+key, value)
}

func Get(profileID, key string) (string, error) {
	v, err := keyring.Get(serviceName, profileID+":"+key)
	if err != nil {
		return "", fmt.Errorf("keyring get: %w", err)
	}
	return v, nil
}
