package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"sort"
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorselection"
)

func collect(ctx context.Context, client *githubClient, cfg config, predecessor string) (predecessorselection.Input, error) {
	runs, err := collectWorkflowRuns(ctx, client, cfg, predecessor)
	if err != nil {
		return predecessorselection.Input{}, err
	}
	input := predecessorselection.Input{Repository: cfg.repository,
		CurrentHeadSHA: cfg.currentHead, PredecessorSHA: predecessor,
		Branch: cfg.branch, Workflow: cfg.workflow}
	for _, run := range runs {
		candidate, ok, err := collectRun(ctx, client, cfg, predecessor, run)
		if err != nil {
			return predecessorselection.Input{}, err
		}
		if ok {
			input.Candidates = append(input.Candidates, candidate)
		}
	}
	sort.Slice(input.Candidates, func(i, j int) bool {
		return input.Candidates[i].RunID < input.Candidates[j].RunID
	})
	return input, nil
}

const workflowRunPageSize = 100

func collectWorkflowRuns(ctx context.Context, client *githubClient, cfg config, predecessor string) ([]workflowRun, error) {
	const maxPages = 1000
	var all []workflowRun
	totalCount := -1
	for page := 1; page <= maxPages; page++ {
		query := url.Values{"branch": {cfg.branch}, "event": {"workflow_run"},
			"status": {"completed"}, "head_sha": {predecessor},
			"per_page": {strconv.Itoa(workflowRunPageSize)}, "page": {strconv.Itoa(page)}}
		endpoint := fmt.Sprintf("/repos/%s/actions/workflows/transformation-effect.yml/runs?%s",
			cfg.repository, query.Encode())
		var runs workflowRunList
		if err := client.getJSON(ctx, endpoint, &runs); err != nil {
			return nil, err
		}
		if totalCount < 0 {
			totalCount = runs.TotalCount
		} else if runs.TotalCount != totalCount {
			return nil, fmt.Errorf("workflow run total count changed during pagination")
		}
		all = append(all, runs.WorkflowRuns...)
		if len(runs.WorkflowRuns) < workflowRunPageSize {
			if totalCount != len(all) {
				return nil, fmt.Errorf("workflow run pagination incomplete")
			}
			return all, nil
		}
	}
	return nil, fmt.Errorf("workflow run pagination exceeded the fail-closed page limit")
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
