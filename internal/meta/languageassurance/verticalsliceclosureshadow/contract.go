package verticalsliceclosureshadow

const (
	Schema            = "gooo/vertical-slice-closure-shadow/v1"
	MetricID          = "gooo.metric.capability.vertical-slice-closure.v1"
	MetaOperation     = "close-vertical-slice"
	PredecessorSHA    = "145b81c8bb8e4b1eb46cb10af0ea21a6b6be51b5"
	AssuranceDigest   = "sha256:13581ebf64e0e3a512d1e8b3ca05de05e14d4453b64f3c7eff8e3b854a89d969"
	DenominatorDigest = "sha256:c3ce8615927ef1529a0e8fafbc7ed5068e21bb6b2a8c9c5c6f17bb06463f2a98"

	DecisionShadowPass  = "SHADOW_PASS"
	DecisionFailClosed  = "FAIL_CLOSED"
	ResolutionExact     = "EXACT"
	ResolutionLower     = "LOWER_RESOLUTION"
	ResolutionInvariant = "INVARIANT_ONLY"
	EnforcementNoEffect = "NO_EFFECT"

	ReasonShadowPass       = "VERTICAL_SLICE_CLOSURE_SHADOW_PROVEN"
	ReasonAssuranceMissing = "VERTICAL_SLICE_ASSURANCE_UNAVAILABLE"
	ReasonAssuranceDigest  = "VERTICAL_SLICE_ASSURANCE_DIGEST_MISMATCH"
	ReasonAssuranceBase    = "VERTICAL_SLICE_ASSURANCE_BASELINE_MISMATCH"
	ReasonDenominator      = "VERTICAL_SLICE_DENOMINATOR_MISMATCH"
	ReasonEvidenceUnknown  = "VERTICAL_SLICE_BOUNDARY_EVIDENCE_UNKNOWN"
	ReasonBoundaryBlocked  = "VERTICAL_SLICE_BOUNDARY_BLOCKED"

	StatusSatisfied = "SATISFIED"
	StatusUnknown   = "UNKNOWN"
	StatusBlocked   = "BLOCKED"

	boundaryTotal        = 6
	linkTotal            = 12
	officialTotal        = 12
	beforeOperating      = 10
	projectedOperating   = 11
	beforeCoverageBPS    = 8333
	projectedCoverageBPS = 9166
)
