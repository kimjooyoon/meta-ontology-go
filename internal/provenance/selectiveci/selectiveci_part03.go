package selectiveci

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"slices"
	"strings"
)

func (r Receipt) canonical() string {
	copy := canonicalReceipt(r)
	copy.Digest = ""
	data, err := marshalReceipt(copy)
	if err != nil {
		return ""
	}
	return string(data)
}
func (r Receipt) expectedDigest() string {
	return semantic.StableHashString("selective-ci-receipt/v1\x00" + r.canonical())
}

// Canonical returns the deterministic receipt payload without its self-digest.
func (r Receipt) Canonical() string { return r.canonical() }

// ExpectedDigest returns the digest that seals the canonical receipt payload.
func (r Receipt) ExpectedDigest() string { return r.expectedDigest() }
func canonicalReceipt(value Receipt) Receipt {
	copy := value
	copy.SelectedCommandIDs = append([]semantic.ID(nil), value.SelectedCommandIDs...)
	copy.ObligationIDs = append([]semantic.ID(nil), value.ObligationIDs...)
	copy.PathIDs = append([]semantic.ID(nil), value.PathIDs...)
	copy.VerifiedCommandIDs = append([]semantic.ID(nil), value.VerifiedCommandIDs...)
	copy.VerifiedObligationIDs = append([]semantic.ID(nil), value.VerifiedObligationIDs...)
	copy.VerifiedPathIDs = append([]semantic.ID(nil), value.VerifiedPathIDs...)
	slices.Sort(copy.SelectedCommandIDs)
	slices.Sort(copy.ObligationIDs)
	slices.Sort(copy.PathIDs)
	slices.Sort(copy.VerifiedCommandIDs)
	slices.Sort(copy.VerifiedObligationIDs)
	slices.Sort(copy.VerifiedPathIDs)
	return copy
}
func validDigest(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil && strings.ToLower(value) == value
}
func normalizeDigest(value, label string) (string, error) {
	value = strings.TrimSpace(value)
	if !validDigest(value) {
		return "", fmt.Errorf("%s must be a lowercase SHA-256 digest", label)
	}
	return value, nil
}
func normalizeID(value semantic.ID, label string) (semantic.ID, error) {
	id, err := semantic.ParseIdentity(value.String())
	if err != nil {
		return "", fmt.Errorf("%s: %w", label, err)
	}
	return id, nil
}
