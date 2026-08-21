package semanticdelta

import (
	"errors"
	"fmt"
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
