package proofchoicejudge

import (
	"encoding/hex"
	"strings"
)

func validate(input receipt) string {
	if input.Schema != schema || input.SourcePath == "" || !validDigest(input.SourceDigest) || input.FixedDenom != fixedDenominator {
		return "RECEIPT_FOUNDATION_UNKNOWN"
	}
	if input.Effects.RepositoryWrites != 0 || input.Effects.MutationAuthority {
		return "READ_ONLY_GUARD_FAILED"
	}
	if len(input.Items) == 0 {
		return "NO_PROOF_CHOICES"
	}
	claims := map[string]string{}
	for _, value := range input.Items {
		if value.ID == "" || value.Statement == "" || !validChoice(value.Choice) {
			return "PROOF_CHOICE_MISSING"
		}
		if unknown(value.Producer, value.Consumer, value.MetaOperation, value.Stage, value.Step, value.Reason) {
			return "UNKNOWN_CONTEXT"
		}
		if value.Kind != "CLAIM" && value.Kind != "METRIC" {
			return "ITEM_KIND_UNKNOWN"
		}
		if value.Kind == "METRIC" && (value.Denominator != fixedDenominator || value.Numerator < 0 || value.Numerator > value.Denominator) {
			return "FIXED_DENOMINATOR_MISMATCH"
		}
		if old, exists := claims[value.ID]; exists && old != value.Choice {
			return "PROOF_CHOICE_CONTRADICTION"
		}
		if value.Kind == "CLAIM" {
			claims[value.ID] = value.Choice
		}
	}
	return validateTransitions(input.Transitions, claims)
}

func validChoice(value string) bool {
	return value == "FOUNDATION" || value == "COHERENCE" || value == "REGRESSION"
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}
