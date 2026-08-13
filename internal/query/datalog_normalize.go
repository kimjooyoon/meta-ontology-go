package query

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func normalizeDatalogQuery(request DatalogQuery) (DatalogQuery, []DatalogRule, error) {
	if len(request.Patterns) == 0 || len(request.Patterns) > MaxDatalogBodyAtoms {
		return DatalogQuery{}, nil, datalogError("query requires 1..%d patterns", MaxDatalogBodyAtoms)
	}
	if len(request.Rules) > MaxDatalogRules {
		return DatalogQuery{}, nil, datalogError("rule count exceeds %d", MaxDatalogRules)
	}
	if request.Limit < 0 || request.Limit > MaxDatalogLimit {
		return DatalogQuery{}, nil, datalogError("limit must be 0..%d", MaxDatalogLimit)
	}
	if request.Limit == 0 {
		request.Limit = DefaultDatalogLimit
	}
	if request.MaxDerivedFacts < 0 || request.MaxDerivedFacts > DefaultDatalogDerivedLimit {
		return DatalogQuery{}, nil, datalogError("max derived facts must be 0..%d", DefaultDatalogDerivedLimit)
	}
	if request.MaxDerivedFacts == 0 {
		request.MaxDerivedFacts = DefaultDatalogDerivedLimit
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
	known := map[string]struct{}{
		string(Used): {}, string(WasGeneratedBy): {}, string(WasDerivedFrom): {}, string(WasAssociatedWith): {},
	}
	for predicate := range knownHeads {
		known[predicate] = struct{}{}
	}
	for _, rule := range rules {
		for _, atom := range rule.Body {
			if _, exists := known[atom.Predicate]; !exists {
				return DatalogQuery{}, nil, datalogError("rule %q references unknown predicate %q", rule.ID, atom.Predicate)
			}
		}
	}
	for _, pattern := range patterns {
		if _, exists := known[pattern.Predicate]; !exists {
			return DatalogQuery{}, nil, datalogError("query references unknown predicate %q", pattern.Predicate)
		}
	}
	sort.Slice(patterns, func(i, j int) bool {
		return datalogAtomCanonical(patterns[i]) < datalogAtomCanonical(patterns[j])
	})
	sort.Slice(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })
	request.Patterns, request.Rules = patterns, rules
	return request, rules, nil
}

func normalizeDatalogRule(rule DatalogRule) (DatalogRule, error) {
	rule.ID = strings.TrimSpace(rule.ID)
	if !validDatalogRuleID(rule.ID) {
		return DatalogRule{}, datalogError("rule ID %q is invalid", rule.ID)
	}
	if len(rule.Body) == 0 || len(rule.Body) > MaxDatalogBodyAtoms {
		return DatalogRule{}, datalogError("rule %q requires 1..%d body atoms", rule.ID, MaxDatalogBodyAtoms)
	}
	head, err := normalizeDatalogAtom(rule.Head)
	if err != nil {
		return DatalogRule{}, datalogError("rule %q head: %v", rule.ID, err)
	}
	body := make([]DatalogAtom, len(rule.Body))
	bound := make(map[string]struct{})
	for index, atom := range rule.Body {
		body[index], err = normalizeDatalogAtom(atom)
		if err != nil {
			return DatalogRule{}, datalogError("rule %q body atom %d: %v", rule.ID, index, err)
		}
		for _, term := range []DatalogTerm{body[index].Subject, body[index].Object} {
			if term.Variable != "" {
				bound[strings.TrimPrefix(term.Variable, "?")] = struct{}{}
			}
		}
	}
	for _, term := range []DatalogTerm{head.Subject, head.Object} {
		if term.Variable != "" {
			if _, exists := bound[strings.TrimPrefix(term.Variable, "?")]; !exists {
				return DatalogRule{}, datalogError("rule %q has unsafe head variable %q", rule.ID, term.Variable)
			}
		}
	}
	sort.Slice(body, func(i, j int) bool {
		return datalogAtomCanonical(body[i]) < datalogAtomCanonical(body[j])
	})
	rule.Head, rule.Body = head, body
	return rule, nil
}

func normalizeDatalogAtom(atom DatalogAtom) (DatalogAtom, error) {
	predicate, err := normalizeDatalogPredicate(atom.Predicate)
	if err != nil {
		return DatalogAtom{}, err
	}
	subject, err := normalizeDatalogTerm(atom.Subject)
	if err != nil {
		return DatalogAtom{}, datalogError("subject: %v", err)
	}
	object, err := normalizeDatalogTerm(atom.Object)
	if err != nil {
		return DatalogAtom{}, datalogError("object: %v", err)
	}
	return DatalogAtom{Predicate: predicate, Subject: subject, Object: object}, nil
}

func normalizeDatalogTerm(term DatalogTerm) (DatalogTerm, error) {
	if term.Variable != "" && term.Constant != "" {
		return DatalogTerm{}, datalogError("term cannot be both variable and constant")
	}
	if term.Variable != "" {
		name := strings.TrimPrefix(strings.TrimSpace(term.Variable), "?")
		if !validDatalogIdentifier(name) {
			return DatalogTerm{}, datalogError("variable %q is invalid", term.Variable)
		}
		return DatalogTerm{Variable: "?" + name}, nil
	}
	if term.Constant == "" {
		return DatalogTerm{}, datalogError("term is empty")
	}
	id, err := ParseID(term.Constant.String())
	if err != nil {
		return DatalogTerm{}, datalogError("constant: %v", err)
	}
	return DatalogTerm{Constant: id}, nil
}

func normalizeDatalogPredicate(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if relation, err := ParseRelation(Relation(raw)); err == nil {
		return string(relation), nil
	}
	if !validDatalogIdentifier(raw) {
		return "", datalogError("predicate %q is invalid", raw)
	}
	return raw, nil
}

func validDatalogIdentifier(value string) bool {
	if value == "" || !((value[0] >= 'a' && value[0] <= 'z') || (value[0] >= 'A' && value[0] <= 'Z') || value[0] == '_') {
		return false
	}
	for _, character := range value[1:] {
		if !((character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '_') {
			return false
		}
	}
	return true
}

func validDatalogRuleID(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if !(character == '/' || character == ':' || character == '.' || character == '-' ||
			(character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '_') {
			return false
		}
	}
	return true
}

func datalogError(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidDatalogQuery, fmt.Sprintf(format, args...))
}

func datalogAtomCanonical(atom DatalogAtom) string {
	return strconv.Quote(atom.Predicate) + "\x00" +
		datalogTermCanonical(atom.Subject) + "\x00" + datalogTermCanonical(atom.Object)
}

func datalogTermCanonical(term DatalogTerm) string {
	if term.Variable != "" {
		return "variable\x00" + term.Variable
	}
	return "constant\x00" + term.Constant.String()
}
