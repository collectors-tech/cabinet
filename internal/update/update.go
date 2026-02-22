package update

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
)

type Channel string

const (
	ChannelStable Channel = "stable"
	ChannelBeta   Channel = "beta"
)

func ParseChannel(v string) Channel {
	switch Channel(v) {
	case ChannelBeta:
		return ChannelBeta
	default:
		return ChannelStable
	}
}

func VerifySignature(publicKeyBase64 string, payload []byte, signatureBase64 string) error {
	rawKey, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return fmt.Errorf("decode public key: %w", err)
	}
	if len(rawKey) != ed25519.PublicKeySize {
		return errors.New("invalid public key length")
	}

	sig, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	if len(sig) != ed25519.SignatureSize {
		return errors.New("invalid signature length")
	}

	ok := ed25519.Verify(ed25519.PublicKey(rawKey), payload, sig)
	if !ok {
		return errors.New("signature verification failed")
	}
	return nil
}
