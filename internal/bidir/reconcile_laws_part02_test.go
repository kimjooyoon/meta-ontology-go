package bidir

import (
	"errors"
)

func hasConflict(err error, want ConflictKind) bool {
	var reconcileErr *ReconcileError
	if !errors.As(err, &reconcileErr) {
		return false
	}
	for _, conflict := range reconcileErr.Conflicts {
		if conflict.Kind == want {
			return true
		}
	}
	return false
}
