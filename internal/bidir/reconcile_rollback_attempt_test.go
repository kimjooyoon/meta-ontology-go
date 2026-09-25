package bidir

import (
	"errors"
	"reflect"
	"testing"
)

func TestReconcileRollbackSeparatesNonAuthoritativeAttempts(t *testing.T) {
	for _, name := range []string{"addition", "removal"} {
		t.Run(name, func(t *testing.T) {
			base, changes, valid := reconcileRollbackFixture(t, name == "removal")
			beforeModel, beforeChanges := base.Normalized(), cloneFactDelta(changes)
			result, err := Reconcile(base, changes)
			var conflict *ReconcileError
			if !errors.As(err, &conflict) || len(conflict.Conflicts) != 1 || conflict.Conflicts[0].Kind != ConflictKindMismatch {
				t.Fatalf("expected one semantic conflict, got %v", err)
			}
			if !reflect.DeepEqual(result.Model, beforeModel) || !result.Delta.IsEmpty() {
				t.Fatalf("rejected transaction changed the returned model: %#v", result)
			}
			if len(result.Accepted) != 0 || len(result.ValidatedBeforeRollback) != 1 || !result.ValidatedBeforeRollback.Contains(valid) {
				t.Fatalf("rolled-back facts escaped into acceptance or lost their attempt: %#v", result)
			}
			if len(result.RawObservation.Added) != len(changes.Added) || len(result.RawObservation.Removed) != len(changes.Removed) || result.RawObservation.EvidenceHash == "" {
				t.Fatalf("rollback lost raw input evidence: %#v", result.RawObservation)
			}
			result.ValidatedBeforeRollback[0].Attributes["mode"] = "mutated-attempt"
			if !reflect.DeepEqual(changes, beforeChanges) || !reflect.DeepEqual(base.Normalized(), beforeModel) {
				t.Fatal("rollback attempt aliases caller inputs")
			}
			raw := append(append(FactSet(nil), result.RawObservation.Added...), result.RawObservation.Removed...)
			for _, fact := range raw {
				if fact.Attributes["mode"] != "observed" {
					t.Fatalf("rollback attempt aliases raw evidence: %#v", fact)
				}
			}
		})
	}
}

func TestReconcileInputRejectionDoesNotManufactureValidatedAttempts(t *testing.T) {
	base, changes, _ := reconcileRollbackFixture(t, false)
	changes.Added[1].Source.End = changes.Added[1].Source.Start - 1
	result, err := Reconcile(base, changes)
	var conflict *ReconcileError
	if !errors.As(err, &conflict) || len(conflict.Conflicts) != 1 || conflict.Conflicts[0].Kind != ConflictInvalidFact {
		t.Fatalf("expected input validation rejection, got %v", err)
	}
	if len(result.Accepted) != 0 || len(result.ValidatedBeforeRollback) != 0 ||
		!reflect.DeepEqual(result.Model, base.Normalized()) || !result.Delta.IsEmpty() {
		t.Fatalf("input rejection manufactured transaction progress: %#v", result)
	}
	if len(result.RawObservation.Added) != 2 || result.RawObservation.EvidenceHash == "" {
		t.Fatalf("input rejection erased observed inputs: %#v", result.RawObservation)
	}
}

func TestReconcileSuccessKeepsAcceptanceWithoutRollbackAttempts(t *testing.T) {
	base, _, valid := reconcileRollbackFixture(t, false)
	result, err := Reconcile(base, FactDelta{Added: FactSet{valid}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Accepted) != 1 || !result.Accepted.Contains(valid) || len(result.ValidatedBeforeRollback) != 0 ||
		len(result.Model.Relations) != len(base.Relations)+1 || len(result.Delta.AddedRelations) != 1 {
		t.Fatalf("successful reconciliation lost its accepted change: %#v", result)
	}
}

func reconcileRollbackFixture(t *testing.T, removal bool) (Model, FactDelta, Fact) {
	t.Helper()
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	valid, conflict := rawEvidenceFact("valid-attempt", 10), rawEvidenceFact("conflicting-attempt", 20)
	conflict.Subject = "billing://entity/order"
	changes := FactDelta{Added: FactSet{valid, conflict}}
	if removal {
		seeded, err := Reconcile(base, FactDelta{Added: FactSet{valid}})
		if err != nil {
			t.Fatal(err)
		}
		base = seeded.Model
		changes = FactDelta{Added: FactSet{conflict}, Removed: FactSet{valid}}
	}
	return base, changes, valid
}
