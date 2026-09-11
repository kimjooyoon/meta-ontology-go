package main

import (
	"context"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackpredecessor"
)

func confirmedFoundation(ctx context.Context, client *githubClient, cfg config, candidates int) (feedbackpredecessor.FoundationEvidence, bool, error) {
	if candidates != 0 || cfg.branch != "main" ||
		cfg.predecessorSHA != feedbackpredecessor.FoundationMissingPredecessorSHA {
		return feedbackpredecessor.FoundationEvidence{}, false, nil
	}
	foundation, err := verifyFoundation(ctx, client, cfg)
	return foundation, true, err
}
