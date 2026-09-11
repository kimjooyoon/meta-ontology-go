package selfimprovementtransport

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func evaluateFixtureWith(t *testing.T, observation, producer []byte, metadata TransportMetadata, archiveDigest string) Report {
	t.Helper()
	repository, _, _, _, _ := fixture(t)
	metadataRaw, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	return Evaluate(repository, "transport.gooo", "kimjooyoon/meta-ontology-go", 101,
		observation, producer, metadataRaw, archiveDigest)
}

func TestProducerPayloadProvenanceClosesExactSubject(t *testing.T) {
	_, observation, producer, metadata, archiveDigest := fixture(t)
	report := evaluateFixtureWith(t, observation, producer, metadata, archiveDigest)
	if err := CheckReadOnly(report); err != nil {
		t.Fatal(err)
	}
	if report.ProvenanceState != ResolutionClosed || report.ResolutionMetrics.CaseDenominator != 10 ||
		report.ResolutionMetrics.CurrentExact != 1 || report.ResolutionMetrics.CurrentUnknownSixField != 0 ||
		report.ResolutionMetrics.CurrentRefuted != 0 {
		t.Fatalf("unexpected exact provenance: %+v", report)
	}
}

func TestProducerPayloadProvenancePreservesUnknownCausality(t *testing.T) {
	_, observation, producer, metadata, archiveDigest := fixture(t)
	metadata.ProducerDeclarationCount = 0
	metadata.ProducerPayloadCount = 0
	report := evaluateFixtureWith(t, observation, producer, metadata, archiveDigest)
	if err := ValidateReport(report); err != nil {
		t.Fatal(err)
	}
	if report.ProvenanceState != ResolutionUnknown || !causalUnknownComplete(report.Provenance.Unknown) ||
		report.ResolutionMetrics.CurrentUnknownSixField != 1 || report.Decision != DecisionObserved {
		t.Fatalf("unexpected UNKNOWN provenance: %+v", report)
	}
}

func TestProducerPayloadProvenanceRefutesContradictions(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*TransportMetadata, *ProducerReceipt)
		reason string
	}{
		{name: "duplicate declaration", mutate: func(metadata *TransportMetadata, _ *ProducerReceipt) {
			metadata.ProducerDeclarationCount = 2
		}, reason: "DUPLICATE_PRODUCER_DECLARATION"},
		{name: "payload subject", mutate: func(metadata *TransportMetadata, _ *ProducerReceipt) {
			metadata.ProducerSubjectSHA = strings.Repeat("b", 40)
		}, reason: "PRODUCER_PAYLOAD_SUBJECT_MISMATCH"},
		{name: "payload digest", mutate: func(_ *TransportMetadata, producer *ProducerReceipt) {
			producer.Subject.Digest = "sha256:" + strings.Repeat("c", 64)
		}, reason: "PRODUCER_PAYLOAD_DIGEST_MISMATCH"},
		{name: "repository", mutate: func(_ *TransportMetadata, producer *ProducerReceipt) {
			producer.RepositoryURI = "https://github.com/other/repository"
		}, reason: "PRODUCER_REPOSITORY_WORKFLOW_MISMATCH"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, observation, producerRaw, metadata, archiveDigest := fixture(t)
			var producer ProducerReceipt
			if err := json.Unmarshal(producerRaw, &producer); err != nil {
				t.Fatal(err)
			}
			test.mutate(&metadata, &producer)
			producerRaw, err := json.Marshal(producer)
			if err != nil {
				t.Fatal(err)
			}
			metadata.ProducerDeclarationDigest = digestBytes(producerRaw)
			report := evaluateFixtureWith(t, observation, producerRaw, metadata, archiveDigest)
			if err := ValidateReport(report); err != nil {
				t.Fatal(err)
			}
			if report.ProvenanceState != ResolutionRefuted || report.Decision != DecisionFailClosed || report.Reason != ReasonKnownMismatch || report.Provenance.Reason != test.reason {
				t.Fatalf("unexpected REFUTED provenance: %+v", report)
			}
		})
	}
}

func TestBindTransportMetadataBindsActualPayloadAndCountsDuplicates(t *testing.T) {
	_, observation, producer, _, _ := fixture(t)
	receipt := LifecycleReceipt{Repository: "kimjooyoon/meta-ontology-go", ExpectedRunID: 101, ExpectedRunAttempt: 2, OrchestrationHeadSHA: strings.Repeat("a", 40), WorkflowPath: ".github/workflows/observation.yml", ArtifactID: 202, ArtifactName: ArtifactName, ArtifactDigest: "sha256:" + strings.Repeat("d", 64), ArtifactSizeBytes: 4096, ArtifactInstanceCount: 1, ArtifactTypeCount: 1}
	archive := zipArchive(t, map[string][][]byte{"first.json": {observation}, "producer.json": {producer}})
	metadata, err := BindTransportMetadata(receipt, archive)
	if err != nil {
		t.Fatal(err)
	}
	if metadata.ProducerDeclarationCount != 1 || metadata.ProducerPayloadCount != 1 || metadata.ProducerPayloadDigest != digestBytes(observation) || metadata.ProducerPayloadBytes != len(observation) {
		t.Fatalf("metadata did not bind actual payload: %+v", metadata)
	}
	duplicateArchive := zipArchive(t, map[string][][]byte{"first.json": {observation}, "producer.json": {producer, producer}})
	duplicate, err := BindTransportMetadata(receipt, duplicateArchive)
	if err != nil {
		t.Fatal(err)
	}
	if duplicate.ProducerDeclarationCount != 2 || duplicate.ProducerSubjectSHA != "" {
		t.Fatalf("duplicate declaration was selected: %+v", duplicate)
	}
}

func zipArchive(t *testing.T, files map[string][][]byte) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, entries := range files {
		for _, data := range entries {
			entry, err := writer.Create(name)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := entry.Write(data); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}
