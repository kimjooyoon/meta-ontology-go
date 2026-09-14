package main

import (
	"bytes"
	"strings"
	"time"
)

// Cost is diagnostic, never part of canonical output or permission evidence.
type metaExecutionCost struct {
	State             string `json:"state"`
	StartedAtEvent    uint64 `json:"started_at_event,omitempty"`
	ElapsedNS         *int64 `json:"elapsed_ns,omitempty"`
	ExecutionMode     string `json:"execution_mode"`
	ToolchainIdentity string `json:"toolchain_identity"`
	Improvement       string `json:"improvement"`
}

type metaExecutionCostKey struct {
	sequence int
	pass     string
	kind     string
	family   string
}

type metaExecutionCostStart struct {
	at    time.Time
	event uint64
}

type metaExecutionCostState struct {
	starts map[metaExecutionCostKey][]metaExecutionCostStart
}

func (state *metaExecutionCostState) observe(event metaExecutionTraceEvent, now time.Time) *metaExecutionCost {
	family, entered := "", false
	switch event.Boundary {
	case "ACTION_ENTERED":
		family, entered = "action", true
	case "ACTION_RETURNED":
		family = "action"
	case "PROCESS_CALL_ENTERED":
		family, entered = "process", true
	case "PROCESS_RETURNED":
		family = "process"
	default:
		return nil
	}
	key := metaExecutionCostKey{event.OperationSequence, event.Pass, event.CommandKind, family}
	if state.starts == nil {
		state.starts = make(map[metaExecutionCostKey][]metaExecutionCostStart)
	}
	cost := &metaExecutionCost{State: "UNKNOWN", ExecutionMode: "OBSERVED_DRIVER_CALL",
		ToolchainIdentity: "UNOBSERVED", Improvement: "UNKNOWN"}
	if entered {
		state.starts[key] = append(state.starts[key], metaExecutionCostStart{now, event.EventSequence})
		cost.State = "STARTED"
		return cost
	}
	stack := state.starts[key]
	if len(stack) == 0 {
		return cost
	}
	start := stack[len(stack)-1]
	state.starts[key] = stack[:len(stack)-1]
	cost.StartedAtEvent = start.event
	elapsed := now.Sub(start.at).Nanoseconds()
	if elapsed >= 0 {
		cost.State, cost.ElapsedNS = "OBSERVED", &elapsed
	}
	return cost
}

// These are bounded stdout observations, not unique packages, test cases,
// executed work, or authority to reuse an earlier verification.
type metaVerifierWork struct {
	Unit                   string `json:"unit"`
	CoverageScope          string `json:"coverage_scope"`
	InputCoverage          string `json:"input_coverage"`
	RawStdoutDigest        string `json:"raw_stdout_digest"`
	StdoutBytes            int    `json:"stdout_bytes"`
	ProcessBinding         string `json:"process_binding"`
	ObservedRows           int    `json:"observed_rows"`
	ElapsedObservedRows    int    `json:"elapsed_observed_rows"`
	ElapsedUnknownRows     int    `json:"elapsed_unknown_rows"`
	OKRows                 int    `json:"ok_rows"`
	FailRows               int    `json:"fail_rows"`
	QuestionRows           int    `json:"question_rows"`
	CachedMarkerRows       int    `json:"cached_marker_rows"`
	NoTestFilesMarkerRows  int    `json:"no_test_files_marker_rows"`
	NoTestsToRunMarkerRows int    `json:"no_tests_to_run_marker_rows"`
	BuildFailedMarkerRows  int    `json:"build_failed_marker_rows"`
	SetupFailedMarkerRows  int    `json:"setup_failed_marker_rows"`
	UnmarkedRows           int    `json:"unmarked_rows"`
	UnrecognizedLines      int    `json:"unrecognized_lines"`
	ExecutionClaims        string `json:"execution_claims"`
	ReuseAuthority         string `json:"reuse_authority"`
	Improvement            string `json:"improvement"`
}

func observeMetaVerifierWork(result processResult) *metaVerifierWork {
	parsed := parseVerifierPackageSummaries(result.Stdout)
	digest := digestBytes(result.Stdout)
	work := &metaVerifierWork{
		Unit: "PACKAGE_SUMMARY_ROWS_NOT_TEST_CASES", CoverageScope: "BOUNDED_STDOUT_PARSE_ONLY",
		InputCoverage: parsed.status(), RawStdoutDigest: digest, StdoutBytes: len(result.Stdout),
		ProcessBinding: "UNOBSERVED", ObservedRows: len(parsed.Rows),
		UnrecognizedLines: parsed.UnrecognizedLineCount,
		ExecutionClaims:   "NOT_INFERRED", ReuseAuthority: "OUTPUT_MARKER_ONLY", Improvement: "UNKNOWN",
	}
	if result.Observation.RawStdoutDigest != "" {
		work.ProcessBinding = "MISMATCH"
		if result.Observation.RawStdoutDigest == digest && result.Observation.StdoutBytes == len(result.Stdout) {
			work.ProcessBinding = "MATCHED"
		}
	}
	for _, row := range parsed.Rows {
		if row.ElapsedNanoseconds == nil {
			work.ElapsedUnknownRows++
		} else {
			work.ElapsedObservedRows++
		}
		switch row.Status {
		case "ok":
			work.OKRows++
		case "FAIL":
			work.FailRows++
		case "?":
			work.QuestionRows++
		}
		switch row.OutputMarker {
		case verifierOutputMarkerCached:
			work.CachedMarkerRows++
		case verifierOutputMarkerNoTestFiles:
			work.NoTestFilesMarkerRows++
		case verifierOutputMarkerNoTestsToRun:
			work.NoTestsToRunMarkerRows++
		case verifierOutputMarkerBuildFailed:
			work.BuildFailedMarkerRows++
		case verifierOutputMarkerSetupFailed:
			work.SetupFailedMarkerRows++
		default:
			work.UnmarkedRows++
		}
	}
	return work
}

const (
	maxVerifierCacheStderrBytes = 64 * 1024
	maxVerifierCacheLineBytes   = 1024
	maxVerifierCacheRows        = 512
	maxVerifierCacheSamples     = 16
)

// Native diagnostic text explains cache decisions; it cannot authorize reuse.
type metaVerifierCache struct {
	Unit                 string   `json:"unit"`
	CoverageScope        string   `json:"coverage_scope"`
	InputCoverage        string   `json:"input_coverage"`
	RawStderrDigest      string   `json:"raw_stderr_digest"`
	StderrBytes          int      `json:"stderr_bytes"`
	ProcessBinding       string   `json:"process_binding"`
	DiagnosticRows       int      `json:"diagnostic_rows"`
	NonDiagnosticLines   int      `json:"non_diagnostic_lines"`
	Samples              []string `json:"samples"`
	NativeInterpretation string   `json:"native_interpretation"`
	ReuseAuthority       string   `json:"reuse_authority"`
	Improvement          string   `json:"improvement"`

	StageRows          map[string]int `json:"stage_rows"`
	SampleStages       []string       `json:"sample_stages"`
	StageBasis         string         `json:"stage_basis"`
	LookupIdentityKind string         `json:"lookup_identity_kind"`
}

func observeMetaVerifierCache(result processResult) *metaVerifierCache {
	digest := digestBytes(result.Stderr)
	cache := &metaVerifierCache{
		Unit: "NATIVE_DIAGNOSTIC_LINES_NOT_TEST_CASES", CoverageScope: "BOUNDED_STDERR_PARSE_ONLY",
		RawStderrDigest: digest, StderrBytes: len(result.Stderr), ProcessBinding: "UNOBSERVED",
		Samples: []string{}, SampleStages: []string{}, NativeInterpretation: "NOT_INFERRED",
		ReuseAuthority: "NONE", Improvement: "UNKNOWN",
		StageBasis: "GO_TESTCACHE_TEXT_V1_NOT_EXECUTION_ATTESTATION", LookupIdentityKind: "NOT_EXPOSED_BY_NATIVE_TEXT",
		StageRows: map[string]int{
			"PRIOR_INPUT_LIST_LOOKUP": 0,
			"INPUT_LOG_PARSE":         0,
			"TEST_OUTPUT_LOOKUP":      0,
			"STORE_ATTEMPT":           0,
			"CONFIGURATION":           0,
			"UNRECOGNIZED":            0,
		},
	}
	if result.Observation.RawStderrDigest != "" {
		cache.ProcessBinding = "MISMATCH"
		if result.Observation.RawStderrDigest == digest && result.Observation.StderrBytes == len(result.Stderr) {
			cache.ProcessBinding = "MATCHED"
		}
	}
	cache.observeLines(result.Stderr)
	return cache
}

func (cache *metaVerifierCache) observeLines(data []byte) {
	truncated := len(data) > maxVerifierCacheStderrBytes
	if truncated {
		data = data[:maxVerifierCacheStderrBytes]
		data = data[:bytes.LastIndexByte(data, '\n')+1]
	}
	for len(data) > 0 {
		line, remaining, _ := bytes.Cut(data, []byte{'\n'})
		data = remaining
		line = bytes.TrimSuffix(line, []byte{'\r'})
		if len(line) > maxVerifierCacheLineBytes {
			truncated = true
			continue
		}
		if len(line) == 0 {
			continue
		}
		if !bytes.HasPrefix(line, []byte("testcache: ")) {
			cache.NonDiagnosticLines++
			continue
		}
		if cache.DiagnosticRows == maxVerifierCacheRows {
			truncated = true
			break
		}
		cache.DiagnosticRows++
		stage := classifyMetaVerifierCacheStage(string(line))
		cache.StageRows[stage]++
		if len(cache.Samples) < maxVerifierCacheSamples {
			cache.Samples = append(cache.Samples, string(line))
			cache.SampleStages = append(cache.SampleStages, stage)
		}
	}
	switch {
	case truncated:
		cache.InputCoverage = "TRUNCATED"
	case cache.NonDiagnosticLines > 0:
		cache.InputCoverage = "NON_DIAGNOSTIC_INPUT"
	case cache.DiagnosticRows == 0:
		cache.InputCoverage = "NO_CACHE_DIAGNOSTICS"
	default:
		cache.InputCoverage = "COMPLETE"
	}
}

// Classify the bounded text shape, not its author, cause, or permission to reuse.
func classifyMetaVerifierCacheStage(line string) string {
	const prefix = "testcache: "
	if !strings.HasPrefix(line, prefix) {
		return "UNRECOGNIZED"
	}
	body := strings.TrimPrefix(line, prefix)
	disabled := "caching disabled for test argument: "
	if strings.HasPrefix(body, disabled) && strings.TrimSpace(strings.TrimPrefix(body, disabled)) != "" {
		return "CONFIGURATION"
	}
	pkg, message, found := strings.Cut(body, ": ")
	if !found || pkg == "" || strings.ContainsAny(pkg, " \t\r\n") {
		return "UNRECOGNIZED"
	}
	switch {
	case message == "input list malformed" || nativeCacheMessageDetail(message, "input list not found: "):
		return "PRIOR_INPUT_LIST_LOOKUP"
	case strings.HasPrefix(message, "input list malformed (") && strings.HasSuffix(message, ")"):
		return "INPUT_LOG_PARSE"
	case message == "test output malformed" || message == "test output expired due to go clean -testcache" ||
		nativeCacheMessageDetail(message, "test output not found: "):
		return "TEST_OUTPUT_LOOKUP"
	default:
		return nativeCacheKeyMessageStage(message)
	}
}

func nativeCacheMessageDetail(message, prefix string) bool {
	return strings.HasPrefix(message, prefix) && strings.TrimSpace(strings.TrimPrefix(message, prefix)) != ""
}

func nativeCacheKeyMessageStage(message string) string {
	fields := strings.Fields(message)
	store := len(fields) > 0 && fields[0] == "save"
	if store {
		fields = fields[1:]
	}
	if (len(fields) != 5 && len(fields) != 9) || fields[0] != "test" || fields[1] != "ID" ||
		!nativeCacheHexText(fields[2]) || fields[3] != "=>" {
		return "UNRECOGNIZED"
	}
	if len(fields) == 5 {
		if store || !nativeCacheHexText(fields[4]) {
			return "UNRECOGNIZED"
		}
		return "PRIOR_INPUT_LIST_LOOKUP"
	}
	if fields[4] != "input" || fields[5] != "ID" || !nativeCacheHexText(fields[6]) ||
		fields[7] != "=>" || !nativeCacheHexText(fields[8]) {
		return "UNRECOGNIZED"
	}
	if store {
		return "STORE_ATTEMPT"
	}
	return "TEST_OUTPUT_LOOKUP"
}

func nativeCacheHexText(value string) bool {
	return value != "" && len(value)%2 == 0 && strings.Trim(value, "0123456789abcdef") == ""
}
