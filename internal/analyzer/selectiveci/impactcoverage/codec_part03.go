package impactcoverage

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func normalizeResult(result Result) (Result, error) {
	if result.Schema != SchemaV1 {
		return Result{}, fmt.Errorf("impact coverage schema %q is invalid", result.Schema)
	}
	if result.Decision != DecisionExact && result.Decision != DecisionUnknown {
		return Result{}, fmt.Errorf("impact coverage decision %q is invalid", result.Decision)
	}
	if !validReason(result.Decision, result.Reason) {
		return Result{}, fmt.Errorf("impact coverage reason %q is invalid", result.Reason)
	}
	if result.FullSuiteRequired != (result.Decision == DecisionUnknown) {
		return Result{}, fmt.Errorf("impact coverage full-suite flag is inconsistent")
	}
	result.ChangedStableIDs = sortedUnique(result.ChangedStableIDs)
	result.UncoveredPaths = sortedUnique(result.UncoveredPaths)
	if result.Decision == DecisionUnknown && len(result.ChangedStableIDs) != 0 {
		return Result{}, fmt.Errorf("UNKNOWN result cannot contain changed stable IDs")
	}
	if result.ChangedStableIDs == nil {
		result.ChangedStableIDs = []string{}
	}
	if result.UncoveredPaths == nil {
		result.UncoveredPaths = []string{}
	}
	return result, nil
}
func validReason(decision Decision, reason Reason) bool {
	if decision == DecisionExact {
		return reason == ReasonComplete || reason == ReasonNoChange
	}
	return reason == ReasonMissingBinding || reason == ReasonAuthorityDrift || reason == ReasonInvalidSnapshot
}
func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	return requireEOF(decoder)
}
