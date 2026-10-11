package bodyexecution

import (
	"context"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func jointSearchCandidateSet(ctx context.Context, filename string, source []byte, activity string,
	prior *bodycodegen.IRBodySearchReceipt) (bodycodegen.SourceSearchCandidateSet, error) {
	if prior.CandidateSet != nil {
		return bodycodegen.RetainedSourceSearchCandidates(ctx, filename, source, activity, prior)
	}
	return bodycodegen.PlanSourceSearchCandidates(ctx, filename, source, activity)
}

func realizeJointSearch(ctx context.Context, filename string, source []byte, slot jointSlot,
	candidateID string) (bodycodegen.SearchCandidate, []byte, error) {
	if slot.searchReceipt != nil {
		return bodycodegen.RealizeRetainedSourceSearchCandidate(ctx, filename, source, slot.activity, slot.searchReceipt, candidateID)
	}
	return bodycodegen.RealizeSourceSearchCandidate(ctx, filename, source, slot.activity, candidateID)
}
