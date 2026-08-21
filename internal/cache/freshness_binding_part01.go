package cache

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// FreshnessJob binds one required PR check to the exact workflow head. It is
// separate from content digests because Git commit IDs are identities, not
// SHA-256 digests of durable payloads.
type FreshnessJob struct {
	ID         string `json:"id"`
	RunID      string `json:"run_id"`
	Attempt    uint64 `json:"attempt"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HeadSHA    string `json:"head_sha"`
}

func validCommitSHA(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	if value != strings.ToLower(value) {
		return false
	}
	if _, err := hex.DecodeString(value); err != nil {
		return false
	}
	return value != strings.Repeat("0", len(value))
}
func validEventRef(value string) bool {
	parts := strings.Split(value, "/")
	if len(parts) != 4 || parts[0] != "refs" || parts[1] != "pull" ||
		(parts[3] != "head" && parts[3] != "merge") || strings.TrimSpace(value) != value {
		return false
	}
	if parts[2] == "" || parts[2] == "0" {
		return false
	}
	for _, char := range parts[2] {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
func validateFreshnessRefs(eventRef, checkoutRef, headSHA string) error {
	if !validEventRef(eventRef) {
		return fmt.Errorf("%w: malformed event ref", ErrInvalidReceipt)
	}
	if !validCommitSHA(checkoutRef) {
		return fmt.Errorf("%w: malformed checkout ref", ErrInvalidReceipt)
	}
	if checkoutRef != headSHA {
		return fmt.Errorf("%w: checkout ref does not match head SHA", ErrInvalidReceipt)
	}
	return nil
}
