package domaincapability

import (
	"strings"
	"testing"
)

func reverseObservationDigest(char string) string {
	return "sha256:" + strings.Repeat(char, 64)
}

func reverseObservationInput() ReverseObservationInput {
	return ReverseObservationInput{
		SourceDigest:            reverseObservationDigest("0"),
		DeclarationDigest:       reverseObservationDigest("1"),
		IRDigest:                reverseObservationDigest("2"),
		GeneratedArtifactDigest: reverseObservationDigest("3"),
		EvidencePrefixDigest:    reverseObservationDigest("4"),
	}
}

func TestObserveReverseObservationBindsGeneratedArtifact(t *testing.T) {
	input := reverseObservationInput()
	input.ObservedArtifactDigest = input.GeneratedArtifactDigest
	receipt := ObserveReverseObservation(input)
	if receipt.Status != ReverseObservationObserved || receipt.FirstMismatch != "" || receipt.MissingStage != "" {
		t.Fatalf("unexpected observed receipt: %+v", receipt)
	}
	if err := receipt.Validate(); err != nil {
		t.Fatalf("observed receipt should validate: %v", err)
	}
}

func TestObserveReverseObservationPreservesMismatch(t *testing.T) {
	input := reverseObservationInput()
	input.ObservedArtifactDigest = reverseObservationDigest("5")
	receipt := ObserveReverseObservation(input)
	if receipt.Status != ReverseObservationMismatch || receipt.FirstMismatch != "reverse_observation" {
		t.Fatalf("unexpected mismatch receipt: %+v", receipt)
	}
	if err := receipt.Validate(); err != nil {
		t.Fatalf("mismatch receipt should validate: %v", err)
	}
}

func TestObserveReverseObservationDefersMissingObservation(t *testing.T) {
	receipt := ObserveReverseObservation(reverseObservationInput())
	if receipt.Status != ReverseObservationDeferred || receipt.MissingStage != "reverse_observation" {
		t.Fatalf("unexpected deferred receipt: %+v", receipt)
	}
	if err := receipt.Validate(); err != nil {
		t.Fatalf("deferred receipt should validate: %v", err)
	}
}
