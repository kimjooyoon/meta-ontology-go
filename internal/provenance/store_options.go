package provenance

import (
	"fmt"
	"strings"
	"time"
)

func checkReadOptions(records []Evidence, options ReadOptions) error {
	expected, err := expectedSourceDigest(options)
	if err != nil {
		return err
	}
	for _, record := range records {
		if expected != "" && record.SourceDigest != expected {
			return &FreshnessError{ID: record.ID, Kind: "source-mismatch", Expected: expected, Actual: record.SourceDigest}
		}
		if err := checkExpiry(record, options); err != nil {
			return err
		}
	}
	return checkVerifiedClaims(records, options.RequiredVerified)
}

func expectedSourceDigest(options ReadOptions) (string, error) {
	expected := strings.ToLower(strings.TrimSpace(options.ExpectedSourceDigest))
	legacy := strings.ToLower(strings.TrimSpace(options.ExpectedSourceHash))
	if expected != "" && legacy != "" && expected != legacy {
		return "", fmt.Errorf("expected source digest and source hash differ")
	}
	if expected == "" {
		expected = legacy
	}
	if expected == "" {
		return "", nil
	}
	return normalizeDigest(expected, "expected source digest")
}

func checkExpiry(record Evidence, options ReadOptions) error {
	if !options.RequireFresh || record.Freshness.ValidUntil == "" {
		return nil
	}
	now := options.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	validUntil, _ := time.Parse(time.RFC3339Nano, record.Freshness.ValidUntil)
	if !now.Before(validUntil) {
		return &FreshnessError{ID: record.ID, Kind: "expired", Actual: validUntil.Format(time.RFC3339Nano)}
	}
	return nil
}

func checkVerifiedClaims(records []Evidence, claims []VerifiedClaim) error {
	for _, claim := range claims {
		semanticID, err := normalizeIdentifier(claim.SemanticID, "verified claim semantic_id")
		if err != nil {
			return &ClaimError{SemanticID: claim.SemanticID, Kind: "invalid", Detail: err.Error()}
		}
		semanticDigest, err := normalizeDigest(claim.SemanticDigest, "verified claim semantic_digest")
		if err != nil {
			return &ClaimError{SemanticID: semanticID, Kind: "invalid", Detail: err.Error()}
		}
		graphDigest, err := normalizeDigest(claim.GraphDigest, "verified claim graph_digest")
		if err != nil {
			return &ClaimError{SemanticID: semanticID, Kind: "invalid", Detail: err.Error()}
		}
		if !claimMatches(records, semanticID, semanticDigest, graphDigest) {
			return &ClaimError{SemanticID: semanticID, Kind: "status-or-digest", Detail: "no verified event has both requested digests"}
		}
	}
	return nil
}

func claimMatches(records []Evidence, semanticID, semanticDigest, graphDigest string) bool {
	for _, record := range records {
		if record.SemanticID == semanticID && record.Status == StatusVerified && record.SemanticDigest == semanticDigest && record.GraphDigest == graphDigest {
			return true
		}
	}
	return false
}
