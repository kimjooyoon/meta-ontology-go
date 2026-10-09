package bodyexecution

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func jointFillSlot(ctx context.Context, filename string, source []byte, activity string, prior *bodycodegen.IRBodyFillReceipt) (jointSlot, error) {
	slot := jointSlot{activity: activity}
	set, err := bodycodegen.PlanSourceFillCandidates(ctx, filename, source, activity)
	if err != nil {
		return slot, err
	}
	if prior.IRPlanSHA256 != set.PlanSHA256 || len(prior.CandidateScores)+len(prior.RejectedCandidates) != len(set.Candidates) {
		return slot, fmt.Errorf("joint source fill contract differs from initial construction")
	}
	slot.fillRejected = len(prior.RejectedCandidates) != 0
	found := false
	for i, candidate := range set.Candidates {
		slot.fillIDs = append(slot.fillIDs, candidate.ID)
		slot.ranking = append(slot.ranking, uint16(i))
		if candidate.ID == prior.SelectedCandidateID {
			slot.initial, found = uint16(i), true
		}
	}
	if !found {
		return slot, fmt.Errorf("initial fill selection is absent from its source contract")
	}
	return slot, nil
}

func jointObservationSchema(slots []jointSlot, attempts []JointAttempt) string {
	// v7 includes typed candidates, local rejections and native arithmetic faults.
	// Older producer schemas keep their original derivation unchanged.
	for _, slot := range slots {
		if len(slot.pathMasks) != 0 {
			return jointPathSchema
		}
	}
	for _, attempt := range attempts {
		if hasCompositionFault(attempt.Runtime) {
			return jointFaultSchema
		}
	}
	schema := jointSchema
	for _, slot := range slots {
		if slot.fillRejected {
			return jointFillRejectionSchema
		}
		if len(slot.fillIDs) != 0 {
			schema = jointFillSchema
		}
		if len(slot.searchIDs) != 0 && schema != jointFillSchema {
			schema = jointMixedSchema
		}
	}
	for _, attempt := range attempts {
		if attempt.Rejection != nil && attempt.Rejection.Stage == "LOCAL_SOURCE_FILL" {
			return jointFillRejectionSchema
		}
	}
	if schema == jointMixedSchema {
		for _, attempt := range attempts {
			if attempt.Rejection != nil {
				return jointRejectionSchema
			}
		}
	}
	return schema
}
