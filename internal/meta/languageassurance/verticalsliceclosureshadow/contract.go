package verticalsliceclosureshadow

const (
	Schema            = "gooo/vertical-slice-closure-shadow/v1"
	MetricID          = "gooo.metric.capability.vertical-slice-closure.v1"
	MetaOperation     = "close-vertical-slice"
	PredecessorSHA    = "145b81c8bb8e4b1eb46cb10af0ea21a6b6be51b5"
	AssuranceDigest   = "sha256:13581ebf64e0e3a512d1e8b3ca05de05e14d4453b64f3c7eff8e3b854a89d969"
	DenominatorDigest = "sha256:b8d336e45d0bc3aaf8e23b33e5f168d73ecd721596b47bd95489ee6a96e41709"
	// DenominatorMigrationDigest is the append-only v22 contract used when the
	// language-syntax corpus exposes the additional capability case.
	DenominatorMigrationDigest = "sha256:ea4821d2b632319b29ffe43276c0b36f1b825fe29c8617c16680cd41b3d5f822"
	// DenominatorMigrationV23Digest is the append-only v23 contract used when
	// the inherited minimal-loop capability is registered in the corpus.
	DenominatorMigrationV23Digest = "sha256:122db17eb5ff8236e6d7fdd6ba175429c70543def3af420240344aa65fe8ca0d"
	// DenominatorMigrationV24Digest is the append-only v24 contract used when
	// the compiler self-improvement capability is registered in the corpus.
	DenominatorMigrationV24Digest = "sha256:e22d8c3ba0ba2440f42d4d2028cf9f61d833ae4db0e0d9f10ddf42a0af4973ce"
	// DenominatorMigrationV25Digest is the append-only v25 contract used when
	// the semantic operation envelope capability is registered in the corpus.
	DenominatorMigrationV25Digest = "sha256:5ef82bd6ca0a861a054d67d37176dd44fd3fd67b25290647fde290c42b916795"
	// DenominatorMigrationV26Digest is the append-only v26 contract used when
	// the compiler operation envelope capability is registered in the corpus.
	DenominatorMigrationV26Digest = "sha256:18cf03b8b7bef829bebde8fefddf59c6a231dbd46760347c17452b867c722009"

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
