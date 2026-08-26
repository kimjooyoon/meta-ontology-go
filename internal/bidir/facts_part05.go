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

// ReconcileResult contains accepted layers, semantic delta, locality, and a
// detached non-authoritative raw observation boundary.
type ReconcileResult struct {
	Model          Model
	Delta          Delta
	Locality       Locality
	RawObservation RawFactObservation
	Accepted       FactSet
	Syntactic      FactSet
	Candidates     FactSet
	Conflicts      []Conflict
}
