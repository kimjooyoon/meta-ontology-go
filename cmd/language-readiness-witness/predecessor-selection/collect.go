package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorselection"
)

func collect(ctx context.Context, client *githubClient, cfg config, predecessor string) (predecessorselection.Input, error) {
	query := url.Values{"branch": {cfg.branch}, "event": {"workflow_run"},
		"status": {"completed"}, "head_sha": {predecessor}, "per_page": {"100"}}
	endpoint := fmt.Sprintf("/repos/%s/actions/workflows/transformation-effect.yml/runs?%s",
		cfg.repository, query.Encode())
	input := predecessorselection.Input{Repository: cfg.repository,
		CurrentHeadSHA: cfg.currentHead, PredecessorSHA: predecessor,
		Branch: cfg.branch, Workflow: cfg.workflow,
		Pagination: predecessorselection.Pagination{Pages: []predecessorselection.PaginationPage{}}}
	runs, pages, failureReason := collectWorkflowRuns(ctx, client, endpoint)
	input.Pagination.Pages = append(input.Pagination.Pages, pages...)
	if failureReason != "" {
		input.Pagination.PageCount = len(input.Pagination.Pages)
		input.Pagination.FailureReason = failureReason
		return input, nil
	}
	for _, run := range runs {
		candidate, ok, runPages, failureReason := collectRun(ctx, client, cfg, predecessor, run)
		input.Pagination.Pages = append(input.Pagination.Pages, runPages...)
		if failureReason != "" {
			input.Pagination.PageCount = len(input.Pagination.Pages)
			input.Pagination.FailureReason = failureReason
			return input, nil
		}
		if ok {
			input.Candidates = append(input.Candidates, candidate)
		}
	}
	input.Pagination.PageCount = len(input.Pagination.Pages)
	input.Pagination.Complete = true
	sort.Slice(input.Candidates, func(i, j int) bool {
		return input.Candidates[i].RunID < input.Candidates[j].RunID
	})
	return input, nil
}

func encodedPayload(ctx context.Context, client *githubClient, repository string,
	metadata artifactMetadata, filename string) (string, error) {
	if metadata.Expired {
		return "", nil
	}
	archive, err := client.get(ctx, fmt.Sprintf("/repos/%s/actions/artifacts/%d/zip",
		repository, metadata.ID))
	if err != nil {
		return "", err
	}
	payload, err := verifiedPayload(archive, filename)
	return base64.StdEncoding.EncodeToString(payload), err
}
