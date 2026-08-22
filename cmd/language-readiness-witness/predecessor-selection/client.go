package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type githubClient struct {
	baseURL string
	token string
	client *http.Client
}

func newGitHubClient(baseURL, token string) *githubClient {
	return &githubClient{baseURL: strings.TrimRight(baseURL, "/"), token: token,
		client: &http.Client{Timeout: 30 * time.Second}}
}

func (client *githubClient) getJSON(ctx context.Context, endpoint string, output any) error {
	raw, err := client.get(ctx, endpoint)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, output); err != nil {
		return fmt.Errorf("decode GitHub response: %w", err)
	}
	return nil
}

func (client *githubClient) get(ctx context.Context, endpoint string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, client.baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+client.token)
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	response, err := client.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("GitHub request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub response status %d", response.StatusCode)
	}
	return io.ReadAll(io.LimitReader(response.Body, 16<<20))
}
