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
	if event.VerifierWork != nil || event.VerifierCache != nil {
		t.Fatalf("process metadata alone invented raw-output accounting: %#v", event)
	}
}

func verifierCacheResult(stderr string, code int) processResult {
	raw := []byte(stderr)
	return processResult{
		Observation: descriptorObservation([]string{"go", "test", "./..."}, nil, raw, code),
		Stderr:      raw,
	}
}

func TestVerifierCacheRetainsNativeTextWithoutInventingMeaning(t *testing.T) {
	line := "testcache: fixture: native diagnostic not interpreted here"
	result := verifierCacheResult(strings.Repeat(line+"\n", 20)+"PRIVATE-NON-CACHE-OUTPUT\n", 1)
	got := observeMetaVerifierCache(result)
	if got.DiagnosticRows != 20 || len(got.Samples) != 16 || got.Samples[0] != line ||
		got.NonDiagnosticLines != 1 || got.InputCoverage != "NON_DIAGNOSTIC_INPUT" ||
		got.Unit != "NATIVE_DIAGNOSTIC_LINES_NOT_TEST_CASES" || got.CoverageScope != "BOUNDED_STDERR_PARSE_ONLY" ||
		got.NativeInterpretation != "NOT_INFERRED" || got.ReuseAuthority != "NONE" || got.Improvement != "UNKNOWN" {
		t.Fatalf("native output became inferred work or permission: %#v", got)
	}
	encoded, err := json.Marshal(got)
	if err != nil || strings.Contains(string(encoded), "PRIVATE-") {
		t.Fatalf("non-cache stderr leaked: %s (%v)", encoded, err)
	}
}

func TestVerifierCacheReportsBoundedAndMissingInput(t *testing.T) {
	line := "testcache: fixture: native line\n"
	for _, test := range []struct {
		name     string
		stderr   string
		coverage string
		rows     int
	}{
		{"empty", "", "NO_CACHE_DIAGNOSTICS", 0},
		{"other", "compiler error\n", "NON_DIAGNOSTIC_INPUT", 0},
		{"native", line, "COMPLETE", 1},
		{"unterminated", strings.TrimSuffix(line, "\n"), "COMPLETE", 1},
		{"rows", strings.Repeat(line, maxVerifierCacheRows+1), "TRUNCATED", maxVerifierCacheRows},
		{"bytes", line + strings.Repeat("x", maxVerifierCacheStderrBytes), "TRUNCATED", 1},
		{"partial-line", strings.Repeat("x", maxVerifierCacheStderrBytes+1), "TRUNCATED", 0},
		{"long-line", "testcache: " + strings.Repeat("x", maxVerifierCacheLineBytes) + "\n" + line, "TRUNCATED", 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := verifierCacheResult(test.stderr, 0)
			got := observeMetaVerifierCache(result)
			if got.InputCoverage != test.coverage || got.DiagnosticRows != test.rows ||
				got.RawStderrDigest != digestBytes(result.Stderr) || got.StderrBytes != len(result.Stderr) ||
				got.ReuseAuthority != "NONE" || got.Improvement != "UNKNOWN" {
				t.Fatalf("bounded observation lost uncertainty: %#v", got)
			}
		})
	}
}

func TestVerifierCacheRequiresExactRawProcessBinding(t *testing.T) {
	for _, condition := range []string{"matched", "missing", "digest", "size"} {
		result := verifierCacheResult("testcache: fixture: native line\n", 0)
		want := "MISMATCH"
		switch condition {
		case "matched":
			want = "MATCHED"
		case "missing":
			result.Observation.RawStderrDigest, want = "", "UNOBSERVED"
		case "digest":
			result.Observation.RawStderrDigest = "sha256:" + strings.Repeat("0", 64)
		case "size":
			result.Observation.StderrBytes++
		}
		if got := observeMetaVerifierCache(result); got.ProcessBinding != want || got.DiagnosticRows != 1 {
			t.Fatalf("%s binding was not preserved: %#v", condition, got)
		}
	}
}

func TestVerifierCacheTracePreservesRawFirstReplayFailure(t *testing.T) {
	var output bytes.Buffer
	trace := metaExecutionTrace{
		headSHA: "head-cache", planDigest: "plan-cache", manifestDigest: "manifest-cache", sequence: 2,
		state: newMetaExecutionTraceStateWithWriter(&output),
	}
	trace.action.Activity = "CollapseAssignReturn"
	trace.action.InputContractSourceDigest = strings.Repeat("a", 64)
	trace.action.InputContractSemanticDigest = strings.Repeat("b", 64)
	for _, pass := range []string{"first", "replay"} {
		output.Reset()
		result := verifierCacheResult("testcache: fixture: input list not found: "+pass+"\n", 1)
		failure := errors.New("verifier failed")
		got, gotErr := observeProcessCall(&trace, pass, "verifier", func() (processResult, error) {
			return result, failure
		})
		if !reflect.DeepEqual(got, result) || gotErr != failure {
			t.Fatal("cache diagnostics changed the raw process result")
		}
		lines := strings.Split(strings.TrimSpace(output.String()), "\n")
		var event metaExecutionTraceEvent
		if len(lines) != 2 {
			t.Fatalf("unexpected event count: %d", len(lines))
		}
		if err := json.Unmarshal([]byte(lines[1]), &event); err != nil {
			t.Fatal(err)
		}
		if event.Pass != pass || event.HeadSHA != trace.headSHA || event.PlanDigest != trace.planDigest ||
			event.ManifestDigest != trace.manifestDigest || event.Activity != trace.action.Activity ||
			event.InputContractSourceDigest != trace.action.InputContractSourceDigest ||
			event.InputContractSemanticDigest != trace.action.InputContractSemanticDigest ||
			event.VerifierCache == nil || event.VerifierCache.ProcessBinding != "MATCHED" ||
			event.VerifierCache.RawStderrDigest != result.Observation.RawStderrDigest ||
			event.VerifierCache.StageRows["PRIOR_INPUT_LIST_LOOKUP"] != 1 ||
			len(event.VerifierCache.SampleStages) != 1 || event.VerifierCache.SampleStages[0] != "PRIOR_INPUT_LIST_LOOKUP" ||
			event.ExitCode == nil || *event.ExitCode != 1 || event.ReturnErrorObserved == nil || !*event.ReturnErrorObserved {
			t.Fatalf("cache record lost its Gooo/process/failure binding: %#v", event)
		}
	}
}

func TestVerifierCacheStagesPreserveLookupBoundariesWithoutInferringCause(t *testing.T) {
	for _, test := range []struct {
		name, message, stage string
	}{
		{"identity", "fixture: test ID aa => bb", "PRIOR_INPUT_LIST_LOOKUP"},
		{"missing-list", "fixture: input list not found: unavailable", "PRIOR_INPUT_LIST_LOOKUP"},
		{"list-envelope", "fixture: input list malformed", "PRIOR_INPUT_LIST_LOOKUP"},
		{"list-record", "fixture: input list malformed (\"bad\")", "INPUT_LOG_PARSE"},
		{"result-key", "fixture: test ID aa => input ID bb => cc", "TEST_OUTPUT_LOOKUP"},
		{"missing-output", "fixture: test output not found: unavailable", "TEST_OUTPUT_LOOKUP"},
		{"output-shape", "fixture: test output malformed", "TEST_OUTPUT_LOOKUP"},
		{"expired", "fixture: test output expired due to go clean -testcache", "TEST_OUTPUT_LOOKUP"},
		{"save-attempt", "fixture: save test ID aa => input ID bb => cc", "STORE_ATTEMPT"},
		{"disabled", "caching disabled for test argument: -test.count=1", "CONFIGURATION"},
		{"future", "fixture: future cache format", "UNRECOGNIZED"},
		{"empty", "", "UNRECOGNIZED"},
		{"empty-package", ": input list not found: unavailable", "UNRECOGNIZED"},
		{"invalid-package", "fake package: input list not found: unavailable", "UNRECOGNIZED"},
		{"incomplete-key", "fixture: test ID aa => input ID => cc", "UNRECOGNIZED"},
		{"not-hex", "fixture: test ID zz => bb", "UNRECOGNIZED"},
		{"save-without-input", "fixture: save test ID aa => bb", "UNRECOGNIZED"},
		{"missing-detail", "fixture: input list not found: ", "UNRECOGNIZED"},
	} {
		t.Run(test.name, func(t *testing.T) {
			line := "testcache: " + test.message
			got := observeMetaVerifierCache(verifierCacheResult(line+"\n", 0))
			total := 0
			for _, count := range got.StageRows {
				total += count
			}
			if got.DiagnosticRows != 1 || total != 1 || len(got.StageRows) != 6 || got.StageRows[test.stage] != 1 ||
				len(got.SampleStages) != 1 || got.SampleStages[0] != test.stage || got.Samples[0] != line ||
				got.StageBasis != "GO_TESTCACHE_TEXT_V1_NOT_EXECUTION_ATTESTATION" ||
				got.LookupIdentityKind != "NOT_EXPOSED_BY_NATIVE_TEXT" || got.ProcessBinding != "MATCHED" ||
				got.NativeInterpretation != "NOT_INFERRED" || got.ReuseAuthority != "NONE" || got.Improvement != "UNKNOWN" {
				t.Fatalf("diagnostic shape became a cause or execution claim: %#v", got)
			}
		})
	}
}

func TestVerifierCacheStageCountsKeepBoundedCoverageBeyondSampleLimit(t *testing.T) {
	line := "testcache: fixture: input list not found: unavailable\n"
	for _, rows := range []int{maxVerifierCacheSamples + 3, maxVerifierCacheRows + 1} {
		got := observeMetaVerifierCache(verifierCacheResult(strings.Repeat(line, rows), 0))
		want, coverage := rows, "COMPLETE"
		if want > maxVerifierCacheRows {
			want, coverage = maxVerifierCacheRows, "TRUNCATED"
		}
		if got.DiagnosticRows != want || got.StageRows["PRIOR_INPUT_LIST_LOOKUP"] != want ||
			got.StageRows["TEST_OUTPUT_LOOKUP"] != 0 || got.StageRows["UNRECOGNIZED"] != 0 ||
			len(got.Samples) != maxVerifierCacheSamples || len(got.SampleStages) != len(got.Samples) ||
			got.InputCoverage != coverage || got.LookupIdentityKind != "NOT_EXPOSED_BY_NATIVE_TEXT" {
			t.Fatalf("sample limit concealed stage counts or lookup uncertainty: %#v", got)
		}
	}
}
