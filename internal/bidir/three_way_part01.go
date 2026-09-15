package bidir

import (
	"fmt"
	"strings"
)

// ThreeWayConflictCode identifies a deterministic semantic merge rejection.
type ThreeWayConflictCode string

const (
	ThreeWayInvalidModel       ThreeWayConflictCode = "invalid-model"
	ThreeWayEndpointInvalid    ThreeWayConflictCode = "endpoint-invalid"
	ThreeWaySameIdentity       ThreeWayConflictCode = "same-identity-changed-differently"
	ThreeWayDeleteVsModify     ThreeWayConflictCode = "delete-vs-modify"
	ThreeWayRelationAttributes ThreeWayConflictCode = "incompatible-relation-attributes"
)

// ThreeWayConflict is a stable, presentation-insensitive merge diagnostic.
type ThreeWayConflict struct {
	Code             ThreeWayConflictCode
	Scope            string
	Identity         string
	Message          string
	BaseFingerprint  string
	LeftFingerprint  string
	RightFingerprint string
}

// ThreeWayConflictError exposes all conflicts in deterministic order.
type ThreeWayConflictError struct {
	Conflicts []ThreeWayConflict
}

func (e *ThreeWayConflictError) Error() string {
	if e == nil || len(e.Conflicts) == 0 {
		return "three-way reconciliation failed"
	}
	parts := make([]string, len(e.Conflicts))
	for index, conflict := range e.Conflicts {
		parts[index] = fmt.Sprintf("%s[%s:%s]: %s", conflict.Code, conflict.Scope, conflict.Identity, conflict.Message)
	}
	return strings.Join(parts, "; ")
}

// ThreeWayResult is the normalized output of a successful three-way merge.
// On conflict, Model is the detached normalized base and Conflicts is filled.
type ThreeWayResult struct {
	Model     Model
	Delta     Delta
	Locality  Locality
	Conflicts []ThreeWayConflict
}

// Succeeded reports whether the merge produced no structured conflicts.
func (r ThreeWayResult) Succeeded() bool { return len(r.Conflicts) == 0 }
