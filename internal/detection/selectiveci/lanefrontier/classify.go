package lanefrontier

import (
	"path"
	"sort"
	"strings"
	"unicode"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// Classify evaluates facts in the documented fixed order and returns a sealed
// canonical result. It has no side effects and does not select a priority.
func Classify(input Input) Output {
	result := outputFromInput(input)
	if input.SchemaVersion != SchemaVersion {
		return seal(result, DecisionUnknown, ReasonUnknownSchema)
	}
	if !hasRequiredFacts(input) {
		return seal(result, DecisionUnknown, ReasonMissingInput)
	}
	if hasInvalidCount(input) {
		return seal(result, DecisionUnknown, ReasonInvalidCount)
	}

	owners, ownerErr := normalizeOwners(input.OwnedPathPrefixes)
	if ownerErr != nil {
		if _, invalid := ownerErr.(errInvalidPath); invalid {
			return seal(result, DecisionUnknown, ReasonMissingInput)
		}
		return seal(result, DecisionUnknown, ReasonAmbiguousOwner)
	}
	paths, pathErr := normalizePaths(input.ChangedPaths)
	if pathErr != nil {
		return seal(result, DecisionUnknown, ReasonMissingInput)
	}
	if !pathsInScope(paths, owners) {
		return seal(result, DecisionIneligible, ReasonPathOutOfScope)
	}
	result.OwnedPathPrefixes = owners
	result.ChangedPaths = paths

	if input.ActiveLeaseCount > 0 {
		return seal(result, DecisionIneligible, ReasonActiveLease)
	}
	if input.OpenPRCount > 0 {
		return seal(result, DecisionIneligible, ReasonActivePR)
	}
	if input.AheadCount > 0 && input.BehindCount > 0 {
		return seal(result, DecisionIneligible, ReasonDivergedBranch)
	}
	if input.AheadCount == 0 && input.BehindCount > 0 {
		return seal(result, DecisionIneligible, ReasonStaleBranch)
	}
	if input.AheadCount > 0 && input.BehindCount == 0 {
		return seal(result, DecisionIneligible, ReasonBranchAhead)
	}
	return seal(result, DecisionEligible, ReasonEligible)
}

// Evaluate is a descriptive alias for callers that treat the classifier as a
// predicate evaluator.
func Evaluate(input Input) Output { return Classify(input) }

func outputFromInput(input Input) Output {
	return Output{
		SchemaVersion:     input.SchemaVersion,
		RegistryDigest:    input.RegistryDigest,
		BaseSHA:           input.BaseSHA,
		LaneHeadSHA:       input.LaneHeadSHA,
		LaneID:            input.LaneID,
		RegisteredBranch:  input.RegisteredBranch,
		OwnedPathPrefixes: append([]string(nil), input.OwnedPathPrefixes...),
		ChangedPaths:      append([]string(nil), input.ChangedPaths...),
		AheadCount:        input.AheadCount,
		BehindCount:       input.BehindCount,
		OpenPRCount:       input.OpenPRCount,
		ActiveLeaseCount:  input.ActiveLeaseCount,
	}
}

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

func normalizePaths(values []string) ([]string, error) {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		clean, err := normalizePath(value, false)
		if err != nil {
			return nil, err
		}
		normalized = append(normalized, clean)
	}
	sort.Strings(normalized)
	return unique(normalized), nil
}

func normalizePath(value string, owner bool) (string, error) {
	if value == "" || strings.Contains(value, "\\") || strings.ContainsRune(value, '\x00') ||
		strings.HasPrefix(value, "/") || strings.TrimSpace(value) != value {
		return "", errInvalidPath{}
	}
	if owner && value != "." {
		value = strings.TrimSuffix(value, "/")
	}
	if value == "" || path.Clean(value) != value || value == ".." || strings.HasPrefix(value, "../") || (!owner && value == ".") {
		return "", errInvalidPath{}
	}
	return value, nil
}

func pathsInScope(paths, owners []string) bool {
	for _, changed := range paths {
		inScope := false
		for _, owner := range owners {
			if pathContains(owner, changed) {
				inScope = true
				break
			}
		}
		if !inScope {
			return false
		}
	}
	return true
}

func pathContains(owner, value string) bool {
	return owner == "." || owner == value || strings.HasPrefix(value, owner+"/")
}

func unique(values []string) []string {
	if len(values) < 2 {
		return values
	}
	result := values[:1]
	for _, value := range values[1:] {
		if value != result[len(result)-1] {
			result = append(result, value)
		}
	}
	return result
}

type errInvalidPath struct{}

func (errInvalidPath) Error() string { return "invalid repository-relative path" }

type errAmbiguousOwner struct{}

func (errAmbiguousOwner) Error() string { return "ambiguous owned path prefixes" }
