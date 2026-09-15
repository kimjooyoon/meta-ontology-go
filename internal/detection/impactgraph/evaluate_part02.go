package impactgraph

import (
	"sort"
)

func inputIDs(ids []string, byID map[string]NodeKind, obligationsOnly bool) ([]string, string) {
	seen := make(map[string]struct{}, len(ids))
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if err := validateID(id); err != nil {
			if obligationsOnly {
				return nil, FailureCodeUnknownExecutedObligation
			}
			return nil, FailureCodeUnknownChangedNode
		}
		kind, registered := byID[id]
		if !registered {
			if obligationsOnly {
				return nil, FailureCodeUnknownExecutedObligation
			}
			return nil, FailureCodeUnknownChangedNode
		}
		if obligationsOnly && kind != NodeKindObligation {
			return nil, FailureCodeUnknownExecutedObligation
		}
		if _, duplicate := seen[id]; duplicate {
			if obligationsOnly {
				return nil, FailureCodeAmbiguousExecutedInput
			}
			return nil, FailureCodeAmbiguousChangedInput
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	sort.Strings(result)
	return result, FailureCodeNone
}
func executedIDs(ids []string, byID map[string]NodeKind) ([]string, string) {
	seen := make(map[string]struct{}, len(ids))
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if err := validateID(id); err != nil {
			return nil, FailureCodeUnknownExecutedObligation
		}
		if kind, registered := byID[id]; registered && kind != NodeKindObligation {
			return nil, FailureCodeUnknownExecutedObligation
		}
		if _, duplicate := seen[id]; duplicate {
			return nil, FailureCodeAmbiguousExecutedInput
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	sort.Strings(result)
	return result, FailureCodeNone
}
