package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSavedReportFixture(t *testing.T, raw []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "result.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestSavedConstructionReportFlags(t *testing.T) {
	flags, err := parseBodyConstruct([]string{"--report", "saved.json"})
	if err != nil || flags["--format"] != "text" {
		t.Fatal(flags, err)
	}
	for _, key := range []string{"--source", "--cases", "--construction-cases", "--attempts", "--entry",
		"--model", "--fill-model", "--construction", "--go-bin", "--out"} {
		if _, err := parseBodyConstruct([]string{"--report", "saved.json", key, "unused"}); err == nil {
			t.Fatal("report accepted execution option", key)
		}
	}
	for _, args := range [][]string{{"--report"}, {"--report", ""}, {"--report", "x", "--report", "y"},
		{"--report", "x", "--format", "other"}, {"--report", "x", "--format", "text", "--format", "json"}} {
		var out, diagnostic bytes.Buffer
		if code := runBodyConstruct(args, &out, &diagnostic); code != exitUsage || out.Len() != 0 {
			t.Fatal("malformed report arguments accepted", args, code)
		}
	}
}

func TestSavedConstructionReportCLIReadsWithoutToolchain(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	saved := constructionReportFixture()
	raw, err := json.MarshalIndent(saved, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	raw = append(raw, '\n')
	path := writeSavedReportFixture(t, raw)
	for _, format := range []string{"text", "markdown", "json"} {
		var out, diagnostic bytes.Buffer
		if code := run([]string{"body-construct", "--report", path, "--format", format}, &out, &diagnostic); code != exitOK || diagnostic.Len() != 0 {
			t.Fatal(code, diagnostic.String())
		}
		if format == "json" {
			if !bytes.Equal(raw, out.Bytes()) {
				t.Fatal("original JSON bytes changed")
			}
			continue
		}
		for _, want := range []string{fmt.Sprintf("%x", sha256.Sum256(raw)), "recorded invocation", "not been reverified",
			"0 model calls and 0 program executions", "PARTIAL_FINITE", "9007199254740995", "9007199254740993"} {
			if !strings.Contains(html.UnescapeString(out.String()), want) {
				t.Fatal(format, want, out.String())
			}
		}
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(after, raw) {
		t.Fatal("saved file changed", err)
	}
}

func TestSavedConstructionReportRejectsIncompleteEnvelope(t *testing.T) {
	for _, raw := range []string{`null`, `{}`, `{"construction":{}}`, `{"evaluation":{}}`,
		`{"generated_now":null,"construction":{},"evaluation":{}}`,
		`{"generated_now":false,"construction":null,"evaluation":{}}`,
		`{"generated_now":true,"construction":{},"evaluation":null}`,
		`{"generated_now":"true","construction":{},"evaluation":{}}`,
		`{"generated_now":true,"construction":{},"evaluation":{},"future":1}`,
		`{"generated_now":true,"construction":{"program_budget":"2"},"evaluation":{}}`,
		`{"generated_now":true,"construction":{},"evaluation":{}} {}`,
		`{"generated_now":true,"construction":{},"evaluation":{}} trailing`, `[`} {
		var out, diagnostic bytes.Buffer
		if code := runBodyConstruct([]string{"--report", writeSavedReportFixture(t, []byte(raw))}, &out, &diagnostic); code != exitFailure || out.Len() != 0 || diagnostic.Len() == 0 {
			t.Fatal("invalid saved observation accepted", raw, code, out.String())
		}
	}
}

func TestSavedConstructionReportKeepsRecordedFailure(t *testing.T) {
	r := constructionReportFixture()
	r.GeneratedNow, r.Construction.Stage, r.Construction.Failure = false, "FAILED", "original toolchain missing"
	r.Evaluation.Runtime.Stage, r.Evaluation.Runtime.Traces = "TOOLCHAIN", nil
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var out, diagnostic bytes.Buffer
	if code := runBodyConstruct([]string{"--report", writeSavedReportFixture(t, raw)}, &out, &diagnostic); code != exitOK {
		t.Fatal("reading failure observation must succeed", code, diagnostic.String())
	}
	for _, want := range []string{"saved construction replay", "original toolchain missing", "not measured", "recorded failures"} {
		if !strings.Contains(out.String(), want) {
			t.Fatal(want, out.String())
		}
	}
	if err := writeSavedConstructionReport(constructionFailWriter{}, raw, r, "text"); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal("writer error lost", err)
	}
	if err := writeSavedConstructionReport(constructionFailWriter{}, raw, r, "json"); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal("JSON writer error lost", err)
	}
}

// Read the once-recorded canonical observation; never repeat its experiment.
func TestSavedConstructionReportRetainedCanonicalObservation(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	z, err := gzip.NewReader(bytes.NewReader(canonicalReportObservation))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	var out, diagnostic bytes.Buffer
	if code := run([]string{"body-construct", "--report", writeSavedReportFixture(t, raw)}, &out, &diagnostic); code != exitOK {
		t.Fatal(code, diagnostic.String())
	}
	for _, want := range []string{"4 recorded candidates; 1 recorded local model calls", "1/3 (33.33%)",
		"conditions 3/3 (100.00%)", "0 model calls and 0 program executions"} {
		if !strings.Contains(out.String(), want) {
			t.Fatal(want, out.String())
		}
	}
}

func TestSavedConstructionReportReadAndWriteErrors(t *testing.T) {
	raw, _ := json.Marshal(constructionReportFixture())
	var diagnostic bytes.Buffer
	if code := runBodyConstruct([]string{"--report", writeSavedReportFixture(t, raw)}, constructionFailWriter{}, &diagnostic); code != exitFailure {
		t.Fatal("command lost writer error", code)
	}
	path := filepath.Join(t.TempDir(), "oversize.json")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Truncate((32 << 20) + 1); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{path, path + ".missing"} {
		var out, diagnostic bytes.Buffer
		if code := runBodyConstruct([]string{"--report", path}, &out, &diagnostic); code != exitFailure || out.Len() != 0 {
			t.Fatal("file read error lost", code, diagnostic.String())
		}
	}
}
