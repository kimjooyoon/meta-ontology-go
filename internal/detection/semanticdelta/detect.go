package semanticdelta

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// ErrScopeViolation identifies a report that contains one or more violations.
var ErrScopeViolation = errors.New("semantic delta is outside allowed scope")

// ScopeError retains the deterministic report behind ErrScopeViolation.
type ScopeError struct {
	Report Report
}

func (e *ScopeError) Error() string {
	if e == nil {
		return ErrScopeViolation.Error()
	}
	return fmt.Sprintf("%s: %d violation(s)", ErrScopeViolation, len(e.Report.Violations))
}

func (e *ScopeError) Unwrap() error { return ErrScopeViolation }

// Detect checks every changed endpoint and predicate against the allowed
// semantic scope. Presentation-only fields never reach this check.
func Detect(delta Delta, scope Scope) (Report, error) {
	normalizedDelta, err := delta.Normalized()
	if err != nil {
		return Report{}, err
	}
	normalizedScope, err := scope.Normalized()
	if err != nil {
		return Report{}, err
	}
	report := Report{Allowed: true}
	for _, node := range normalizedDelta.AddedNodes {
		report.checkNode(OperationAdd, node, normalizedScope)
	}
	for _, node := range normalizedDelta.RemovedNodes {
		report.checkNode(OperationRemove, node, normalizedScope)
	}
	for _, fact := range normalizedDelta.AddedFacts {
		report.checkFact(OperationAdd, fact, normalizedScope)
	}
	for _, fact := range normalizedDelta.RemovedFacts {
		report.checkFact(OperationRemove, fact, normalizedScope)
	}
	report.Normalize()
	return report, nil
}

// Check is a concise alias for Detect.
func Check(delta Delta, scope Scope) (Report, error) { return Detect(delta, scope) }

func (r *Report) checkNode(operation Operation, node Node, scope Scope) {
	if scope.AllowsID(node.ID) {
		return
	}
	r.Allowed = false
	r.Violations = append(r.Violations, Violation{
		Operation: operation, Change: ChangeNode, ID: node.ID, Kind: node.Kind,
		Endpoint: "node", Reason: "node identity is outside allowed scope",
	})
}

func (r *Report) checkFact(operation Operation, fact Fact, scope Scope) {
	if !scope.AllowsID(fact.Subject) {
		r.addFactViolation(operation, fact, "subject", "fact subject is outside allowed scope")
	}
	if !scope.AllowsID(fact.Object) {
		r.addFactViolation(operation, fact, "object", "fact object is outside allowed scope")
	}
	if !scope.AllowsPredicate(fact.Predicate) {
		r.addFactViolation(operation, fact, "predicate", "fact predicate is outside allowed scope")
	}
}

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

// Normalize sorts a report and returns a detached copy.
func (r *Report) Normalize() {
	if r == nil {
		return
	}
	sort.SliceStable(r.Violations, func(i, j int) bool {
		left, right := r.Violations[i], r.Violations[j]
		if left.Operation != right.Operation {
			return left.Operation < right.Operation
		}
		if left.Change != right.Change {
			return changeRank(left.Change) < changeRank(right.Change)
		}
		if left.ID != right.ID {
			return left.ID < right.ID
		}
		if left.Subject != right.Subject {
			return left.Subject < right.Subject
		}
		if left.Predicate != right.Predicate {
			return left.Predicate < right.Predicate
		}
		if left.Object != right.Object {
			return left.Object < right.Object
		}
		if left.Endpoint != right.Endpoint {
			return left.Endpoint < right.Endpoint
		}
		return left.Reason < right.Reason
	})
}

func changeRank(kind ChangeKind) int {
	if kind == ChangeNode {
		return 0
	}
	return 1
}

// Allowed reports whether the detector found no scope violation.
func (r Report) Passes() bool { return r.Allowed && len(r.Violations) == 0 }
