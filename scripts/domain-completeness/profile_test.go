package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestProfileGoooRoundTripsItsReceiptSchema(t *testing.T) {
	source, err := os.ReadFile("profile.gooo")
	if err != nil {
		t.Fatal(err)
	}
	model, err := compileProfile("profile.gooo", source)
	if err != nil {
		t.Fatal(err)
	}
	generated := []byte(renderProfile(model))
	generatedHash, err := semanticHash("profile.generated.gooo", generated)
	if err != nil {
		t.Fatal(err)
	}
	if generatedHash != model.SemanticHash {
		t.Fatalf("generated structure semantic hash %q differs from contract %q", generatedHash, model.SemanticHash)
	}
	if model.Entities["DomainCompletenessReceipt"].ID != ReceiptSchema {
		t.Fatalf("receipt entity schema = %q, want %q", model.Entities["DomainCompletenessReceipt"].ID, ReceiptSchema)
	}
	want := []string{
		"declaration_coverage", "generation_coverage", "reverse_observation_coverage",
		"use_case_coverage", "boundary_coverage", "provenance_integrity",
	}
	if len(model.Dimensions) != len(want) {
		t.Fatalf("profile dimension count = %d, want %d", len(model.Dimensions), len(want))
	}
	for index, id := range want {
		if model.Dimensions[index].ID != id {
			t.Fatalf("profile dimension %d = %q, want %q", index, model.Dimensions[index].ID, id)
		}
	}
}

func TestProfileVectorOrderControlsMeasurementOrder(t *testing.T) {
	source, err := os.ReadFile("profile.gooo")
	if err != nil {
		t.Fatal(err)
	}
	old := "DeclarationCoverage, GenerationCoverage, ReverseObservationCoverage"
	replacement := "GenerationCoverage, DeclarationCoverage, ReverseObservationCoverage"
	if !bytes.Contains(source, []byte(old)) {
		t.Fatalf("profile vector signature does not contain %q", old)
	}
	changed := bytes.Replace(source, []byte(old), []byte(replacement), 1)
	model, err := compileProfile("profile-reordered.gooo", changed)
	if err != nil {
		t.Fatal(err)
	}
	dimensions := measureDimensions(model, loadedInputs{}, "", false, 0, 0)
	if len(dimensions) != len(model.Dimensions) || dimensions[0].ID != "generation_coverage" || dimensions[1].ID != "declaration_coverage" {
		t.Fatalf("measurement order did not follow the Gooo vector: %#v", dimensions)
	}
}

func TestProfileRejectsUnregisteredOrDisconnectedMetric(t *testing.T) {
	source, err := os.ReadFile("profile.gooo")
	if err != nil {
		t.Fatal(err)
	}
	t.Run("unregistered computes identifier", func(t *testing.T) {
		changed := bytes.Replace(source,
			[]byte("gooo.metric.domain-completeness.declaration-coverage.v1"),
			[]byte("gooo.metric.domain-completeness.unknown-coverage.v1"), 1)
		if _, err := compileProfile("profile-unknown.gooo", changed); err == nil || !strings.Contains(err.Error(), "unsupported metric") {
			t.Fatalf("unsupported metric = %v", err)
		}
	})
	t.Run("unbound vector measurement", func(t *testing.T) {
		changed := bytes.Replace(source,
			[]byte("DeclarationCoverage, GenerationCoverage, ReverseObservationCoverage, UseCaseCoverage"),
			[]byte("DeclarationCoverage, GenerationCoverage, ReverseObservationCoverage"), 1)
		if _, err := compileProfile("profile-disconnected.gooo", changed); err == nil || !strings.Contains(err.Error(), "not connected to the vector") {
			t.Fatalf("disconnected measurement = %v", err)
		}
	})
	t.Run("missing required axis", func(t *testing.T) {
		changed := bytes.Replace(source,
			[]byte(", UseCaseCoverage, BoundaryCoverage"), []byte(", BoundaryCoverage"), 1)
		changed = bytes.Replace(changed,
			[]byte("activity MeasureUseCaseCoverage(UseCaseEvidence) -> UseCaseCoverage computes \"gooo.metric.domain-completeness.use-case-coverage.v1\"\n"),
			nil, 1)
		if _, err := compileProfile("profile-missing-axis.gooo", changed); err == nil || !strings.Contains(err.Error(), "dimension denominator") {
			t.Fatalf("missing required axis = %v", err)
		}
	})
}

func testProfileDimensions(t *testing.T) []dimensionSpec {
	t.Helper()
	source, err := os.ReadFile("profile.gooo")
	if err != nil {
		t.Fatal(err)
	}
	model, err := compileProfile("profile.gooo", source)
	if err != nil {
		t.Fatal(err)
	}
	return model.Dimensions
}
