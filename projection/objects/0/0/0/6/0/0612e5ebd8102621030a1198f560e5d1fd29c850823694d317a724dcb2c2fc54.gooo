package query

import (
	"sort"
)

func normalizeDatalogQuery(request DatalogQuery) (DatalogQuery, []DatalogRule, error) {
	var err error
	request, err = normalizeDatalogBounds(request)
	if err != nil {
		return DatalogQuery{}, nil, err
	}

	patterns := make([]DatalogAtom, len(request.Patterns))
	for index, pattern := range request.Patterns {
		var err error
		patterns[index], err = normalizeDatalogAtom(pattern)
		if err != nil {
			return DatalogQuery{}, nil, err
		}
	}
	rules := make([]DatalogRule, len(request.Rules))
	knownHeads := make(map[string]struct{}, len(request.Rules))
	seenRules := make(map[string]struct{}, len(request.Rules))
	for index, rule := range request.Rules {
		normalized, err := normalizeDatalogRule(rule)
		if err != nil {
			return DatalogQuery{}, nil, err
		}
		if _, exists := seenRules[normalized.ID]; exists {
			return DatalogQuery{}, nil, datalogError("duplicate rule ID %q", normalized.ID)
		}
		seenRules[normalized.ID] = struct{}{}
		knownHeads[normalized.Head.Predicate] = struct{}{}
		rules[index] = normalized
	}
	if err := validateDatalogPredicates(patterns, rules, knownHeads); err != nil {
		return DatalogQuery{}, nil, err
	}
	sort.Slice(patterns, func(i, j int) bool {
		return datalogAtomCanonical(patterns[i]) < datalogAtomCanonical(patterns[j])
	})
	sort.Slice(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })
	request.Patterns, request.Rules = patterns, rules
	return request, rules, nil
}
