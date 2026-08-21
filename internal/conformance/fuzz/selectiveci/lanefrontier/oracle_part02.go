package lanefrontier

import (
	"sort"
	"strings"
)

func decide(input Input) (Decision, Reason) {
	if input.Schema != SchemaV1 {
		return Unknown, UnknownSchema
	}
	if missingInput(input) {
		return Unknown, MissingInput
	}
	if invalidCount(input) {
		return Unknown, InvalidCount
	}
	prefixes, valid := canonicalPrefixes(input.OwnedPathPrefixes)
	if !valid || ambiguousPrefixes(prefixes) {
		if !valid {
			return Unknown, MissingInput
		}
		return Unknown, AmbiguousOwner
	}
	if !validChangedPaths(input.ChangedPaths) {
		return Unknown, MissingInput
	}
	if !pathsInScope(input.ChangedPaths, prefixes) {
		return Ineligible, PathOutOfScope
	}
	if input.ActiveLeaseCount > 0 {
		return Ineligible, ActiveLease
	}
	if input.OpenPRCount > 0 {
		return Ineligible, ActivePR
	}
	if input.AheadCount > 0 && input.BehindCount > 0 {
		return Ineligible, DivergedBranch
	}
	if input.AheadCount == 0 && input.BehindCount > 0 {
		return Ineligible, StaleBranch
	}
	if input.AheadCount > 0 && input.BehindCount == 0 {
		return Ineligible, BranchAhead
	}
	return Eligible, Clean
}
func missingInput(input Input) bool {
	values := []string{input.RegistryDigest, input.BaseSHA, input.LaneHeadSHA, input.LaneStableID, input.RegisteredBranch}
	for _, value := range values {
		if strings.TrimSpace(value) == "" || value != strings.TrimSpace(value) {
			return true
		}
	}
	return len(input.OwnedPathPrefixes) == 0 || input.ChangedPaths == nil
}
func invalidCount(input Input) bool {
	return input.AheadCount < 0 || input.BehindCount < 0 || input.OpenPRCount < 0 || input.ActiveLeaseCount < 0
}
func canonicalPrefixes(prefixes []string) ([]string, bool) {
	result := make([]string, len(prefixes))
	for i, prefix := range prefixes {
		canonical, ok := canonicalPath(prefix, true)
		if !ok {
			return nil, false
		}
		result[i] = canonical
	}
	sort.Strings(result)
	return result, true
}
