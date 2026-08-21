package bindingcoverage

import (
	"sort"
	"strings"
)

func missingEvidence(input Input) ([]string, []string) {
	seen := make(map[string]map[string]bool, len(input.RequiredBindings))
	for _, binding := range input.RequiredBindings {
		seen[binding.ID] = map[string]bool{PolarityMatch: false, PolarityMismatch: false}
	}
	for _, partition := range input.Partitions {
		if seen[partition.BindingID] != nil {
			seen[partition.BindingID][partition.Polarity] = true
		}
	}
	missingMatch, missingMismatch := make([]string, 0), make([]string, 0)
	for id, polarities := range seen {
		if !polarities[PolarityMatch] {
			missingMatch = append(missingMatch, id)
		}
		if !polarities[PolarityMismatch] {
			missingMismatch = append(missingMismatch, id)
		}
	}
	sort.Strings(missingMatch)
	sort.Strings(missingMismatch)
	return missingMatch, missingMismatch
}
func incompleteReason(match, mismatch []string) string {
	if len(match) != 0 && len(mismatch) != 0 {
		return "MISSING_MATCH_AND_MISMATCH"
	}
	if len(match) != 0 {
		return "MISSING_MATCH"
	}
	return "MISSING_MISMATCH"
}
func finish(vector Vector, decision, reason string) Result {
	vector.Decision = decision
	vector.Reason = reason
	vector.MissingMatch = append([]string{}, vector.MissingMatch...)
	vector.MissingMismatch = append([]string{}, vector.MissingMismatch...)
	return Result{Vector: vector, CanonicalDigest: digestVector(vector)}
}
func endpointReferenceCount(bindings []Binding) (int64, bool) {
	return safeAdd(int64(len(bindings)), int64(len(bindings)))
}
func validKind(kind string) bool {
	return kind == KindExactValue || kind == KindExactDigest || kind == KindSetEqual || kind == KindDerivedDigest
}
func validStableID(value string) bool {
	if !strings.HasPrefix(value, "sid:") || len(value) < 7 {
		return false
	}
	body := value[4:]
	if body[0] == '-' || body[len(body)-1] == '-' || strings.Contains(body, "--") {
		return false
	}
	for _, char := range body {
		if !(char >= 'a' && char <= 'z') && !(char >= '0' && char <= '9') && char != '-' {
			return false
		}
	}
	return true
}
func validToken(value, prefix string) bool {
	return strings.HasPrefix(value, prefix) && validStableID("sid:"+value[len(prefix):])
}
