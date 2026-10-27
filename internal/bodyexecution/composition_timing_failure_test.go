package bodyexecution

import (
	"context"
	"errors"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodytiming"
)

func TestCompositionTimingPreservesCancellationAndSourceFailure(t *testing.T) {
	source, suite, prior := retainedFieldFixture(t)
	for _, kind := range []string{"canceled", "changed-source"} {
		t.Run(kind, func(t *testing.T) {
			recorder := bodytiming.NewRecorder()
			ctx, cancel := context.WithCancel(bodytiming.WithRecorder(context.Background(), recorder))
			defer cancel()
			candidate := prior
			failedPhase := "source_replay"
			if kind == "canceled" {
				cancel()
				failedPhase = "executor_gate"
			} else {
				candidate.GeneratedSHA256 = "changed"
			}
			owner := NewExecutor()
			defer owner.Close()
			r, err := owner.ExecuteComposition(ctx, "fields.gooo", source, candidate, suite, nativeTool())
			if err == nil || r.Build.Started || len(r.Runs) != 0 ||
				(kind == "canceled" && !errors.Is(err, context.Canceled)) {
				t.Fatal("diagnostic changed refusal", err)
			}
			s := recorder.Snapshot()
			if s.Status != "OBSERVED" || len(s.Phases) == 0 ||
				s.Phases[len(s.Phases)-1].Name != failedPhase || s.Phases[len(s.Phases)-1].Outcome != "failed" {
				t.Fatal("failure interval was lost or left active", s)
			}
		})
	}
}
