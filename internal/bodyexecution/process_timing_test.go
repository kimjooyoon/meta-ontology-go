package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRuntimeProcessSeparatesStartFailureFromWait(t *testing.T) {
	_, r, err := process(context.Background(), t.TempDir(), filepath.Join(t.TempDir(), "missing"), nil)
	if err == nil || r.Started || r.FailurePhase != "START" || r.Timing == nil ||
		r.Timing.StartNS < 0 || r.Timing.WaitNS != 0 || r.Timing.StartNS != r.WallNS ||
		r.Timing.ContextAtStartReturn != "ACTIVE" || r.Timing.DeadlineRemainingNS != nil {
		t.Fatalf("start failure timing lost: %+v %v", r, err)
	}
}

func TestRuntimeLegacyObservationKeepsTimingUnobserved(t *testing.T) {
	var r ProcessObservation
	if err := json.Unmarshal([]byte(`{"started":true,"completed":false,"timed_out":true,"wall_ns":2953490000}`), &r); err != nil {
		t.Fatal(err)
	}
	if r.Timing != nil || r.FailurePhase != "" || r.WallNS != 2953490000 {
		t.Fatal("historical process acquired phase measurements", r)
	}
}

func TestRuntimeProcessSuccessRetainsOutputAndPhaseTiming(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	raw, r, err := process(context.Background(), t.TempDir(), binary, nil,
		"-test.run=^TestRuntimeProcessHelper$", "--", "body-helper-ok")
	if err != nil || string(raw) != "ready" || !r.Started || !r.Completed || r.FailurePhase != "" ||
		r.Timing == nil || r.Timing.StartNS+r.Timing.WaitNS != r.WallNS || r.Timing.ContextAtStartReturn != "ACTIVE" {
		t.Fatalf("successful command changed: %s %+v %v", raw, r, err)
	}
}

func TestRuntimeProcessKeepsExpiredStartContext(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, expire := range []string{"CANCELED", "DEADLINE_EXCEEDED"} {
		t.Run(expire, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if expire == "DEADLINE_EXCEEDED" {
				ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
				defer cancel()
			}
			_, r, runErr := process(ctx, t.TempDir(), binary, nil)
			if runErr == nil || r.Started || r.FailurePhase != "START" || r.Timing == nil ||
				r.Timing.ContextAtStartReturn != expire || r.Timing.WaitNS != 0 || r.Timing.StartNS != r.WallNS {
				t.Fatalf("expired start was misreported: %+v %v", r, runErr)
			}
			if expire == "DEADLINE_EXCEEDED" && (r.Timing.DeadlineRemainingNS == nil || *r.Timing.DeadlineRemainingNS >= 0 || !r.TimedOut) {
				t.Fatal("expired budget was reset", r)
			}
		})
	}
}

func TestRuntimeProcessWaitTimingPreservesOriginalFailure(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, r, err := process(ctx, t.TempDir(), binary, nil, "-test.run=^TestRuntimeProcessHelper$", "--", "body-helper-exit")
	if err == nil || !r.Started || r.Completed || r.ExitCode == nil || *r.ExitCode != 7 ||
		r.FailurePhase != "WAIT" || r.Failure != "exit status 7" || r.Timing == nil ||
		r.Timing.StartNS < 0 || r.Timing.WaitNS < 0 || r.Timing.StartNS+r.Timing.WaitNS != r.WallNS ||
		r.Timing.ContextAtStartReturn != "ACTIVE" || r.Timing.DeadlineRemainingNS == nil || *r.Timing.DeadlineRemainingNS > int64(2*time.Second) {
		t.Fatalf("wait failure changed or phase lost: %+v %v", r, err)
	}
}
