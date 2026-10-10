package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/outcomedelta"
)

func TestBodyOutcomesMarkdownCLI(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "saved.json")
	input := `{"schema":"gooo/body-composition-runtime/v1","stage":"COMPLETE","finite_passed":1,"finite_total":1,"traces":[{"case_index":0,"deliveries":[{"activity_id":"test://main","input":9007199254740993,"actual":16,"expected":16,"passed":true}]}]}`
	if err := os.WriteFile(path, []byte(input), 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"--before", path, "--after", path}
	var out, stderr bytes.Buffer
	if code := runBodyOutcomesDelta(append(args, "--markdown"), &out, &stderr); code != exitOK || stderr.Len() != 0 {
		t.Fatalf("%d %s", code, &stderr)
	}
	for _, want := range []string{"# Workflow outcome comparison", "9007199254740993", "STILL&#95;MATCHING", "0/1 (0.00%)", "Before input SHA256", "No model calls or program executions"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q in %s", want, &out)
		}
	}
	for _, flags := range [][]string{{"--markdown", "--markdown"}, {"--json", "--markdown"}, {"--markdown", "--json"}} {
		out.Reset()
		stderr.Reset()
		if code := runBodyOutcomesDelta(append(args, flags...), &out, &stderr); code != exitUsage || out.Len() != 0 {
			t.Fatalf("invalid flags accepted: %v", flags)
		}
	}
}

func TestOutcomeMarkdownDenominatorsAndEscaping(t *testing.T) {
	statuses := []string{"REGRESSION", "STILL_MATCHING", "IMPROVEMENT", "STILL_FAILING", "REQUIREMENT_CHANGED", "UNOBSERVED", "AMBIGUOUS"}
	r := &outcomedelta.OutcomeDelta{Counts: map[string]int{}}
	for _, status := range statuses {
		r.Rows = append(r.Rows, outcomedelta.OutcomeChange{ActivityID: "test://<b>|`x`\nnext", Assessment: status})
		r.Counts["assessment:"+status]++
	}
	var out bytes.Buffer
	if err := writeOutcomeMarkdown(&out, r); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"4/7 (57.14%)", "1/2 (50.00%)", "1/7 (14.29%)", "&lt;b&gt;&#124;", "&#10;next", "(absent)"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q in %s", want, &out)
		}
	}
	if strings.Contains(out.String(), "<b>") || strings.Contains(out.String(), "\nnext") {
		t.Fatal("input changed Markdown structure")
	}
	if strings.Count(out.String(), "| 1/2 (50.00%) |") != 2 {
		t.Fatal("regressions and improvements need their own eligible populations")
	}
}

func TestOutcomeMarkdownSmallNonzeroShare(t *testing.T) {
	r := &outcomedelta.OutcomeDelta{Rows: make([]outcomedelta.OutcomeChange, 8192), Counts: map[string]int{
		"assessment:REGRESSION": 1, "assessment:STILL_MATCHING": 8191,
	}}
	var out bytes.Buffer
	writeOutcomeMarkdownMetrics(&out, r)
	if !strings.Contains(out.String(), "1/8192 (0.01%)") {
		t.Fatal("rounded an observed regression down to zero:", &out)
	}
}

func TestOutcomeMarkdownValuePresenceAndSummary(t *testing.T) {
	r := &outcomedelta.OutcomeDelta{Rows: []outcomedelta.OutcomeChange{{
		Before: []map[string]any{{"actual": nil}},
		After:  []map[string]any{{"expected": strings.Repeat("한", 130)}},
	}}}
	var out bytes.Buffer
	if err := writeOutcomeMarkdown(&out, r); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"<code>null</code>", "(unobserved or unavailable)", "… (full value in --json)"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing value distinction %q", want)
		}
	}
}

func TestOutcomeMarkdownEmptyAndWriteFailure(t *testing.T) {
	var out bytes.Buffer
	if err := writeOutcomeMarkdown(&out, &outcomedelta.OutcomeDelta{}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "%") || !strings.Contains(out.String(), "n/a (no eligible groups)") || !strings.Contains(out.String(), "No activity/input groups were observed.") {
		t.Fatal(out.String())
	}
	if err := writeOutcomeMarkdown(outcomeClosedWriter{}, &outcomedelta.OutcomeDelta{}); err != io.ErrClosedPipe {
		t.Fatalf("lost output error: %v", err)
	}
}

type outcomeClosedWriter struct{}

func (outcomeClosedWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }
