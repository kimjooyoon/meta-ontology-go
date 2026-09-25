package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestUnresolvedDetailsStayOutsideAdapterAuthorityAndRejectTampering(t *testing.T) {
	source := SourceFile{Filename: "unknown.go", PackagePath: "billing", Source: []byte(`package billing

//gooo:semantic activity id="billing://activity/run" namespace=billing
func Run() { missingIdentifier }
`)}
	base := semantic.NewIR("billing", semantic.Namespace("billing"))
	baseBefore := irSnapshot(base)
	result, err := AnalyzeAndAdaptSemantic(SourceSemanticAdapterInput{
		Base: base, Sources: []SourceFile{source}, Registry: NewRegistry(),
		Policy: emptyPolicy(t), Producer: semantic.GoHostedCompilerID,
		EvidenceKind: semantic.CompilerRunEvidence, ToolchainIdentity: "go1.26.5|test/amd64",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := irSnapshot(base); got != baseBefore {
		t.Fatalf("base IR changed during deferred adaptation: before=%q after=%q", baseBefore, got)
	}
	if len(result.IR.Graph.DeterministicFacts()) != 0 || len(result.IR.Graph.Candidates()) != 0 {
		t.Fatalf("unknown detail entered semantic graph: %#v", result.IR.Graph)
	}
	if len(result.NormalizedDelta.DeferredDetails) != 1 ||
		result.NormalizedDelta.DeferredDetails[0].Detail.IdentityState != IdentityUnresolved {
		t.Fatalf("normalized deferred details = %#v", result.NormalizedDelta.DeferredDetails)
	}
	before := irSnapshot(result.IR)
	result.NormalizedDelta.DeferredDetails[0].Detail.IdentityState = IdentityState("tampered")
	reconcile := ReconcileSemantic(result, result.IR, result.SourceDigest, result.PolicyDigest,
		result.ToolchainDigest, result.ImplementationObservationDigest)
	if reconcile.Accepted || reconcile.WriteEffect != ReconcileNoWrite || reconcile.FailureCode != "invalid-delta-binding" {
		t.Fatalf("tampered deferred detail reconcile = %#v, want no-write rejection", reconcile)
	}
	if got := irSnapshot(result.IR); got != before {
		t.Fatalf("reconcile mutated IR: before=%q after=%q", before, got)
	}
}
func TestTopLevelUnresolvedDetailTamperFailsReconcileWithoutWrite(t *testing.T) {
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
	result.ImplementationDetails[0].Reference = "tamperedReference"

	reconcile := ReconcileSemantic(result, result.IR, result.SourceDigest, result.PolicyDigest,
		result.ToolchainDigest, result.ImplementationObservationDigest)
	if reconcile.Accepted || reconcile.DeltaValid || reconcile.WriteEffect != ReconcileNoWrite ||
		reconcile.FailureCode != "invalid-delta-binding" {
		t.Fatalf("top-level detail tamper reconcile = %#v, want invalid no-write", reconcile)
	}
	if got := irSnapshot(result.IR); got != before {
		t.Fatalf("reconcile mutated IR: before=%q after=%q", before, got)
	}
}
