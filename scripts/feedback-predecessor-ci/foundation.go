package main

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackpredecessor"
)

func verifyFoundation(ctx context.Context, client *githubClient, cfg config) (feedbackpredecessor.FoundationEvidence, error) {
	expected := feedbackpredecessor.FoundationEvidenceForConfirmedGap()
	if cfg.branch != "main" || cfg.workflow != "CI [push full]" {
		return feedbackpredecessor.FoundationEvidence{}, fmt.Errorf("foundation recovery requires the canonical main push workflow")
	}
	var run foundationRun
	if err := client.getJSON(ctx, fmt.Sprintf("/repos/%s/actions/runs/%d", cfg.repository, expected.LastKnownGoodRunID), &run); err != nil {
		return feedbackpredecessor.FoundationEvidence{}, err
	}
	if !validFoundationRun(run, cfg.workflow, expected) {
		return feedbackpredecessor.FoundationEvidence{}, fmt.Errorf("last-known-good CI run is not exact")
	}
	if err := verifyFoundationArtifact(ctx, client, cfg.repository, run.ID, expected); err != nil {
		return feedbackpredecessor.FoundationEvidence{}, err
	}
	if err := verifyFoundationHistory(ctx, client, cfg.repository, expected); err != nil {
		return feedbackpredecessor.FoundationEvidence{}, err
	}
	return expected, nil
}

func validFoundationRun(run foundationRun, workflow string, expected feedbackpredecessor.FoundationEvidence) bool {
	return run.ID == expected.LastKnownGoodRunID && run.Name == workflow &&
		run.Path == ".github/workflows/ci.yml" && run.HeadBranch == "main" &&
		run.HeadSHA == expected.LastKnownGoodSHA && run.Event == "push" &&
		run.Status == "completed" && run.Conclusion == "success" && run.RunAttempt == 1
}
