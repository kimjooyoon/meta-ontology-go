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
	linkTotal            = 13
	officialTotal        = 12
	beforeOperating      = 10
	projectedOperating   = 11
	beforeCoverageBPS    = 8333
	projectedCoverageBPS = 9166
	// DenominatorMigrationV46Digest binds the capability-79 denominator snapshot.
	DenominatorMigrationV46Digest = "sha256:e4e6a4a603c035a1fded313beed3b0ed121a0fe93a80541f691845f44135e2a0"
	// DenominatorMigrationV47Digest binds the expanded four-platform release denominator.
	DenominatorMigrationV47Digest = "sha256:f2e749bb032eb8e83d3e467f049091ec18ecb4562517788b16697bd095597dd4"
)

const DenominatorMigrationV48Digest = "sha256:2b9ac5c2293e7c61f23de5e0fbb1ca59d843e43ed9b78cba21054173fdaabfab"

// V49 registers the ordered batch planner and cross-activity list handoff.
const DenominatorMigrationV49Digest = "sha256:aafb7676a5c22606f120f179a23af86ab760ea545ed8cfd31ece275221fd5669"

const DenominatorMigrationV50Digest = "sha256:23dc86403c183f670592830888982d6f0d5400f9b3872802f1bed4966c475a1f"

const DenominatorMigrationV51Digest = "sha256:a686959b74edb42d15e8614d5c55dca1c10a587b8f8a9c8fac52d71a971e83c3"

const DenominatorMigrationV52Digest = "sha256:14d07aa72214a1487878fef1d0bdfd28ca745ba8f9bfdb63e5341a881454e1af"

const DenominatorMigrationV53Digest = "sha256:c37c31268a9e1be706d4fd6ef50dca4d495db9c8311b824df9dc6b8331b229ba"

const DenominatorMigrationV54Digest = "sha256:ae130583ec041538e455cf6377312bf5af1a9398a217af785205a1e0019edfbf"

const DenominatorMigrationV55Digest = "sha256:9dc1c5e506a4ea33ba64b62d1f65d6c1f2c16a6c26ef1aedeb51fafee876b787"

const DenominatorMigrationV56Digest = "sha256:7d943a9f8f804b2fb765dc462cba20e320a382f54dccaaf6432a9acd88681262"

const DenominatorMigrationV57Digest = "sha256:d5fd5b6bf8203b4d1466c03d2d6f4112881002e6a281791564bf689bbe93f6c2"

const DenominatorMigrationV58Digest = "sha256:55f249dc9d663931fd5373322890cee91a2f721f30634db993d208a614ac1ec6"

const DenominatorMigrationV59Digest = "sha256:05a004c06f8a91a8dfaef944d107290bca1d43f08b120f4636fa12c081639cea"

const DenominatorMigrationV60Digest = "sha256:b9036337d010ab4bbbd7fcc68a401e94a498c5fd3f929e7395c3ab671b47c349"

const DenominatorMigrationV61Digest = "sha256:7ab8dba830e9f909f9607f6f0f52afe84e7b03af78e82cb9ceac4971cb264351"

const DenominatorMigrationV62Digest = "sha256:1f8860e090a1655306c24b4843ad6d35e2e3491c86b0be9545ebecd2b61d7e16"

const DenominatorMigrationV63Digest = "sha256:b30ea616a631e1f12b519794ceff0a982dceaa82b89d5f0b2c4d85717b4ee95e"

const DenominatorMigrationV64Digest = "sha256:ba8110fd264fb7d8f9c2c1e65b504d7b9465911488eb32681b484bba03bb4169"

const DenominatorMigrationV65Digest = "sha256:992845cfa77e3d927b97f047ad9879127f36b972636b4776cedcc991984b3e38"

const DenominatorMigrationV66Digest = "sha256:6f275e2485b3a741ec3f0a7301fdd50b76a9db0dffeb82837a728c4d217fbba9"

const DenominatorMigrationV67Digest = "sha256:3917b9bd0ba18fe1d54fcbae8a1b7730d97ed7ea0597ac8c86f75c9c4af71d96"

const DenominatorMigrationV68Digest = "sha256:bb19d506c2a50758aa02dfcc6231f4d521ac7326281982a15f60e6212554e145"

const DenominatorMigrationV69Digest = "sha256:b0529a6277c7278a6c88618b3691d7d0e035aa00186f068c535edc06c59c1cad"

const DenominatorMigrationV70Digest = "sha256:da53a55cd60bedc139226e9aa44b0392a981b8cab6b7b84c2b30b60922ca8dda"

const DenominatorMigrationV71Digest = "sha256:8aab49767149c43a77b57f0d67773b6c4b724bd5f5e2890e30fefb0e64c8aa8c"

const DenominatorMigrationV72Digest = "sha256:a2c5f1b122d393aff2fc4f4b5ab21889f9548a4f3eb86559a9ece4a755738351"

const DenominatorMigrationV73Digest = "sha256:08c612ba137581d648e49339dcd2e6b95b391b97ce0c28e8204c83091e8d9616"

const DenominatorMigrationV74Digest = "sha256:c3ec2da26aa7dc284a32b3d41301063109ed0439c4dfbdaa6722a36bcbe5b6ff"

const DenominatorMigrationV75Digest = "sha256:0ae7b66bcbd26ae26c42873ebf57976fbc418524791d6a5753aa0b618ab168e7"

const DenominatorMigrationV76Digest = "sha256:e90d02bd6a93f618174dcd05b1edb767f375fe06f3fe857b249a322c2fc5f937"

const DenominatorMigrationV77Digest = "sha256:f61eca0e1226482e5c43d99de91adef42dc1c03f8626223cd7e6e56f9c566b6c"

const DenominatorMigrationV78Digest = "sha256:20eaf87a2ab8a20eb15d217a12b00258949728755f12fb6252735baf1b05e7f5"

const DenominatorMigrationV79Digest = "sha256:f36ee4a7435d5bf180751fc149855beb949e935a4af6bacd3673d8244f3753bd"

const DenominatorMigrationV80Digest = "sha256:d3c410815e67e063e3ac504ba601c24fec05a127349432f6255dd5b2bdb342e4"

const DenominatorMigrationV81Digest = "sha256:a556742a019b1787035f3a5ee8ca8f73e0ac17bf376c1260804554f7f2dcce70"

const DenominatorMigrationV82Digest = "sha256:d77385641c6816491b8684fa7e5d8d6ea965a872ef9f423b6a65b3aed33cd8d0"

const DenominatorMigrationV83Digest = "sha256:81a14c7762b47d902ee5b9aa3158b9c6920625a8d66fb5254cdadedbbacedd1c"

const DenominatorMigrationV84Digest = "sha256:b334c5bde5b254f2d1495a2d9c06f76dd6692dbc0b524970f3d0956d52550255"

const DenominatorMigrationV85Digest = "sha256:e73d5da14c96f0a27e895262af04df9e2310946ef1e0a0357077b7509524ed0f"

const DenominatorMigrationV86Digest = "sha256:fbd80da3405b52e9d8779616b93e7958c5d3bb763c24144a80b4596e57c2a473"

const DenominatorMigrationV87Digest = "sha256:4014ac4a0b3c1c0251ce4581e762eaeea994483a47cc44200c834658d5ce099c"

const DenominatorMigrationV88Digest = "sha256:c08229ea827f03eb0b5120b604d53f2ae1de4c45008f0cb589a7fe2ac78dba61"

const DenominatorMigrationV89Digest = "sha256:3a9fc35e19678a1401c4fb773d30384b1c87d87ae480d56c960de79e20c8a0cb"

const DenominatorMigrationV90Digest = "sha256:b9392f196bf14426e1526b29d7b6f06e36d1b74629f94314172f0339a5af1c6a"

const DenominatorMigrationV91Digest = "sha256:1d3d5f2f544495cedfc23ee510c9aaa1a94969779f0c26926bab9a21b50dc7a3"
