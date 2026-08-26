package proposalpredecessor

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

const maximumResponseBytes = 16 << 20

type githubRun struct {
	ID         int64  `json:"id"`
	RunAttempt int    `json:"run_attempt"`
	HeadSHA    string `json:"head_sha"`
	Event      string `json:"event"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	Name       string `json:"name"`
}

type runsEnvelope struct {
	TotalCount   int         `json:"total_count"`
	WorkflowRuns []githubRun `json:"workflow_runs"`
}

type githubArtifact struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	Expired            bool   `json:"expired"`
	ArchiveDownloadURL string `json:"archive_download_url"`
}

type artifactsEnvelope struct {
	TotalCount int              `json:"total_count"`
	Artifacts  []githubArtifact `json:"artifacts"`
}

func getBytes(ctx context.Context, client *http.Client, targetURL, token string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("GitHub response status %d", response.StatusCode)
	}
	payload, err := io.ReadAll(io.LimitReader(response.Body, maximumResponseBytes+1))
	if err != nil || len(payload) > maximumResponseBytes {
		return nil, fmt.Errorf("GitHub response exceeds fixed bound: %w", err)
	}
	return payload, nil
}
