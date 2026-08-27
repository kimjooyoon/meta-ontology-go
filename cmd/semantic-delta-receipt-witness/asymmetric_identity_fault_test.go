package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"

	producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceipt"
	consumer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceiptconsumer"
)

func TestAsymmetricIdentityFaultReceiptProbesUseRawInputsOnly(t *testing.T) {
	input := asymmetricIdentityFaultInputs(t)
	producerBaseline := producer.IdentityFaultReceiptFromFiles(input.producer)
	consumerBaseline := consumer.IdentityFaultReceiptFromFiles(input.consumer)

	producerTampered := producerBaseline.Receipt
	producerTampered.Graph.Mapping[0].NewStableID, producerTampered.Graph.Mapping[1].NewStableID = producerTampered.Graph.Mapping[1].NewStableID, producerTampered.Graph.Mapping[0].NewStableID
	producerTampered.Graph.MappingDigest = jsonDigest(producerTampered.Graph.Mapping)
	producerTampered.Graph.RewrittenReferenceCount = 0
	producerTampered.Graph.DanglingReferenceCount = 1
	producerTampered.FaultGraphClosed = false
	producerTampered.Reason = "COHERENT_PRODUCER_RECEIPT_TAMPER"
	if _, err := json.Marshal(producerTampered); err != nil {
		t.Fatalf("marshal producer tamper: %v", err)
	}
	consumerAfterProducerTamper := consumer.IdentityFaultReceiptFromFiles(input.consumer)
	if !bytes.Equal(consumerBaseline.ComparisonBytes, consumerAfterProducerTamper.ComparisonBytes) {
		t.Fatal("consumer changed when only a producer receipt was tampered")
	}

	consumerTampered := consumerBaseline.Receipt
	consumerTampered.Graph.Mapping[0].NewStableID, consumerTampered.Graph.Mapping[1].NewStableID = consumerTampered.Graph.Mapping[1].NewStableID, consumerTampered.Graph.Mapping[0].NewStableID
	consumerTampered.Graph.MappingDigest = jsonDigest(consumerTampered.Graph.Mapping)
	consumerTampered.Graph.RewrittenReferenceCount = 0
	consumerTampered.Graph.DanglingReferenceCount = 1
	consumerTampered.FaultGraphClosed = false
	consumerTampered.Reason = "COHERENT_CONSUMER_RECEIPT_TAMPER"
	if _, err := json.Marshal(consumerTampered); err != nil {
		t.Fatalf("marshal consumer tamper: %v", err)
	}
	producerAfterConsumerTamper := producer.IdentityFaultReceiptFromFiles(input.producer)
	if !bytes.Equal(producerBaseline.ComparisonBytes, producerAfterConsumerTamper.ComparisonBytes) {
		t.Fatal("producer changed when only a consumer receipt was tampered")
	}
}

type asymmetricIdentityFaultInput struct {
	producer producer.IdentityFaultInput
	consumer consumer.IdentityFaultInput
}

func asymmetricIdentityFaultInputs(t *testing.T) asymmetricIdentityFaultInput {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "../../examples/semantic-delta-receipt")
	sha := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	return asymmetricIdentityFaultInput{
		producer: producer.IdentityFaultInput{
			Baseline:     producer.Input{CaseID: "persistence-probe", SubjectSHA: sha, ObservedCheckoutSHA: sha, BeforePath: filepath.Join(root, "before.gooo"), AfterPath: filepath.Join(root, "equivalent-after.gooo")},
			Alternate:    producer.Input{CaseID: "persistence-probe", SubjectSHA: sha, ObservedCheckoutSHA: sha, BeforePath: filepath.Join(root, "persistence-equivalent-before.gooo"), AfterPath: filepath.Join(root, "persistence-equivalent-after.gooo")},
			ArtifactPath: filepath.Join(root, "claim-identity-fault.json"),
		},
		consumer: consumer.IdentityFaultInput{
			Baseline:     consumer.Input{CaseID: "persistence-probe", SubjectSHA: sha, ObservedCheckoutSHA: sha, BeforePath: filepath.Join(root, "before.gooo"), AfterPath: filepath.Join(root, "equivalent-after.gooo")},
			Alternate:    consumer.Input{CaseID: "persistence-probe", SubjectSHA: sha, ObservedCheckoutSHA: sha, BeforePath: filepath.Join(root, "persistence-equivalent-before.gooo"), AfterPath: filepath.Join(root, "persistence-equivalent-after.gooo")},
			ArtifactPath: filepath.Join(root, "claim-identity-fault.json"),
		},
	}
}

func jsonDigest(value any) string {
	raw, _ := json.Marshal(value)
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}
