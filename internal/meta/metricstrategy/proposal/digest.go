package proposal

import (
	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	strategyverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/verify"
)

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
