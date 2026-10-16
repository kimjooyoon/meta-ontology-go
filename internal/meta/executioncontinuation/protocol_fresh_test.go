package executioncontinuation

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrepareOriginsFresh1617(t *testing.T) {
	for _, same := range []bool{false, true} {
		f := fixtureConsumer1617(t)
		if same {
			f.grantRun.Head = f.prepared.Metadata.Source.SubjectSHA
			f.grantArtifact.Run.Head = f.grantRun.Head
			f.prepared.Metadata.Source.WorkflowSHA = f.grantRun.Head
		}
		client := client1617(t, f, "")
		e := f.prepared.Metadata.Source
		p, err := client.Prepare(context.Background(), e, f.consumerRun.ID, 1)
		if err != nil {
			t.Fatal("source preparation", same, err)
		}
		if p.Metadata.Source.SubjectSHA != e.SubjectSHA || p.Metadata.ConsumerWorkflowSHA != f.consumerRun.Head || p.Metadata.ContractWorkflowSHA != f.contractRun.Head {
			t.Fatal("collapsed origins")
		}
		if _, err := Reserve(client.Root, p); err != nil {
			t.Fatal("source reservation", err)
		}
	}
}

func TestPrepareNegativeFresh1617(t *testing.T) {
	f := fixtureConsumer1617(t)
	for _, c := range []struct{ name, state string }{{"missing-grant", "UNKNOWN"}, {"duplicate-grant", "REFUTED"}, {"wrong-attempt", "REFUTED"}, {"expired-grant", "UNKNOWN"}, {"changed-archive", "REFUTED"}, {"api-unavailable", "UNKNOWN"}} {
		t.Run(c.name, func(t *testing.T) {
			client := client1617(t, f, c.name)
			_, err := client.Prepare(context.Background(), f.prepared.Metadata.Source, f.consumerRun.ID, 1)
			state, _ := errorState(err)
			if err == nil || state != c.state {
				t.Fatal(c, err)
			}
		})
	}
}

func TestDispatchStagesFresh1617(t *testing.T) {
	f := fixtureConsumer1617(t)
	for _, c := range []struct{ name, state string }{{"", "TARGET_RUN_OBSERVED"}, {"dispatch-no-reference", "DISPATCH_SENT"}, {"wrong-target", "CONTINUATION_REFUTED"}, {"api-unavailable", "DISPATCH_UNKNOWN"}} {
		t.Run(c.name, func(t *testing.T) {
			client := client1617(t, f, c.name)
			r, _ := client.Dispatch(context.Background(), f.prepared.Metadata.Source)
			if r.State != c.state {
				t.Fatal(c, r)
			}
			if r.State == "DISPATCH_SENT" && r.Unknown == nil {
				t.Fatal("target not observed but unknown lost")
			}
		})
	}
	e := f.prepared.Metadata.Source
	e.Wait = "COMPLETED"
	r, err := client1617(t, f, "").Dispatch(context.Background(), e)
	if err == nil || r.State != "CONTINUATION_REFUTED" {
		t.Fatal("cyclic wait permitted")
	}
}

func TestSubjectContradictionFresh1617(t *testing.T) {
	f := fixtureConsumer1617(t)
	e := f.prepared.Metadata.Source
	e.SubjectSHA = strings.Repeat("e", 40)
	f.grantArtifact.Name = "self-improvement-execution-grant-" + e.SubjectSHA
	_, err := client1617(t, f, "").Prepare(context.Background(), e, f.consumerRun.ID, 1)
	state, reason := errorState(err)
	if state != "REFUTED" || reason != "SUBJECT_REFUTED" {
		t.Fatal("declared subject replaced payload", err)
	}
}

func TestLocalReservationClaimFresh1617(t *testing.T) {
	f := fixtureConsumer1617(t)
	root := os.DirFS("../../..")
	reservation, err := Reserve(root, f.prepared)
	if err != nil {
		t.Fatal(err)
	}
	claim := filepath.Join(t.TempDir(), "claimed")
	execution, err := Acquire(root, f.prepared, reservation, claim)
	if err != nil || execution == nil {
		t.Fatal(err)
	}
	if _, err := Acquire(root, f.prepared, reservation, claim); err == nil {
		t.Fatal("duplicate local claim accepted")
	}
	reservation.ConsumerRunID++
	if _, err := Acquire(root, f.prepared, reservation, filepath.Join(t.TempDir(), "invalid")); err == nil {
		t.Fatal("foreign reservation accepted")
	}
}

// Only original CI runs this new positive consumption path. Local protocol
// checks do not manually replay the existing five-value native experiment.
func TestGoConsumptionOriginalCIFresh1617(t *testing.T) {
	f := fixtureConsumer1617(t)
	root := os.DirFS("../../..")
	r, err := Reserve(root, f.prepared)
	if err != nil {
		t.Fatal(err)
	}
	execution, err := Acquire(root, f.prepared, r, filepath.Join(t.TempDir(), "claim"))
	if err != nil {
		t.Fatal(err)
	}
	report, err := execution.Run()
	if err != nil || report.ExecutionCount != 1 || report.Invocations != 5 || report.PassedCases != 5 || report.Stage != "EXECUTION_OBSERVED" {
		t.Fatal("Go source execution", report, err)
	}
	if _, err := execution.Run(); err == nil {
		t.Fatal("execution handle reused")
	}
}
