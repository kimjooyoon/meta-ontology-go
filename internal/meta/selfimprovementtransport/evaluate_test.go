package selfimprovementtransport

import (
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"
)

const contractFixture = `package selfimprovementtransport
namespace selfimprovementtransport
entity TransportInput id "gooo://self-improvement/transport/entity/input"
entity SourceIdentityEvidence id "gooo://self-improvement/transport/evidence/source-identity"
entity CheckoutBindingEvidence id "gooo://self-improvement/transport/evidence/checkout-binding"
entity ProducerIdentityEvidence id "gooo://self-improvement/transport/evidence/producer-identity"
entity LogicalDigestEvidence id "gooo://self-improvement/transport/evidence/logical-digest"
entity ImmutableLocatorEvidence id "gooo://self-improvement/transport/evidence/immutable-locator"
entity ArchiveDigestEvidence id "gooo://self-improvement/transport/evidence/archive-digest"
entity ConsumerReplayEvidence id "gooo://self-improvement/transport/evidence/consumer-replay"
entity ProducerAttestationEvidence id "gooo://self-improvement/transport/evidence/producer-attestation"
entity TransportReceipt id "gooo://self-improvement/transport/entity/receipt"
activity ObserveSourceIdentity(TransportInput) -> SourceIdentityEvidence
activity ObserveCheckoutBinding(TransportInput) -> CheckoutBindingEvidence
activity ObserveProducerIdentity(TransportInput) -> ProducerIdentityEvidence
activity ObserveLogicalDigest(TransportInput) -> LogicalDigestEvidence
activity ObserveImmutableLocator(TransportInput) -> ImmutableLocatorEvidence
activity ObserveArchiveDigest(TransportInput) -> ArchiveDigestEvidence
activity ObserveConsumerReplay(TransportInput) -> ConsumerReplayEvidence
activity ObserveProducerAttestation(TransportInput) -> ProducerAttestationEvidence
activity ReduceTransport(TransportInput) -> TransportReceipt
`

func fixture(t *testing.T) (fstest.MapFS, []byte, []byte, TransportMetadata, string) {
	t.Helper()
	repository := fstest.MapFS{"transport.gooo": {Data: []byte(contractFixture)}}
	sha := strings.Repeat("a", 40)
	observation := []byte(`{"schema":"gooo/self-improvement-language-observation/v1","subject_sha":"` + sha + `","decision":"OBSERVED"}`)
	producer, err := Produce(repository, "transport.gooo", ProducerInput{
		Repository: "kimjooyoon/meta-ontology-go", SubjectSHA: sha, CheckoutSHA: sha,
		WorkflowRef: "repo/.github/workflows/observation.yml@refs/heads/dev", WorkflowSHA: strings.Repeat("b", 40),
		RunID: 101, RunAttempt: 2, Job: "observation", ArtifactName: ArtifactName,
	}, observation)
	if err != nil {
		t.Fatal(err)
	}
	producerRaw, _ := json.Marshal(producer)
	archiveDigest := digestBytes([]byte("archive"))
	metadata := TransportMetadata{
		Schema: MetadataSchema, Repository: "kimjooyoon/meta-ontology-go", ProducerRunID: 101,
		ProducerRunAttempt: 2, OrchestrationHeadSHA: strings.Repeat("c", 40), WorkflowPath: ".github/workflows/observation.yml",
		ArtifactID: 202, ArtifactName: ArtifactName, ArtifactDigest: archiveDigest, ArtifactSizeBytes: 4096,
	}
	return repository, observation, producerRaw, metadata, archiveDigest
}

func evaluateFixture(t *testing.T, metadata TransportMetadata, archiveDigest string) Report {
	t.Helper()
	repository, observation, producer, _, _ := fixture(t)
	metadataRaw, _ := json.Marshal(metadata)
	return Evaluate(repository, "transport.gooo", "kimjooyoon/meta-ontology-go", 101,
		observation, producer, metadataRaw, archiveDigest)
}

func TestUnsignedTransportLowersResolution(t *testing.T) {
	_, _, _, metadata, archiveDigest := fixture(t)
	report := evaluateFixture(t, metadata, archiveDigest)
	if err := CheckReadOnly(report); err != nil {
		t.Fatal(err)
	}
	if report.Metrics.CoverageBasisPoints != 8750 || report.Coordinate.Stage != "ATTEST" ||
		report.Coordinate.Step != "verify-producer-identity" {
		t.Fatalf("unexpected lowered receipt: %+v", report)
	}
}

func TestVerifiedAttestationClosesEHT8(t *testing.T) {
	_, _, _, metadata, archiveDigest := fixture(t)
	metadata.Attestation = Attestation{Status: "VERIFIED", Digest: digestBytes([]byte("attestation")), ProducerIdentity: "github-actions"}
	report := evaluateFixture(t, metadata, archiveDigest)
	if err := ValidateReport(report); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionPass || report.Metrics.VerifiedTotal != 8 || report.Metrics.CoverageBasisPoints != 10000 {
		t.Fatalf("unexpected exact receipt: %+v", report)
	}
}

func TestKnownTransportMismatchFailsClosed(t *testing.T) {
	_, _, _, metadata, _ := fixture(t)
	report := evaluateFixture(t, metadata, digestBytes([]byte("different archive")))
	if err := ValidateReport(report); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionFailClosed || report.Metrics.FalseTotal != 1 ||
		report.Coordinate.Stage != "TRANSPORT" || report.Coordinate.Step != "verify-archive-digest" {
		t.Fatalf("known mismatch was not fail-closed: %+v", report)
	}
}
