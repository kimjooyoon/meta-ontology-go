package freshness

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
)

const digestLength = sha256.Size * 2

// HashBytes returns a lowercase SHA-256 digest.
func HashBytes(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

// HashFile returns the lowercase SHA-256 digest of path.
func HashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return HashBytes(data), nil
}

// DigestInputs computes the stable digest recorded on artifacts and
// evidence. Input order does not matter; duplicate IDs are rejected.
func DigestInputs(ids []string, digests map[string]string) (string, error) {
	ordered := append([]string(nil), ids...)
	sort.Strings(ordered)
	if err := validateInputIDs(ordered, digests); err != nil {
		return "", err
	}
	buffer := make([]byte, 0, len(ordered)*(digestLength+2))
	for _, id := range ordered {
		buffer = append(buffer, id...)
		buffer = append(buffer, 0)
		buffer = append(buffer, digests[id]...)
		buffer = append(buffer, '\n')
	}
	return HashBytes(buffer), nil
}

func validateInputIDs(ids []string, digests map[string]string) error {
	for i, id := range ids {
		if id == "" {
			return fmt.Errorf("input ID must not be empty")
		}
		if i > 0 && ids[i-1] == id {
			return fmt.Errorf("duplicate input ID %q", id)
		}
		digest, ok := digests[id]
		if !ok {
			return fmt.Errorf("input %q is unavailable", id)
		}
		if !ValidDigest(digest) {
			return fmt.Errorf("input %q has invalid digest", id)
		}
	}
	return nil
}

// ValidDigest reports whether value is a lowercase SHA-256 hexadecimal
// digest, the format used by all freshness records.
func ValidDigest(value string) bool {
	if len(value) != digestLength {
		return false
	}
	for _, char := range value {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}
	return true
}
