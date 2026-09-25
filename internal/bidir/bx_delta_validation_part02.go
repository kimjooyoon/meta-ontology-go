package bidir

import (
	"crypto/sha256"
	"encoding/hex"
)

func isSHA256(value string) bool {
	if len(value) != hex.EncodedLen(sha256.Size) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
