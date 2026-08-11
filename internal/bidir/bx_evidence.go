package bidir

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// ReconciliationFixture supplies the source view and two Go-side updates for
// a measurable bidirectional contract experiment.
type ReconciliationFixture interface {
	Name() string
	Document() Document
	AcceptedDelta() FactDelta
	PartialDelta() FactDelta
}

// BXConflictEvidence records an expected partial-information rejection.
type BXConflictEvidence struct {
	Kind          ConflictKind
	Count         int
	Transactional bool
}

// BXEvidence is the stable, reviewable output of one reconciliation fixture.
type BXEvidence struct {
	Fixture              string
	GetPutPassed         bool
	PutGetPassed         bool
	SemanticEquivalent   bool
	AcceptedRelationAdds int
	Locality             Locality
	PartialConflict      BXConflictEvidence
}

// MeasureBXFixture runs Get-Put, Put-Get, locality, and partial conflict checks.
func MeasureBXFixture(fixture ReconciliationFixture) (BXEvidence, error) {
	if fixture == nil {
		return BXEvidence{}, errors.New("reconciliation fixture is nil")
	}
	evidence := BXEvidence{Fixture: strings.TrimSpace(fixture.Name())}
	if evidence.Fixture == "" {
		return BXEvidence{}, errors.New("reconciliation fixture name is empty")
	}
	document := fixture.Document()
	base, err := Get(document)
	if err != nil {
		return evidence, fmt.Errorf("get base document: %w", err)
	}
	evidence.GetPutPassed = CheckGetPut(document) == nil
	accepted, err := Reconcile(base, fixture.AcceptedDelta())
	if err != nil {
		return evidence, fmt.Errorf("reconcile accepted delta: %w", err)
	}
	evidence.AcceptedRelationAdds = len(accepted.Delta.AddedRelations)
	evidence.Locality = accepted.Locality
	updatedDocument, err := Put(document, accepted.Model)
	if err != nil {
		return evidence, fmt.Errorf("put accepted model: %w", err)
	}
	observed, err := Get(updatedDocument)
	if err != nil {
		return evidence, fmt.Errorf("get updated document: %w", err)
	}
	evidence.PutGetPassed = true
	evidence.SemanticEquivalent = SemanticEquivalent(accepted.Model, observed)
	evidence.PartialConflict = measurePartialConflict(base, fixture.PartialDelta())
	return evidence, nil
}

func measurePartialConflict(base Model, delta FactDelta) BXConflictEvidence {
	result, err := Reconcile(base, delta)
	evidence := BXConflictEvidence{Transactional: SemanticEquivalent(base, result.Model)}
	var reconcileErr *ReconcileError
	if !errors.As(err, &reconcileErr) {
		return evidence
	}
	evidence.Count = len(reconcileErr.Conflicts)
	if evidence.Count > 0 {
		evidence.Kind = reconcileErr.Conflicts[0].Kind
	}
	return evidence
}

// Canonical returns line-oriented golden evidence with deterministic IDs.
func (e BXEvidence) Canonical() string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "fixture=%s\n", e.Fixture)
	fmt.Fprintf(&builder, "get_put=%s\n", evidenceStatus(e.GetPutPassed))
	fmt.Fprintf(&builder, "put_get=%s\n", evidenceStatus(e.PutGetPassed))
	fmt.Fprintf(&builder, "semantic_equivalence=%s\n", evidenceStatus(e.SemanticEquivalent))
	fmt.Fprintf(&builder, "accepted_relation_adds=%d\n", e.AcceptedRelationAdds)
	fmt.Fprintf(&builder, "touched=%s\n", joinIDs(e.Locality.Touched))
	fmt.Fprintf(&builder, "affected=%s\n", joinIDs(e.Locality.Affected))
	fmt.Fprintf(&builder, "partial_conflict=%s\n", conflictStatus(e.PartialConflict.Kind))
	fmt.Fprintf(&builder, "partial_conflict_count=%d\n", e.PartialConflict.Count)
	fmt.Fprintf(&builder, "partial_transactional=%s\n", evidenceStatus(e.PartialConflict.Transactional))
	return builder.String()
}

func evidenceStatus(passed bool) string {
	if passed {
		return "pass"
	}
	return "fail"
}

func conflictStatus(kind ConflictKind) string {
	if kind == "" {
		return "none"
	}
	return string(kind)
}

func joinIDs(ids []ID) string {
	copyIDs := append([]ID(nil), ids...)
	sort.Slice(copyIDs, func(i, j int) bool { return copyIDs[i] < copyIDs[j] })
	values := make([]string, len(copyIDs))
	for index, id := range copyIDs {
		values[index] = string(id)
	}
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ",")
}
