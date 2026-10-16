package executioncontinuation

import (
	"encoding/json"
	"io/fs"
	"os"
	"sync/atomic"
)

// Execution is an invocation-local handle. JSON cannot recreate its authority.
type Execution struct {
	root        fs.FS
	prepared    Prepared
	reservation Reservation
	valid       bool
	used        atomic.Bool
}

// Acquire creates an exclusive caller-owned claim before any value evaluation.
// Serialized workflow concurrency and the retained reservation artifact cover
// separate workflow invocations; this claim also prevents local retries.
func Acquire(root fs.FS, p Prepared, r Reservation, claimPath string) (*Execution, error) {
	raw, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	var detached Prepared
	if decode(raw, &detached) != nil {
		return nil, refuted("PREPARED_SOURCE_INVALID")
	}
	if _, err := validatePrepared(root, detached); err != nil {
		return nil, err
	}
	if err := validateReservation(detached, r); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(claimPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, refuted("EXECUTION_ALREADY_CLAIMED_OR_UNAVAILABLE")
	}
	content, _ := json.Marshal(r)
	_, writeErr := file.Write(content)
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		return nil, refuted("EXECUTION_CLAIM_UNAVAILABLE")
	}
	return &Execution{root: root, prepared: detached, reservation: r, valid: true}, nil
}

func (e *Execution) Run() (Consumption, error) {
	if e == nil || !e.valid || !e.used.CompareAndSwap(false, true) {
		return Consumption{}, refuted("EXECUTION_HANDLE_UNAVAILABLE")
	}
	return consume(e.root, e.prepared, e.reservation)
}
