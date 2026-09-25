package bidir

import "fmt"

// bxMemorySourceObserver is an observer-owned source boundary for fixtures.
// It has no producer-facing before/after injection or write operation.
type bxMemorySourceObserver struct {
	snapshot BXFileSnapshot
}

// NewBXMemoryRejectedWriteObserver creates a deterministic no-write observer.
func NewBXMemoryRejectedWriteObserver(document Document) BXRejectedWriteObserver {
	return &bxMemorySourceObserver{snapshot: memorySourceSnapshot(document)}
}

func (observer *bxMemorySourceObserver) Kind() string { return "memory-source" }

func (observer *bxMemorySourceObserver) ObserveRejected(operation func() error) (BXWriteObservation, error) {
	if operation == nil {
		return BXWriteObservation{}, fmt.Errorf("rejected operation is nil")
	}
	before := cloneSnapshot(observer.snapshot)
	if err := operation(); err != nil {
		// Reconciliation errors are expected for this observer; the caller
		// records them separately from observer failures.
	}
	after := cloneSnapshot(observer.snapshot)
	return BXWriteObservation{Observed: true, Before: before, After: after}, nil
}

func memorySourceSnapshot(document Document) BXFileSnapshot {
	path := document.Package
	if path == "" {
		path = "fixture"
	}
	bytes := documentSourceBytes(document)
	return BXFileSnapshot{Bytes: bytes, LStat: BXLStat{
		Path: path + ".gooo", Size: int64(len(bytes)), Mode: 0o644, Exists: true,
	}}
}

func cloneSnapshot(snapshot BXFileSnapshot) BXFileSnapshot {
	return BXFileSnapshot{Bytes: append([]byte(nil), snapshot.Bytes...), LStat: snapshot.LStat}
}
