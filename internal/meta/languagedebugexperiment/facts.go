package languagedebugexperiment

import (
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/languagedebug"
)

func collectFacts(input Input) (facts, string) {
	if !validSHA(input.SubjectSHA) || !validDigest(input.ExecutableDigest) {
		return facts{Unknowns: 1, UnknownCases: []Uncertainty{unknownCase(
			"DEBUGGING", "READ_SUBJECT", "DEBUG_SUBJECT_UNKNOWN", "MISSING_IDENTITY",
			"BIND_EXACT_HEAD_AND_BINARY", "SUBJECT_OR_BINARY_DIGEST")}}, "DEBUG_SUBJECT_UNKNOWN"
	}
	if !validGraph(input.Graph) {
		return facts{RefutedCases: []Refutation{{Stage: "GRAPH_BINDING", Step: "VERIFY_GOOO_GRAPH", Reason: "GOOO_GRAPH_ACTIVITY_OR_EDGE_CONTRADICTION"}}}, "GOOO_GRAPH_BINDING_REFUTED"
	}
	if missingReceipt(input.First) || missingReceipt(input.Second) {
		return facts{Unknowns: 1, UnknownCases: []Uncertainty{unknownCase(
			"DEBUGGING", "READ_SECOND_RECEIPT", "SECOND_DEBUG_RECEIPT_MISSING", "MISSING_EVIDENCE",
			"REEXECUTE_DEBUG_PATH_TWICE", "SECOND_RECEIPT")}}, "DEBUG_RECEIPT_MISSING"
	}
	firstDeterministic := input.First.DeterministicDigest()
	secondDeterministic := input.Second.DeterministicDigest()
	if firstDeterministic != secondDeterministic {
		return facts{RefutedCases: []Refutation{{Stage: "DETERMINISTIC_REPLAY", Step: "COMPARE_SEMANTIC_RECEIPT_DIGESTS", Reason: "DEBUG_DETERMINISTIC_DIGEST_CONTRADICTION"}}}, "DEBUG_DETERMINISTIC_DIGEST_CONTRADICTION"
	}
	if input.First.Decision != languagedebug.DecisionPass || input.Second.Decision != languagedebug.DecisionPass {
		return facts{Unknowns: 1, UnknownCases: []Uncertainty{unknownCase(
			"DEBUGGING", "CLASSIFY_TOP_DECISION", "DEBUG_DECISION_UNKNOWN", "UNKNOWN_DECISION",
			"REQUIRE_PASS_RECEIPT", "TOP_DECISION")}}, "DEBUG_DECISION_UNKNOWN"
	}
	if languagedebug.Validate(input.First) != nil || languagedebug.Validate(input.Second) != nil ||
		languagedebug.Validate(input.UnknownBreakpoint) != nil {
		return facts{Unknowns: 1, UnknownCases: []Uncertainty{unknownCase(
			"DEBUGGING", "VALIDATE_RECEIPTS", "DEBUG_RECEIPT_MALFORMED", "MALFORMED_EVIDENCE",
			"REEXECUTE_DEBUG_PATH", "RECEIPT_SCHEMA_OR_DIGEST")}}, "DEBUG_RECEIPT_INVALID"
	}
	if uncertainty := runtimeUncertainty(input); uncertainty != nil {
		return facts{Unknowns: 1, UnknownCases: []Uncertainty{*uncertainty}}, "DEBUG_RUNTIME_UNKNOWN"
	}
	result := facts{DebugReceipts: 2, Go127Runtimes: len(input.RuntimeObservations), ReplayMatches: 1, ResourceObservations: len(input.RuntimeObservations)}
	positive := []languagedebug.Receipt{input.First, input.Second}
	digests := map[string]bool{}
	for _, receipt := range positive {
		if receipt.State == languagedebug.StatePaused {
			result.PausedSessions++
		}
		if receipt.CurrentEvent != nil && receipt.CurrentEvent.Kind == receipt.Breakpoint {
			result.BreakpointsReached++
			result.CurrentEvents++
		}
		result.TraceEvents += len(receipt.Trace)
		result.RemainingEvents += receipt.RemainingEvents
		result.RepositoryWrites += receipt.Effects.RepositoryWrites
		result.MutationAuthority = result.MutationAuthority || receipt.Effects.MutationAuthority
		digests[receipt.ExecutionDigest] = true
	}
	result.ExecutionDigestVariants = len(digests)
	result.SubjectCoherence = coherence(input.First, input.Second)
	if slices.Equal(input.First.NonClaims, languagedebug.CanonicalNonClaims()) &&
		slices.Equal(input.Second.NonClaims, languagedebug.CanonicalNonClaims()) {
		result.NonClaims = len(languagedebug.CanonicalNonClaims())
	}
	if input.UnknownBreakpoint.Decision == languagedebug.DecisionFailClosed &&
		input.UnknownBreakpoint.Reason == "DEBUG_BREAKPOINT_NOT_REACHED" {
		result.UnknownBreakpointRejections = 1
	}
	return result, ""
}

func unknownCase(stage, step, reason, class, next, blocked string) Uncertainty {
	return Uncertainty{Stage: stage, Step: step, Reason: reason, UnknownClass: class,
		NextOperation: next, BlockedBy: blocked}
}

func missingReceipt(receipt languagedebug.Receipt) bool {
	return receipt.Schema == "" || receipt.Digest == ""
}

func validGraph(graph GraphObservation) bool {
	return graph.Schema == "gooo-graph/v1" && validDigest(graph.ProgramDigest) && validHexDigest(graph.GraphHash) &&
		graph.ActivityCount == 44 && graph.EdgeCount == 88 && graph.DebugActivityCount == 2 &&
		graph.DebugOutputCount == 2 && graph.DebugUsedEdgeCount == 2 && graph.DebugGeneratedEdgeCount == 2 &&
		validDebugGraph(graph)
}

func validDebugGraph(graph GraphObservation) bool {
	expectedActivities := map[string]bool{
		"languageutility://activity/observe-debugging-deterministic-replay": true,
		"languageutility://activity/observe-debugging-resource-observed":    true,
	}
	if len(graph.DebugActivityIDs) != len(expectedActivities) {
		return false
	}
	seenActivities := map[string]bool{}
	for _, activity := range graph.DebugActivityIDs {
		if !expectedActivities[activity] || seenActivities[activity] {
			return false
		}
		seenActivities[activity] = true
	}
	expectedEdges := map[string]bool{
		"used\x00languageutility://activity/observe-debugging-deterministic-replay\x00gooo://meta/language-utility/entity/cell":               true,
		"wasGeneratedBy\x00gooo://meta/language-utility/entity/evidence\x00languageutility://activity/observe-debugging-deterministic-replay": true,
		"used\x00languageutility://activity/observe-debugging-resource-observed\x00gooo://meta/language-utility/entity/cell":                  true,
		"wasGeneratedBy\x00gooo://meta/language-utility/entity/evidence\x00languageutility://activity/observe-debugging-resource-observed":    true,
	}
	if len(graph.DebugCausalEdges) != len(expectedEdges) {
		return false
	}
	seenEdges := map[string]bool{}
	for _, edge := range graph.DebugCausalEdges {
		key := edge.Relation + "\x00" + edge.Subject + "\x00" + edge.Object
		if !expectedEdges[key] || seenEdges[key] {
			return false
		}
		seenEdges[key] = true
	}
	return len(seenEdges) == len(expectedEdges)
}

func validHexDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, char := range value {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}
	return true
}

func validMeasurement(value Measurement, expectedName string) bool {
	return value.Name == expectedName && value.Executed && value.WallNS > 0 &&
		value.WallMS == (value.WallNS+999999)/1000000 && value.WallMS > 0 &&
		value.PeakRSSKiB > 0 && value.CacheState != ""
}

func runtimeUncertainty(input Input) *Uncertainty {
	if len(input.RuntimeObservations) != input.Contract.ExpectedResourceObservations {
		value := unknownCase("RESOURCE_OBSERVED", "READ_RUNTIME_OBSERVATION", "RUNTIME_OBSERVATION_MISSING", "MISSING_EVIDENCE", "REEXECUTE_DEBUG_PATH_WITH_RESOURCES", "RUNTIME_OBSERVATIONS")
		return &value
	}
	if !validMeasurement(input.Build, "debug-producer-build") || !validMeasurement(input.EvaluatorBuild, "debug-evaluator-build") ||
		!validMeasurement(input.Test, "debug-relevant-tests") {
		value := unknownCase("RESOURCE_OBSERVED", "READ_BUILD_TEST_MEASUREMENTS", "BUILD_OR_TEST_RESOURCE_MISSING", "MISSING_EVIDENCE", "RECORD_CI_MEASUREMENTS", "BUILD_TEST_RECEIPTS")
		return &value
	}
	positive := []languagedebug.Receipt{input.First, input.Second}
	for index, runtime := range input.RuntimeObservations {
		if runtime.Run != index+1 || runtime.RuntimeReceiptSchema != RuntimeReceiptSchema || runtime.Runner == "" || !strings.Contains(runtime.Toolchain, "go1.27") ||
			!validDigest(runtime.SourceRawDigest) || !validDigest(runtime.SourceSemanticDigest) ||
			!validDigest(runtime.BinaryDigest) || len(runtime.Arguments) == 0 || !validSHA(runtime.SubjectSHA) ||
			runtime.SubjectSHA != input.SubjectSHA || !validDigest(runtime.OutputDigest) || runtime.WallNS <= 0 ||
			runtime.WallMS <= 0 || runtime.WallMS != (runtime.WallNS+999999)/1000000 || runtime.PeakRSSKiB <= 0 ||
			runtime.BinaryDigest != input.ExecutableDigest || runtime.SourceRawDigest != positive[index].SourceDigest ||
			runtime.SourceSemanticDigest != positive[index].SemanticDigest {
			value := unknownCase("RESOURCE_OBSERVED", "VALIDATE_RUNTIME_OBSERVATION", "RUNTIME_FIELD_MISSING_OR_CONTRADICTED", "INCOMPLETE_EVIDENCE", "REEXECUTE_DEBUG_PATH_WITH_COMPLETE_RECEIPT", "RUNNER_RESOURCE_RECEIPT")
			return &value
		}
	}
	return nil
}

func coherence(first, second languagedebug.Receipt) int {
	if first.Filename == second.Filename && first.SourceDigest == second.SourceDigest &&
		first.SemanticDigest == second.SemanticDigest && first.ExecutionDigest == second.ExecutionDigest {
		return 2
	}
	return 0
}
