package extractor

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

func assertCallbackCounterexampleClaim(t *testing.T, events []CallbackPackageTestEvent, state string) {
	t.Helper()
	run := CallbackPackageRun{Variant: "final", ExitCode: 1, Events: events}
	claim := callbackPackageFailureFrontier(run)
	if claim.State != state || claim.Stage != "CALLBACK_OBSERVATION" ||
		claim.Step != "RUN_FINAL_PACKAGE" || claim.BlockedBy == nil {
		t.Fatalf("counterexample frontier=%+v, want state=%s", claim, state)
	}
	if state == "REFUTED" {
		if claim.Reason != "PACKAGE_TEST_COUNTEREXAMPLE_OBSERVED" || claim.UnknownClass != "" ||
			claim.NextOperation != "PRESERVE_PACKAGE_TEST_COUNTEREXAMPLE" {
			t.Fatalf("known counterexample lost its causal identity: %+v", claim)
		}
	} else if claim.Reason != "PACKAGE_TEST_OBSERVATION_INCOMPLETE" || claim.UnknownClass != "DIRECT_MISSING" ||
		claim.NextOperation != "RESOLVE_PACKAGE_TEST_OBSERVATION" {
		t.Fatalf("missing evidence was promoted to a counterexample: %+v", claim)
	}
	assertCallbackCounterexampleRecord(t, run, claim)
}

func assertCallbackCounterexampleRecord(t *testing.T, run CallbackPackageRun, claim CallbackExtractionClaim) {
	t.Helper()
	contract, err := generation.LoadCallbackExtractionContract()
	if err != nil {
		t.Fatal(err)
	}
	observation := CallbackExtractionObservation{
		Scope: "PACKAGE_TEST_EVENTS_ONLY", Decision: claim.State, Frontier: claim,
		Runs: []CallbackPackageRun{run}, OperationAdmission: "UNKNOWN", ApplyPermission: "FORBIDDEN",
	}
	if err := bindCallbackPackageObservation(&observation, contract); err != nil {
		t.Fatal(err)
	}
	if observation.Record.Fields["ObservationDecision"] != claim.State ||
		observation.Record.Fields["Scope"] != observation.Scope ||
		observation.AttemptedRuns != 1 || observation.CompletedTestRuns != 0 || observation.RequiredTestRuns != 2 ||
		observation.OperationAdmission != "UNKNOWN" || observation.ApplyPermission != "FORBIDDEN" {
		t.Fatalf("Gooo-bound incomplete observation lost its decision or authority boundary: %+v", observation)
	}
}
