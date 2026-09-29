package main

import (
	"os"
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
}
