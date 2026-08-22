package proposalpredecessor

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const workflowName = "Metric counterfactual conformance"

func Collect(ctx context.Context, client *http.Client, apiURL, token, repository, predecessorSHA string) (Collection, error) {
	collection := Collection{}
	if client == nil || apiURL == "" || token == "" || repository == "" || !validSHA(predecessorSHA) {
		return collection, fmt.Errorf("proposal predecessor collector identity is invalid")
	}
	runsURL := fmt.Sprintf("%s/repos/%s/actions/workflows/metric-counterfactual.yml/runs?head_sha=%s&event=push&status=success&per_page=100", strings.TrimRight(apiURL, "/"), repository, url.QueryEscape(predecessorSHA))
	var runs runsEnvelope
	if err := getJSON(ctx, client, runsURL, token, &runs); err != nil {
		return collection, err
	}
	collection.ObservedRuns = len(runs.WorkflowRuns)
	if runs.TotalCount != len(runs.WorkflowRuns) {
		return collection, fmt.Errorf("proposal predecessor run pagination is unresolved")
	}
	for _, run := range runs.WorkflowRuns {
		if run.HeadSHA != predecessorSHA {
			continue
		}
		collection.ExactRuns++
		if !canonicalRun(run) {
			collection.Unresolved++
			continue
		}
		if err := collectRun(ctx, client, apiURL, token, repository, predecessorSHA, run, &collection); err != nil {
			return collection, err
		}
	}
	return collection, nil
}

func getJSON(ctx context.Context, client *http.Client, targetURL, token string, target any) error {
	payload, err := getBytes(ctx, client, targetURL, token)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(payload, target); err != nil {
		return fmt.Errorf("decode GitHub response: %w", err)
	}
	return nil
}

func readArchiveFile(file *zip.File) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	payload, err := io.ReadAll(io.LimitReader(reader, maximumResponseBytes+1))
	if err != nil || len(payload) > maximumResponseBytes {
		return nil, fmt.Errorf("proposal archive member exceeds fixed bound: %w", err)
	}
	return payload, nil
}

func canonicalRun(run githubRun) bool {
	return run.ID > 0 && run.RunAttempt > 0 && run.Event == "push" && run.Status == "completed" && run.Conclusion == "success" && run.Name == workflowName
}
