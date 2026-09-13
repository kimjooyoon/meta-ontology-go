package main

import "time"

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
