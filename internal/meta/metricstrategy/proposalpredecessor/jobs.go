package proposalpredecessor

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

const synthesisJobName = "strategy"

type githubJob struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}

type jobsEnvelope struct {
	TotalCount int         `json:"total_count"`
	Jobs       []githubJob `json:"jobs"`
}

func collectSynthesisJob(
	ctx context.Context, client *http.Client, apiURL, token, repository string,
	runID int64, collection *Collection,
) (githubJob, bool, error) {
	target := fmt.Sprintf("%s/repos/%s/actions/runs/%d/jobs?filter=latest&per_page=100",
		strings.TrimRight(apiURL, "/"), repository, runID)
	var response jobsEnvelope
	if err := getJSON(ctx, client, target, token, &response); err != nil {
		return githubJob{}, false, err
	}
	collection.ObservedJobs += len(response.Jobs)
	if response.TotalCount != len(response.Jobs) {
		return githubJob{}, false, &Failure{Reason: ReasonJobPaginationIncomplete}
	}
	matches := make([]githubJob, 0, 1)
	for _, job := range response.Jobs {
		if job.Name == synthesisJobName {
			collection.ExactJobs++
			matches = append(matches, job)
		}
	}
	if len(matches) != 1 {
		return githubJob{}, false, nil
	}
	job := matches[0]
	ready := job.ID > 0 && job.Status == "completed" && job.Conclusion == "success"
	return job, ready, nil
}
