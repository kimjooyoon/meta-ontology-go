package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"runtime"
	"testing"
)

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
