package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorselection"
)

const paginationPageCap = 1000

func pageObservation(page githubPage, endpointClass string, pageNumber int) predecessorselection.PaginationPage {
	return predecessorselection.PaginationPage{
		EndpointClass:  endpointClass,
		URL:            page.URL,
		PageNumber:     pageNumber,
		HTTPStatus:     page.StatusCode,
		BodyDigest:     digestBytes(page.Body),
		BodyBytes:      len(page.Body),
		NextLinkDigest: digestBytes([]byte(page.Link)),
	}
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func parseNextLink(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	var next string
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if !strings.HasPrefix(item, "<") {
			return "", fmt.Errorf("Link entry has no opening URL delimiter")
		}
		end := strings.IndexByte(item, '>')
		if end <= 1 {
			return "", fmt.Errorf("Link entry has no URL")
		}
		urlValue := item[1:end]
		params := strings.Split(item[end+1:], ";")
		foundRel := false
		isNext := false
		for _, param := range params {
			param = strings.TrimSpace(param)
			if param == "" {
				continue
			}
			parts := strings.SplitN(param, "=", 2)
			if len(parts) != 2 || strings.TrimSpace(parts[0]) != "rel" {
				continue
			}
			foundRel = true
			relation := strings.Trim(strings.TrimSpace(parts[1]), "\"")
			if relation == "next" {
				isNext = true
			}
		}
		if !foundRel {
			return "", fmt.Errorf("Link entry has no rel")
		}
		if isNext {
			if next != "" {
				return "", fmt.Errorf("Link header has multiple next links")
			}
			next = urlValue
		}
	}
	return next, nil
}

func decodeJSONEOF(raw []byte, output any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(output); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("response has trailing JSON")
		}
		return fmt.Errorf("response has trailing bytes: %w", err)
	}
	return nil
}

func collectWorkflowRuns(ctx context.Context, client *githubClient, endpoint string) ([]workflowRun, []predecessorselection.PaginationPage, string) {
	return collectPaged(ctx, client, endpoint, "WORKFLOW_RUNS", decodeWorkflowRunPage, workflowRunID)
}

func collectArtifacts(ctx context.Context, client *githubClient, endpoint string) ([]artifactMetadata, []predecessorselection.PaginationPage, string) {
	return collectPaged(ctx, client, endpoint, "ARTIFACTS", decodeArtifactPage, artifactID)
}

func collectJobs(ctx context.Context, client *githubClient, endpoint string) ([]workflowJob, []predecessorselection.PaginationPage, string) {
	return collectPaged(ctx, client, endpoint, "JOBS", decodeJobPage, workflowJobID)
}

func collectPaged[T any](ctx context.Context, client *githubClient, endpoint, endpointClass string, decode func([]byte) ([]T, int, error), identity func(T) int64) ([]T, []predecessorselection.PaginationPage, string) {
	var values []T
	pages := make([]predecessorselection.PaginationPage, 0)
	seenURLs := map[string]struct{}{}
	seenIDs := map[int64]struct{}{}
	next := endpoint
	for pageNumber := 1; ; pageNumber++ {
		if pageNumber > paginationPageCap {
			return values, pages, predecessorReason(endpointClass, "PAGE_CAP_EXCEEDED")
		}
		page, err := client.getPage(ctx, next)
		pages = append(pages, pageObservation(page, endpointClass, pageNumber))
		if err != nil {
			return values, pages, predecessorReason(endpointClass, "HTTP_FAILURE")
		}
		if _, exists := seenURLs[page.URL]; exists {
			return values, pages, predecessorReason(endpointClass, "NEXT_LINK_REPEATED")
		}
		seenURLs[page.URL] = struct{}{}
		pageValues, total, err := decode(page.Body)
		if err != nil {
			return values, pages, predecessorReason(endpointClass, "RESPONSE_MALFORMED")
		}
		for _, value := range pageValues {
			id := identity(value)
			if id <= 0 {
				return values, pages, predecessorReason(endpointClass, "RESPONSE_MALFORMED")
			}
			if _, exists := seenIDs[id]; exists {
				return values, pages, predecessorReason(endpointClass, "DUPLICATE_ID")
			}
			seenIDs[id] = struct{}{}
			values = append(values, value)
		}
		link, err := parseNextLink(page.Link)
		if err != nil {
			return values, pages, predecessorReason(endpointClass, "LINK_MALFORMED")
		}
		if link == "" {
			if total != len(values) {
				return values, pages, predecessorReason(endpointClass, "PAGINATION_INCOMPLETE")
			}
			return values, pages, ""
		}
		if pageNumber == paginationPageCap {
			return values, pages, predecessorReason(endpointClass, "PAGE_CAP_EXCEEDED")
		}
		if _, exists := seenURLs[link]; exists {
			return values, pages, predecessorReason(endpointClass, "NEXT_LINK_REPEATED")
		}
		next = resolveNextURL(page.URL, link)
	}
}

func resolveNextURL(base, next string) string {
	baseURL, baseErr := url.Parse(base)
	nextURL, nextErr := url.Parse(next)
	if baseErr != nil || nextErr != nil || baseURL == nil || nextURL == nil {
		return next
	}
	return baseURL.ResolveReference(nextURL).String()
}

func predecessorReason(endpointClass, suffix string) string {
	prefix := map[string]string{"WORKFLOW_RUNS": "WORKFLOW_RUN", "ARTIFACTS": "ARTIFACT", "JOBS": "JOB"}[endpointClass]
	if prefix == "" {
		prefix = "PAGINATION"
	}
	return prefix + "_" + suffix
}

func decodeWorkflowRunPage(raw []byte) ([]workflowRun, int, error) {
	var page workflowRunList
	if err := decodeJSONEOF(raw, &page); err != nil || page.TotalCount < 0 || page.WorkflowRuns == nil {
		return nil, 0, fmt.Errorf("malformed workflow run page")
	}
	return page.WorkflowRuns, page.TotalCount, nil
}

func decodeArtifactPage(raw []byte) ([]artifactMetadata, int, error) {
	var page artifactList
	if err := decodeJSONEOF(raw, &page); err != nil || page.TotalCount < 0 || page.Artifacts == nil {
		return nil, 0, fmt.Errorf("malformed artifact page")
	}
	return page.Artifacts, page.TotalCount, nil
}

func decodeJobPage(raw []byte) ([]workflowJob, int, error) {
	var page workflowJobList
	if err := decodeJSONEOF(raw, &page); err != nil || page.TotalCount < 0 || page.Jobs == nil {
		return nil, 0, fmt.Errorf("malformed job page")
	}
	return page.Jobs, page.TotalCount, nil
}

func workflowRunID(value workflowRun) int64   { return value.ID }
func artifactID(value artifactMetadata) int64 { return value.ID }
func workflowJobID(value workflowJob) int64   { return value.ID }
