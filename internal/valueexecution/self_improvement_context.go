package valueexecution

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

// ImprovementEvaluationContext excludes candidate source and semantic
// fingerprints so two different candidates can be compared under the same
// toolchain, contract, model, skill, and gateway conditions.
type ImprovementEvaluationContext struct {
	ToolchainDigest     string `json:"toolchain_digest"`
	ContractDigest      string `json:"contract_digest"`
	ModelDigest         string `json:"model_digest"`
	SkillDigest         string `json:"skill_digest"`
	GatewayPolicyDigest string `json:"gateway_policy_digest"`
}

type ImprovementContextStatus string

const (
	ImprovementContextStatusUnknown ImprovementContextStatus = "UNKNOWN"
	ImprovementContextStatusReady   ImprovementContextStatus = "READY"
)

// ImprovementContextObservation is the fixed evaluation boundary for a
// candidate comparison. It is evidence, not authorization.
type ImprovementContextObservation struct {
	Schema         string                       `json:"schema"`
	Inputs         ImprovementEvaluationContext `json:"inputs"`
	Status         ImprovementContextStatus     `json:"status"`
	Missing        []string                     `json:"missing,omitempty"`
	Reason         string                       `json:"reason"`
	Digest         string                       `json:"digest,omitempty"`
	NonAuthorizing bool                         `json:"non_authorizing"`
}

const ImprovementContextObservationSchema = "gooo/self-improvement-context/v1"

// ObserveImprovementEvaluationContext projects only the conditions that must
// remain fixed while source-backed candidates are compared.
func ObserveImprovementEvaluationContext(environment EnvironmentObservation) ImprovementContextObservation {
	inputs := ImprovementEvaluationContext{
		ToolchainDigest:     strings.TrimSpace(environment.Inputs.ToolchainDigest),
		ContractDigest:      strings.TrimSpace(environment.Inputs.ContractDigest),
		ModelDigest:         strings.TrimSpace(environment.Inputs.ModelDigest),
		SkillDigest:         strings.TrimSpace(environment.Inputs.SkillDigest),
		GatewayPolicyDigest: strings.TrimSpace(environment.Inputs.GatewayPolicyDigest),
	}
	observation := ImprovementContextObservation{
		Schema:         ImprovementContextObservationSchema,
		Inputs:         inputs,
		Status:         ImprovementContextStatusUnknown,
		Reason:         "EVALUATION_CONTEXT_INCOMPLETE",
		NonAuthorizing: true,
		Missing:        missingImprovementContextInputs(inputs),
	}
	if len(observation.Missing) == 0 {
		observation.Status = ImprovementContextStatusReady
		observation.Reason = "EVALUATION_CONTEXT_BOUND"
		observation.Digest = improvementContextDigest(observation)
	}
	return observation
}

func CompareImprovementContexts(before, after ImprovementContextObservation) EnvironmentTransition {
	if before.Status != ImprovementContextStatusReady || after.Status != ImprovementContextStatusReady ||
		!validDigest(before.Digest) || !validDigest(after.Digest) {
		return EnvironmentTransitionUnknown
	}
	if before.Digest == after.Digest {
		return EnvironmentTransitionUnchanged
	}
	return EnvironmentTransitionChanged
}

// ObserveSelfImprovementWithContext compares candidates whose source and
// semantic identities may differ while their evaluation context remains
// stable. The returned EnvironmentDigest is the fixed context digest.
func ObserveSelfImprovementWithContext(
	beforeOrigin, afterOrigin provenance.OriginChainObservation,
	beforeContext, afterContext ImprovementContextObservation,
	beforeMetric, afterMetric ImprovementMetric,
) SelfImprovementObservation {
	beforeMetric = normalizeImprovementMetric(beforeMetric)
	afterMetric = normalizeImprovementMetric(afterMetric)
	originTransition := provenance.CompareOriginChains(beforeOrigin, afterOrigin)
	contextTransition := CompareImprovementContexts(beforeContext, afterContext)
	observation := SelfImprovementObservation{
		Schema:                     SelfImprovementObservationSchema,
		OriginTransition:           originTransition,
		EnvironmentTransition:      contextTransition,
		BeforeOriginDigest:         beforeOrigin.Digest,
		AfterOriginDigest:          afterOrigin.Digest,
		EnvironmentDigest:          afterContext.Digest,
		MetricName:                 beforeMetric.Name,
		BeforeMetricValue:          beforeMetric.Value,
		AfterMetricValue:           afterMetric.Value,
		BeforeMetricEvidenceDigest: beforeMetric.EvidenceDigest,
		AfterMetricEvidenceDigest:  afterMetric.EvidenceDigest,
		Status:                     SelfImprovementStatusUnknown,
		NonAuthorizing:             true,
	}
	switch {
	case originTransition == provenance.OriginChainTransitionUnknown:
		observation.Reason = "ORIGIN_COMPARISON_UNKNOWN"
	case contextTransition == EnvironmentTransitionUnknown:
		observation.Reason = "EVALUATION_CONTEXT_UNKNOWN"
	case contextTransition == EnvironmentTransitionChanged:
		observation.Reason = "EVALUATION_CONTEXT_CHANGED"
	case !validImprovementMetric(beforeMetric) || !validImprovementMetric(afterMetric):
		observation.Reason = "METRIC_EVIDENCE_INCOMPLETE"
	case beforeMetric.Name != afterMetric.Name:
		observation.Reason = "METRIC_NAME_MISMATCH"
	case beforeMetric.LowerIsBetter != afterMetric.LowerIsBetter:
		observation.Reason = "METRIC_DIRECTION_MISMATCH"
	case beforeMetric.Value == afterMetric.Value:
		observation.Status = SelfImprovementStatusUnchanged
		observation.Reason = "METRIC_UNCHANGED"
	case beforeMetric.LowerIsBetter && afterMetric.Value < beforeMetric.Value:
		observation.Status = SelfImprovementStatusImproved
		observation.Reason = "METRIC_IMPROVED"
	case !beforeMetric.LowerIsBetter && afterMetric.Value > beforeMetric.Value:
		observation.Status = SelfImprovementStatusImproved
		observation.Reason = "METRIC_IMPROVED"
	default:
		observation.Status = SelfImprovementStatusRegressed
		observation.Reason = "METRIC_REGRESSED"
	}
	observation.Digest = selfImprovementObservationDigest(observation)
	return observation
}

func missingImprovementContextInputs(inputs ImprovementEvaluationContext) []string {
	missing := make([]string, 0, 5)
	values := []struct {
		name  string
		value string
	}{
		{name: "toolchain_digest", value: inputs.ToolchainDigest},
		{name: "contract_digest", value: inputs.ContractDigest},
		{name: "model_digest", value: inputs.ModelDigest},
		{name: "skill_digest", value: inputs.SkillDigest},
		{name: "gateway_policy_digest", value: inputs.GatewayPolicyDigest},
	}
	for _, item := range values {
		if !validDigest(item.value) {
			missing = append(missing, item.name)
		}
	}
	return missing
}

func improvementContextDigest(observation ImprovementContextObservation) string {
	observation.Digest = ""
	return digestValue(observation)
}
