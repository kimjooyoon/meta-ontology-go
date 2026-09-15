package authorizationfoundation

const (
	FoundationSchema = "gooo/external-capability-policy-foundation/v1"
	BootstrapSchema  = "gooo/external-capability-authorization-receipt/v1"
	ReceiptSchema    = "gooo/external-capability-authorization-foundation-receipt/v1"
	SuiteSchema      = "gooo/external-capability-authorization-foundation-suite/v1"

	ExpectedRepository    = "kimjooyoon/meta-ontology-go"
	ExpectedRunID         = int64(33034283265)
	ExpectedRunAttempt    = 1
	ExpectedArtifactID    = int64(9631439149)
	ExpectedSubjectSHA    = "768bd7c4cf974b61de46e57e6309108b52375641"
	ExpectedArtifactName  = "external-capability-authorization-" + ExpectedSubjectSHA
	ExpectedArchiveDigest = "sha256:d8757edaa0651bc1bd606a8dd0fa99b8cefd3a7b9ffb3009d4d27c352be48b5f"
	ExpectedFileDigest    = "sha256:a5b162b11d98a266b276c7344cbf4e97339d344b7441f5b19aa97a3a66ce9c3e"
	ExpectedReceiptDigest = "sha256:25641552b4aa33969dd86de6de5135d01e6928da4912e6a5b7e105d7795043fd"
	ExpectedSourceDigest  = "sha256:e24c7c72a486b502b594327b8a72738c3b2ea065466eb4a26637a59a6f22ef7b"
	ExpectedTreeDigest    = "sha256:4a64eb24dc3b06f3dac2a0a5a4b2005fa85b5e95a9818502a37d78b246cbb2d6"

	PolicyMetric = "gooo.metric.external-capability-authorization-policy-foundation.v1"
	PolicyClaim  = "gooo.claim.external-capability-authorization-policy-foundation.v1"
	PolicyStage  = "AUTHORIZE/policy-foundation"
)
