package languagesourcebindingpromotion

import (
	"encoding/json"
	"fmt"
)

func producerFixture(head, receiptDigest, artifactDigest string) producerEnvelope {
	cases := fmt.Sprintf(`[{"id":"execute-billing","status":"SATISFIED","evidence_digest":%q}]`, artifactDigest)
	value := producerEnvelope{Schema: "gooo/language-source-execution-artifact/v1", HeadSHA: head,
		Decision: DecisionPass, Resolution: ResolutionExact, Reason: "SOURCE_EXECUTION_CONTRACT_SATISFIED",
		ContractDigest: receiptDigest, Cases: json.RawMessage(cases),
		Summary: json.RawMessage(`{"cases_satisfied":4,"cases_total":4}`), Indicators: json.RawMessage(`[]`),
		Proofs: json.RawMessage(`[]`), NotClaimed: json.RawMessage(`[]`)}
	value.Digest = digestJSON(value)
	return value
}

func oracleFixture(head string, receipt receiptEnvelope, artifactDigest string) oracleEnvelope {
	cases := []oracleCase{{ID: "genuine-source-bound", Status: "SATISFIED", ObservedDecision: DecisionPass,
		ObservedResolution: ResolutionExact, ObservedReason: "ARTIFACT_SOURCE_PROJECTION_EXACT",
		SourceDigest: receipt.SourceDigest, ArtifactDigest: artifactDigest}, {}, {}, {}}
	casesRaw, _ := json.Marshal(cases)
	value := oracleEnvelope{Schema: "gooo/language-artifact-oracle/v1", Scope: "SOURCE_EXECUTION_ARTIFACT_BINDING_ONLY",
		HeadSHA: head, Decision: DecisionPass, Resolution: ResolutionExact, Reason: "ARTIFACT_ORACLE_CONTRACT_SATISFIED",
		ContractDigest: receipt.Digest, IndependenceDigest: receipt.Digest, LegacyDigest: receipt.Digest,
		Cases: casesRaw, Summary: json.RawMessage(`{"cases_satisfied":4,"cases_total":4,"producer_dependencies":0,"semantic_correctness_claims":0}`),
		Indicators: json.RawMessage(`[]`), NotClaimed: json.RawMessage(`[]`)}
	value.Digest = digestJSON(value)
	return value
}
