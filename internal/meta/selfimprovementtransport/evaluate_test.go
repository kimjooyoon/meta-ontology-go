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
entity ProducerDeclarationEvidence id "gooo://self-improvement/transport/evidence/producer-declaration"
entity TransportIndexEvidence id "gooo://self-improvement/transport/evidence/transport-index"
entity ConsumerResolutionEvidence id "gooo://self-improvement/transport/evidence/consumer-resolution"
entity ArtifactMetadataEvidence id "gooo://self-improvement/transport/evidence/artifact-metadata"
entity ArtifactValidationEvidence id "gooo://self-improvement/transport/evidence/artifact-validation"
entity ArchiveDownloadEvidence id "gooo://self-improvement/transport/evidence/archive-download"
entity SourceIdentityEvidence id "gooo://self-improvement/transport/evidence/source-identity"
entity CheckoutBindingEvidence id "gooo://self-improvement/transport/evidence/checkout-binding"
entity ProducerIdentityEvidence id "gooo://self-improvement/transport/evidence/producer-identity"
entity LogicalDigestEvidence id "gooo://self-improvement/transport/evidence/logical-digest"
entity ImmutableLocatorEvidence id "gooo://self-improvement/transport/evidence/immutable-locator"
entity ArchiveDigestEvidence id "gooo://self-improvement/transport/evidence/archive-digest"
entity ConsumerReplayEvidence id "gooo://self-improvement/transport/evidence/consumer-replay"
entity ProducerAttestationEvidence id "gooo://self-improvement/transport/evidence/producer-attestation"
entity TransportReceipt id "gooo://self-improvement/transport/entity/receipt"
activity DeclareProducerSubject(TransportInput) -> ProducerDeclarationEvidence computes "meta.transport.producer-declaration/v1"
activity BindTransportIndex(ProducerDeclarationEvidence) -> TransportIndexEvidence computes "meta.transport.artifact-index/v1"
activity ResolveConsumerSubject(TransportIndexEvidence) -> ConsumerResolutionEvidence computes "meta.transport.consumer-resolution/v1;resolution-schema=gooo/self-improvement-transport-resolution-policy/v1;resolution-states=CLOSED,UNKNOWN,REFUTED;resolution-causal-fields=stage,step,reason,unknown_class,next_operation,blocked_by;resolution-transition=producer-declaration>transport-index;resolution-transition=transport-index>consumer-resolution;resolution-artifact-identity=artifact_id,artifact_name,artifact_digest,producer_run_id,producer_run_attempt,producer_subject_sha,producer_declaration_digest,producer_payload_name,producer_payload_digest;resolution-refuted-dominates-unknown=true;resolution-metric=active_root_before|1;resolution-metric=active_root_after|0;resolution-metric=exact_resolutions_before|0;resolution-metric=exact_resolutions_after|1;resolution-metric=unknown_six_field_before|0;resolution-metric=unknown_six_field_after|3;resolution-metric=refuted_contradictions_before|0;resolution-metric=refuted_contradictions_after|4;resolution-metric=fallback_accepted_before|0;resolution-metric=fallback_accepted_after|0;resolution-metric=artifact_instances_before|1;resolution-metric=artifact_instances_after|1;resolution-metric=artifact_types_before|1;resolution-metric=artifact_types_after|1;resolution-metric=independent_replay_comparisons_before|0;resolution-metric=independent_replay_comparisons_after|1;resolution-case=EXACT_PRODUCER_DECLARATION|CLOSED|CONSUME|resolve-producer-subject||EXACT_PRODUCER_SUBJECT_PAYLOAD_MATCH;resolution-case=EXACT_PAYLOAD_SUBJECT|CLOSED|CONSUME|resolve-producer-subject||EXACT_PAYLOAD_SUBJECT_MATCH;resolution-case=EXACT_NONCURRENT_SUBJECT|CLOSED|CONSUME|resolve-producer-subject||EXACT_NONCURRENT_SUBJECT_MATCH;resolution-case=EXPIRED_ARTIFACT|UNKNOWN|LOCATE|read-artifact-metadata|EXPIRED|ARTIFACT_EXPIRED;resolution-case=MISSING_PRODUCER_DECLARATION|UNKNOWN|CONSUME|resolve-producer-declaration|DIRECT_MISSING|PRODUCER_DECLARATION_MISSING;resolution-case=MISSING_PAYLOAD|UNKNOWN|CONSUME|resolve-payload|DIRECT_MISSING|PRODUCER_PAYLOAD_MISSING;resolution-case=DUPLICATE_DECLARATION|REFUTED|CONSUME|resolve-producer-declaration||DUPLICATE_PRODUCER_DECLARATION;resolution-case=PAYLOAD_SUBJECT_CONTRADICTION|REFUTED|CONSUME|resolve-payload||PRODUCER_PAYLOAD_SUBJECT_MISMATCH;resolution-case=PAYLOAD_DIGEST_MISMATCH|REFUTED|CONSUME|resolve-payload||PRODUCER_PAYLOAD_DIGEST_MISMATCH;resolution-case=REPOSITORY_WORKFLOW_CONTRADICTION|REFUTED|CONSUME|resolve-producer-identity||PRODUCER_REPOSITORY_WORKFLOW_MISMATCH"
activity ReadArtifactMetadata(TransportInput) -> ArtifactMetadataEvidence computes "meta.artifact.lifecycle.read-metadata:v1"
activity ResolveArtifact(ArtifactMetadataEvidence) -> ImmutableLocatorEvidence computes "meta.artifact.lifecycle.resolve-artifact:v1"
activity ValidateArtifactMetadata(ImmutableLocatorEvidence) -> ArtifactValidationEvidence computes "meta.artifact.lifecycle.validate-metadata:v1"
activity DownloadArtifactArchive(ArtifactValidationEvidence) -> ArchiveDownloadEvidence computes "meta.artifact.lifecycle.download-archive:v1"
activity VerifyArtifactArchiveDigest(ArchiveDownloadEvidence) -> ArchiveDigestEvidence computes "meta.artifact.lifecycle.verify-archive-digest:v1"
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
	producer, err := Produce(repository, "transport.gooo", ProducerInput{Repository: "kimjooyoon/meta-ontology-go", SubjectSHA: sha, CheckoutSHA: sha, WorkflowRef: "repo/.github/workflows/observation.yml@refs/heads/dev", WorkflowSHA: strings.Repeat("b", 40), RunID: 101, RunAttempt: 2, Job: "observation", ArtifactName: ArtifactName}, observation)
	if err != nil {
		t.Fatal(err)
	}
	producerRaw, _ := json.Marshal(producer)
	archiveDigest := digestBytes([]byte("archive"))
	metadata := TransportMetadata{Schema: MetadataSchema, Repository: "kimjooyoon/meta-ontology-go", ProducerRunID: 101, ProducerRunAttempt: 2, OrchestrationHeadSHA: sha, WorkflowPath: ".github/workflows/observation.yml", ArtifactID: 202, ArtifactName: ArtifactName, ArtifactDigest: archiveDigest, ArtifactSizeBytes: 4096, ArtifactInstanceCount: 1, ArtifactTypeCount: 1, ProducerDeclarationCount: 1, ProducerDeclarationDigest: digestBytes(producerRaw), ProducerSubjectSHA: sha, ProducerPayloadCount: 1, ProducerPayloadName: "first.json", ProducerPayloadDigest: digestBytes(observation), ProducerPayloadBytes: len(observation)}
	return repository, observation, producerRaw, metadata, archiveDigest
}

func evaluateFixture(t *testing.T, metadata TransportMetadata, archiveDigest string) Report {
	t.Helper()
	repository, observation, producer, _, _ := fixture(t)
	metadataRaw, _ := json.Marshal(metadata)
	return Evaluate(repository, "transport.gooo", "kimjooyoon/meta-ontology-go", 101, observation, producer, metadataRaw, archiveDigest)
}

func TestUnsignedTransportLowersResolution(t *testing.T) {
	_, _, _, metadata, archiveDigest := fixture(t)
	report := evaluateFixture(t, metadata, archiveDigest)
	if err := CheckReadOnly(report); err != nil {
		t.Fatal(err)
	}
	if report.Metrics.CoverageBasisPoints != 8750 || report.Coordinate.Stage != "ATTEST" || report.Coordinate.Step != "verify-producer-identity" {
		t.Fatalf("unexpected lowered receipt: %+v", report)
	}
}
