package bodyexecution

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

func TestNativeWaitBudgetStartsAfterHostCreation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		parent, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		ctx, cancel := context.WithCancelCause(parent)
		defer cancel(nil)
		timing, wall, phase, err := timedProcessCalls(ctx,
			func() error { time.Sleep(3 * time.Second); return nil },
			func() error { time.Sleep(100 * time.Millisecond); return nil }, 2*time.Second, cancel)
		if err != nil || phase != "" || timing.ContextAtStartReturn != "ACTIVE" ||
			timing.StartNS != int64(3*time.Second) || timing.WaitNS != int64(100*time.Millisecond) ||
			timing.WaitLimitNS != int64(2*time.Second) || wall != timing.StartNS+timing.WaitNS ||
			timing.DeadlineRemainingNS == nil || *timing.DeadlineRemainingNS != int64(10*time.Second) {
			t.Fatal("host creation spent the native wait budget", timing, wall, phase, err)
		}
		time.Sleep(3 * time.Second)
		if ctx.Err() != nil {
			t.Fatal("completed process left an active timeout", context.Cause(ctx))
		}
	})
}

func TestNativeWaitBudgetStillExpiresAndKeepsParentDeadline(t *testing.T) {
	for _, duration := range []time.Duration{4 * time.Second, 10 * time.Second} {
		synctest.Test(t, func(t *testing.T) {
			parent, stop := context.WithTimeout(context.Background(), duration)
			defer stop()
			ctx, cancel := context.WithCancelCause(parent)
			defer cancel(nil)
			timing, _, phase, err := timedProcessCalls(ctx,
				func() error { time.Sleep(3 * time.Second); return nil },
				func() error { <-ctx.Done(); return ctx.Err() }, 2*time.Second, cancel)
			wantWait := min(duration-3*time.Second, 2*time.Second)
			if err == nil || phase != "WAIT" || timing.WaitNS != int64(wantWait) ||
				!errors.Is(context.Cause(ctx), context.DeadlineExceeded) {
				t.Fatal("native wait extended its own or parent's budget", timing, err, context.Cause(ctx))
			}
		})
	}
}

func TestNativeWaitBudgetJoinsStartedChildAfterParentExpires(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		parent, stop := context.WithTimeout(context.Background(), time.Second)
		defer stop()
		ctx, cancel := context.WithCancelCause(parent)
		defer cancel(nil)
		waited := false
		timing, _, phase, err := timedProcessCalls(ctx,
			func() error { time.Sleep(3 * time.Second); return nil },
			func() error { waited = true; return nil }, 2*time.Second, cancel)
		if !waited || !errors.Is(err, context.DeadlineExceeded) || phase != "WAIT" ||
			timing.ContextAtStartReturn != "DEADLINE_EXCEEDED" || timing.WaitNS != 0 {
			t.Fatal("expired parent was reset or child was not joined", timing, waited, err)
		}
	})
}

func TestNativeWaitBudgetDoesNotWaitOrArmAfterStartFailure(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancelCause(context.Background())
		defer cancel(nil)
		failure := errors.New("start failed")
		timing, _, phase, err := timedProcessCalls(ctx, func() error { return failure },
			func() error { t.Fatal("wait after unsuccessful start"); return nil }, 2*time.Second, cancel)
		time.Sleep(3 * time.Second)
		if err != failure || phase != "START" || timing.WaitNS != 0 || timing.WaitLimitNS != 0 || ctx.Err() != nil {
			t.Fatal(timing, phase, err, ctx.Err())
		}
	})
}

func TestNativeWaitBudgetKeepsExplicitParentCancellation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		parent, stop := context.WithCancel(context.Background())
		defer stop()
		ctx, cancel := context.WithCancelCause(parent)
		defer cancel(nil)
		time.AfterFunc(time.Second, stop)
		waited := false
		timing, _, phase, err := timedProcessCalls(ctx,
			func() error { time.Sleep(3 * time.Second); return nil },
			func() error { waited = true; return nil }, 2*time.Second, cancel)
		if !waited || !errors.Is(err, context.Canceled) || phase != "WAIT" ||
			timing.ContextAtStartReturn != "CANCELED" || timing.WaitNS != 0 {
			t.Fatal("parent cancellation was lost or wait was skipped", timing, waited, err)
		}
	})
}

func TestNativeWaitLimitTerminatesRealChild(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, r, err := processWithWaitLimit(ctx, 100*time.Millisecond, t.TempDir(), binary, nil,
		"-test.run=^TestRuntimeProcessHelper$", "--", "body-helper-sleep")
	if err == nil || !r.Started || r.Completed || !r.TimedOut || r.Canceled || r.ExitCode == nil ||
		r.FailurePhase != "WAIT" || r.Timing == nil || r.Timing.WaitLimitNS != int64(100*time.Millisecond) ||
		r.Timing.StartNS+r.Timing.WaitNS != r.WallNS {
		t.Fatalf("native wait did not stop and join the child: %+v %v", r, err)
	}
}

func TestNativeWaitLegacyPhaseHasNoPostStartBudget(t *testing.T) {
	var old ProcessTiming
	raw := `{"start_ns":2854721800,"wait_ns":510000,"deadline_remaining_ns":2000000000,"context_at_start_return":"DEADLINE_EXCEEDED"}`
	if err := json.Unmarshal([]byte(raw), &old); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(old)
	if err != nil || old.WaitLimitNS != 0 || strings.Contains(string(encoded), "wait_limit_ns") ||
		old.StartNS != 2854721800 || *old.DeadlineRemainingNS != 2000000000 {
		t.Fatal("historical shared budget was reinterpreted as a post-start budget", old, err)
	}
}
