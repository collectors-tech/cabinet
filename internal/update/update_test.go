package update

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"
)

func TestParseChannel(t *testing.T) {
	t.Parallel()

	if got := ParseChannel("beta"); got != ChannelBeta {
		t.Fatalf("expected beta, got %q", got)
	}
	if got := ParseChannel("stable"); got != ChannelStable {
		t.Fatalf("expected stable, got %q", got)
	}
	if got := ParseChannel("unknown"); got != ChannelStable {
		t.Fatalf("expected fallback stable, got %q", got)
	}
}

func TestVerifySignature(t *testing.T) {
	t.Parallel()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	payload := []byte("cabinet-update-manifest")
	sig := ed25519.Sign(priv, payload)

	pubB64 := base64.StdEncoding.EncodeToString(pub)
	sigB64 := base64.StdEncoding.EncodeToString(sig)

	if err := VerifySignature(pubB64, payload, sigB64); err != nil {
		t.Fatalf("VerifySignature() error = %v", err)
	}

	if err := VerifySignature(pubB64, []byte("tampered"), sigB64); err == nil {
		t.Fatal("expected signature verification failure for tampered payload")
	}
}
