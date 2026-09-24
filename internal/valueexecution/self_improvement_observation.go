package valueexecution

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

// ImprovementMetric is an evidence-backed measurement used for one candidate
// comparison. The value is descriptive data, not an authorization decision.
type ImprovementMetric struct {
	Name           string `json:"name"`
	Value          int64  `json:"value"`
	EvidenceDigest string `json:"evidence_digest"`
	LowerIsBetter  bool   `json:"lower_is_better"`
}

// SelfImprovementStatus describes the measured direction of one comparison.
type SelfImprovementStatus string

const (
	SelfImprovementStatusUnknown   SelfImprovementStatus = "UNKNOWN"
	SelfImprovementStatusImproved  SelfImprovementStatus = "IMPROVED"
	SelfImprovementStatusRegressed SelfImprovementStatus = "REGRESSED"
	SelfImprovementStatusUnchanged SelfImprovementStatus = "UNCHANGED"
)

// SelfImprovementObservation binds a candidate comparison to source origin,
// an unchanged environment, and two supplied metric evidence records. It does
// not infer causality, authorize a candidate, or hide a regression.
type SelfImprovementObservation struct {
	Schema                     string                           `json:"schema"`
	OriginTransition           provenance.OriginChainTransition `json:"origin_transition"`
	EnvironmentTransition      EnvironmentTransition            `json:"environment_transition"`
	BeforeOriginDigest         string                           `json:"before_origin_digest,omitempty"`
	AfterOriginDigest          string                           `json:"after_origin_digest,omitempty"`
	EnvironmentDigest          string                           `json:"environment_digest,omitempty"`
	MetricName                 string                           `json:"metric_name,omitempty"`
	BeforeMetricValue          int64                            `json:"before_metric_value"`
	AfterMetricValue           int64                            `json:"after_metric_value"`
	BeforeMetricEvidenceDigest string                           `json:"before_metric_evidence_digest,omitempty"`
	AfterMetricEvidenceDigest  string                           `json:"after_metric_evidence_digest,omitempty"`
	Status                     SelfImprovementStatus            `json:"status"`
	Reason                     string                           `json:"reason"`
	NonAuthorizing             bool                             `json:"non_authorizing"`
	Digest                     string                           `json:"digest"`
}

const SelfImprovementObservationSchema = "gooo/self-improvement-observation/v1"

// ObserveSelfImprovement compares two complete source-backed observations.
// An environment change, incomplete origin, or incomplete metric evidence
// remains UNKNOWN so a lower number cannot be mistaken for improvement.
func ObserveSelfImprovement(
	beforeOrigin, afterOrigin provenance.OriginChainObservation,
	beforeEnvironment, afterEnvironment EnvironmentObservation,
	beforeMetric, afterMetric ImprovementMetric,
) SelfImprovementObservation {
	beforeMetric = normalizeImprovementMetric(beforeMetric)
	afterMetric = normalizeImprovementMetric(afterMetric)
	originTransition := provenance.CompareOriginChains(beforeOrigin, afterOrigin)
	environmentTransition := CompareEnvironments(beforeEnvironment, afterEnvironment)
	observation := SelfImprovementObservation{
		Schema:                     SelfImprovementObservationSchema,
		OriginTransition:           originTransition,
		EnvironmentTransition:      environmentTransition,
		BeforeOriginDigest:         beforeOrigin.Digest,
		AfterOriginDigest:          afterOrigin.Digest,
		EnvironmentDigest:          afterEnvironment.Digest,
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
	case environmentTransition == EnvironmentTransitionUnknown:
		observation.Reason = "ENVIRONMENT_COMPARISON_UNKNOWN"
	case environmentTransition == EnvironmentTransitionChanged:
		observation.Reason = "ENVIRONMENT_CHANGED"
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

func normalizeImprovementMetric(metric ImprovementMetric) ImprovementMetric {
	metric.Name = strings.TrimSpace(metric.Name)
	metric.EvidenceDigest = strings.TrimSpace(metric.EvidenceDigest)
	return metric
}

func validImprovementMetric(metric ImprovementMetric) bool {
	return metric.Name != "" && validDigest(metric.EvidenceDigest)
}

func selfImprovementObservationDigest(observation SelfImprovementObservation) string {
	observation.Digest = ""
	return digestValue(observation)
}
