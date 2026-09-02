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

func Collect(ctx context.Context, client *http.Client, apiURL, token, repository, predecessorSHA, requestedRoute string) (Collection, error) {
	collection := Collection{RequestedRoute: requestedRoute}
	if client == nil || apiURL == "" || token == "" || repository == "" || !validSHA(predecessorSHA) {
		return collection, fmt.Errorf("proposal predecessor collector identity is invalid")
	}
	if !validRoute(requestedRoute) {
		return collection, &Failure{Reason: ReasonRouteUnknown, Err: fmt.Errorf("requested route is not an allowed branch")}
	}
	runsURL := fmt.Sprintf("%s/repos/%s/actions/workflows/metric-counterfactual.yml/runs?head_sha=%s&event=push&status=completed&per_page=100", strings.TrimRight(apiURL, "/"), repository, url.QueryEscape(predecessorSHA))
	var runs runsEnvelope
	if err := getJSON(ctx, client, runsURL, token, &runs); err != nil {
		return collection, err
	}
	collection.ObservedRuns = len(runs.WorkflowRuns)
	if runs.TotalCount != len(runs.WorkflowRuns) {
		return collection, &Failure{Reason: ReasonRunPaginationIncomplete}
	}
	for _, run := range runs.WorkflowRuns {
		if run.HeadSHA != predecessorSHA {
			continue
		}
		if run.HeadBranch == "" {
			collection.RouteUnknownRuns++
			collection.Unresolved++
			continue
		}
		if run.HeadBranch != requestedRoute {
			collection.OtherRouteRuns++
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
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	if err := decoder.Decode(target); err != nil {
		return &Failure{Reason: ReasonResponseMalformed, Err: err}
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return &Failure{Reason: ReasonResponseMalformed, Err: fmt.Errorf("trailing JSON")}
		}
		return &Failure{Reason: ReasonResponseMalformed, Err: err}
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
	terminal := run.Conclusion == "success" || run.Conclusion == "failure"
	return run.ID > 0 && run.RunAttempt > 0 && run.Event == "push" &&
		run.Status == "completed" && terminal && run.Name == workflowName
}
