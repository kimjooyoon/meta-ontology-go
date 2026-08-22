package main

import (
	"context"
	"fmt"
	"net/url"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackpredecessor"
)

func collect(ctx context.Context, client *githubClient, cfg config) (collection, error) {
	query := url.Values{"branch": {cfg.branch}, "event": {"push"}, "status": {"completed"},
		"head_sha": {cfg.predecessorSHA}, "per_page": {"100"}}
	endpoint := fmt.Sprintf("/repos/%s/actions/workflows/ci.yml/runs?%s",
		cfg.repository, query.Encode())
	var runs workflowRunList
	if err := client.getJSON(ctx, endpoint, &runs); err != nil {
		return collection{}, err
	}
	if runs.TotalCount != len(runs.WorkflowRuns) {
		return collection{}, fmt.Errorf("workflow run pagination is incomplete")
	}
	input := feedbackpredecessor.Input{Repository: cfg.repository,
		PredecessorSHA: cfg.predecessorSHA, CanonicalBranch: cfg.branch,
		CanonicalWorkflow: cfg.workflow}
	expectedName := "artifact-feedback-resolution-" + cfg.predecessorSHA
	for _, run := range runs.WorkflowRuns {
		if run.HeadSHA != cfg.predecessorSHA || run.HeadBranch != cfg.branch ||
			run.Event != "push" || run.Status != "completed" {
			continue
		}
		var artifacts artifactList
		endpoint := fmt.Sprintf("/repos/%s/actions/runs/%d/artifacts?per_page=100",
			cfg.repository, run.ID)
		if err := client.getJSON(ctx, endpoint, &artifacts); err != nil {
			return collection{}, err
		}
		if artifacts.TotalCount != len(artifacts.Artifacts) {
			return collection{}, fmt.Errorf("artifact pagination is incomplete")
		}
		for _, artifact := range artifacts.Artifacts {
			if artifact.Name != expectedName {
				continue
			}
			receipt := archivedReceipt{}
			if !artifact.Expired && run.Conclusion == "success" {
				archive, err := client.get(ctx, fmt.Sprintf(
					"/repos/%s/actions/artifacts/%d/zip", cfg.repository, artifact.ID))
				if err != nil {
					return collection{}, err
				}
				receipt = decodeReceipt(archive)
			}
			input.Candidates = append(input.Candidates, feedbackpredecessor.Candidate{
				ArtifactID: artifact.ID, RunID: run.ID, RunAttempt: run.RunAttempt,
				ArtifactName: artifact.Name, HeadSHA: run.HeadSHA, HeadBranch: run.HeadBranch,
				Workflow: run.Name, Event: run.Event, Conclusion: run.Conclusion,
				Expired: artifact.Expired, ReceiptDigest: receipt.ReceiptDigest,
				RepositoryWrites: receipt.RepositoryWrites,
			})
		}
	}
	sort.Slice(input.Candidates, func(i, j int) bool {
		left, right := input.Candidates[i], input.Candidates[j]
		return left.RunID < right.RunID ||
			left.RunID == right.RunID && left.ArtifactID < right.ArtifactID
	})
	return collection{Input: input}, nil
}
