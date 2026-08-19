package analyzer

import (
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestUnresolvedDetailMissingSourceFileFailsReconcileWithoutWrite(t *testing.T) {
	source := SourceFile{Filename: "unknown.go", PackagePath: "billing", Source: []byte(`package billing

//gooo:semantic activity id="billing://activity/run" namespace=billing
func Run() { missingIdentifier }
`)}
	result, err := AnalyzeAndAdaptSemantic(SourceSemanticAdapterInput{
		Base:    semantic.NewIR("billing", semantic.Namespace("billing")),
		Sources: []SourceFile{source}, Registry: NewRegistry(), Policy: emptyPolicy(t),
		Producer: semantic.GoHostedCompilerID, EvidenceKind: semantic.CompilerRunEvidence,
		ToolchainIdentity: "go1.26.5|test/amd64",
	})
	if err != nil {
		t.Fatal(err)
	}
	before := irSnapshot(result.IR)
	result.NormalizedDelta.DeferredDetails[0].Detail.Span.Filename = ""
	result.NormalizedDelta.Digest = result.NormalizedDelta.StableHash()
	result.BindingDigest = semanticAdapterBindingDigest(result)

	reconcile := ReconcileSemantic(result, result.IR, result.SourceDigest, result.PolicyDigest,
		result.ToolchainDigest, result.ImplementationObservationDigest)
	if reconcile.Accepted || reconcile.DeltaValid || reconcile.WriteEffect != ReconcileNoWrite ||
		reconcile.FailureCode != "invalid-delta-binding" {
		t.Fatalf("missing source file reconcile = %#v, want invalid no-write", reconcile)
	}
	if got := irSnapshot(result.IR); got != before {
		t.Fatalf("reconcile mutated IR: before=%q after=%q", before, got)
	}
}
func TestMalformedSourceRejectsBeforeDeferredCommit(t *testing.T) {
	base := semantic.NewIR("billing", semantic.Namespace("billing"))
	before := irSnapshot(base)
	_, err := AnalyzeAndAdaptSemantic(SourceSemanticAdapterInput{
		Base: base, Sources: []SourceFile{{Filename: "broken.go", Source: []byte("package billing\nfunc Run(\n")}},
		Registry: NewRegistry(), Policy: emptyPolicy(t), Producer: semantic.GoHostedCompilerID,
		EvidenceKind: semantic.CompilerRunEvidence, ToolchainIdentity: "go1.26.5|test/amd64",
	})
	if err == nil {
		t.Fatal("malformed source was accepted")
	}
	var adapterErr AdapterError
	if errors.As(err, &adapterErr) {
		t.Fatalf("malformed source returned adapter error instead of parse error: %v", err)
	}
	if got := irSnapshot(base); got != before {
		t.Fatalf("malformed source changed base IR: before=%q after=%q", before, got)
	}
}
