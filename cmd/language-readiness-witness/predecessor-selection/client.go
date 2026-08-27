package main

import (
	"bytes"
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
	token   string
	client  *http.Client
}

func newGitHubClient(baseURL, token string) *githubClient {
	return &githubClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (client *githubClient) getJSON(ctx context.Context, endpoint string, output any) error {
	raw, err := client.get(ctx, endpoint)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(output); err != nil {
		return fmt.Errorf("decode GitHub response: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("GitHub response has trailing JSON")
		}
		return fmt.Errorf("GitHub response has trailing bytes: %w", err)
	}
	return nil
}

type githubPage struct {
	URL        string
	StatusCode int
	Body       []byte
	Link       string
}

func (client *githubClient) getPage(ctx context.Context, endpoint string) (githubPage, error) {
	target := endpoint
	if strings.HasPrefix(target, "/") {
		target = client.baseURL + target
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return githubPage{URL: target}, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+client.token)
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	response, err := client.client.Do(request)
	if err != nil {
		return githubPage{URL: target}, fmt.Errorf("GitHub request: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 16<<20))
	if err != nil {
		return githubPage{URL: target, StatusCode: response.StatusCode, Body: body, Link: response.Header.Get("Link")}, err
	}
	if response.StatusCode != http.StatusOK {
		return githubPage{URL: target, StatusCode: response.StatusCode, Body: body, Link: response.Header.Get("Link")}, fmt.Errorf("GitHub response status %d", response.StatusCode)
	}
	return githubPage{URL: target, StatusCode: response.StatusCode, Body: body, Link: response.Header.Get("Link")}, nil
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
