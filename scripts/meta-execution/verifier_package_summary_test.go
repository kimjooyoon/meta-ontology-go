package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
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

func TestCompactVerifierPackageSummaryRowsPrioritizesMarkersAndElapsedRows(t *testing.T) {
	timedRow := func(packageName string, elapsed int64) verifierPackageSummaryRow {
		return verifierPackageSummaryRow{
			Package:            packageName,
			Status:             "ok",
			ElapsedToken:       "1s",
			ElapsedStatus:      "PARSED_EXACTLY",
			ElapsedUnit:        "NANOSECONDS",
			ElapsedNanoseconds: &elapsed,
		}
	}
	rows := []verifierPackageSummaryRow{
		{Package: "pkg/cached", Status: "ok", ElapsedStatus: "UNKNOWN", OutputMarker: verifierOutputMarkerCached},
		{Package: "pkg/no-test-files", Status: "?", ElapsedStatus: "UNKNOWN", OutputMarker: verifierOutputMarkerNoTestFiles},
		{Package: "pkg/no-tests-to-run", Status: "?", ElapsedStatus: "UNKNOWN", OutputMarker: verifierOutputMarkerNoTestsToRun},
		{Package: "pkg/build-failed", Status: "FAIL", ElapsedStatus: "UNKNOWN", OutputMarker: verifierOutputMarkerBuildFailed},
		{Package: "pkg/setup-failed", Status: "FAIL", ElapsedStatus: "UNKNOWN", OutputMarker: verifierOutputMarkerSetupFailed},
		{Package: "pkg/fail", Status: "FAIL", ElapsedStatus: "UNKNOWN"},
	}
	for _, item := range []struct {
		packageName string
		elapsed     int64
	}{
		{packageName: "pkg/timed-high", elapsed: 100},
		{packageName: "pkg/timed-90", elapsed: 90},
		{packageName: "pkg/timed-80", elapsed: 80},
		{packageName: "pkg/timed-70", elapsed: 70},
		{packageName: "pkg/timed-60", elapsed: 60},
		{packageName: "pkg/timed-50", elapsed: 50},
		{packageName: "pkg/timed-40", elapsed: 40},
		{packageName: "pkg/timed-30", elapsed: 30},
		{packageName: "pkg/timed-20", elapsed: 20},
		{packageName: "pkg/timed-10", elapsed: 10},
		{packageName: "pkg/timed-1", elapsed: 1},
		{packageName: "pkg/timed-0", elapsed: 0},
		{packageName: "pkg/timed-low", elapsed: -1},
	} {
		rows = append(rows, timedRow(item.packageName, item.elapsed))
	}

	first := compactVerifierPackageSummaryRows(rows, maxVerifierPackageSummarySampleRows)
	second := compactVerifierPackageSummaryRows(rows, maxVerifierPackageSummarySampleRows)
	if len(first) != maxVerifierPackageSummarySampleRows || !reflect.DeepEqual(first, second) {
		t.Fatalf("non-deterministic bounded sample: first=%#v second=%#v", first, second)
	}
	for _, marker := range []string{
		verifierOutputMarkerBuildFailed,
		verifierOutputMarkerSetupFailed,
		verifierOutputMarkerCached,
		verifierOutputMarkerNoTestFiles,
		verifierOutputMarkerNoTestsToRun,
	} {
		found := false
		for _, row := range first {
			if row.OutputMarker == marker {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("bounded sample omitted marker %q: %#v", marker, first)
		}
	}
	seenHigh := false
	seenLow := false
	for _, row := range first {
		seenHigh = seenHigh || row.Package == "pkg/timed-high"
		seenLow = seenLow || row.Package == "pkg/timed-low"
	}
	if !seenHigh || seenLow {
		t.Fatalf("elapsed priority was not applied: %#v", first)
	}
}

func TestVerifierPackageSummaryRowLessOrdersTimedRowsTransitively(t *testing.T) {
	twenty := int64(20)
	ten := int64(10)
	rows := []verifierPackageSummaryRow{
		{Package: "pkg/z", ElapsedNanoseconds: &twenty},
		{Package: "pkg/m", ElapsedStatus: "UNKNOWN"},
		{Package: "pkg/a", ElapsedNanoseconds: &ten},
	}
	for first := range rows {
		if verifierPackageSummaryRowLess(rows[first], rows[first]) {
			t.Fatalf("row comparator is not irreflexive for row %d", first)
		}
		for second := range rows {
			if verifierPackageSummaryRowLess(rows[first], rows[second]) && verifierPackageSummaryRowLess(rows[second], rows[first]) {
				t.Fatalf("row comparator has a cycle: %d and %d", first, second)
			}
			for third := range rows {
				if verifierPackageSummaryRowLess(rows[first], rows[second]) && verifierPackageSummaryRowLess(rows[second], rows[third]) && !verifierPackageSummaryRowLess(rows[first], rows[third]) {
					t.Fatalf("row comparator is not transitive: %d, %d, %d", first, second, third)
				}
			}
		}
	}
	if !verifierPackageSummaryRowLess(rows[0], rows[2]) || !verifierPackageSummaryRowLess(rows[2], rows[1]) || !verifierPackageSummaryRowLess(rows[0], rows[1]) {
		t.Fatalf("timed rows did not precede unknown elapsed rows: %#v", rows)
	}

	fifty := int64(50)
	fiftyAgain := int64(50)
	ties := []verifierPackageSummaryRow{
		{Package: "pkg/b", Status: "ok", OutputMarker: "z", ElapsedNanoseconds: &fifty},
		{Package: "pkg/a", Status: "ok", OutputMarker: "z", ElapsedNanoseconds: &fiftyAgain},
		{Package: "pkg/a", Status: "FAIL", OutputMarker: "z", ElapsedNanoseconds: &fifty},
		{Package: "pkg/a", Status: "ok", OutputMarker: "a", ElapsedNanoseconds: &fiftyAgain},
	}
	if !verifierPackageSummaryRowLess(ties[2], ties[3]) || !verifierPackageSummaryRowLess(ties[3], ties[1]) || !verifierPackageSummaryRowLess(ties[1], ties[0]) {
		t.Fatalf("equal-duration lexical tie-break changed: %#v", ties)
	}
}

func TestCompactVerifierPackageSummaryRowsPrioritizesFailureOverBenignMarkers(t *testing.T) {
	rows := []verifierPackageSummaryRow{
		{Package: "pkg/cached", Status: "ok", ElapsedStatus: "UNKNOWN", OutputMarker: verifierOutputMarkerCached},
		{Package: "pkg/fail", Status: "FAIL", ElapsedStatus: "UNKNOWN"},
	}
	sampled := compactVerifierPackageSummaryRows(rows, 1)
	if len(sampled) != 1 || sampled[0].Status != "FAIL" || sampled[0].Package != "pkg/fail" {
		t.Fatalf("failure exemplar was not prioritized: %#v", sampled)
	}
}

func TestBoundedVerifierPackageSummaryPayloadRetainsAllRecordSamples(t *testing.T) {
	records := make([]verifierPackageSummaryRecord, 0, 4)
	for index := 0; index < 4; index++ {
		packages := make([]verifierPackageSummaryRow, 0, maxVerifierPackageSummaryRowsPerInvocation)
		for rowIndex := 0; rowIndex < maxVerifierPackageSummaryRowsPerInvocation; rowIndex++ {
			packages = append(packages, verifierPackageSummaryRow{
				Package:       "github.com/kimjooyoon/meta-ontology-go/internal/" + strings.Repeat("package", 30),
				Status:        "ok",
				ElapsedStatus: "UNKNOWN",
			})
		}
		records = append(records, verifierPackageSummaryRecord{
			InvocationID:      "invocation-" + strings.Repeat("i", 80),
			ActionIndicatorID: "action-" + strings.Repeat("a", 80),
			Activity:          "selected",
			MetaOperation:     "gooo/meta/operation/summary",
			Subject:           "internal/summary.go:1:Summary",
			OperationSequence: index + 1,
			Pass:              "pass-" + strings.Repeat("p", 80),
			CommandKind:       "verifier",
			ParseStatus:       "COMPLETE",
			Packages:          packages,
		})
	}
	document := verifierPackageSummaryDocument{
		Schema:            verifierPackageSummarySchema,
		DiagnosticOnly:    "DIAGNOSTIC_ONLY",
		Authenticity:      "AUTHENTICITY_UNVERIFIED",
		Improvement:       "UNKNOWN",
		Records:           records,
	}
	payload, err := boundedVerifierPackageSummaryPayload(document)
	if err != nil || len(payload) > maxVerifierPackageSummaryFileBytes {
		t.Fatalf("bounded sampled payload = %d bytes, err=%v", len(payload), err)
	}
	var got verifierPackageSummaryDocument
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("decode bounded sampled payload: %v", err)
	}
	if !got.Truncated || !slices.Contains(got.DiagnosticMarkers, "TRUNCATED_OUTPUT_FILE") {
		t.Fatalf("global truncation marker missing: %#v", got)
	}
	if len(got.Records) != len(records) {
		t.Fatalf("record order/count changed: got=%d want=%d", len(got.Records), len(records))
	}
	for index, record := range got.Records {
		if record.InvocationID != records[index].InvocationID || record.ActionIndicatorID != records[index].ActionIndicatorID || record.OperationSequence != index+1 || record.Pass != records[index].Pass || record.CommandKind != "verifier" {
			t.Errorf("record %d binding fields changed: %#v", index, record)
		}
		if len(record.Packages) == 0 || len(record.Packages) > maxVerifierPackageSummarySampleRows || !record.Truncated || !slices.Contains(record.DiagnosticMarkers, "TRUNCATED_OUTPUT_FILE") {
			t.Errorf("record %d sample/truncation fields = %#v", index, record)
		}
	}
}

func TestBoundedVerifierPackageSummaryPayloadPreservesCompleteShape(t *testing.T) {
	rows := []verifierPackageSummaryRow{
		{Package: "pkg/first", Status: "ok", ElapsedStatus: "UNKNOWN"},
		{Package: "pkg/second", Status: "FAIL", ElapsedStatus: "UNKNOWN"},
	}
	document := verifierPackageSummaryDocument{
		Schema:            verifierPackageSummarySchema,
		DiagnosticOnly:    "DIAGNOSTIC_ONLY",
		Authenticity:      "AUTHENTICITY_UNVERIFIED",
		Improvement:       "UNKNOWN",
		DiagnosticMarkers: []string{"OBSERVED"},
		Records: []verifierPackageSummaryRecord{{
			InvocationID:      "invocation-1",
			ActionIndicatorID: "action-1",
			OperationSequence: 1,
			Pass:              "first",
			CommandKind:       "verifier",
			ParseStatus:       "COMPLETE",
			Packages:          rows,
		}},
	}
	want, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	want = append(want, '\n')
	got, err := boundedVerifierPackageSummaryPayload(document)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("complete payload changed: err=%v got=%q want=%q", err, got, want)
	}
}

func TestCompactVerifierPackageSummaryRecordsDistinguishesEmptyAndOmittedRows(t *testing.T) {
	rows := []verifierPackageSummaryRow{
		{Package: "pkg/one", Status: "ok", ElapsedStatus: "UNKNOWN"},
		{Package: "pkg/two", Status: "ok", ElapsedStatus: "UNKNOWN"},
	}
	records := []verifierPackageSummaryRecord{
		{InvocationID: "empty"},
		{InvocationID: "sampled", Packages: rows},
	}
	compacted := compactVerifierPackageSummaryRecords(records, 1)
	if len(compacted[0].Packages) != 0 || compacted[0].Truncated || slices.Contains(compacted[0].DiagnosticMarkers, "TRUNCATED_OUTPUT_FILE") {
		t.Fatalf("source-empty record was marked omitted: %#v", compacted[0])
	}
	if len(compacted[1].Packages) != 1 || !compacted[1].Truncated || compacted[1].ParseStatus != "TRUNCATED" || !slices.Contains(compacted[1].DiagnosticMarkers, "TRUNCATED_OUTPUT_FILE") {
		t.Fatalf("omitted sample was not marked as truncated: %#v", compacted[1])
	}
}
