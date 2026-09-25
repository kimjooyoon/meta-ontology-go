package analyzer

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestDeferredFactTamperFailsReconcileWithoutWrite(t *testing.T) {
	source := SourceFile{Filename: "deferred.go", PackagePath: "billing", Source: []byte(`package billing

//gooo:semantic activity id="billing://activity/run" namespace=billing
func Run(order Order) Payment { return Payment{} }

//gooo:semantic entity id="billing://entity/order" namespace=billing
type Order struct{}

//gooo:semantic entity id="billing://entity/payment" namespace=billing
type Payment struct{}
`)}
	result, err := AnalyzeAndAdaptSemantic(SourceSemanticAdapterInput{
		Base:    semantic.NewIR("billing", semantic.Namespace("billing")),
		Sources: []SourceFile{source}, Registry: NewRegistry(), Policy: billingPolicy(t, RelationUses),
		Producer: semantic.GoHostedCompilerID, EvidenceKind: semantic.CompilerRunEvidence,
		ToolchainIdentity: "go1.26.5|test/amd64",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.DeferredFacts) != 1 {
		t.Fatalf("deferred facts = %d, want one", len(result.DeferredFacts))
	}
	before := irSnapshot(result.IR)
	result.DeferredFacts[0].Object.ID = "billing://entity/tampered"

	reconcile := ReconcileSemantic(result, result.IR, result.SourceDigest, result.PolicyDigest,
		result.ToolchainDigest, result.ImplementationObservationDigest)
	if reconcile.Accepted || reconcile.DeltaValid || reconcile.WriteEffect != ReconcileNoWrite ||
		reconcile.FailureCode != "invalid-delta-binding" {
		t.Fatalf("tampered deferred fact reconcile = %#v, want invalid no-write", reconcile)
	}
	if got := irSnapshot(result.IR); got != before {
		t.Fatalf("reconcile mutated IR: before=%q after=%q", before, got)
	}
}
