package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type githubClient struct {
	baseURL      string
	token        string
	client       *http.Client
	observations *observationStore
}

func newGitHubClient(baseURL, token string) *githubClient {
	result := &githubClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
	result.client.CheckRedirect = func(request *http.Request, _ []*http.Request) error {
		base, err := url.Parse(result.baseURL)
		if err != nil || base == nil || base.Scheme == "" || base.Host == "" || base.User != nil || base.Fragment != "" ||
			request == nil || request.URL == nil || request.URL.User != nil || request.URL.Fragment != "" ||
			!strings.EqualFold(request.URL.Scheme, base.Scheme) || !strings.EqualFold(request.URL.Host, base.Host) {
			return pageRedirectFailure{}
		}
		return nil
	}
	return result
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

type pageRedirectFailure struct{}

func (pageRedirectFailure) Error() string { return "REDIRECT_ORIGIN_MISMATCH" }

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
	if _, ok := errors.AsType[pageRedirectFailure](err); ok {
		return predecessorReason(endpointClass, "REDIRECT_ORIGIN_MISMATCH")
	}
	var failure pageURLFailure
	if !errors.As(err, &failure) {
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
	if client.observations != nil && client.observations.replay {
		observed, err := client.observations.next("PAGE", target)
		if err != nil {
			return githubPage{URL: target}, err
		}
		page := githubPage{URL: observed.URL, StatusCode: observed.StatusCode,
			Body: append([]byte(nil), observed.Body...), Link: observed.Link}
		if observed.Failure != "" {
			return page, replayHTTPFailure(observed)
		}
		return page, nil
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
		client.recordHTTPObservation("PAGE", target, 0, nil, "", err)
		return githubPage{URL: target}, fmt.Errorf("GitHub request: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 16<<20))
	if err != nil {
		client.recordHTTPObservation("PAGE", target, response.StatusCode, body,
			response.Header.Get("Link"), err)
		return githubPage{URL: target, StatusCode: response.StatusCode, Body: body, Link: response.Header.Get("Link")}, err
	}
	if response.StatusCode != http.StatusOK {
		client.recordHTTPObservation("PAGE", target, response.StatusCode, body,
			response.Header.Get("Link"), fmt.Errorf("status"))
		return githubPage{URL: target, StatusCode: response.StatusCode, Body: body, Link: response.Header.Get("Link")}, fmt.Errorf("GitHub response status %d", response.StatusCode)
	}
	client.recordHTTPObservation("PAGE", target, response.StatusCode, body,
		response.Header.Get("Link"), nil)
	return githubPage{URL: target, StatusCode: response.StatusCode, Body: body, Link: response.Header.Get("Link")}, nil
}

func (client *githubClient) get(ctx context.Context, endpoint string) ([]byte, error) {
	target := client.baseURL + endpoint
	if client.observations != nil && client.observations.replay {
		observed, err := client.observations.next("GET", target)
		if err != nil {
			return nil, err
		}
		if observed.Failure != "" {
			return nil, replayHTTPFailure(observed)
		}
		return append([]byte(nil), observed.Body...), nil
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+client.token)
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	response, err := client.client.Do(request)
	if err != nil {
		client.recordHTTPObservation("GET", target, 0, nil, "", err)
		return nil, fmt.Errorf("GitHub request: %w", err)
	}
	defer response.Body.Close()
	body, readErr := io.ReadAll(io.LimitReader(response.Body, 16<<20))
	if readErr != nil {
		client.recordHTTPObservation("GET", target, response.StatusCode, body, "", readErr)
		return nil, readErr
	}
	if response.StatusCode != http.StatusOK {
		client.recordHTTPObservation("GET", target, response.StatusCode, body, "",
			fmt.Errorf("status"))
		return nil, fmt.Errorf("GitHub response status %d", response.StatusCode)
	}
	client.recordHTTPObservation("GET", target, response.StatusCode, body, "", nil)
	return body, nil
}

func (client *githubClient) recordHTTPObservation(kind, target string, status int,
	body []byte, link string, err error) {
	if client.observations == nil || client.observations.replay {
		return
	}
	failure := ""
	if err != nil {
		if _, ok := errors.AsType[pageRedirectFailure](err); ok {
			failure = "REDIRECT_ORIGIN_MISMATCH"
		} else if status != http.StatusOK {
			failure = "HTTP_STATUS"
		} else {
			failure = "HTTP_FAILURE"
		}
	}
	client.observations.record(observedResponse{Kind: kind, URL: target,
		StatusCode: status, Body: append([]byte(nil), body...), Link: link,
		Failure: failure})
}

func replayHTTPFailure(observed observedResponse) error {
	if observed.Failure == "REDIRECT_ORIGIN_MISMATCH" {
		return pageRedirectFailure{}
	}
	if observed.Failure == "HTTP_STATUS" {
		return fmt.Errorf("GitHub response status %d", observed.StatusCode)
	}
	return fmt.Errorf("GitHub request replay failed")
}
