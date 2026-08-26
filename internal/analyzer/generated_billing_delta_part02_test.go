package analyzer

import (
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
	"testing"
)

func TestGeneratedBillingMutationReturnsNoWriteAndChangesDelta(t *testing.T) {
	source := generatedBillingSource(t)
	policy := generatedBillingPolicy(t)
	registry := generatedBillingRegistry(t)
	first := adaptGeneratedBillingSource(t, source, registry, policy)
	mutated := source
	mutated.Source = []byte(strings.Replace(
		string(source.Source), "type Order struct{}", "type Order struct{ _ byte }", 1,
	))
	second := adaptGeneratedBillingSource(t, mutated, generatedBillingRegistry(t), policy)
	before := irSnapshot(first.IR)
	reconcile := ReconcileSemantic(second, first.IR, first.SourceDigest, first.PolicyDigest,
		first.ToolchainDigest, first.ImplementationObservationDigest)
	if first.NormalizedDelta.Digest == second.NormalizedDelta.Digest {
		t.Fatal("generated body mutation retained normalized delta digest")
	}
	if reconcile.Accepted || reconcile.WriteEffect != ReconcileNoWrite ||
		reconcile.FailureCode != "source-or-observation-mismatch" {
		t.Fatalf("mutation reconcile = %#v, want source-bound no-write", reconcile)
	}
	if got := irSnapshot(first.IR); got != before {
		t.Fatalf("rejected mutation changed expected IR: before=%q after=%q", before, got)
	}
}
func TestGeneratedBillingPromotedCandidateReturnsNoWrite(t *testing.T) {
	source := generatedBillingSource(t)
	policy := generatedBillingPolicy(t)
	observed := adaptGeneratedBillingSource(t, source, ambiguousGeneratedBillingRegistry(t), policy)
	key := observed.IR.Graph.Candidates()[0].Key()
	if _, err := observed.IR.Graph.PromoteCandidate(key); err != nil {
		t.Fatalf("promote test candidate: %v", err)
	}
	reconcile := ReconcileSemantic(observed, declaredBillingContract(t), observed.SourceDigest,
		observed.PolicyDigest, observed.ToolchainDigest, observed.ImplementationObservationDigest)
	if reconcile.Accepted || reconcile.AuthoritySafe || reconcile.WriteEffect != ReconcileNoWrite {
		t.Fatalf("promoted candidate reconcile = %#v, want no-write rejection", reconcile)
	}
}
func TestGeneratedBillingWrongIdentityReturnsNoWrite(t *testing.T) {
	source := generatedBillingSource(t)
	policy := generatedBillingPolicy(t)
	observed := adaptGeneratedBillingSource(
		t, source, generatedBillingRegistryWithOrder(t, "billing://entity/wrong-order"), policy,
	)
	reconcile := ReconcileSemantic(observed, declaredBillingContract(t), observed.SourceDigest,
		observed.PolicyDigest, observed.ToolchainDigest, observed.ImplementationObservationDigest)
	if reconcile.Accepted || reconcile.WriteEffect != ReconcileNoWrite ||
		reconcile.FailureCode != "identity-mismatch" {
		t.Fatalf("wrong identity reconcile = %#v, want identity no-write", reconcile)
	}
}
func TestGeneratedBillingDuplicateSourceReportsNoWrite(t *testing.T) {
	source := generatedBillingSource(t)
	_, err := AnalyzeAndAdaptSemantic(SourceSemanticAdapterInput{
		Base:    semantic.NewIR("billing", semantic.Namespace("billing")),
		Sources: []SourceFile{source, source}, Registry: generatedBillingRegistry(t),
		Policy: generatedBillingPolicy(t), Producer: semantic.GoHostedCompilerID,
		EvidenceKind: semantic.CompilerRunEvidence, ToolchainIdentity: billingToolchain(),
	})
	var adapterErr AdapterError
	if !errors.As(err, &adapterErr) || adapterErr.Code != AdapterSourceConfig ||
		adapterErr.WriteEffect != ReconcileNoWrite {
		t.Fatalf("duplicate source error = %#v, want source-config/no-write", err)
	}
}
