package bodytiming

import (
	"context"
	"testing"
	"time"
)

func TestRecorderOwnsOrderedCurrentWallIntervals(t *testing.T) {
	now := time.Unix(0, 0)
	r := newRecorder(func() time.Time { return now })
	ctx := WithRecorder(context.Background(), r)
	now = now.Add(2 * time.Nanosecond)
	a := Start(ctx, "request_decode")
	now = now.Add(5 * time.Nanosecond)
	a.End(true)
	a.End(false)
	now = now.Add(time.Nanosecond)
	b := Start(ctx, "generation")
	now = now.Add(9 * time.Nanosecond)
	b.End(false)
	now = now.Add(3 * time.Nanosecond)
	s := r.Snapshot()
	if s.Status != "OBSERVED" || s.CaptureNS != 20 || s.UnassignedNS != 6 || len(s.Phases) != 2 ||
		s.Phases[0].StartNS != 2 || s.Phases[0].EndNS != 7 || s.Phases[0].Outcome != "completed" ||
		s.Phases[1].Outcome != "failed" {
		t.Fatal(s)
	}
	s.Phases[0].Name = "changed"
	if r.Snapshot().Phases[0].Name != "request_decode" {
		t.Fatal("snapshot changed retained data")
	}
}

func TestRecorderFlagsIncompleteNestedAndOverCapacityObservations(t *testing.T) {
	for _, kind := range []string{"unfinished", "nested", "overflow", "unknown"} {
		t.Run(kind, func(t *testing.T) {
			r := NewRecorder()
			ctx := WithRecorder(context.Background(), r)
			switch kind {
			case "unfinished":
				Start(ctx, "generation")
			case "nested":
				a := Start(ctx, "generation")
				b := Start(ctx, "source_replay")
				b.End(true)
				a.End(true)
			case "overflow":
				for range 33 {
					Start(ctx, "generation").End(true)
				}
			case "unknown":
				Start(ctx, "invented_phase").End(true)
			}
			if r.Snapshot().Status != "INCOMPLETE" {
				t.Fatal("invalid measurement accepted", r.Snapshot())
			}
		})
	}
}

func TestDisabledAndCanceledContextDoNotChangeBusinessExecution(t *testing.T) {
	Start(nil, "generation").End(true)
	Start(context.Background(), "generation").End(true)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := NewRecorder()
	Start(WithRecorder(ctx, r), "generation").End(false)
	if r.Snapshot().Status != "OBSERVED" || ctx.Err() != context.Canceled {
		t.Fatal("diagnostic altered cancellation")
	}
}
