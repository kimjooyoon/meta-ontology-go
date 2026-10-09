package bodyexecution

import (
	"context"
	"os/exec"
	"time"
)

// ProcessTiming separates the host's Start call from Wait, including pipe I/O.
// Neither interval is a measurement of when generated user code began running.
type ProcessTiming struct {
	StartNS              int64  `json:"start_ns"`
	WaitNS               int64  `json:"wait_ns"`
	DeadlineRemainingNS  *int64 `json:"deadline_remaining_ns,omitempty"`
	ContextAtStartReturn string `json:"context_at_start_return"`
}

func timedProcess(ctx context.Context, cmd *exec.Cmd) (ProcessTiming, int64, string, error) {
	var timing ProcessTiming
	start := time.Now()
	if deadline, ok := ctx.Deadline(); ok {
		remaining := deadline.Sub(start).Nanoseconds()
		timing.DeadlineRemainingNS = &remaining
	}
	err := cmd.Start()
	started := time.Now()
	timing.StartNS = started.Sub(start).Nanoseconds()
	timing.ContextAtStartReturn = processContextState(ctx)
	if err != nil {
		return timing, timing.StartNS, "START", err
	}
	err = cmd.Wait()
	timing.WaitNS = time.Since(started).Nanoseconds()
	phase := ""
	if err != nil {
		phase = "WAIT"
	}
	return timing, timing.StartNS + timing.WaitNS, phase, err
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
