package artifactfeedback

const (
	ResolutionFeedbackSchema = "gooo/meta-artifact-feedback-resolution/v1"

	DecisionLowerResolution         = "LOWER_RESOLUTION"
	ReasonCoverageDecisionUnknown   = "FEEDBACK_COVERAGE_DECISION_UNKNOWN"
	NextOperationReevaluateFeedback = "reevaluate-artifact-feedback"

	MetricResolutionRecovery = "gooo.metric.meta.semantic-resolution-recovery.coverage-bps.v1"
	MetricConflictObservation = "gooo.metric.meta.semantic-conflict-observation.coverage-bps.v1"
	MetricMonotoneDescent     = "gooo.metric.meta.monotone-resolution-descent.coverage-bps.v1"
	MetricTransitionReplay    = "gooo.metric.meta.resolution-transition-replay.coverage-bps.v1"
	MetricFalseFixedPoint     = "gooo.metric.meta.false-fixed-point-decisions.guardrail.v1"
	MetricResolutionDescents  = "gooo.metric.meta.semantic-resolution-descents.guardrail.v1"
	MetricResolutionWrites    = "gooo.metric.meta.semantic-resolution-writes.guardrail.v1"
)
