package bodyexecution

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func materializeJointPath(ctx context.Context, filename string, source []byte, slot jointSlot,
	mask uint16, index int, attempt *JointAttempt) ([]byte, error) {
	if !slices.Contains(slot.pathMasks, mask) {
		return nil, fmt.Errorf("typed path mask exceeds the source attempt cap")
	}
	candidate, selected, err := bodycodegen.RealizeSourcePathCandidate(ctx, filename, source, slot.activity, slot.pathReceipt, mask)
	attempt.PathCandidates = append(attempt.PathCandidates, candidate)
	if err != nil {
		if rejected, ok := errors.AsType[*bodycodegen.SourcePathCandidateRejection](err); ok && ctx.Err() == nil {
			attempt.Rejection = &JointCandidateRejection{Stage: "LOCAL_TYPED_PATH", Slot: index,
				Activity: slot.activity, CandidateID: fmt.Sprintf("path_%04x", mask), Reason: rejected.Error()}
			return nil, nil
		}
		return nil, fmt.Errorf("joint typed path activity %s: %w", slot.activity, err)
	}
	attempt.LocalPassed += candidate.Passed
	attempt.LocalTotal += candidate.Total
	return selected, nil
}
