package bidir

import (
	"fmt"
)

func (e *ReconcileError) Error() string {
	if len(e.Conflicts) == 0 {
		return "bidir reconciliation failed"
	}
	return fmt.Sprintf("bidir reconciliation rejected %d fact(s): %s", len(e.Conflicts), e.Conflicts[0].Message)
}

// ReconcileResult contains a transactional model and non-authoritative observations.
// Accepted records handled facts only for a successful reconciliation; it does
// not authorize external writes or policy adoption.
type ReconcileResult struct {
	Model          Model
	Delta          Delta
	Locality       Locality
	RawObservation RawFactObservation
	Accepted       FactSet
	Syntactic      FactSet
	Candidates     FactSet
	Conflicts      []Conflict

	// ValidatedBeforeRollback preserves detached local processing from a rejected
	// transaction. These facts are diagnostic attempts, not applied changes.
	ValidatedBeforeRollback FactSet
}
