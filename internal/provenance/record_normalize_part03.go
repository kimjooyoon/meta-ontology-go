package provenance

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func normalizePredecessor(evidence *Evidence) error {
	if evidence.Predecessor == nil {
		return nil
	}
	link := *evidence.Predecessor
	var err error
	link.ID, err = normalizeIdentifier(link.ID, "predecessor.id")
	if err != nil {
		return err
	}
	link.Digest, err = normalizeDigest(link.Digest, "predecessor.digest")
	if err != nil {
		return err
	}
	evidence.Predecessor = &link
	return nil
}
func normalizeStatus(status EvidenceStatus) EvidenceStatus {
	switch strings.ToLower(strings.TrimSpace(string(status))) {
	case "passed", "pass":
		return StatusVerified
	case "not-run", "not_run":
		return StatusDeferred
	default:
		return EvidenceStatus(strings.ToLower(strings.TrimSpace(string(status))))
	}
}
func validStatus(status EvidenceStatus) bool {
	switch status {
	case StatusVerified, StatusCandidate, StatusDeferred, StatusFailed, StatusRejected:
		return true
	default:
		return false
	}
}
func normalizeIdentifier(value, field string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s is required", field)
	}
	if strings.ContainsAny(value, "\r\n") {
		return "", fmt.Errorf("%s must not contain line breaks", field)
	}
	return value, nil
}
func normalizeDigest(value, field string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) != sha256.Size*2 {
		return "", fmt.Errorf("%s must be a %d-character SHA-256 hex digest", field, sha256.Size*2)
	}
	if _, err := hex.DecodeString(value); err != nil {
		return "", fmt.Errorf("%s must be a SHA-256 hex digest: %w", field, err)
	}
	return value, nil
}
