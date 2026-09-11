package main

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackpredecessor"
)

func verifyFoundationHistory(ctx context.Context, client *githubClient, repository string, expected feedbackpredecessor.FoundationEvidence) error {
	var comparison compareResult
	endpoint := fmt.Sprintf("/repos/%s/compare/%s...%s", repository,
		expected.LastKnownGoodSHA, expected.MissingPredecessorSHA)
	if err := client.getJSON(ctx, endpoint, &comparison); err != nil {
		return err
	}
	if comparison.Status != "ahead" || comparison.AheadBy != len(expected.GapCommitSHAs) || comparison.BehindBy != 0 {
		return fmt.Errorf("last-known-good is not the exact confirmed ancestor")
	}
	baseSHAs := []string{expected.LastKnownGoodSHA, expected.GapCommitSHAs[0], expected.GapCommitSHAs[1]}
	for index, number := range expected.GapPRNumbers {
		var pull foundationPull
		if err := client.getJSON(ctx, fmt.Sprintf("/repos/%s/pulls/%d", repository, number), &pull); err != nil {
			return err
		}
		if pull.Number != number || !pull.Merged || pull.MergeCommitSHA != expected.GapCommitSHAs[index] ||
			pull.Base.Ref != "main" || pull.Base.SHA != baseSHAs[index] {
			return fmt.Errorf("foundation gap PR #%d is not exact", number)
		}
	}
	return nil
}
