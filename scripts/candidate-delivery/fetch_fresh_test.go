package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	delivery "github.com/kimjooyoon/meta-ontology-go/internal/meta/candidatedelivery"
)

type roundTrip1613 func(*http.Request) (*http.Response, error)

func (f roundTrip1613) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestFetchDiagnosticsFresh1613(t *testing.T) {
	for _, name := range []string{"missing-role", "duplicate-role", "wrong-attempt", "missing-members", "api-unavailable", "pagination-incomplete"} {
		t.Run(name, func(t *testing.T) {
			var raw bytes.Buffer
			z := zip.NewWriter(&raw)
			_ = z.Close()
			archive := raw.Bytes()
			previous := http.DefaultTransport
			defer func() { http.DefaultTransport = previous }()
			calls := 0
			http.DefaultTransport = roundTrip1613(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.URL.Host != "api.github.com" || r.Method != "GET" {
					t.Fatal("unexpected delivery request", r.URL.Host)
				}
				status, body := response1613(name, r.URL.Path, archive)
				return &http.Response{StatusCode: status, Body: io.NopCloser(bytes.NewReader(body)), Header: http.Header{}, Request: r}, nil
			})
			dir := filepath.Join(t.TempDir(), "fresh")
			err := run(context.Background(), options{mode: "fetch", output: dir, repository: "kimjooyoon/meta-ontology-go", run: 9161310, attempt: 2})
			var receipt delivery.Receipt
			body, readErr := os.ReadFile(filepath.Join(dir, "delivery.json"))
			if readErr != nil || json.Unmarshal(body, &receipt) != nil {
				t.Fatal("diagnostic missing", readErr)
			}
			want := "UNKNOWN"
			if name == "duplicate-role" || name == "wrong-attempt" {
				want = "REFUTED"
			}
			if receipt.State != want || (err != nil) != (want == "REFUTED") {
				t.Fatalf("%s: %+v / %v", name, receipt, err)
			}
			if _, err := os.Stat(filepath.Join(dir, "candidate.json")); !os.IsNotExist(err) {
				t.Fatal("unbound candidate exposed")
			}
			if calls > 9 {
				t.Fatalf("unbounded lookup: %d requests", calls)
			}
		})
	}
}

func response1613(name, path string, archive []byte) (int, []byte) {
	if name == "api-unavailable" {
		return 503, nil
	}
	if strings.HasSuffix(path, "/zip") {
		return 200, archive
	}
	head := strings.Repeat("b", 40)
	r := delivery.Run{ID: 9161310, Attempt: 2, Head: head, Event: "workflow_run", Path: ".github/workflows/self-improvement-language-candidate.yml", Status: "completed", Conclusion: "success"}
	r.Repository.Name = "kimjooyoon/meta-ontology-go"
	if name == "wrong-attempt" {
		r.Attempt = 3
	}
	if !strings.HasSuffix(path, "/artifacts") {
		body, _ := json.Marshal(r)
		return 200, body
	}
	a := delivery.Artifact{ID: 9161320, Name: "self-improvement-nonexecuting-candidate-" + strings.Repeat("1", 40), Digest: delivery.Digest(archive), Size: int64(len(archive))}
	a.Run.ID, a.Run.Head = r.ID, head
	artifacts, total := []delivery.Artifact{a}, 1
	switch name {
	case "missing-role":
		artifacts, total = []delivery.Artifact{}, 0
	case "duplicate-role":
		artifacts, total = append(artifacts, a), 2
	case "pagination-incomplete":
		artifacts, total = []delivery.Artifact{}, 901
	}
	body, _ := json.Marshal(map[string]any{"total_count": total, "artifacts": artifacts})
	return 200, body
}

func TestFetchRejectsStaleDirectoryFresh1613(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "candidate.json")
	if err := os.WriteFile(path, []byte("old candidate"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run(context.Background(), options{mode: "fetch", output: dir}); err == nil {
		t.Fatal("existing output reused")
	}
	raw, _ := os.ReadFile(path)
	if string(raw) != "old candidate" {
		t.Fatal("existing caller data changed")
	}
}

func TestFetchCancellationFresh1613(t *testing.T) {
	previous := http.DefaultTransport
	defer func() { http.DefaultTransport = previous }()
	http.DefaultTransport = roundTrip1613(func(r *http.Request) (*http.Response, error) { <-r.Context().Done(); return nil, r.Context().Err() })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	dir := filepath.Join(t.TempDir(), "fresh")
	if err := run(ctx, options{mode: "fetch", output: dir, repository: "kimjooyoon/meta-ontology-go", run: 9161310, attempt: 2}); err != nil {
		t.Fatal(err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("cancelled delivery did not return promptly")
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "delivery.json"))
	var receipt delivery.Receipt
	_ = json.Unmarshal(raw, &receipt)
	if receipt.State != "UNKNOWN" {
		t.Fatal("cancelled request lost UNKNOWN")
	}
}
