package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

type pageURLFailure string

const (
	pageURLMalformed      pageURLFailure = "LINK_MALFORMED"
	pageURLOriginMismatch pageURLFailure = "LINK_ORIGIN_MISMATCH"
)

func (failure pageURLFailure) Error() string { return string(failure) }

func (client *githubClient) resolvePageURL(reference, endpoint string) (string, error) {
	base, err := url.Parse(client.baseURL)
	if err != nil || base == nil || base.Scheme == "" || base.Host == "" || base.User != nil || base.Fragment != "" {
		return "", pageURLFailure(pageURLMalformed)
	}
	target, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || target == nil {
		return "", pageURLFailure(pageURLMalformed)
	}
	if reference == "" {
		reference = base.String()
	}
	ref, err := url.Parse(reference)
	if err != nil || ref == nil || ref.Scheme == "" || ref.Host == "" {
		return "", pageURLFailure(pageURLMalformed)
	}
	resolved := ref.ResolveReference(target)
	if resolved == nil || resolved.Scheme == "" || resolved.Host == "" {
		return "", pageURLFailure(pageURLMalformed)
	}
	if resolved.User != nil || resolved.Fragment != "" ||
		!strings.EqualFold(resolved.Scheme, base.Scheme) || !strings.EqualFold(resolved.Host, base.Host) {
		return "", pageURLFailure(pageURLOriginMismatch)
	}
	return resolved.String(), nil
}

func pageURLFailureReason(err error, endpointClass string) string {
	failure, ok := err.(pageURLFailure)
	if !ok {
		return ""
	}
	switch failure {
	case pageURLMalformed:
		return predecessorReason(endpointClass, "LINK_MALFORMED")
	case pageURLOriginMismatch:
		return predecessorReason(endpointClass, "LINK_ORIGIN_MISMATCH")
	default:
		return ""
	}
}

func (client *githubClient) getPage(ctx context.Context, endpoint string) (githubPage, error) {
	target, resolveErr := client.resolvePageURL("", endpoint)
	if resolveErr != nil {
		return githubPage{URL: endpoint}, resolveErr
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
