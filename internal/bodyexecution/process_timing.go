package bodyexecution

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ProcessTiming separates the host's Start call from Wait, including pipe I/O.
// Neither interval is a measurement of when generated user code began running.
type ProcessTiming struct {
	StartNS              int64  `json:"start_ns"`
	WaitNS               int64  `json:"wait_ns"`
	WaitLimitNS          int64  `json:"wait_limit_ns,omitempty"`
	DeadlineRemainingNS  *int64 `json:"deadline_remaining_ns,omitempty"`
	ContextAtStartReturn string `json:"context_at_start_return"`
}

func timedProcessCalls(ctx context.Context, startProcess, waitProcess func() error,
	waitLimit time.Duration, cancel context.CancelCauseFunc) (ProcessTiming, int64, string, error) {
	var timing ProcessTiming
	start := time.Now()
	if deadline, ok := ctx.Deadline(); ok {
		remaining := deadline.Sub(start).Nanoseconds()
		timing.DeadlineRemainingNS = &remaining
	}
	err := startProcess()
	started := time.Now()
	timing.StartNS = started.Sub(start).Nanoseconds()
	timing.ContextAtStartReturn = processContextState(ctx)
	if err != nil {
		return timing, timing.StartNS, "START", processContextFailure(ctx, err)
	}
	timing.WaitLimitNS = int64(waitLimit)
	err = processContextFailure(ctx, boundedProcessWait(ctx, waitProcess, waitLimit, cancel))
	timing.WaitNS = time.Since(started).Nanoseconds()
	phase := ""
	if err != nil {
		phase = "WAIT"
	}
	return timing, timing.StartNS + timing.WaitNS, phase, err
}

// Keep interruption and operating-system failure available to errors.Is/As.
// exec.Cmd.Wait commonly returns an ExitError after cancellation kills a child.
// The cause is attached after waiting so the child is still always joined.
func processContextFailure(ctx context.Context, failure error) error {
	cause := context.Cause(ctx)
	if cause == nil || errors.Is(failure, cause) {
		return failure
	}
	if failure == nil {
		return cause
	}
	return fmt.Errorf("%w: %w", cause, failure)
}

// The parent's deadline is never replaced. Native calls add a shorter budget
// only after Start returns; toolchain and build calls keep their parent budget.
func boundedProcessWait(ctx context.Context, wait func() error, limit time.Duration, cancel context.CancelCauseFunc) error {
	if limit <= 0 {
		return wait()
	}
	done := make(chan struct{})
	timer := time.AfterFunc(limit, func() {
		cancel(context.DeadlineExceeded)
		close(done)
	})
	err := wait()
	if !timer.Stop() {
		<-done
	}
	if err == nil {
		err = context.Cause(ctx)
	}
	return err
}

func processContextState(ctx context.Context) string {
	switch ctx.Err() {
	case context.DeadlineExceeded:
		return "DEADLINE_EXCEEDED"
	case context.Canceled:
		return "CANCELED"
	default:
		return "ACTIVE"
	}
}
