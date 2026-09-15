package guardedpromotion

import (
	"context"
	"fmt"
	"net/url"
)

func (collector *Collector) collectPredecessor(
	ctx context.Context, repositoryPath string, source *Source,
) error {
	switch source.Workflow.Event {
	case "push":
		return collector.collectPushPredecessor(ctx, repositoryPath, source)
	case "pull_request":
		return collector.collectPullRequestPredecessor(ctx, repositoryPath, source)
	default:
		return fmt.Errorf("source workflow event %q is unknown", source.Workflow.Event)
	}
}

func (collector *Collector) collectPushPredecessor(
	ctx context.Context, repositoryPath string, source *Source,
) error {
	var response commitResponse
	path := repositoryPath + "/commits/" + url.PathEscape(source.CurrentHeadSHA)
	if err := collector.getJSON(ctx, path, &response); err != nil {
		return err
	}
	if len(response.Parents) == 0 || !validSHA(response.Parents[0].SHA) {
		return fmt.Errorf("push predecessor is unavailable")
	}
	source.PredecessorSHA = response.Parents[0].SHA
	return nil
}

func (collector *Collector) collectPullRequestPredecessor(
	ctx context.Context, repositoryPath string, source *Source,
) error {
	var run workflowRunResponse
	path := fmt.Sprintf("%s/actions/runs/%d", repositoryPath, source.Workflow.RunID)
	if err := collector.getJSON(ctx, path, &run); err != nil {
		return err
	}
	if len(run.PullRequests) != 1 {
		return fmt.Errorf("pull request predecessor candidates = %d, want 1", len(run.PullRequests))
	}
	var response pullResponse
	path = fmt.Sprintf("%s/pulls/%d", repositoryPath, run.PullRequests[0].Number)
	if err := collector.getJSON(ctx, path, &response); err != nil {
		return err
	}
	if !validSHA(response.Base.SHA) {
		return fmt.Errorf("pull request base sha is unavailable")
	}
	source.PredecessorSHA = response.Base.SHA
	return nil
}
