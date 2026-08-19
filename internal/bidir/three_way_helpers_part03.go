package bidir

import (
	"sort"
)

func unionRelationKeys(groups ...map[string]Relation) []string {
	keys := make(map[string]struct{})
	for _, group := range groups {
		for key := range group {
			keys[key] = struct{}{}
		}
	}
	result := make([]string, 0, len(keys))
	for key := range keys {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}
func conflictFor(code ThreeWayConflictCode, scope, identity string, fingerprints [3]string, message string) *ThreeWayConflict {
	return &ThreeWayConflict{Code: code, Scope: scope, Identity: identity, Message: message, BaseFingerprint: fingerprints[0], LeftFingerprint: fingerprints[1], RightFingerprint: fingerprints[2]}
}
func sortThreeWayConflicts(conflicts []ThreeWayConflict) {
	sort.SliceStable(conflicts, func(i, j int) bool {
		left, right := conflicts[i], conflicts[j]
		if left.Code != right.Code {
			return left.Code < right.Code
		}
		if left.Scope != right.Scope {
			return left.Scope < right.Scope
		}
		if left.Identity != right.Identity {
			return left.Identity < right.Identity
		}
		return left.Message < right.Message
	})
}
