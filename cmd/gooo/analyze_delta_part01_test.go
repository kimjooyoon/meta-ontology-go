package main

import (
	"bytes"
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer"
	"os"
	"testing"
)

func TestRunAnalyzeBillingGeneratedDeltaIsCanonicalAndReadOnly(t *testing.T) {
	authority, generated := billingAnalyzeFiles(t, billingAnalyzeAuthority)
	beforeSource, beforeInfo := snapshotFile(t, generated)
	first, firstCode, firstErr := runAnalyzePaths(authority, generated)
	second, secondCode, secondErr := runAnalyzePaths(authority, generated)
	if firstCode != exitOK || secondCode != exitOK || firstErr != "" || secondErr != "" {
		t.Fatalf("billing analyze = %d/%d, stderr=%q/%q", firstCode, secondCode, firstErr, secondErr)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("billing analyze replay changed output:\nfirst=%s\nsecond=%s", first, second)
	}
	var output analyzeDeltaOutput
	if err := json.Unmarshal(first, &output); err != nil {
		t.Fatalf("decode semantic delta: %v; output=%s", err, first)
	}
	if output.SchemaVersion != "analyzer-semantic-delta/v1" || output.Digest == "" || output.AuthoritySemanticDigest == "" || output.AuthoritySemanticDigest != output.ObservedSemanticDigest || !output.SemanticEqual || output.WriteEffect != analyzer.ReconcileNoWrite {
		t.Fatalf("incomplete billing analyze output: %#v", output)
	}
	if len(output.SignatureFacts) != 3 || len(output.CandidateFacts) != 0 || len(output.DeferredImplementation) != 1 || len(output.DeferredSlots) != 1 {
		t.Fatalf("billing delta classes = %d signature, %d candidate, %d implementation, %d slots", len(output.SignatureFacts), len(output.CandidateFacts), len(output.DeferredImplementation), len(output.DeferredSlots))
	}
	for _, fact := range output.SignatureFacts {
		if fact.Fact.Span.File == "" || fact.Fact.Span.End.Offset <= fact.Fact.Span.Start.Offset || fact.Evidence.Span != fact.Fact.Span {
			t.Fatalf("signature fact lost exact source span: %#v", fact)
		}
	}
	if output.DeferredImplementation[0].Origin != analyzer.OriginImplementation || output.DeferredImplementation[0].Object.ID != "billing://entity/payment" {
		t.Fatalf("implementation observation = %#v", output.DeferredImplementation)
	}
	if output.DeferredSlots[0].SlotID != "billing://activity/pay-order/implementation" || output.DeferredSlots[0].Span.End.Offset <= output.DeferredSlots[0].Span.Start.Offset || output.DeferredSlots[0].BodySpan.End.Offset <= output.DeferredSlots[0].BodySpan.Start.Offset {
		t.Fatalf("protected slot = %#v", output.DeferredSlots[0])
	}
	afterSource, afterInfo := snapshotFile(t, generated)
	if !bytes.Equal(beforeSource, afterSource) || !os.SameFile(beforeInfo, afterInfo) || beforeInfo.ModTime() != afterInfo.ModTime() || beforeInfo.Mode() != afterInfo.Mode() {
		t.Fatal("analyze mutated generated Go input")
	}
}
func TestRunAnalyzePreservesStableIDsAcrossDisplayRename(t *testing.T) {
	_, generated := billingAnalyzeFiles(t, billingAnalyzeAuthority)
	authority, _ := writeAnalyzeFile(t, "renamed.gooo", billingAnalyzeRenamedAuthority)
	output, code, stderr := runAnalyzePaths(authority, generated)
	if code != exitOK || stderr != "" {
		t.Fatalf("renamed billing analyze = %d, stderr=%q, output=%s", code, stderr, output)
	}
	var delta analyzeDeltaOutput
	if err := json.Unmarshal(output, &delta); err != nil {
		t.Fatal(err)
	}
	if !delta.SemanticEqual || len(delta.SignatureFacts) != 3 {
		t.Fatalf("renamed billing delta = %#v", delta)
	}
	for _, fact := range delta.SignatureFacts {
		if fact.Fact.Subject.String() != "billing://activity/pay-order" && fact.Fact.Object.String() != "billing://activity/pay-order" {
			t.Fatalf("display rename changed stable semantic IDs: %#v", fact.Fact)
		}
	}
}
