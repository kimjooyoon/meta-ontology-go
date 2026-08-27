package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestCollectWorkflowRunsPaginates(t *testing.T) {
	const predecessor = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/repos/owner/repo/actions/workflows/transformation-effect.yml/runs" {
			t.Fatalf("unexpected path %q", request.URL.Path)
		}
		if request.URL.Query().Get("per_page") != strconv.Itoa(workflowRunPageSize) {
			t.Fatalf("unexpected page size %q", request.URL.Query().Get("per_page"))
		}
		page, err := strconv.Atoi(request.URL.Query().Get("page"))
		if err != nil {
			t.Fatalf("decode page: %v", err)
		}
		var runs []workflowRun
		switch page {
		case 1:
			runs = []workflowRun{{ID: 1, HeadSHA: predecessor, Status: "completed"}}
		case 2:
			runs = []workflowRun{{ID: 2, HeadSHA: predecessor, Status: "completed"}}
		}
		response.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(response).Encode(workflowRunList{TotalCount: 2, WorkflowRuns: runs}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client := newGitHubClient(server.URL, "token")
	runs, err := collectWorkflowRuns(context.Background(), client, config{repository: "owner/repo", branch: "dev"}, predecessor)
	if err != nil {
		t.Fatalf("collect workflow runs: %v", err)
	}
	if len(runs) != 2 || runs[0].ID != 1 || runs[1].ID != 2 {
		t.Fatalf("unexpected runs: %#v", runs)
	}
}
