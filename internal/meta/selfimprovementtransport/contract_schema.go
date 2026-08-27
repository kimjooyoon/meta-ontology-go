package selfimprovementtransport

var expectedEntities = map[string]string{
	"TransportInput":              "gooo://self-improvement/transport/entity/input",
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

var expectedActivities = map[string][2]string{
	"ObserveSourceIdentity":      {"TransportInput", "SourceIdentityEvidence"},
	"ObserveCheckoutBinding":     {"TransportInput", "CheckoutBindingEvidence"},
	"ObserveProducerIdentity":    {"TransportInput", "ProducerIdentityEvidence"},
	"ObserveLogicalDigest":       {"TransportInput", "LogicalDigestEvidence"},
	"ObserveImmutableLocator":    {"TransportInput", "ImmutableLocatorEvidence"},
	"ObserveArchiveDigest":       {"TransportInput", "ArchiveDigestEvidence"},
	"ObserveConsumerReplay":      {"TransportInput", "ConsumerReplayEvidence"},
	"ObserveProducerAttestation": {"TransportInput", "ProducerAttestationEvidence"},
	"ReduceTransport":            {"TransportInput", "TransportReceipt"},
}
