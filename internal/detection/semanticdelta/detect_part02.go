package semanticdelta

import (
	"strings"
)

func (r *Report) addFactViolation(operation Operation, fact Fact, endpoint, reason string) {
	r.Allowed = false
	r.Violations = append(r.Violations, Violation{
		Operation: operation, Change: ChangeFact, Subject: fact.Subject,
		Predicate: fact.Predicate, Object: fact.Object, Endpoint: endpoint, Reason: reason,
	})
}

// AllowsID reports whether an exact identity or configured prefix contains id.
func (s Scope) AllowsID(id string) bool {
	for _, candidate := range s.IDs {
		if candidate == id {
			return true
		}
	}
	for _, prefix := range s.Prefixes {
		if strings.HasPrefix(id, prefix) {
			return true
		}
	}
	return false
}

// AllowsPredicate reports whether the scope permits predicate. No predicate
// restriction is applied when Scope.Predicates is empty.
func (s Scope) AllowsPredicate(predicate string) bool {
	if len(s.Predicates) == 0 {
		return true
	}
	for _, candidate := range s.Predicates {
		if candidate == predicate {
			return true
		}
	}
	return false
}
