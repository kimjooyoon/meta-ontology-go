package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newProvenanceCLIFixture(t *testing.T) provenanceCLIFixture {
	t.Helper()
	directory := t.TempDir()
	sourcePath := filepath.Join(directory, "main.gooo")
	evidencePath := filepath.Join(directory, "evidence.json")
	storePath := filepath.Join(directory, "provenance.jsonl")
	source := []byte(`package billing
namespace billing

entity Order id "billing://entity/order"
entity PaymentMethod id "billing://entity/payment-method"
entity Payment id "billing://entity/payment"

activity PayOrder(Order, PaymentMethod) -> Payment
`)
	if err := os.WriteFile(sourcePath, source, 0o640); err != nil {
		t.Fatal(err)
	}
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if diagnostics.HasErrors() {
		t.Fatalf("fixture parse = %v", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		t.Fatalf("fixture lower = %v", err)
	}
	sourceDigest := sha256Digest(source)
	produced := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	records := []provenance.Evidence{
		{
			ID: "billing://event/order-created", SemanticID: "billing://entity/order", Producer: "gooo://producer/cli-fixture",
			Kind: provenance.KindObservation, Status: provenance.StatusVerified,
			SourceSpan:   provenance.SourceSpan{URI: sourcePath, Start: provenance.Position{Offset: 0, Line: 1, Column: 1}, End: provenance.Position{Offset: len(source), Line: 10, Column: 1}},
			SourceDigest: sourceDigest, SemanticDigest: ir.StableHash(), GraphDigest: ir.Graph.StableHash(),
			Freshness:  provenance.NewFreshness(sourceDigest, produced, produced.Add(2*time.Hour)),
			Attributes: map[string]string{"fixture": "billing", "status": "verified"},
		},
		{
			ID: "billing://event/payment-created", SemanticID: "billing://entity/payment", Producer: "gooo://producer/cli-fixture",
			Kind: provenance.KindObservation, Status: provenance.StatusVerified,
			SourceSpan:   provenance.SourceSpan{URI: sourcePath, Start: provenance.Position{Offset: 0, Line: 1, Column: 1}, End: provenance.Position{Offset: len(source), Line: 10, Column: 1}},
			SourceDigest: sourceDigest, SemanticDigest: ir.StableHash(), GraphDigest: ir.Graph.StableHash(),
			Freshness:  provenance.NewFreshness(sourceDigest, produced, produced.Add(2*time.Hour)),
			Attributes: map[string]string{"fixture": "billing", "status": "verified"},
		},
	}
	return provenanceCLIFixture{
		directory: directory, sourcePath: sourcePath, evidencePath: evidencePath, storePath: storePath,
		sourceDigest: sourceDigest, semanticDigest: ir.StableHash(), graphDigest: ir.Graph.StableHash(), records: records,
	}
}
