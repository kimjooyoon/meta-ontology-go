package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	maxGetAttempts       = 3
	maxResponseBodyBytes = 16 << 20
)

type githubClient struct {
	baseURL string
	token   string
	client  *http.Client
}

func newGitHubClient(baseURL, token string) *githubClient {
	return &githubClient{baseURL: strings.TrimRight(baseURL, "/"), token: token,
		client: &http.Client{Timeout: 30 * time.Second}}
}

func (client *githubClient) getJSON(ctx context.Context, endpoint string, output any) error {
	data, err := client.get(ctx, endpoint)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, output); err != nil {
		return fmt.Errorf("decode GitHub response: %w", err)
	}
	return nil
}

func (client *githubClient) get(ctx context.Context, endpoint string) ([]byte, error) {
	requestURL := client.baseURL + endpoint
	statuses := make([]int, 0, maxGetAttempts)
	for attempt := 1; attempt <= maxGetAttempts; attempt++ {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
		if err != nil {
			return nil, fmt.Errorf("GitHub GET request construction failed after %d attempt(s); statuses [%s]", attempt-1, formatStatuses(statuses))
		}
		request.Header.Set("Accept", "application/vnd.github+json")
		request.Header.Set("Authorization", "Bearer "+client.token)
		request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
		response, err := client.client.Do(request)
		if err != nil {
			if contextErr := ctx.Err(); contextErr != nil {
				return nil, fmt.Errorf("GitHub GET request failed without an HTTP status after %d attempt(s); prior statuses [%s]: %w", attempt, formatStatuses(statuses), contextErr)
			}
			return nil, fmt.Errorf("GitHub GET request failed without an HTTP status after %d attempt(s); prior statuses [%s]", attempt, formatStatuses(statuses))
		}
		if response.StatusCode != http.StatusOK {
			status := response.StatusCode
			statuses = append(statuses, status)
			_ = response.Body.Close()
			if retryableGitHubStatus(status) && attempt < maxGetAttempts {
				log.Printf("GitHub GET received transient HTTP status %d; retrying attempt %d/%d", status, attempt+1, maxGetAttempts)
				if err := waitBeforeRetry(ctx, retryDelay(attempt)); err != nil {
					return nil, fmt.Errorf("GitHub response status %d after attempts [%s]; retry cancelled: %w", status, formatStatuses(statuses), err)
				}
				continue
			}
			return nil, fmt.Errorf("GitHub response status %d after %d attempt(s); statuses [%s]", status, len(statuses), formatStatuses(statuses))
		}

		data, readErr := io.ReadAll(io.LimitReader(response.Body, maxResponseBodyBytes+1))
		_ = response.Body.Close()
		if readErr != nil {
			if contextErr := ctx.Err(); contextErr != nil {
				return nil, fmt.Errorf("GitHub response body read failed after %d attempt(s); prior statuses [%s]: %w", attempt, formatStatuses(statuses), contextErr)
			}
			return nil, fmt.Errorf("GitHub response body read failed after %d attempt(s)", attempt)
		}
		if len(data) > maxResponseBodyBytes {
			return nil, fmt.Errorf("GitHub response body exceeds %d MiB limit after %d attempt(s)", maxResponseBodyBytes>>20, attempt)
		}
		return data, nil
	}
	return nil, fmt.Errorf("GitHub GET exhausted %d attempts; statuses [%s]", maxGetAttempts, formatStatuses(statuses))
}

func retryableGitHubStatus(status int) bool {
	return status == http.StatusBadGateway || status == http.StatusServiceUnavailable || status == http.StatusGatewayTimeout
}

func retryDelay(completedAttempt int) time.Duration {
	if completedAttempt <= 1 {
		return 50 * time.Millisecond
	}
	return 100 * time.Millisecond
}

func waitBeforeRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func formatStatuses(statuses []int) string {
	parts := make([]string, len(statuses))
	for index, status := range statuses {
		parts[index] = fmt.Sprint(status)
	}
	return strings.Join(parts, ",")
}
