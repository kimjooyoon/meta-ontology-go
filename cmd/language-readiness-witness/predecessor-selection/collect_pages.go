package main

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

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
