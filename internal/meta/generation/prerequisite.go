package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

// CandidatePrerequisite records a read-only observation, not a completed repair.
// SourceDigest identifies the bytes inspected; MetricsDigest includes topology.
type CandidatePrerequisite struct {
	IndicatorID            string   `json:"indicator_id"`
	HeadSHA                string   `json:"head_sha"`
	SourceDigest           string   `json:"source_digest,omitempty"`
	MetricsDigest          string   `json:"metrics_digest"`
	ContractSourceDigest   string   `json:"contract_source_digest"`
	ContractSemanticDigest string   `json:"contract_semantic_digest"`
	State                  string   `json:"state"`
	Reason                 string   `json:"reason"`
	NextOperation          string   `json:"next_operation"`
	Stage                  string   `json:"stage,omitempty"`
	Step                   string   `json:"step,omitempty"`
	UnknownClass           string   `json:"unknown_class,omitempty"`
	BlockedBy              []string `json:"blocked_by,omitempty"`
	Diagnostics            []string `json:"diagnostics,omitempty"`
}

// PrerequisiteObserver is called lazily for source rewrite candidates and operations
// without a native execution route, before their independence group is claimed.
// It must not modify source or metrics.
type PrerequisiteObserver func(Action) CandidatePrerequisite

func NewCandidatePrerequisite(action Action, head, sourceDigest, metricsDigest, reason string) CandidatePrerequisite {
	state, next := prerequisiteDisposition(reason)
	return CandidatePrerequisite{IndicatorID: action.IndicatorID, HeadSHA: head,
		SourceDigest: sourceDigest, MetricsDigest: metricsDigest,
		ContractSourceDigest:   action.InputContractSourceDigest,
		ContractSemanticDigest: action.InputContractSemanticDigest,
		State:                  state, Reason: reason, NextOperation: next}
}

func prerequisiteDisposition(reason string) (string, string) {
	switch reason {
	case "EXECUTION_CONSUMER_UNAVAILABLE":
		return "UNKNOWN", "connect-native-execution-consumer"
	case "SPLIT_PLAN_READY":
		return "READY", "execute-observed-split"
	case "EXTRACTION_PLAN_READY":
		return "READY", "execute-observed-extraction"
	case "RENDERED_DECLARATION_CAPACITY":
		return "BLOCKED", "extract-function-before-split"
	case "LOGICAL_DIRECTORY_CAPACITY", "MIXED_LOGICAL_TOPOLOGY":
		return "BLOCKED", "partition-logical-directory-before-split"
	case "SPLIT_NOT_REQUIRED":
		return "BLOCKED", "reconcile-source-metrics"
	case "DIRECTORY_OBSERVATION_MISSING":
		return "UNKNOWN", "observe-logical-directory"
	case "SUCCESSOR_PROJECTION_REQUIRED":
		return "DEFERRED", "materialize-and-verify-successor-projection"
	case "SOURCE_PLANNING_UNAVAILABLE", "SOURCE_CHANGED_DURING_PLANNING":
		return "UNKNOWN", "observe-source-prerequisites"
	default:
		return "", ""
	}
}

func validCandidatePrerequisite(observed CandidatePrerequisite, action Action, head string) bool {
	state, next := prerequisiteDisposition(observed.Reason)
	validState := validPrerequisiteDomain(observed, action) && state != "" &&
		observed.State == state && observed.NextOperation == next &&
		observed.Stage == "" && observed.Step == "" && observed.UnknownClass == "" &&
		len(observed.BlockedBy) == 0 && len(observed.Diagnostics) == 0
	if action.Operation == sourcepolicy.OperationExtractFunction && observed.Reason != "EXECUTION_CONSUMER_UNAVAILABLE" {
		validState = validExtractionPrerequisite(observed)
	}
	return validState &&
		observed.IndicatorID == action.IndicatorID && observed.HeadSHA == head &&
		observed.ContractSourceDigest == action.InputContractSourceDigest &&
		observed.ContractSemanticDigest == action.InputContractSemanticDigest &&
		validDigest(observed.MetricsDigest) &&
		(validDigest(observed.SourceDigest) || observed.State == "UNKNOWN" && observed.SourceDigest == "")
}

func validPrerequisiteDomain(observed CandidatePrerequisite, action Action) bool {
	if observed.Reason == "EXECUTION_CONSUMER_UNAVAILABLE" {
		return NativeExecutionRoute(action.Operation) == "" && observed.SourceDigest == ""
	}
	return action.Operation == sourcepolicy.OperationSplitGo && observed.Reason != "EXTRACTION_PLAN_READY"
}
