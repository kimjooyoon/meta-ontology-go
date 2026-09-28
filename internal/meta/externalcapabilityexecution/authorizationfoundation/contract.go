package authorizationfoundation

const (
	FoundationSchema = "gooo/external-capability-policy-foundation/v1"
	BootstrapSchema  = "gooo/external-capability-authorization-receipt/v1"
	ReceiptSchema    = "gooo/external-capability-authorization-foundation-receipt/v1"
	SuiteSchema      = "gooo/external-capability-authorization-foundation-suite/v1"

	ExpectedRepository    = "kimjooyoon/meta-ontology-go"
	ExpectedRunID         = int64(36350185463)
	ExpectedRunAttempt    = 1
	ExpectedArtifactID    = int64(10942435307)
	ExpectedSubjectSHA    = "6bb57ce8f9713e8e8d9c45bfea357e9cf3727d27"
	ExpectedArtifactName  = "external-capability-authorization-" + ExpectedSubjectSHA
	ExpectedArchiveDigest = "sha256:c75f527306d4e7983e58021d6e3cab06f2dfb6b553c2cc793529d52fd954b86f"
	ExpectedFileDigest    = "sha256:f73a537b590da5ec5afdcd3b1d9b68f5191edaf8727949c84cda918dcbd515b3"
	ExpectedReceiptDigest = "sha256:13b559138259e61def967387c8f368c443b3709dfefe90ea6a6e06b1305b9c89"
	ExpectedSourceDigest  = "sha256:e24c7c72a486b502b594327b8a72738c3b2ea065466eb4a26637a59a6f22ef7b"
	ExpectedTreeDigest    = "sha256:4a64eb24dc3b06f3dac2a0a5a4b2005fa85b5e95a9818502a37d78b246cbb2d6"

	PolicyMetric = "gooo.metric.external-capability-authorization-policy-foundation.v1"
	PolicyClaim  = "gooo.claim.external-capability-authorization-policy-foundation.v1"
	PolicyStage  = "AUTHORIZE/policy-foundation"
)
