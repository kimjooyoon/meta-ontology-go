package sourceauthoritypromotion

import (
	"encoding/json"
	"fmt"
	"testing"
)

func validInput(t *testing.T) Input {
	t.Helper()
	subject := "fixture-subject-sha"
	definitions := make([]assuranceDefinition, 12)
	obligations := make([]assuranceObligation, 12)
	for index := 0; index < 12; index++ {
		metric := fmt.Sprintf("fixture.metric.%d", index)
		definitions[index] = assuranceDefinition{MetricID: metric, Priority: "P1", Class: "DRIVER", ProofChoice: "FOUNDATION", RequiredMetaOperation: "fixture"}
		obligations[index] = assuranceObligation{MetricID: metric, Status: "NOT_IMPLEMENTED", Resolution: "NONE"}
		if index < 6 { obligations[index].Status, obligations[index].Resolution = "OPERATING", ResolutionExact }
	}
	definitions[6] = assuranceDefinition{MetricID: SourceMetric, Priority: "P1", Class: "DRIVER", ProofChoice: "FOUNDATION", RequiredMetaOperation: SourceOperation}
	obligations[6] = assuranceObligation{MetricID: SourceMetric, Status: "NOT_IMPLEMENTED", Resolution: "NONE"}
	assurance := assuranceDocument{Schema: AssuranceSchema, SubjectSHA: subject,
		DenominatorID: AssuranceDenominator, DenominatorDigest: AssuranceDigest,
		AssuranceDecision: "PARTIAL", CandidateDecision: "ALLOW_LIMITED", Denominator: definitions, Obligations: obligations,
		Summary: assuranceSummary{DenominatorTotal: 12, Operating: 6, NotImplemented: 6, ImplementationCoverageBPS: 5000}}
	snapshot := &upstreamSnapshot{Digest: "sha256:29362aa311de0f24c66f41cc65a8b6ffd996baf37e048b5a72db63172aae5bf2", Bytes: 77,
		Authority: upstreamAuthority{Repository: "cosmos72/gomacro", Revision: "cf0d4bf32da393dbda97e3572f216731013ffa55", Path: "README.md"}, upstreamSelection: upstreamSelection{StartLine: 1, EndLine: 1}}
	indicators := []upstreamIndicator{{Class: "OUTCOME", ProofChoice: "COHERENCE", Satisfied: true},
		{Class: "DRIVER", ProofChoice: "FOUNDATION", Satisfied: true}, {Class: "DRIVER", ProofChoice: "FOUNDATION", Satisfied: true},
		{Class: "GUARDRAIL", ProofChoice: "REGRESSION", Satisfied: true}, {Class: "GUARDRAIL", ProofChoice: "FOUNDATION", Satisfied: true},
		{Class: "GUARDRAIL", ProofChoice: "COHERENCE", Satisfied: true}}
	upstream := upstreamDocument{Schema: UpstreamSchema, SubjectSHA: subject, Decision: "PASS", Resolution: ResolutionExact,
		DenominatorID: UpstreamDenominator, DenominatorDigest: UpstreamDigest,
		Summary: upstreamSummary{CasesTotal: 3, CasesPassed: 3, ExactAllow: 1, FailClosed: 2, CoverageBPS: 10000}}
	upstream.Cases = []upstreamCase{fixtureCase("exact", "SATISFIED", ResolutionExact, "ALLOW", "SOURCE_SNAPSHOT_EXACT", snapshot, indicators),
		fixtureCase("digest-mismatch", "UNKNOWN", ResolutionInvariantOnly, DecisionBlock, "SOURCE_DIGEST_MISMATCH", snapshot, nil),
		fixtureCase("authority-mismatch", "UNKNOWN", ResolutionInvariantOnly, DecisionBlock, "AUTHORITY_SCOPE_MISMATCH", nil, nil)}
	assuranceJSON, _ := json.Marshal(assurance)
	upstreamJSON, _ := json.Marshal(upstream)
	return Input{SubjectSHA: subject, AssuranceJSON: assuranceJSON, UpstreamJSON: upstreamJSON}
}

func fixtureCase(id, observation, resolution, enforcement, reason string, snapshot *upstreamSnapshot, indicators []upstreamIndicator) upstreamCase {
	return upstreamCase{ID: id, ExpectedObservation: observation, ExpectedResolution: resolution,
		ExpectedEnforcement: enforcement, ExpectedReason: reason, Passed: true,
		Receipt: upstreamReceipt{SubjectSHA: "fixture-subject-sha", Observation: observation, Resolution: resolution,
			Enforcement: enforcement, Reason: reason, Snapshot: snapshot, Indicators: indicators}}
}
