package proposalpredecessor

import (
	"context"
	"net/http"
)

// Collect preserves the original completed-run query and observation replay.
func Collect(ctx context.Context, client *http.Client, apiURL, token, repository, predecessorSHA, requestedRoute string) (Collection, error) {
	return collectRuns(ctx, client, apiURL, token, repository, predecessorSHA, requestedRoute, false)
}

// CollectPending also observes in-flight runs; they never become candidates.
func CollectPending(ctx context.Context, client *http.Client, apiURL, token, repository, predecessorSHA, requestedRoute string) (Collection, error) {
	return collectRuns(ctx, client, apiURL, token, repository, predecessorSHA, requestedRoute, true)
}

func isInFlightPredecessorRun(run githubRun) bool {
	if run.ID <= 0 || run.RunAttempt <= 0 || run.Event != "push" ||
		run.Name != workflowName || run.Conclusion != "" {
		return false
	}
	switch run.Status {
	case "queued", "in_progress", "waiting", "pending", "requested":
		return true
	default:
		return false
	}
}
