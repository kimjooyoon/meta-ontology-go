package analyzer

import (
	"encoding/json"
	"errors"
	"runtime"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestGeneratedBillingNormalizedDeltaIsStableAcrossRunsAndOrder(t *testing.T) {
	source := generatedBillingSource(t)
	policy := generatedBillingPolicy(t)
	first := adaptGeneratedBillingSource(t, source, generatedBillingRegistry(t), policy)
	repeated := adaptGeneratedBillingSource(t, source, generatedBillingRegistry(t), policy)
	permuted := adaptGeneratedBillingSource(t, source, reversedGeneratedBillingRegistry(t), policy)

	for _, result := range []SemanticAdapterResult{first, repeated, permuted} {
		if len(result.NormalizedDelta.SignatureFacts) != 3 || len(result.NormalizedDelta.CandidateFacts) != 0 {
			t.Fatalf("normalized delta classes = %d signature, %d candidate",
				len(result.NormalizedDelta.SignatureFacts), len(result.NormalizedDelta.CandidateFacts))
		}
		if len(result.NormalizedDelta.DeferredImplementation) != 1 {
			t.Fatalf("deferred implementation observations = %d, want 1", len(result.NormalizedDelta.DeferredImplementation))
		}
		assertGeneratedDeltaBindings(t, result)
	}
	if first.NormalizedDelta.Digest != repeated.NormalizedDelta.Digest ||
		first.NormalizedDelta.Digest != permuted.NormalizedDelta.Digest {
		t.Fatal("normalized delta changed across repeat or registration-order permutation")
	}
	deferred := ReconcileSemantic(first, declaredBillingContract(t), first.SourceDigest,
		first.PolicyDigest, first.ToolchainDigest, "")
	if deferred.Accepted || deferred.WriteEffect != ReconcileNoWrite ||
		deferred.FailureCode != "source-or-observation-mismatch" {
		t.Fatalf("deferred implementation reconcile = %#v, want no-write", deferred)
	}
	payload, err := json.Marshal(first.NormalizedDelta)
	if err != nil {
		t.Fatalf("marshal normalized delta: %v", err)
	}
	for _, field := range []string{
		"schema_version", "signature_facts", "candidate_facts", "deferred_implementation",
		"deferred_details", "deferred_slots", "registry_digest", "digest",
	} {
		if !strings.Contains(string(payload), `"`+field+`"`) {
			t.Fatalf("machine-readable delta omitted %q: %s", field, payload)
		}
	}
}

func TestGeneratedBillingCandidateStaysNonAuthoritative(t *testing.T) {
	source := generatedBillingSource(t)
	policy := generatedBillingPolicy(t)
	observed := adaptGeneratedBillingSource(t, source, ambiguousGeneratedBillingRegistry(t), policy)
	if len(observed.NormalizedDelta.CandidateFacts) != 1 {
		t.Fatalf("candidate facts = %d, want 1", len(observed.NormalizedDelta.CandidateFacts))
	}
	candidate := observed.NormalizedDelta.CandidateFacts[0]
	if len(candidate.Facts) != 2 || len(candidate.Evidence) != 2 {
		t.Fatalf("candidate handoff = %d facts, %d evidence; want two each", len(candidate.Facts), len(candidate.Evidence))
	}
	for _, fact := range candidate.Facts {
		if fact.Status != semantic.FactCandidate || observed.IR.Graph.HasFact(fact.Key()) ||
			!observed.IR.Graph.HasCandidate(fact.Key()) {
			t.Fatalf("candidate crossed authority boundary: %#v", fact)
		}
	}
	reconcile := ReconcileSemantic(observed, declaredBillingContract(t), observed.SourceDigest,
		observed.PolicyDigest, observed.ToolchainDigest, observed.ImplementationObservationDigest)
	if reconcile.Accepted || !reconcile.AuthoritySafe || reconcile.WriteEffect != ReconcileNoWrite {
		t.Fatalf("candidate reconcile = %#v, want no-write rejection", reconcile)
	}
}

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

func adaptGeneratedBillingSource(
	t *testing.T, source SourceFile, registry *Registry, policy MappingPolicy,
) SemanticAdapterResult {
	t.Helper()
	result, err := AnalyzeAndAdaptSemantic(SourceSemanticAdapterInput{
		Base: semantic.NewIR("billing", semantic.Namespace("billing")), Sources: []SourceFile{source},
		Registry: registry, Policy: policy, Producer: semantic.GoHostedCompilerID,
		EvidenceKind: semantic.CompilerRunEvidence, ToolchainIdentity: billingToolchain(),
	})
	if err != nil {
		t.Fatalf("adapt generated billing source: %v", err)
	}
	return result
}

func assertGeneratedDeltaBindings(t *testing.T, result SemanticAdapterResult) {
	t.Helper()
	base := result.NormalizedDelta.SignatureFacts[0].Binding.BaseDigest
	if !validDigest(result.RegistryDigest) || result.RegistryDigest != generatedBillingRegistry(t).Digest() {
		t.Fatalf("registry binding = %q", result.RegistryDigest)
	}
	for _, fact := range result.NormalizedDelta.SignatureFacts {
		binding := fact.Binding
		if !binding.complete() || binding.BaseDigest != base || binding.RegistryDigest != result.RegistryDigest ||
			fact.Fact.Span.File == "" || fact.Evidence.ID == "" {
			t.Fatalf("signature binding = %#v", binding)
		}
	}
	for _, observation := range result.NormalizedDelta.DeferredImplementation {
		if observation.BaseDigest != base || !validDigest(observation.SourceDigest) ||
			observation.RegistryDigest != result.RegistryDigest || observation.SourceFile == "" ||
			!validDigest(observation.Fingerprint()) {
			t.Fatalf("deferred observation binding = %#v, base=%q fingerprint=%q", observation, base, observation.Fingerprint())
		}
	}
}

func reversedGeneratedBillingRegistry(t *testing.T) *Registry {
	t.Helper()
	entries := []Registration{
		{Ref: billingRef("Payment"), Kind: KindEntity, Identity: NewIdentity("billing", "billing://entity/payment")},
		{Ref: billingRef("PaymentMethod"), Kind: KindEntity,
			Identity: NewIdentity("billing", "billing://entity/payment-method")},
		{Ref: billingRef("Order"), Kind: KindEntity, Identity: NewIdentity("billing", "billing://entity/order")},
		{Ref: billingRef("PayOrder"), Kind: KindActivity, Identity: NewIdentity("billing", "billing://activity/pay-order")},
	}
	registry := NewRegistry()
	for _, entry := range entries {
		if err := registry.Register(entry); err != nil {
			t.Fatal(err)
		}
	}
	return registry
}

func ambiguousGeneratedBillingRegistry(t *testing.T) *Registry {
	t.Helper()
	registry := generatedBillingRegistry(t)
	if err := registry.Register(Registration{
		Ref: billingRef("Order"), Kind: KindEntity,
		Identity: NewIdentity("billing-alt", "billing://entity/alternate-order"),
	}); err != nil {
		t.Fatal(err)
	}
	return registry
}

func billingToolchain() string {
	return "go1.26.5|" + runtime.GOOS + "/" + runtime.GOARCH
}
