package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

type rejectNetwork struct{ calls int }

func (transport *rejectNetwork) RoundTrip(*http.Request) (*http.Response, error) {
	transport.calls++
	return nil, fmt.Errorf("saved-report query attempted HTTP")
}

func TestSavedReportCommandUsesLocalFilesAndStructuredResults(t *testing.T) {
	prior := filepath.Join("..", "..", "transport-origins", "testdata", "retained-readonly-report.json")
	raw, err := os.ReadFile(prior)
	if err != nil {
		t.Fatal(err)
	}
	trap, previous := &rejectNetwork{}, http.DefaultTransport
	http.DefaultTransport = trap
	t.Cleanup(func() { http.DefaultTransport = previous })
	for _, pair := range []bool{false, true} {
		args := []string{"-reassess", prior, "-material", filepath.Join(t.TempDir(), "unused")}
		if pair {
			args = append(args, "-read-report", prior)
		}
		var output, diagnostics bytes.Buffer
		if err := execute(args, &output, &diagnostics); err != nil {
			t.Fatal(err, diagnostics.String())
		}
		var result struct {
			Prior   string `json:"prior_report_sha256"`
			Current string `json:"current_report_sha256"`
			State   string `json:"state"`
		}
		if err := json.Unmarshal(output.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		digest := fmt.Sprintf("sha256:%x", sha256.Sum256(raw))
		if result.Prior != digest || result.State != "UNRESOLVED_SOURCE_PAIR" ||
			pair && result.Current != digest || !pair && result.Current != "" || trap.calls != 0 {
			t.Fatal("retained older source must remain unresolved without HTTP", result, trap.calls)
		}
	}
}

func TestSavedReportCommandRejectsConflictingModesAndKeepsHelp(t *testing.T) {
	trap, previous := &rejectNetwork{}, http.DefaultTransport
	http.DefaultTransport = trap
	t.Cleanup(func() { http.DefaultTransport = previous })
	for _, args := range [][]string{
		{"-read-report", "current.json"},
		{"-reassess", "prior.json", "-apply"},
		{"unexpected.json"},
		{"-reassess", "", "-material", "../testdata/manifest.json"},
		{"-read-report", "", "-material", "../testdata/manifest.json"},
		{"-reassess", "prior.json", "-read-report", ""},
	} {
		var output, diagnostics bytes.Buffer
		if err := execute(args, &output, &diagnostics); err == nil || output.Len() != 0 ||
			errors.Is(err, os.ErrNotExist) || trap.calls != 0 {
			t.Fatal("invalid command must stop before execution", args, err)
		}
	}
	var output, diagnostics bytes.Buffer
	if err := execute([]string{"-h"}, &output, &diagnostics); !errors.Is(err, flag.ErrHelp) ||
		!bytes.Contains(diagnostics.Bytes(), []byte("-reassess")) {
		t.Fatal("help is an ordinary command result", err, diagnostics.String())
	}
}

func TestSavedReportReaderRetainsOnlyTheBoundedPrefix(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.json")
	if err := os.WriteFile(path, []byte("abcdef"), 0600); err != nil {
		t.Fatal(err)
	}
	raw, err := readBoundedFile(path, 3)
	if err != nil || string(raw) != "abcd" {
		t.Fatal("keep one extra byte for the source bound check", err, string(raw))
	}
}
