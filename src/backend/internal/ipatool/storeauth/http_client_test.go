package storeauth

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/pem"
	"testing"
)

func TestEmbeddedAppleRootCAMatchesPinnedHash(t *testing.T) {
	block, _ := pem.Decode(appleRootCAPEM)
	if block == nil {
		t.Fatal("embedded Apple root PEM does not decode")
	}

	sum := sha256.Sum256(block.Bytes)
	if got := hex.EncodeToString(sum[:]); got != appleRootDERHash {
		t.Fatalf("embedded Apple root DER SHA-256 = %s, want %s", got, appleRootDERHash)
	}
}
