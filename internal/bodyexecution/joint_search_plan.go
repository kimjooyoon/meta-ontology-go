package bodyexecution

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func jointSearchCandidateSet(ctx context.Context, filename string, source []byte, activity string,
	prior *bodycodegen.IRBodySearchReceipt) (bodycodegen.SourceSearchCandidateSet, error) {
	if prior.CandidateSet != nil {
		return bodycodegen.RetainedSourceSearchCandidates(ctx, filename, source, activity, prior)
	}
	return bodycodegen.PlanSourceSearchCandidates(ctx, filename, source, activity)
}

func jointSearchSlot(ctx context.Context, filename string, source []byte, activity string,
	search *bodycodegen.IRBodySearchReceipt) (jointSlot, error) {
	set, err := jointSearchCandidateSet(ctx, filename, source, activity, search)
	if err != nil {
		return jointSlot{}, err
	}
	if search.IRPlanSHA256 != set.PlanSHA256 || search.AttemptBudget == nil ||
		*search.AttemptBudget != set.AttemptBudget || search.CandidateCount != len(set.Candidates) {
		return jointSlot{}, fmt.Errorf("joint source search contract differs from initial construction")
	}
	slot := jointSlot{activity: activity}
	if search.CandidateSet != nil {
		slot.searchReceipt = search
	}
	found := false
	for i, candidate := range set.Candidates[:min(set.AttemptBudget, len(set.Candidates))] {
		slot.searchIDs = append(slot.searchIDs, candidate.ID)
		slot.ranking = append(slot.ranking, uint16(i))
		if candidate.ID == search.SelectedCandidateID {
			slot.initial, found = uint16(i), true
		}
	}
	if !found {
		return jointSlot{}, fmt.Errorf("initial search selection exceeds the source budget")
	}
	return slot, nil
}

func realizeJointSearch(ctx context.Context, filename string, source []byte, slot jointSlot,
	candidateID string) (bodycodegen.SearchCandidate, []byte, error) {
	if slot.searchReceipt != nil {
		return bodycodegen.RealizeRetainedSourceSearchCandidate(ctx, filename, source, slot.activity, slot.searchReceipt, candidateID)
	}
	return bodycodegen.RealizeSourceSearchCandidate(ctx, filename, source, slot.activity, candidateID)
}
