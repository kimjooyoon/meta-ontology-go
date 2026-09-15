package generation

import (
	"fmt"
	"strings"
)

func buildSemanticOperationIR(source []byte, scenarioID string) (SemanticOperationIR, EnvelopeMetrics, error) {
	plan, err := semanticOperationScenario(scenarioID)
	if err != nil {
		return SemanticOperationIR{}, EnvelopeMetrics{}, err
	}
	authorityDigest := envelopeDigestBytes(source)
	intent := OperationIntent{
		OperationID:   "operation-envelope/" + scenarioID,
		ScenarioID:    scenarioID,
		RequestedMode: "semantic-decision",
	}
	sourceRevision := SourceRevision{ID: "source-r1", Digest: authorityDigest}
	grant := EffectGrant{
		GrantID:  "grant/" + scenarioID,
		ParentID: "root-grant",
		Effects:  append([]string(nil), plan.GrantedEffects...),
	}
	var request *EffectRequest
	if len(plan.RequestedEffects) > 0 {
		request = &EffectRequest{
			RequestID:      "request/" + scenarioID,
			OperationID:    intent.OperationID,
			SourceRevision: sourceRevision.ID,
			Effects:        append([]string(nil), plan.RequestedEffects...),
			ReplayIdentity: "replay/" + scenarioID,
		}
		request.PayloadDigest = envelopeDigestString(fmt.Sprintf("%s|%s|%s", request.RequestID, request.SourceRevision, strings.Join(request.Effects, ",")))
	}
	var result *EffectResult
	if plan.ResultPresent {
		result = &EffectResult{
			ResultID:       "result/" + scenarioID,
			RequestID:      "request/" + scenarioID,
			SourceRevision: plan.ResultSourceRevision,
			Effects:        append([]string(nil), plan.ResultEffects...),
			PayloadDigest:  envelopeDigestString(fmt.Sprintf("result|%s|%s", scenarioID, plan.ResultSourceRevision)),
			ArtifactDigest: envelopeDigestString("artifact|" + scenarioID),
		}
	}
	replay := ReplayIdentity{Identity: "replay/" + scenarioID}
	if request != nil && plan.ReplayCompared {
		replay.Compared = true
		replay.CurrentRequestDigest = envelopeDigestJSON(*request)
		replay.PreviousRequestDigest = replay.CurrentRequestDigest
		if plan.PreviousRequestDigest != "" {
			replay.PreviousRequestDigest = plan.PreviousRequestDigest
		}
	}
	decision := classifySemanticOperation(request, result, grant, sourceRevision, replay)
	ir := SemanticOperationIR{
		Schema:          SemanticOperationEnvelopeSchema,
		Intent:          intent,
		Source:          sourceRevision,
		Grant:           grant,
		Request:         request,
		Result:          result,
		Replay:          replay,
		Decision:        decision,
		Activities:      SemanticOperationActivityNames(),
		AuthorityDigest: authorityDigest,
		ToolchainDigest: envelopeDigestString(semanticOperationToolchainDigest),
	}
	metrics := semanticOperationMetrics(ir)
	metrics.InputGoooPhysicalLines = countPhysicalLines(source)
	return ir, metrics, nil
}

func classifySemanticOperation(request *EffectRequest, result *EffectResult, grant EffectGrant, source SourceRevision, replay ReplayIdentity) OperationDecision {
	if request != nil && result != nil && !envelopeSubset(result.Effects, grant.Effects) {
		return OperationDecision{Decision: "REFUTED", Reason: "EFFECT_ESCALATION"}
	}
	if replay.Compared && replay.CurrentRequestDigest != replay.PreviousRequestDigest {
		return OperationDecision{Decision: "REFUTED", Reason: "REPLAY_COLLISION"}
	}
	if request != nil && result == nil {
		return OperationDecision{
			Decision: "UNKNOWN",
			Reason:   "DIRECT_MISSING",
			Unknown: &EnvelopeUnknownState{
				Stage: "result-verification", Step: "RecordEffectResult", Reason: "DIRECT_MISSING",
				UnknownClass: "DIRECT_MISSING", NextOperation: "submit-effect-result", BlockedBy: []string{"effect-result"},
			},
		}
	}
	if request != nil && result != nil && result.SourceRevision != source.ID {
		return OperationDecision{
			Decision: "UNKNOWN",
			Reason:   "STALE",
			Unknown: &EnvelopeUnknownState{
				Stage: "source-freshness", Step: "BindSourceRevision", Reason: "STALE",
				UnknownClass: "STALE", NextOperation: "bind-current-source-revision", BlockedBy: []string{"source-revision"},
			},
		}
	}
	return OperationDecision{Decision: "CLOSED", Reason: "EXACT_RESULT"}
}

func semanticOperationMetrics(ir SemanticOperationIR) EnvelopeMetrics {
	metrics := EnvelopeMetrics{
		OperationRequests:      boolToInt(ir.Request != nil),
		OperationResults:       boolToInt(ir.Result != nil),
		EffectsGranted:         len(ir.Grant.Effects),
		InputDescendantDirs:    0,
		InputRegularFiles:      1,
		InputGoPhysicalLines:   0,
		InputGoooPhysicalLines: 0,
		OutputArtifactFiles:    6,
		PeakRSSKib:             0,
		WallMS:                 0,
		RepositoryWrites:       0,
		LocalTestExecutions:    0,
	}
	if ir.Result != nil && ir.Result.SourceRevision == ir.Source.ID {
		metrics.EffectsUsed = len(ir.Result.Effects)
	}
	if ir.Replay.Compared {
		metrics.ReplayComparisons = 1
		if ir.Replay.CurrentRequestDigest != ir.Replay.PreviousRequestDigest {
			metrics.ReplayMismatches = 1
		}
	}
	if ir.Result != nil && ir.Result.SourceRevision != ir.Source.ID {
		metrics.StaleRejections = 1
	}
	if ir.Result != nil && !envelopeSubset(ir.Result.Effects, ir.Grant.Effects) {
		metrics.EffectEscalationsRefuted = 1
	}
	return metrics
}
