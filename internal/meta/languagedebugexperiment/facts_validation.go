package languagedebugexperiment

import "github.com/kimjooyoon/meta-ontology-go/internal/languagedebug"

func validateInput(input Input) (facts, string) {
	if !validSHA(input.SubjectSHA) || !validDigest(input.ExecutableDigest) {
		return unknownFacts("DEBUG_SUBJECT_UNKNOWN", "DEBUGGING", "READ_SUBJECT", "DIRECT_MISSING", "BIND_EXACT_HEAD_AND_BINARY", "SUBJECT_OR_BINARY_DIGEST")
	}
	if !validGraph(input.Graph) {
		return refutedFacts("GRAPH_BINDING", "VERIFY_GOOO_GRAPH", "GOOO_GRAPH_ACTIVITY_OR_EDGE_CONTRADICTION")
	}
	if missingReceipt(input.First) || missingReceipt(input.Second) {
		return unknownFacts("DEBUG_RECEIPT_MISSING", "DEBUGGING", "READ_SECOND_RECEIPT", "DIRECT_MISSING", "REEXECUTE_DEBUG_PATH_TWICE", "SECOND_RECEIPT")
	}
	if input.First.DeterministicDigest() != input.Second.DeterministicDigest() {
		return refutedFacts("DETERMINISTIC_REPLAY", "COMPARE_SEMANTIC_RECEIPT_DIGESTS", "DEBUG_DETERMINISTIC_DIGEST_CONTRADICTION")
	}
	if input.First.Decision != languagedebug.DecisionPass || input.Second.Decision != languagedebug.DecisionPass {
		return unknownFacts("DEBUG_DECISION_UNKNOWN", "DEBUGGING", "CLASSIFY_TOP_DECISION", "UNKNOWN_DECISION", "REQUIRE_PASS_RECEIPT", "TOP_DECISION")
	}
	if languagedebug.Validate(input.First) != nil || languagedebug.Validate(input.Second) != nil || languagedebug.Validate(input.UnknownBreakpoint) != nil {
		return unknownFacts("DEBUG_RECEIPT_INVALID", "DEBUGGING", "VALIDATE_RECEIPTS", "MALFORMED_EVIDENCE", "REEXECUTE_DEBUG_PATH", "RECEIPT_SCHEMA_OR_DIGEST")
	}
	if uncertainty := runtimeUncertainty(input); uncertainty != nil {
		return facts{Unknowns: 1, UnknownCases: []Uncertainty{*uncertainty}}, "DEBUG_RUNTIME_UNKNOWN"
	}
	return facts{}, ""
}

func unknownFacts(reason, stage, step, class, next, blocked string) (facts, string) {
	return facts{Unknowns: 1, UnknownCases: []Uncertainty{unknownCase(stage, step, reason, class, next, blocked)}}, reason
}

func refutedFacts(stage, step, reason string) (facts, string) {
	return facts{RefutedCases: []Refutation{{Stage: stage, Step: step, Reason: reason}}}, reason
}

func unknownCase(stage, step, reason, class, next, blocked string) Uncertainty {
	frontier := []string{}
	if class == "DEPENDENCY_BLOCKED" && blocked != "" {
		frontier = []string{blocked}
	}
	return Uncertainty{Stage: stage, Step: step, Reason: reason, UnknownClass: class, NextOperation: next, BlockedBy: frontier}
}

func missingReceipt(receipt languagedebug.Receipt) bool {
	return receipt.Schema == "" || receipt.Digest == ""
}

func validGraph(graph GraphObservation) bool {
	return graph.Schema == "gooo-graph/v1" && validDigest(graph.ProgramDigest) && validHexDigest(graph.GraphHash) && graph.ActivityCount == 44 && graph.EdgeCount == 88 && graph.DebugActivityCount == 2 && graph.DebugOutputCount == 2 && graph.DebugUsedEdgeCount == 2 && graph.DebugGeneratedEdgeCount == 2 && validDebugGraph(graph)
}
