package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestCostSeparatesPassesAndRetainsFailureDuration(t *testing.T) {
	var state metaExecutionCostState
	start := time.Unix(0, 0)
	event := metaExecutionTraceEvent{Boundary: "PROCESS_CALL_ENTERED", OperationSequence: 1,
		Pass: "first", CommandKind: "verifier", EventSequence: 1}
	state.observe(event, start)
	event.Pass, event.EventSequence = "replay", 2
	state.observe(event, start.Add(time.Second))
	event.Boundary, event.Pass = "PROCESS_RETURNED", "first"
	failed, code := true, 1
	event.ReturnErrorObserved, event.ExitCode = &failed, &code
	got := state.observe(event, start.Add(3*time.Second))
	if got.State != "OBSERVED" || got.ElapsedNS == nil || *got.ElapsedNS != 3000000000 || got.StartedAtEvent != 1 {
		t.Fatalf("first pass cost: %+v", got)
	}
	if got.Improvement != "UNKNOWN" || got.ToolchainIdentity != "UNOBSERVED" {
		t.Fatalf("cost invented stronger evidence: %+v", got)
	}
	event.Pass = "replay"
	got = state.observe(event, start.Add(5*time.Second))
	if got.ElapsedNS == nil || *got.ElapsedNS != 4000000000 || got.StartedAtEvent != 2 {
		t.Fatalf("replay cost: %+v", got)
	}
}

func TestCostMissingStartAndNegativeClockStayUnknown(t *testing.T) {
	var state metaExecutionCostState
	event := metaExecutionTraceEvent{Boundary: "ACTION_RETURNED"}
	if got := state.observe(event, time.Unix(2, 0)); got.State != "UNKNOWN" || got.ElapsedNS != nil {
		t.Fatalf("unpaired return: %+v", got)
	}
	event.Boundary, event.EventSequence = "ACTION_ENTERED", 1
	state.observe(event, time.Unix(2, 0))
	event.Boundary = "ACTION_RETURNED"
	if got := state.observe(event, time.Unix(1, 0)); got.State != "UNKNOWN" || got.ElapsedNS != nil {
		t.Fatalf("negative elapsed: %+v", got)
	}
}

func verifierWorkResult(stdout string, code int) processResult {
	raw := []byte(stdout)
	return processResult{
		Observation: descriptorObservation([]string{"go", "test", "./..."}, raw, nil, code),
		Stdout:      raw,
	}
}

func TestVerifierWorkCountsOutputAxesWithoutInferringExecutedTests(t *testing.T) {
	// Repeated paths deliberately remain separate output observations.
	pkg := verifierPackageSummaryModulePrefix + "/internal/work"
	result := verifierWorkResult(strings.Join([]string{
		"ok " + pkg + " 0.250s",
		"ok " + pkg + " (cached)",
		"? " + pkg + " [no test files]",
		"ok " + pkg + " 0.002s [no tests to run]",
		"FAIL " + pkg + " [build failed]",
		"FAIL " + pkg + " [setup failed]",
		"FAIL " + pkg + " 0.003s",
		"ok " + pkg,
		"ok " + pkg + " SECRET-UNKNOWN-TAIL",
		"SECRET-NON-SUMMARY",
	}, "\n"), 1)
	got := observeMetaVerifierWork(result)
	if got.ObservedRows != 9 || got.ElapsedObservedRows != 3 || got.ElapsedUnknownRows != 6 ||
		got.OKRows != 5 || got.FailRows != 3 || got.QuestionRows != 1 ||
		got.CachedMarkerRows != 1 || got.NoTestFilesMarkerRows != 1 || got.NoTestsToRunMarkerRows != 1 ||
		got.BuildFailedMarkerRows != 1 || got.SetupFailedMarkerRows != 1 || got.UnmarkedRows != 4 ||
		got.UnrecognizedLines != 2 || got.InputCoverage != "UNRECOGNIZED_INPUT" {
		t.Fatalf("output observation axes = %#v", got)
	}
	if got.ObservedRows != got.ElapsedObservedRows+got.ElapsedUnknownRows ||
		got.ObservedRows != got.OKRows+got.FailRows+got.QuestionRows ||
		got.ObservedRows != got.CachedMarkerRows+got.NoTestFilesMarkerRows+got.NoTestsToRunMarkerRows+
			got.BuildFailedMarkerRows+got.SetupFailedMarkerRows+got.UnmarkedRows {
		t.Fatalf("an observation axis lost rows: %#v", got)
	}
	if got.Unit != "PACKAGE_SUMMARY_ROWS_NOT_TEST_CASES" || got.CoverageScope != "BOUNDED_STDOUT_PARSE_ONLY" ||
		got.ExecutionClaims != "NOT_INFERRED" || got.ReuseAuthority != "OUTPUT_MARKER_ONLY" ||
		got.Improvement != "UNKNOWN" || got.ProcessBinding != "MATCHED" ||
		got.RawStdoutDigest != digestBytes(result.Stdout) || got.StdoutBytes != len(result.Stdout) {
		t.Fatalf("work observation invented stronger authority or lost its input: %#v", got)
	}
	encoded, err := json.Marshal(got)
	if err != nil || strings.Contains(string(encoded), "SECRET-") {
		t.Fatalf("unknown stdout contents leaked: %s (%v)", encoded, err)
	}
}

func TestVerifierWorkPreservesParsedCountsBeforePresentationSampling(t *testing.T) {
	pkg := verifierPackageSummaryModulePrefix + "/internal/work"
	result := verifierWorkResult(strings.Repeat("ok "+pkg+" 1s\n", 20)+
		strings.Repeat("ok "+pkg+" (cached)\n", 20), 0)
	parsed := parseVerifierPackageSummaries(result.Stdout)
	sample := compactVerifierPackageSummaryRows(parsed.Rows, maxVerifierPackageSummarySampleRows)
	got := observeMetaVerifierWork(result)
	if len(sample) != 16 || got.InputCoverage != "COMPLETE" || got.ObservedRows != 40 ||
		got.ElapsedObservedRows != 20 || got.CachedMarkerRows != 20 || got.UnmarkedRows != 20 {
		t.Fatalf("presentation sampling changed observed counts: sample=%d work=%#v", len(sample), got)
	}
}

func TestVerifierWorkBoundedInputsDoNotClaimACompleteDenominator(t *testing.T) {
	line := "ok " + verifierPackageSummaryModulePrefix + "/internal/work 1s\n"
	for _, test := range []struct {
		name   string
		stdout string
		rows   int
	}{
		{name: "rows", stdout: strings.Repeat(line, maxVerifierPackageSummaryRowsPerInvocation+1), rows: maxVerifierPackageSummaryRowsPerInvocation},
		{name: "bytes", stdout: strings.Repeat("x", maxVerifierPackageSummaryStdoutBytes+1), rows: 0},
		{name: "line", stdout: strings.Repeat("x", maxVerifierPackageSummaryLineBytes+1) + "\n", rows: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := verifierWorkResult(test.stdout, 0)
			got := observeMetaVerifierWork(result)
			if got.InputCoverage != "TRUNCATED" || got.ObservedRows != test.rows ||
				got.RawStdoutDigest != digestBytes(result.Stdout) || got.StdoutBytes != len(result.Stdout) ||
				got.ExecutionClaims != "NOT_INFERRED" || got.Improvement != "UNKNOWN" {
				t.Fatalf("bounded input was promoted to complete coverage: %#v", got)
			}
		})
	}
}

func TestVerifierWorkEmptyOrUnrecognizedInputDoesNotMeanZeroTests(t *testing.T) {
	for _, test := range []struct {
		stdout   string
		coverage string
	}{
		{stdout: "", coverage: "NO_PACKAGE_SUMMARIES"},
		{stdout: "not a package summary\n", coverage: "UNRECOGNIZED_INPUT"},
	} {
		got := observeMetaVerifierWork(verifierWorkResult(test.stdout, 0))
		if got.ObservedRows != 0 || got.InputCoverage != test.coverage ||
			got.ExecutionClaims != "NOT_INFERRED" || got.ReuseAuthority != "OUTPUT_MARKER_ONLY" {
			t.Fatalf("missing package output became a zero-test claim: %#v", got)
		}
	}
}

func TestVerifierWorkReportsMissingAndContradictoryProcessBindings(t *testing.T) {
	result := verifierWorkResult("ok "+verifierPackageSummaryModulePrefix+"/internal/work 1s\n", 0)
	missing := result
	missing.Observation.RawStdoutDigest = ""
	badDigest := result
	badDigest.Observation.RawStdoutDigest = "sha256:" + strings.Repeat("0", 64)
	badSize := result
	badSize.Observation.StdoutBytes++
	for _, test := range []struct {
		name    string
		result  processResult
		binding string
	}{
		{name: "matched", result: result, binding: "MATCHED"},
		{name: "missing-digest", result: missing, binding: "UNOBSERVED"},
		{name: "wrong-digest", result: badDigest, binding: "MISMATCH"},
		{name: "wrong-size", result: badSize, binding: "MISMATCH"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := observeMetaVerifierWork(test.result)
			if got.ProcessBinding != test.binding || got.ObservedRows != 1 ||
				got.RawStdoutDigest != digestBytes(result.Stdout) || got.ExecutionClaims != "NOT_INFERRED" {
				t.Fatalf("process binding = %#v", got)
			}
		})
	}
}

func TestVerifierWorkTracePreservesGoooBindingResultsAndFailure(t *testing.T) {
	var output bytes.Buffer
	trace := metaExecutionTrace{
		headSHA: "head-1", planDigest: "plan-1", manifestDigest: "manifest-1", sequence: 3,
		state: newMetaExecutionTraceStateWithWriter(&output),
	}
	trace.action.IndicatorID = "indicator-1"
	trace.action.Activity = "CollapseAssignReturn"
	trace.action.Operation = "collapse-assign-return"
	trace.action.Subject = "fixture.go:3:Fixture"
	trace.action.InputContractSourceDigest = strings.Repeat("a", 64)
	trace.action.InputContractSemanticDigest = strings.Repeat("b", 64)
	failure := errors.New("verifier failure")
	first := verifierWorkResult("FAIL "+verifierPackageSummaryModulePrefix+"/internal/work 1s\n", 1)
	got, gotErr := observeProcessCall(&trace, "first", "verifier", func() (processResult, error) {
		return first, failure
	})
	if !reflect.DeepEqual(got, first) || gotErr != failure {
		t.Fatal("diagnostic work accounting changed the process result or error")
	}
	replay := verifierWorkResult("ok "+verifierPackageSummaryModulePrefix+"/internal/work (cached)\n", 0)
	_, _ = observeProcessCall(&trace, "replay", "verifier", func() (processResult, error) {
		return replay, nil
	})
	_, _ = observeProcessCall(&trace, "first", "executor", func() (processResult, error) {
		return first, failure
	})
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 6 {
		t.Fatalf("trace event count = %d", len(lines))
	}
	for index, line := range lines {
		var event metaExecutionTraceEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatal(err)
		}
		if index != 1 && index != 3 {
			if event.VerifierWork != nil {
				t.Fatalf("non-verifier-return event acquired work evidence: %#v", event)
			}
			continue
		}
		if event.VerifierWork == nil || event.HeadSHA != "head-1" || event.PlanDigest != "plan-1" ||
			event.ManifestDigest != "manifest-1" || event.OperationSequence != 3 ||
			event.ActionIndicatorID != trace.action.IndicatorID || event.Activity != trace.action.Activity ||
			event.InputContractSourceDigest != trace.action.InputContractSourceDigest ||
			event.InputContractSemanticDigest != trace.action.InputContractSemanticDigest ||
			!event.DiagnosticOnly || event.SemanticEffect != "UNOBSERVED" || event.Permission != "UNOBSERVED" ||
			event.Cost == nil || event.Cost.State != "OBSERVED" || event.Cost.Improvement != "UNKNOWN" {
			t.Fatalf("work trace lost its meta binding or gained authority: %#v", event)
		}
		expected := first
		if index == 3 {
			expected = replay
		}
		if event.VerifierWork.RawStdoutDigest != expected.Observation.RawStdoutDigest ||
			event.VerifierWork.ProcessBinding != "MATCHED" || event.ExitCode == nil ||
			*event.ExitCode != expected.Observation.ExitCode || event.ReturnErrorObserved == nil ||
			*event.ReturnErrorObserved != (index == 1) {
			t.Fatalf("first/replay input or failure attribution changed: %#v", event)
		}
	}
}

func TestVerifierWorkLegacyReturnWithoutRawOutputStaysUnobserved(t *testing.T) {
	var output bytes.Buffer
	trace := metaExecutionTrace{state: newMetaExecutionTraceStateWithWriter(&output)}
	result := verifierWorkResult("ok "+verifierPackageSummaryModulePrefix+"/internal/work (cached)\n", 0)
	trace.emitProcessReturned("first", "verifier", result.Observation, nil)
	var event metaExecutionTraceEvent
	if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &event); err != nil {
		t.Fatal(err)
	}
	if event.VerifierWork != nil {
		t.Fatalf("process metadata alone invented raw-output accounting: %#v", event)
	}
}
