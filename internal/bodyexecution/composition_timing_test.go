package bodyexecution

import (
	"context"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodytiming"
)

func TestCompositionTimingDoesNotControlExecution(t *testing.T) {
	source, suite, prior := retainedFieldFixture(t)
	for _, mode := range []string{"disabled", "enabled", "invalid"} {
		t.Run(mode, func(t *testing.T) {
			ctx := context.Background()
			var recorder *bodytiming.Recorder
			if mode != "disabled" {
				recorder = bodytiming.NewRecorder()
				ctx = bodytiming.WithRecorder(ctx, recorder)
			}
			if mode == "invalid" {
				bodytiming.Start(ctx, "unknown diagnostic phase").End(true)
			}
			owner := NewExecutor()
			defer owner.Close()
			r, err := owner.ExecuteComposition(ctx, "fields.gooo", source, prior, suite, nativeTool())
			if err != nil || r.FiniteTotal != 14 || r.FinitePassed != 14 || !r.RuntimeReplayed ||
				r.SelectedSourceSHA256 != prior.SelectedSourceSHA256 || r.ModelCalls != 0 {
				t.Fatal("diagnostic changed source or execution", mode, err)
			}
			if recorder == nil {
				return
			}
			observed := recorder.Snapshot()
			want := "OBSERVED"
			if mode == "invalid" {
				want = "INCOMPLETE"
			}
			if observed.Status != want {
				t.Fatal(observed)
			}
			assertCompositionTimingIntervals(t, observed)
		})
	}
}

func assertCompositionTimingIntervals(t *testing.T, snapshot bodytiming.Snapshot) {
	t.Helper()
	end, total := int64(0), int64(0)
	seen := map[string]bool{}
	for _, phase := range snapshot.Phases {
		if phase.StartNS < end || phase.EndNS < phase.StartNS || phase.Outcome != "completed" {
			t.Fatal("wall intervals overlap or invent a completion", snapshot)
		}
		seen[phase.Name] = true
		end, total = phase.EndNS, total+phase.EndNS-phase.StartNS
	}
	if total+snapshot.UnassignedNS != snapshot.CaptureNS || end > snapshot.CaptureNS {
		t.Fatal("wall interval accounting differs", snapshot)
	}
	for _, name := range []string{"executor_gate", "composition_receipt", "source_replay", "native_projection",
		"native_run_1", "native_run_2", "runtime_decode_1", "runtime_decode_2", "input_separation"} {
		if !seen[name] {
			t.Fatal("missing current phase", name)
		}
	}
}
