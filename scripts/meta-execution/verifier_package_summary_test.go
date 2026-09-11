package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestParseVerifierPackageSummariesPreservesBoundedRows(t *testing.T) {
	stdout := strings.Join([]string{
		"ok   github.com/kimjooyoon/meta-ontology-go/internal/meta 0.123s",
		"ok   github.com/kimjooyoon/meta-ontology-go/internal/cache (cached)",
		"?    github.com/kimjooyoon/meta-ontology-go/examples/demo [no test files]",
		"FAIL github.com/kimjooyoon/meta-ontology-go/internal/broken 2.000s",
	}, "\n")

	parsed := parseVerifierPackageSummaries([]byte(stdout))
	if parsed.status() != "COMPLETE" || len(parsed.Rows) != 4 {
		t.Fatalf("parsed package summaries = %#v", parsed)
	}
	if parsed.Rows[0].ElapsedToken != "0.123s" || parsed.Rows[0].ElapsedStatus != "PARSED_EXACTLY" || parsed.Rows[0].ElapsedUnit != "NANOSECONDS" || parsed.Rows[0].ElapsedNanoseconds == nil || *parsed.Rows[0].ElapsedNanoseconds != 123000000 {
		t.Fatalf("exact elapsed row = %#v", parsed.Rows[0])
	}
	if parsed.Rows[1].OutputMarker != verifierOutputMarkerCached || parsed.Rows[1].OutputMarkerAuthority != verifierOutputMarkerAuthority || parsed.Rows[1].ElapsedStatus != "UNKNOWN" {
		t.Fatalf("cached marker row = %#v", parsed.Rows[1])
	}
	if parsed.Rows[2].Status != "?" || parsed.Rows[2].OutputMarker != verifierOutputMarkerNoTestFiles || parsed.Rows[2].OutputMarkerAuthority != verifierOutputMarkerAuthority {
		t.Fatalf("no-test-files row = %#v", parsed.Rows[2])
	}
	if parsed.Rows[3].ElapsedToken != "2.000s" || parsed.Rows[3].ElapsedNanoseconds == nil || *parsed.Rows[3].ElapsedNanoseconds != 2000000000 {
		t.Fatalf("failure duration row = %#v", parsed.Rows[3])
	}
}

func TestParseVerifierPackageSummariesRejectsModulePrefixSpoofing(t *testing.T) {
	stdout := strings.Join([]string{
		"ok github.com/kimjooyoon/meta-ontology-go.evil/pkg 1s",
		"ok github.com/kimjooyoon/meta-ontology-go/../evil 1s",
		"ok github.com/kimjooyoon/meta-ontology-go/internal/valid 1s",
	}, "\n")

	parsed := parseVerifierPackageSummaries([]byte(stdout))
	if parsed.status() != "UNRECOGNIZED_INPUT" || parsed.UnrecognizedLineCount != 2 || len(parsed.Rows) != 1 {
		t.Fatalf("module boundary result = %#v", parsed)
	}
	if parsed.Rows[0].Package != "github.com/kimjooyoon/meta-ontology-go/internal/valid" {
		t.Fatalf("accepted package = %#v", parsed.Rows[0])
	}
}

func TestParseVerifierPackageSummariesMarksGlobalBounds(t *testing.T) {
	line := "ok github.com/kimjooyoon/meta-ontology-go/internal/valid 1s\n"
	parsed := parseVerifierPackageSummaries([]byte(strings.Repeat(line, maxVerifierPackageSummaryRowsPerInvocation+1)))
	if !parsed.Truncated || parsed.status() != "TRUNCATED" || len(parsed.Rows) != maxVerifierPackageSummaryRowsPerInvocation {
		t.Fatalf("row bound result = %#v", parsed)
	}

	longLine := "ok github.com/kimjooyoon/meta-ontology-go/internal/valid " + strings.Repeat("x", maxVerifierPackageSummaryLineBytes)
	parsed = parseVerifierPackageSummaries([]byte(longLine))
	if !parsed.Truncated || parsed.status() != "TRUNCATED" || len(parsed.Rows) != 0 {
		t.Fatalf("line bound result = %#v", parsed)
	}
}

func TestExactVerifierDurationUsesUnsignedDecimalSecondsOnly(t *testing.T) {
	valid := map[string]int64{
		"0s":                    0,
		"0.000000001s":          1,
		"2.000s":                2000000000,
		"9223372036.854775807s": 1<<63 - 1,
	}
	for token, want := range valid {
		got, ok := exactVerifierDuration(token)
		if !ok || got != want {
			t.Errorf("exactVerifierDuration(%q) = %d, %t; want %d, true", token, got, ok, want)
		}
	}
	for _, token := range []string{
		"0.5ns",
		"0.000000001ms",
		"1.0000000001s",
		"1m2.000s",
		"+1s",
		"-1s",
		"1e2s",
		"9223372036.854775808s",
		"2ms",
	} {
		if got, ok := exactVerifierDuration(token); ok {
			t.Errorf("exactVerifierDuration(%q) = %d, true; want rejected", token, got)
		}
	}
}

func TestVerifierPackageSummaryMarkersAreFiniteAndUnknownTailsAreNotRetained(t *testing.T) {
	module := "github.com/kimjooyoon/meta-ontology-go/internal/marker"
	stdout := strings.Join([]string{
		"ok " + module + " (cached)",
		"? " + module + " [no test files]",
		"? " + module + " [no tests to run]",
		"FAIL " + module + " [build failed]",
		"FAIL " + module + " [setup failed]",
		"ok " + module + " [unknown marker]",
		"ok " + module + " arbitrary-tail-secret",
	}, "\n")
	parsed := parseVerifierPackageSummaries([]byte(stdout))
	if parsed.status() != "UNRECOGNIZED_INPUT" || parsed.UnrecognizedLineCount != 2 || len(parsed.Rows) != 7 {
		t.Fatalf("marker parse = %#v", parsed)
	}
	wantMarkers := []string{
		verifierOutputMarkerCached,
		verifierOutputMarkerNoTestFiles,
		verifierOutputMarkerNoTestsToRun,
		verifierOutputMarkerBuildFailed,
		verifierOutputMarkerSetupFailed,
		"",
		"",
	}
	for index, want := range wantMarkers {
		if parsed.Rows[index].OutputMarker != want {
			t.Errorf("row %d marker = %q; want %q", index, parsed.Rows[index].OutputMarker, want)
		}
		if index < 5 && parsed.Rows[index].OutputMarkerAuthority != verifierOutputMarkerAuthority {
			t.Errorf("row %d marker authority = %q", index, parsed.Rows[index].OutputMarkerAuthority)
		}
	}
	if parsed.Rows[5].ElapsedToken != "" || parsed.Rows[5].OutputMarkerAuthority != "" || parsed.Rows[6].ElapsedToken != "" {
		t.Fatalf("unknown tails retained = %#v", parsed.Rows)
	}
	encoded, err := json.Marshal(parsed.Rows)
	if err != nil || strings.Contains(string(encoded), "arbitrary-tail-secret") || strings.Contains(string(encoded), "unknown marker") {
		t.Fatalf("unknown tail leaked into rows: %s (%v)", encoded, err)
	}
}

func TestVerifierPackageSummaryCollectorBindsVerifierOnlyAndPreservesProcessResult(t *testing.T) {
	state := newMetaExecutionTraceStateWithWriter(io.Discard)
	state.invocationID = "invocation-1"
	state.verifierPackageSummary = newVerifierPackageSummaryCollector(filepath.Join(t.TempDir(), "self-improvement-execution.json"))
	trace := metaExecutionTrace{
		action: generation.Action{
			IndicatorID: "action-1",
			Activity:    "selected",
			Subject:     "internal/marker.go:1:Marker",
			Operation:   sourcepolicy.Operation("gooo/meta/operation/marker"),
		},
		sequence: 7,
		state:    state,
	}
	firstError := errors.New("verifier failed")
	first := processResult{
		Observation: generation.ProcessObservation{
			ExitCode:        1,
			StdoutBytes:     75,
			RawStdoutDigest: "sha256:" + strings.Repeat("a", 64),
			StdoutDigest:    "sha256:" + strings.Repeat("b", 64),
		},
		Stdout: []byte("FAIL github.com/kimjooyoon/meta-ontology-go/internal/marker 1.000s\n"),
		Stderr: []byte("SECRET-STDERR"),
	}
	got, gotErr := observeProcessCall(&trace, "first", "verifier", func() (processResult, error) {
		return first, firstError
	})
	if !reflect.DeepEqual(got, first) || gotErr != firstError {
		t.Fatalf("process result/error changed: got=%#v err=%v", got, gotErr)
	}
	replay := first
	replay.Observation.RawStdoutDigest = "sha256:" + strings.Repeat("c", 64)
	replay.Observation.StdoutDigest = "sha256:" + strings.Repeat("d", 64)
	if _, err := observeProcessCall(&trace, "replay", "verifier", func() (processResult, error) {
		return replay, nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := observeProcessCall(&trace, "first", "executor", func() (processResult, error) {
		return first, nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := observeProcessCall(nil, "first", "verifier", func() (processResult, error) {
		return first, nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(state.verifierPackageSummary.records) != 2 {
		t.Fatalf("collected non-verifier or nil-trace record: %#v", state.verifierPackageSummary.records)
	}
	for index, wantPass := range []string{"first", "replay"} {
		record := state.verifierPackageSummary.records[index]
		if record.InvocationID != "invocation-1" || record.ActionIndicatorID != "action-1" || record.Activity != "selected" || record.MetaOperation != string(trace.action.Operation) || record.Subject != trace.action.Subject || record.OperationSequence != 7 || record.Pass != wantPass || record.CommandKind != "verifier" || record.StdoutBytes != first.Observation.StdoutBytes {
			t.Errorf("record %d join fields = %#v", index, record)
		}
	}
	if state.verifierPackageSummary.records[0].RawStdoutDigest == state.verifierPackageSummary.records[1].RawStdoutDigest || state.verifierPackageSummary.records[0].StdoutDigest == state.verifierPackageSummary.records[1].StdoutDigest {
		t.Fatal("first/replay stdout digests were not preserved separately")
	}
	if err := state.verifierPackageSummary.write(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(state.verifierPackageSummary.path)
	if err != nil || strings.Contains(string(data), "SECRET-STDERR") {
		t.Fatalf("sidecar retained stderr sentinel: err=%v data=%s", err, data)
	}
}

func TestVerifierPackageSummaryCollectorBoundsAndUsesCallerOutputOwner(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "self-improvement-execution.json")
	collector := newVerifierPackageSummaryCollector(outputPath)
	if collector.path != outputPath+verifierPackageSummarySidecarSuffix {
		t.Fatalf("sidecar path = %q", collector.path)
	}
	state := &metaExecutionTraceState{invocationID: "invocation-bounds"}
	trace := metaExecutionTrace{state: state}
	line := "ok github.com/kimjooyoon/meta-ontology-go/internal/bounds 1.000s\n"
	result := processResult{
		Observation: generation.ProcessObservation{
			ExitCode:        0,
			StdoutBytes:     len(line),
			RawStdoutDigest: "sha256:" + strings.Repeat("e", 64),
			StdoutDigest:    "sha256:" + strings.Repeat("f", 64),
		},
		Stdout: []byte(strings.Repeat(line, maxVerifierPackageSummaryRowsPerInvocation)),
	}
	for pass := range maxVerifierPackageSummaryRecords + 1 {
		trace.sequence = pass + 1
		collector.observe(trace, "first", result)
	}
	if !collector.truncated || len(collector.records) != maxVerifierPackageSummaryRecords || collector.totalRows > maxVerifierPackageSummaryTotalRows {
		t.Fatalf("aggregate bounds = records=%d rows=%d truncated=%t", len(collector.records), collector.totalRows, collector.truncated)
	}
	document := verifierPackageSummaryDocument{
		Schema:            verifierPackageSummarySchema,
		DiagnosticOnly:    "DIAGNOSTIC_ONLY",
		Authenticity:      "AUTHENTICITY_UNVERIFIED",
		Improvement:       "UNKNOWN",
		DiagnosticMarkers: collector.markers(),
		Truncated:         collector.truncated,
		Records:           collector.records,
	}
	payload, err := boundedVerifierPackageSummaryPayload(document)
	if err != nil || len(payload) > maxVerifierPackageSummaryFileBytes {
		t.Fatalf("bounded sidecar payload = %d bytes, err=%v", len(payload), err)
	}
	if err := collector.write(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(collector.path)
	if err != nil || len(data) > maxVerifierPackageSummaryFileBytes || strings.Contains(string(data), "SECRET-STDERR") {
		t.Fatalf("sidecar owner/write = %d bytes, err=%v", len(data), err)
	}
}
