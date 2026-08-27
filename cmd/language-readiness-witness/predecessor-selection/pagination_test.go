package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
)

type paginationFixtureFile struct {
	Schema      string              `json:"schema"`
	SourceClass string              `json:"source_class"`
	ExpectedIDs []string            `json:"expected_ids"`
	Cases       []paginationFixture `json:"cases"`
}

type paginationFixture struct {
	ID             string        `json:"id"`
	ExpectedReason string        `json:"expected_reason"`
	Pages          []fixturePage `json:"pages"`
}

type fixturePage struct {
	Path   string `json:"path"`
	Status int    `json:"status"`
	Body   string `json:"body"`
	Link   string `json:"link"`
}

func readPaginationFixtures(t *testing.T) paginationFixtureFile {
	t.Helper()
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	fixturePath := filepath.Join(filepath.Dir(sourceFile), "../../../examples/causal-ci-selection/pagination-fixtures.json")
	raw, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	var fixtures paginationFixtureFile
	if err := decoder.Decode(&fixtures); err != nil {
		t.Fatal(err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		t.Fatalf("fixture has trailing data: %v", err)
	}
	if fixtures.Schema != "gooo/causal-ci-selection-pagination-fixtures/v1" || fixtures.SourceClass != "SYNTHETIC_FIXTURE" {
		t.Fatalf("fixture identity = %#v", fixtures)
	}
	return fixtures
}

func TestPaginationFixturesExecuteParserAndHTTPClient(t *testing.T) {
	fixtures := readPaginationFixtures(t)
	expectedIDs := []string{
		"normal-multi-page", "inconsistent-total", "repeated-next-link", "page-cap-exceeded",
		"malformed-url", "malformed-header", "duplicate-run-id", "last-page",
	}
	if !reflect.DeepEqual(fixtures.ExpectedIDs, expectedIDs) {
		t.Fatalf("fixture denominator = %v, want %v", fixtures.ExpectedIDs, expectedIDs)
	}
	if len(fixtures.Cases) != len(expectedIDs) {
		t.Fatalf("fixture cases = %d, want %d", len(fixtures.Cases), len(expectedIDs))
	}
	seen := map[string]bool{}
	for _, fixture := range fixtures.Cases {
		if seen[fixture.ID] {
			t.Fatalf("duplicate fixture ID %q", fixture.ID)
		}
		seen[fixture.ID] = true
	}
	for _, id := range expectedIDs {
		if !seen[id] {
			t.Fatalf("missing fixture ID %q", id)
		}
	}

	for _, fixture := range fixtures.Cases {
		fixture := fixture
		t.Run(fixture.ID, func(t *testing.T) {
			var requestCount atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				requestCount.Add(1)
				path := request.URL.RequestURI()
				if fixture.ID == "page-cap-exceeded" {
					pageNumber, err := strconv.Atoi(request.URL.Query().Get("page"))
					if err != nil || pageNumber < 1 {
						http.Error(writer, "bad page", http.StatusBadRequest)
						return
					}
					body := fixture.Pages[0].Body
					link := fixture.Pages[0].Link
					if pageNumber != 1 {
						body = fmt.Sprintf("{\"total_count\":1001,\"workflow_runs\":[{\"id\":%d}]}", pageNumber)
						link = fmt.Sprintf("</runs?page=%d>; rel=\"next\"", pageNumber+1)
					}
					writer.Header().Set("Link", link)
					writer.WriteHeader(fixture.Pages[0].Status)
					_, _ = writer.Write([]byte(body))
					return
				}
				for _, page := range fixture.Pages {
					if page.Path == path {
						writer.Header().Set("Link", page.Link)
						writer.WriteHeader(page.Status)
						_, _ = writer.Write([]byte(page.Body))
						return
					}
				}
				http.NotFound(writer, request)
			}))
			defer server.Close()

			client := newGitHubClient(server.URL, "fixture-secret")
			values, pages, reason := collectWorkflowRuns(context.Background(), client, server.URL+"/runs?page=1")
			if reason != fixture.ExpectedReason {
				t.Fatalf("reason = %q, want %q", reason, fixture.ExpectedReason)
			}
			if fixture.ID == "page-cap-exceeded" {
				if len(pages) != paginationPageCap || len(values) != paginationPageCap {
					t.Fatalf("cap inventory values/pages = %d/%d, want %d/%d", len(values), len(pages), paginationPageCap, paginationPageCap)
				}
			} else if len(pages) != len(fixture.Pages) {
				t.Fatalf("page inventory = %d, want %d", len(pages), len(fixture.Pages))
			}
			for index, expected := range fixture.Pages {
				if index >= len(pages) {
					break
				}
				observedURL, err := url.Parse(pages[index].URL)
				if err != nil || observedURL.RequestURI() != expected.Path {
					t.Fatalf("page %d URL = %q, want path %q", index+1, pages[index].URL, expected.Path)
				}
				if pages[index].HTTPStatus != expected.Status || pages[index].BodyDigest != digestBytes([]byte(expected.Body)) || pages[index].BodyBytes != len([]byte(expected.Body)) || pages[index].NextLinkDigest != digestBytes([]byte(expected.Link)) {
					t.Fatalf("page %d raw inventory = %#v, expected body/link digests from fixture", index+1, pages[index])
				}
			}
			if requestCount.Load() == 0 {
				t.Fatal("fixture did not execute an HTTP request")
			}
		})
	}
}

func TestPaginationRejectsCrossOriginNextLinkBeforeRequest(t *testing.T) {
	var externalHits atomic.Int32
	external := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		externalHits.Add(1)
		if request.Header.Get("Authorization") != "" {
			t.Errorf("authorization reached cross-origin server")
		}
		writer.WriteHeader(http.StatusOK)
	}))
	defer external.Close()

	base := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Link", fmt.Sprintf("<%s/runs?page=2>; rel=\"next\"", external.URL))
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("{\"total_count\":2,\"workflow_runs\":[{\"id\":1}]}"))
	}))
	defer base.Close()

	_, pages, reason := collectWorkflowRuns(context.Background(), newGitHubClient(base.URL, "fixture-secret"), base.URL+"/runs?page=1")
	if reason != "WORKFLOW_RUN_LINK_ORIGIN_MISMATCH" {
		t.Fatalf("reason = %q, want origin mismatch", reason)
	}
	if len(pages) != 1 || externalHits.Load() != 0 {
		t.Fatalf("pages/external hits = %d/%d, want 1/0", len(pages), externalHits.Load())
	}
}

func TestPaginationRejectsOtherOriginCoordinates(t *testing.T) {
	cases := map[string]func(string) string{
		"scheme":   func(base string) string { return "https://" + strings.TrimPrefix(base, "http://") + "/runs?page=2" },
		"userinfo": func(base string) string { return strings.Replace(base, "http://", "http://user@", 1) + "/runs?page=2" },
		"fragment": func(base string) string { return base + "/runs?page=2#fragment" },
	}
	for name, link := range cases {
		name, link := name, link
		t.Run(name, func(t *testing.T) {
			var base *httptest.Server
			base = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Link", fmt.Sprintf("<%s>; rel=\"next\"", link(base.URL)))
				writer.WriteHeader(http.StatusOK)
				_, _ = writer.Write([]byte("{\"total_count\":2,\"workflow_runs\":[{\"id\":1}]}"))
			}))
			defer base.Close()

			_, pages, reason := collectWorkflowRuns(context.Background(), newGitHubClient(base.URL, "fixture-secret"), base.URL+"/runs?page=1")
			if reason != "WORKFLOW_RUN_LINK_ORIGIN_MISMATCH" || len(pages) != 1 {
				t.Fatalf("reason/pages = %q/%d, want origin mismatch/1", reason, len(pages))
			}
		})
	}
}

func TestPaginationFailureReasonAllowlist(t *testing.T) {
	suffixes := []string{"HTTP_FAILURE", "PAGINATION_INCOMPLETE", "NEXT_LINK_REPEATED", "PAGE_CAP_EXCEEDED", "LINK_MALFORMED", "LINK_ORIGIN_MISMATCH", "DUPLICATE_ID", "RESPONSE_MALFORMED"}
	for _, prefix := range []string{"WORKFLOW_RUN", "ARTIFACT", "JOB"} {
		for _, suffix := range suffixes {
			if !knownPaginationFailureReason(prefix + "_" + suffix) {
				t.Errorf("allowlist rejects %s_%s", prefix, suffix)
			}
		}
	}
	for _, reason := range []string{"WORKFLOW_RUN_ANYTHING", "ARTIFACT_HTTP_FAILURE_EXTRA", "arbitrary"} {
		if knownPaginationFailureReason(reason) {
			t.Errorf("allowlist accepts arbitrary reason %q", reason)
		}
	}
}
