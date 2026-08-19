package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"runtime"
	"testing"
)

func TestGeneratedBillingProjectionLiftsAgainstDeclaredContract(t *testing.T) {
	source := generatedBillingSource(t)
	policy := generatedBillingPolicy(t)
	registry := generatedBillingRegistry(t)
	base := semantic.NewIR("billing", semantic.Namespace("billing"))
	toolchain := "go1.26.5|" + runtime.GOOS + "/" + runtime.GOARCH
	input := SourceSemanticAdapterInput{
		Base: base, Sources: []SourceFile{source}, Registry: registry, Policy: policy,
		Producer: semantic.GoHostedCompilerID, EvidenceKind: semantic.CompilerRunEvidence,
		ToolchainIdentity: toolchain,
	}
	adapted, err := AnalyzeAndAdaptSemantic(input)
	if err != nil {
		t.Fatalf("lift generated billing projection: %v", err)
	}
	delta, err := AnalyzePackage([]SourceFile{source}, registry)
	if err != nil {
		t.Fatalf("analyze generated billing projection: %v", err)
	}
	if len(delta.Delta.Added) != 4 || len(delta.Delta.Candidates) != 0 {
		t.Fatalf("generated delta = %#v, want four facts and no candidates", delta.Delta)
	}
	for _, fact := range delta.Delta.Added[:3] {
		if fact.Origin != OriginSignature {
			t.Fatalf("contract fact origin = %q, want signature", fact.Origin)
		}
	}
	if delta.Delta.Added[3].Origin != OriginImplementation {
		t.Fatalf("implementation fact origin = %q, want implementation", delta.Delta.Added[3].Origin)
	}
	expected := declaredBillingContract(t)
	reconcile := ReconcileSemantic(adapted, expected, adapted.SourceDigest, policy.Digest(),
		ToolchainDigest(toolchain), "")
	if !reconcile.Comparison.SemanticEqual {
		t.Fatalf("lifted semantic facts do not match DSL contract: %#v", reconcile.Comparison)
	}
	if reconcile.Accepted || reconcile.Comparison.ProvenanceEqual {
		t.Fatal("synthetic or absent external provenance was accepted")
	}
	if len(adapted.IR.Graph.DeterministicFacts()) != 3 {
		t.Fatalf("authoritative facts = %d, want 3", len(adapted.IR.Graph.DeterministicFacts()))
	}
	if len(adapted.DeferredFacts) != 1 || adapted.DeferredFacts[0].Origin != OriginImplementation {
		t.Fatalf("implementation observation was not deferred: %#v", adapted.DeferredFacts)
	}
	if len(adapted.ImplementationObservations) != 1 {
		t.Fatalf("implementation observations = %d, want 1", len(adapted.ImplementationObservations))
	}
	observation := adapted.ImplementationObservations[0]
	if observation.SourceFile != source.Filename || observation.BaseDigest != base.StableHash() {
		t.Fatalf("implementation observation binding = %#v", observation)
	}
	if observation.Fingerprint() == "" || !validDigest(adapted.ImplementationObservationDigest) {
		t.Fatalf("missing implementation observation digest: %#v", adapted)
	}
	if len(adapted.DeferredCandidates) != 0 || len(adapted.ImplementationDetails) != 0 {
		t.Fatalf("generated projection produced unexpected candidate/detail evidence")
	}
	if !validDigest(adapted.SourceDigest) || !validDigest(adapted.BindingDigest) {
		t.Fatalf("missing source binding: source=%q binding=%q", adapted.SourceDigest, adapted.BindingDigest)
	}
	digest, err := SourceBundleDigest([]SourceFile{source})
	if err != nil || adapted.SourceDigest != digest {
		t.Fatalf("source digest binding = %q, want %q (err=%v)", adapted.SourceDigest, digest, err)
	}
	if err := adapted.IR.Validate(); err != nil {
		t.Fatalf("lifted IR invalid: %v", err)
	}
}
