package lanefrontier

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
	"strings"
	"unicode"
)

func seal(result Output, decision Decision, reason Reason) Output {
	result = normalizeOutput(result)
	result.Decision = decision
	result.Reason = reason
	result.CanonicalDigest = result.StableDigest()
	return result
}
func hasRequiredFacts(input Input) bool {
	return validDigest(input.RegistryDigest) && validSHA(input.BaseSHA) &&
		validSHA(input.LaneHeadSHA) && validStableID(input.LaneID) &&
		validToken(input.RegisteredBranch) && input.OwnedPathPrefixes != nil &&
		input.ChangedPaths != nil
}
func hasInvalidCount(input Input) bool {
	return input.AheadCount < 0 || input.BehindCount < 0 ||
		input.OpenPRCount < 0 || input.ActiveLeaseCount < 0
}
func validDigest(value string) bool { return validHex(value, 64) }
func validSHA(value string) bool {
	return validHex(value, 40) || validHex(value, 64)
}
func validHex(value string, size int) bool {
	if len(value) != size {
		return false
	}
	for _, char := range value {
		if !(char >= '0' && char <= '9') && !(char >= 'a' && char <= 'f') {
			return false
		}
	}
	return true
}
func validToken(value string) bool {
	if value == "" || strings.TrimSpace(value) != value {
		return false
	}
	for _, char := range value {
		if unicode.IsSpace(char) || unicode.IsControl(char) {
			return false
		}
	}
	return true
}
func validStableID(value string) bool {
	parsed, err := semantic.ParseIdentity(value)
	return err == nil && parsed.String() == value
}
func normalizeOwners(values []string) ([]string, error) {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		owner, err := normalizePath(value, true)
		if err != nil {
			return nil, err
		}
		normalized = append(normalized, owner)
	}
	sort.Strings(normalized)
	for index := 1; index < len(normalized); index++ {
		if pathContains(normalized[index-1], normalized[index]) {
			return nil, errAmbiguousOwner{}
		}
	}
	return normalized, nil
}
