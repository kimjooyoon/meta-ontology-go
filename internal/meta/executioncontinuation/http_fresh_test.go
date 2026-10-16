package executioncontinuation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/candidatedelivery"
)

type transport1617 func(*http.Request) (*http.Response, error)

func (f transport1617) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func client1617(t *testing.T, f fixture1617, variant string) Client {
	t.Helper()
	tr := transport1617(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "api.github.com" || r.Header.Get("X-GitHub-Api-Version") != "2026-03-10" {
			t.Fatal("unexpected protocol origin/version")
		}
		status := 200
		var raw []byte
		if variant == "api-unavailable" {
			status = 503
		} else if r.Method == "POST" {
			var payload struct {
				Ref    string            `json:"ref"`
				Inputs map[string]string `json:"inputs"`
			}
			if json.NewDecoder(r.Body).Decode(&payload) != nil || payload.Ref != "dev" || payload.Inputs["grant_request_digest"] != f.prepared.Metadata.Source.RequestDigest {
				t.Fatal("dispatch payload changed intent")
			}
			response := map[string]any{"workflow_run_id": f.consumerRun.ID}
			if variant == "dispatch-no-reference" {
				response = map[string]any{}
			}
			raw, _ = json.Marshal(response)
		} else {
			raw = response1617(t, f, r.URL.Path, variant)
		}
		return &http.Response{StatusCode: status, Body: io.NopCloser(bytes.NewReader(raw)), Header: http.Header{}, Request: r}, nil
	})
	return Client{Repository: f.prepared.Metadata.Source.Repository, Token: "synthetic-token", Root: os.DirFS("../../.."), HTTP: &http.Client{Transport: tr}}
}

func response1617(t *testing.T, f fixture1617, path, variant string) []byte {
	t.Helper()
	var value any
	switch {
	case path == fmt.Sprintf("/repos/%s/actions/runs/%d", f.prepared.Metadata.Source.Repository, f.grantRun.ID):
		run := f.grantRun
		if variant == "wrong-attempt" {
			run.Attempt = 2
		}
		value = run
	case strings.HasSuffix(path, fmt.Sprintf("/actions/runs/%d", f.contractRun.ID)):
		value = f.contractRun
	case strings.HasSuffix(path, fmt.Sprintf("/actions/runs/%d", f.consumerRun.ID)):
		run := f.consumerRun
		if variant == "wrong-target" {
			run.Path = "wrong.yml"
		}
		value = run
	case strings.Contains(path, "/attempts/1/jobs"):
		value = map[string]any{"total_count": 1, "jobs": []any{map[string]string{"name": "execution-grant", "status": "completed", "conclusion": "success"}}}
	case strings.HasSuffix(path, fmt.Sprintf("/runs/%d/artifacts", f.grantRun.ID)):
		artifacts := []candidatedelivery.Artifact{f.grantArtifact}
		if variant == "missing-grant" {
			artifacts = []candidatedelivery.Artifact{}
		}
		if variant == "duplicate-grant" {
			artifacts = append(artifacts, f.grantArtifact)
		}
		if variant == "expired-grant" {
			artifacts[0].Expired = true
		}
		value = map[string]any{"total_count": len(artifacts), "artifacts": artifacts}
	case strings.HasSuffix(path, fmt.Sprintf("/runs/%d/artifacts", f.contractRun.ID)):
		value = map[string]any{"total_count": 1, "artifacts": []candidatedelivery.Artifact{f.contractArtifact}}
	case strings.HasSuffix(path, fmt.Sprintf("/artifacts/%d/zip", f.grantArtifact.ID)):
		raw := append([]byte(nil), f.grantArchive...)
		if variant == "changed-archive" {
			raw[len(raw)/2] ^= 1
		}
		return raw
	case strings.HasSuffix(path, fmt.Sprintf("/artifacts/%d/zip", f.contractArtifact.ID)):
		return f.contractArchive
	case strings.Contains(path, "/contents/"):
		raw := []byte(f.prepared.SourceBytes)
		value = map[string]any{"content": base64.StdEncoding.EncodeToString(raw), "encoding": "base64", "sha": f.prepared.Metadata.SubjectBlobSHA, "size": len(raw)}
	case strings.HasSuffix(path, "/self-improvement-execution-consumer.yml/runs"):
		value = map[string]any{"total_count": 0, "workflow_runs": []Run{}}
	default:
		t.Fatal("unexpected protocol path", path)
	}
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
