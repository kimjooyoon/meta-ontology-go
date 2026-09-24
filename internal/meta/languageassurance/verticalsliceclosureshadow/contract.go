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
	// DenominatorMigrationV27Digest is the append-only v27 contract used when
	// the public self-observation discovery capabilities are registered in the corpus.
	DenominatorMigrationV27Digest = "sha256:be246116defd88b0150cbe221b17b2d30286482c68d6cfbe95b8629fdf264fe0"
	// V28 binds the existing 56 language capabilities. CI continuation is
	// a governance observation and does not increase this denominator.
	DenominatorMigrationV28Digest = "sha256:65fcb1dae77aacfd2e092fc4ab036099f89fae4504171874a154d656c2cf397d"
	// V29 binds the existing 57 language capabilities after the runtime
	// binding fixture was registered in the syntax corpus.
	DenominatorMigrationV29Digest = "sha256:554dde79ad5108152351782ce36fd5e60b024b2d5ccb88ba9a8ac192db3af6d3"
	// V30 binds the existing 58 language capabilities after the source-bound
	// record transport fixture was registered in the syntax corpus.
	DenominatorMigrationV30Digest = "sha256:9c19b37f8f28fb534f82b82570ddccc5ca31dbd36e50df5a810b8cb920a201ed"
	// V31 binds the two domain-observation capabilities registered in the syntax corpus.
	DenominatorMigrationV31Digest = "sha256:dc340f0762355614683760a0c3ac0152cbad73c847ccaae2f1e720c87e3d1f50"
	// V32 binds the three observed domain-syntax capabilities registered in the syntax corpus.
	DenominatorMigrationV32Digest = "sha256:9f06d115bc42f534e17fb34d276ad729e00e905eb547a6776be83befe9c92989"
	// V33 binds the int.sign value witness registered in the syntax corpus.
	DenominatorMigrationV33Digest = "sha256:99bf0a1c500684969a3fbe208bb18583e985ef3c37a38f535072ab25158cbfcd"
	// V34 binds the int.mod value witness registered in the syntax and semantic corpora.
	DenominatorMigrationV34Digest = "sha256:15015d1f028d254513ee97ab36fed60ddf9cbcf97f61e19957a4e7bb974795e5"
	// V35 binds the int.neg value witness registered in the syntax and semantic corpora.
	DenominatorMigrationV35Digest = "sha256:ccbe1f2aec8512d535446b1626ab752b2b442c7f26fbafe125c2f93a44f969a7"
	// V36 binds the int.abs value witness registered in the syntax and semantic corpora.
	DenominatorMigrationV36Digest = "sha256:07ef17811672785448a4ef38baec9b970d116024ce6fac5afae8d5b187a82b1c"
	// V37 binds the parameterized int.add:2 value witness registered in the syntax and semantic corpora.
	DenominatorMigrationV37Digest = "sha256:e014188ea0aac72592d17a30fbb6210d50553995bb9628d19eb57291c97abbe2"
	// V38 binds the int.mul:3 value witness registered in the syntax and semantic corpora.
	DenominatorMigrationV38Digest = "sha256:f2db6d69f4bb0ba9fb50a2e8e8b4054bdacbe048470ce21a16e0b8cf72575952"
	// V39 binds the int.div:2 value witness and its fail-closed divisor boundary.
	DenominatorMigrationV39Digest = "sha256:aa81b4ccb99a2a23179f59f0706a534f4fef0a69ad9336fb3f94207e4d787b33"
	// V40 binds the int.max:0 comparison witness registered in the syntax and semantic corpora.
	DenominatorMigrationV40Digest = "sha256:cd8913a5b59c3ef5854722f2049ed94e1b9b153b8b26ff2764ae385c5be40a6d"
	// V41 binds the inverse int.min:0 comparison witness and corrects the max scope provenance.
	DenominatorMigrationV41Digest = "sha256:f16924dbb2385fb88a64d497de5d193266a09d3df4c2bbcc14d7b1c2c226f425"
	// V42 binds the typed Integer-to-Boolean int.iszero:0 comparison witness.
	DenominatorMigrationV42Digest = "sha256:1538f5a331e6e436aee9ac1e0b29d5940b56dbf416645f9acab8afe30bf8d68f"
	// V43 binds the existing 75 language capabilities after the syntax registry denominator was restored.
	DenominatorMigrationV43Digest = "sha256:cc71098af93f003e6e5c01e845ac5dba204acfdb866bf2a46ada1fc29ff2142c"
	// V44 binds the two Boolean-and language capabilities registered after the v43 snapshot.
	DenominatorMigrationV44Digest = "sha256:3af46b88345fb50ba3b032a9112cab9781ac1cb9deefff40adabbd4cd4c13b08"
	// V45 binds the corresponding two semantic cases added with the Boolean-and capabilities.
	DenominatorMigrationV45Digest = "sha256:676ca467e080dad459a9a7f5c5b0d82f6ba2eb136ba73a8ec13d9165db4ad624"

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