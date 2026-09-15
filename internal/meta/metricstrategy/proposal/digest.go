package proposal

import (
	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	strategyverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/verify"
)

var registry = []CoordinateSpec{
	{ID: "exact-strategy-subject", Class: "DRIVER", ProofChoice: "FOUNDATION", MetaOperation: "bind-exact-strategy-subject"},
	{ID: "verified-strategy-receipt", Class: "DRIVER", ProofChoice: "FOUNDATION", MetaOperation: "verify-strategy-receipt"},
	{ID: "deterministic-strategy-replay", Class: "GUARDRAIL", ProofChoice: "REGRESSION", MetaOperation: "replay-strategy-plan"},
	{ID: "concept-governed-trilemma", Class: "DRIVER", ProofChoice: "COHERENCE", MetaOperation: "bind-concept-trilemma-selection"},
	{ID: "actionable-generation-plan", Class: "OUTCOME", ProofChoice: "COHERENCE", MetaOperation: "propose-independent-meta-operations"},
	{ID: "independent-action-groups", Class: "DRIVER", ProofChoice: "REGRESSION", MetaOperation: "select-independent-action-groups"},
	{ID: "executable-conformance-obligations", Class: "DRIVER", ProofChoice: "FOUNDATION", MetaOperation: "bind-executor-evaluator-receipts"},
	{ID: "read-only-non-authorizing-boundary", Class: "GUARDRAIL", ProofChoice: "FOUNDATION", MetaOperation: "preserve-proposal-boundary"},
}

func Registry() []CoordinateSpec {
	return append([]CoordinateSpec(nil), registry...)
}

func makeCoordinate(index int, status, reason string, evidence any) (Coordinate, error) {
	digest, err := artifact.Digest(evidence)
	return Coordinate{CoordinateSpec: registry[index], Status: status, Reason: reason, EvidenceDigest: digest}, err
}

func validReceipt(value strategyverify.Receipt) bool {
	digest := value.Digest
	value.Digest = ""
	expected, err := artifact.Digest(value)
	return err == nil && digest == expected
}

func registryDigest() (string, error) {
	return artifact.Digest(struct {
		Schema      string           `json:"schema"`
		Coordinates []CoordinateSpec `json:"coordinates"`
	}{RegistrySchema, registry})
}

func sealReport(value Report) (Report, error) {
	value.ReportDigest = ""
	digest, err := artifact.Digest(value)
	value.ReportDigest = digest
	return value, err
}

func decisionFor(summary Summary) (string, string) {
	if summary.Unresolved != 0 {
		return "FAIL_CLOSED", "CHANGE_PROPOSAL_EVIDENCE_UNKNOWN"
	}
	if summary.Satisfied != summary.Total {
		return "NOT_READY", "CHANGE_PROPOSAL_CONTRACT_INCOMPLETE"
	}
	return "PASS", "CHANGE_PROPOSAL_CONTRACT_READY"
}

func coordinateStatus(satisfied, unresolved bool, successReason string) (string, string) {
	if unresolved {
		return "UNRESOLVED", "EVIDENCE_RESOLUTION_UNKNOWN"
	}
	if !satisfied {
		return "NOT_SATISFIED", "CONTRACT_COORDINATE_NOT_PROVEN"
	}
	return "SATISFIED", successReason
}
