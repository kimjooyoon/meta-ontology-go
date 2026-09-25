package adapter

import (
	"crypto/sha256"
	"fmt"
)

type observerStamp struct {
	digest [sha256.Size]byte
}
type NoWriteObserver struct {
	binding          ObservationBinding
	paths            ObserverPaths
	before           FilesystemState
	workflow         WorkflowBinding
	workflowCaptured bool
	mutation         MutationEvidence
	mutationCaptured bool
	stamp            *observerStamp
	finished         bool
}

// NewNoWriteObserver captures the pre-invocation state using os.Lstat and bytes.
func NewNoWriteObserver(binding ObservationBinding, paths ObserverPaths) (*NoWriteObserver, error) {
	if err := binding.validate(); err != nil {
		return nil, err
	}
	normalized, err := normalizeObserverPaths(paths)
	if err != nil {
		return nil, err
	}
	before, err := captureState(normalized)
	if err != nil {
		return nil, fmt.Errorf("capture before state: %w", err)
	}
	return &NoWriteObserver{
		binding: binding, paths: normalized, before: before,
		workflow: missingWorkflowBinding(), mutation: missingMutationEvidence(),
		stamp: &observerStamp{},
	}, nil
}

// Finish captures the post-invocation state. Differences are verified by the oracle.
func (o *NoWriteObserver) Finish() (NoWriteObservation, error) {
	return o.finish("")
}

// CaptureRejected closes an observer around a cancelled or closed transaction.
func (o *NoWriteObserver) CaptureRejected(reason RejectionKind) (NoWriteObservation, error) {
	if !validRejectionKind(reason) {
		return NoWriteObservation{}, fmt.Errorf("unsupported rejection kind %q", reason)
	}
	return o.finish(reason)
}
func (o *NoWriteObserver) finish(reason RejectionKind) (NoWriteObservation, error) {
	if o == nil || o.stamp == nil || o.finished {
		return NoWriteObservation{}, fmt.Errorf("observer is not initialized")
	}
	o.finished = true
	after, err := captureState(o.paths)
	if err != nil {
		return NoWriteObservation{}, fmt.Errorf("capture after state: %w", err)
	}
	observation := NoWriteObservation{
		Binding: o.binding, Paths: o.paths, Workflow: o.workflow,
		Mutation: o.mutation, Reason: reason, Before: o.before, After: after,
		stamp: o.stamp,
	}
	o.stamp.digest = observationSeal(observation)
	return observation, nil
}
