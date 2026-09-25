package query

import (
	"sort"
	"strings"
)

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
