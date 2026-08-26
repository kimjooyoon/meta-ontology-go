package artifactemit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

func symbolicReaderBytesDigest(payload []byte) string {
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func symbolicReaderReachabilityDigest(value SymbolicValueReachability) string {
	value.Digest = ""
	payload, _ := json.Marshal(value)
	return symbolicReaderBytesDigest(payload)
}

func symbolicReaderProjectionDigest(value SymbolicValueReaderProjection) string {
	value.Digest = ""
	payload, _ := json.Marshal(value)
	return symbolicReaderBytesDigest(payload)
}

func symbolicReaderValidDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != 71 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}
