package bindingcoverage

import (
	"encoding/hex"
	"strings"
)

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	_, err := hex.DecodeString(value[len("sha256:"):])
	return err == nil
}
func evidenceKey(stage, reason string) string { return stage + "\x00" + reason }
func safeAdd(left, right int64) (int64, bool) {
	const maxInt64 = int64(1<<63 - 1)
	if right > 0 && left > maxInt64-right {
		return 0, true
	}
	return left + right, false
}
