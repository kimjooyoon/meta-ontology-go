package selfimprovementtransport

type expectedActivity struct {
	Input   string
	Output  string
	Program string
}

var expectedEntities = map[string]string{
	"TransportInput":              "gooo://self-improvement/transport/entity/input",
	"ArtifactMetadataEvidence":    "gooo://self-improvement/transport/evidence/artifact-metadata",
	"ArtifactValidationEvidence":  "gooo://self-improvement/transport/evidence/artifact-validation",
	"ArchiveDownloadEvidence":     "gooo://self-improvement/transport/evidence/archive-download",
	"SourceIdentityEvidence":      "gooo://self-improvement/transport/evidence/source-identity",
	"CheckoutBindingEvidence":     "gooo://self-improvement/transport/evidence/checkout-binding",
	"ProducerIdentityEvidence":    "gooo://self-improvement/transport/evidence/producer-identity",
	"LogicalDigestEvidence":       "gooo://self-improvement/transport/evidence/logical-digest",
	"ImmutableLocatorEvidence":    "gooo://self-improvement/transport/evidence/immutable-locator",
	"ArchiveDigestEvidence":       "gooo://self-improvement/transport/evidence/archive-digest",
	"ConsumerReplayEvidence":      "gooo://self-improvement/transport/evidence/consumer-replay",
	"ProducerAttestationEvidence": "gooo://self-improvement/transport/evidence/producer-attestation",
	"TransportReceipt":            "gooo://self-improvement/transport/entity/receipt",
}

var expectedActivities = map[string]expectedActivity{
	"ReadArtifactMetadata": {"TransportInput", "ArtifactMetadataEvidence",
		"meta.artifact.lifecycle.read-metadata:v1"},
	"ResolveArtifact": {"ArtifactMetadataEvidence", "ImmutableLocatorEvidence",
		"meta.artifact.lifecycle.resolve-artifact:v1"},
	"ValidateArtifactMetadata": {"ImmutableLocatorEvidence", "ArtifactValidationEvidence",
		"meta.artifact.lifecycle.validate-metadata:v1"},
	"DownloadArtifactArchive": {"ArtifactValidationEvidence", "ArchiveDownloadEvidence",
		"meta.artifact.lifecycle.download-archive:v1"},
	"VerifyArtifactArchiveDigest": {"ArchiveDownloadEvidence", "ArchiveDigestEvidence",
		"meta.artifact.lifecycle.verify-archive-digest:v1"},
	"ObserveSourceIdentity":      {Input: "TransportInput", Output: "SourceIdentityEvidence"},
	"ObserveCheckoutBinding":     {Input: "TransportInput", Output: "CheckoutBindingEvidence"},
	"ObserveProducerIdentity":    {Input: "TransportInput", Output: "ProducerIdentityEvidence"},
	"ObserveLogicalDigest":       {Input: "TransportInput", Output: "LogicalDigestEvidence"},
	"ObserveImmutableLocator":    {Input: "TransportInput", Output: "ImmutableLocatorEvidence"},
	"ObserveArchiveDigest":       {Input: "TransportInput", Output: "ArchiveDigestEvidence"},
	"ObserveConsumerReplay":      {Input: "TransportInput", Output: "ConsumerReplayEvidence"},
	"ObserveProducerAttestation": {Input: "TransportInput", Output: "ProducerAttestationEvidence"},
	"ReduceTransport":            {Input: "TransportInput", Output: "TransportReceipt"},
}
