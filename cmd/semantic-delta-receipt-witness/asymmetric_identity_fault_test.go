package main

import (
	"bytes"
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
	tamperProducerIdentityFaultReceipt(&producerTampered)
	consumerAfterProducerTamper := consumer.IdentityFaultReceiptFromFiles(input.consumer)
	if !bytes.Equal(consumerBaseline.ComparisonBytes, consumerAfterProducerTamper.ComparisonBytes) {
		t.Fatal("consumer changed when only a producer receipt was tampered")
	}
	producerTamperedEvidence := producer.MarshalIdentityFaultReceiptEvidence(producerTampered)
	if compareOpaqueIdentityFaultEvidence(producerTamperedEvidence, consumerAfterProducerTamper.ComparisonBytes) {
		t.Fatal("consumer accepted tampered producer receipt evidence")
	}

	consumerTampered := consumerBaseline.Receipt
	tamperConsumerIdentityFaultReceipt(&consumerTampered)
	producerAfterConsumerTamper := producer.IdentityFaultReceiptFromFiles(input.producer)
	if !bytes.Equal(producerBaseline.ComparisonBytes, producerAfterConsumerTamper.ComparisonBytes) {
		t.Fatal("producer changed when only a consumer receipt was tampered")
	}
	consumerTamperedEvidence := consumer.MarshalIdentityFaultReceiptEvidence(consumerTampered)
	if compareOpaqueIdentityFaultEvidence(consumerTamperedEvidence, producerAfterConsumerTamper.ComparisonBytes) {
		t.Fatal("producer accepted tampered consumer receipt evidence")
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
