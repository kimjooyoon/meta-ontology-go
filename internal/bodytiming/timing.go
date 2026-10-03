// Package bodytiming records bounded, sequential wall intervals for optional
// diagnostics. These observations do not authorize or change construction.
package bodytiming

import (
	"context"
	"sync"
	"time"
)

const Capacity = 32

type Phase struct {
	Name    string `json:"name"`
	StartNS int64  `json:"start_ns"`
	EndNS   int64  `json:"end_ns"`
	Outcome string `json:"outcome"`
}

type Snapshot struct {
	Status       string  `json:"status"`
	CaptureNS    int64   `json:"capture_ns"`
	UnassignedNS int64   `json:"unassigned_ns"`
	Phases       []Phase `json:"phases"`
}

type Recorder struct {
	mu      sync.Mutex
	now     func() time.Time
	origin  time.Time
	phases  [Capacity]Phase
	count   int
	active  bool
	invalid bool
}

type contextKey struct{}

func NewRecorder() *Recorder { return newRecorder(time.Now) }

func newRecorder(now func() time.Time) *Recorder {
	return &Recorder{now: now, origin: now()}
}

func WithRecorder(ctx context.Context, r *Recorder) context.Context {
	return context.WithValue(ctx, contextKey{}, r)
}

type Span struct {
	owner *Recorder
	index int
	done  bool // guarded by owner.mu
}

// Start is a no-op without a recorder. No clock is read in that case.
func Start(ctx context.Context, name string) *Span {
	if ctx == nil {
		return nil
	}
	r, _ := ctx.Value(contextKey{}).(*Recorder)
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.active || r.count == Capacity || !KnownPhase(name) {
		r.invalid = true
		return nil
	}
	start := r.now().Sub(r.origin).Nanoseconds()
	if start < 0 || (r.count > 0 && start < r.phases[r.count-1].EndNS) {
		r.invalid = true
	}
	index := r.count
	r.phases[index] = Phase{Name: name, StartNS: start, EndNS: start, Outcome: "incomplete"}
	r.count++
	r.active = true
	return &Span{owner: r, index: index}
}

// End is nil-safe and idempotent, including concurrent duplicate calls.
func (s *Span) End(success bool) {
	if s == nil {
		return
	}
	r := s.owner
	r.mu.Lock()
	defer r.mu.Unlock()
	if s.done {
		return
	}
	s.done = true
	p := &r.phases[s.index]
	p.EndNS = r.now().Sub(r.origin).Nanoseconds()
	if p.EndNS < p.StartNS {
		r.invalid = true
	}
	p.Outcome = "failed"
	if success {
		p.Outcome = "completed"
	}
	r.active = false
}

// Snapshot copies the retained intervals. Unassigned time includes inter-stage
// bookkeeping; it is not attributed to any child process or CPU measurement.
func (r *Recorder) Snapshot() Snapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := Snapshot{Status: "OBSERVED", CaptureNS: r.now().Sub(r.origin).Nanoseconds(),
		Phases: append([]Phase{}, r.phases[:r.count]...)}
	var total int64
	for _, p := range s.Phases {
		total += p.EndNS - p.StartNS
	}
	s.UnassignedNS = s.CaptureNS - total
	if r.invalid || r.active || s.UnassignedNS < 0 || s.CaptureNS < 0 ||
		(r.count > 0 && s.CaptureNS < r.phases[r.count-1].EndNS) {
		s.Status = "INCOMPLETE"
	}
	return s
}

func KnownPhase(name string) bool {
	switch name {
	case "request_decode", "source_plan", "generation", "generation_parent_encode",
		"executor_gate", "parent_receipt_validate", "source_replay", "runtime_suite_prepare",
		"go_tool_hash", "toolchain_bind", "executable_prepare", "executable_hash_1",
		"executable_hash_2", "native_run_1", "native_run_2", "runtime_decode_1", "runtime_decode_2",
		"runtime_receipt", "artifact_save", "artifact_bind":
		return true
	}
	return false
}
