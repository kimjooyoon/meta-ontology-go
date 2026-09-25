package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func makeCouplingInput(delta, candidate bool) Input {
	profile := ProfileConfig{ID: "gooo-profile/coupling", Version: "v1", Digest: digestText("profile-v1")}
	input := Input{
		Schema: SchemaV1, FixtureID: "fixture-label/initial",
		Config:                EvaluationConfig{ToolchainDigest: digestText("go1.26.5"), Profile: profile},
		AuthoritySourceBefore: "package billing\nactivity PayOrder(Order, PaymentMethod) -> Payment\n",
		AuthoritySourceAfter:  "package billing\nactivity PayOrder(Order, PaymentMethod) -> Payment\n",
		SemanticBefore:        baseSemantic(false), SemanticAfter: baseSemantic(false),
		Registry: []CodeBinding{
			{RegisteredSurfaceID: "urn:gooo:surface:billing/pay-order", CodeSymbolID: "urn:gooo:code:billing/pay-order", SemanticOwnerID: "urn:gooo:owner:billing/pay-order", SourceMapID: "sm.billing.pay-order", PackageLabel: "billing", FileLabel: "billing/generated.go", SourceSpan: "12:1-18:2"},
			{RegisteredSurfaceID: "urn:gooo:surface:billing/pay-order-helper", CodeSymbolID: "urn:gooo:code:billing/pay-order-helper", SemanticOwnerID: "urn:gooo:owner:billing/pay-order-helper", SourceMapID: "sm.billing.pay-order-helper", PackageLabel: "billing", FileLabel: "billing/helper.go", SourceSpan: "3:1-5:2"},
		},
		Changes: []CodeChange{{CodeSymbolID: "urn:gooo:code:billing/pay-order", BeforeDigest: digestText("code-before"), AfterDigest: digestText("code-after")}, {CodeSymbolID: "urn:gooo:code:billing/pay-order-helper", BeforeDigest: digestText("helper"), AfterDigest: digestText("helper")}},
		Roots:   []string{"urn:gooo:source:billing"},
	}
	for i := range input.Registry {
		input.Registry[i].BindingDigest = bindingDigest(input.Registry[i])
	}
	registry, _ := normalizeRegistry(input.Registry)
	input.RegistryDigest = registry.digest
	if delta {
		input.SemanticAfter = baseSemantic(true)
		input.AuthoritySourceAfter = "package billing\nactivity PayOrder(Order, PaymentMethod) -> Payment\nactivity AuthorizePayment(Payment) -> Receipt\n"
	}
	before, _ := normalizeSemantic(input.SemanticBefore)
	after, _ := normalizeSemantic(input.SemanticAfter)
	input.Manifest = SourceManifest{Complete: true, BeforeSnapshotDigest: stateSnapshotDigest(input.AuthoritySourceBefore, before.digest, registry.digest, input.Config), AfterSnapshotDigest: stateSnapshotDigest(input.AuthoritySourceAfter, after.digest, registry.digest, input.Config), ToolchainDigest: input.Config.ToolchainDigest, ProfileDigest: input.Config.Profile.Digest, RegistryDigest: registry.digest}
	deltaText, _, _ := semanticDelta(before.facts, after.facts)
	if delta {
		input.Receipts, input.Path = makePathAndReceipt(input, registry, before.digest, after.digest, deltaText, ClaimDelta, candidate)
	} else {
		input.Receipts, input.Path = makePathAndReceipt(input, registry, before.digest, after.digest, deltaText, ClaimNoDelta, candidate)
	}
	resourcesSnapshot := snapshotDigest(input, before.digest, after.digest, registry.digest)
	providerID, observerID := "urn:gooo:resource-provider:runner", "urn:gooo:resource-observer:coupling"
	input.Config.ResourceBinding = ResourceBindingConfig{ProviderID: providerID, ObserverID: observerID, ProviderDigest: resourceProviderDigest(providerID), ObserverDigest: resourceObserverDigest(observerID), SnapshotDigest: resourcesSnapshot, SourceDigest: resourceSourceDigest(providerID, observerID, resourcesSnapshot)}
	input.ResourceRegistry = input.Config.ResourceBinding
	input.ResourceReceipts = makeResourceReceipts(input.Config.ResourceBinding)
	return input
}
func baseSemantic(withDelta bool) SemanticIR {
	ir := SemanticIR{Nodes: []SemanticNode{
		{ID: "urn:gooo:entity:order", Kind: semantic.Entity.String(), Namespace: "billing", Name: "Order", Aliases: []string{"Purchase"}},
		{ID: "urn:gooo:entity:payment-method", Kind: semantic.Entity.String(), Namespace: "billing", Name: "PaymentMethod"},
		{ID: "urn:gooo:entity:payment", Kind: semantic.Entity.String(), Namespace: "billing", Name: "Payment"},
		{ID: "urn:gooo:activity:pay-order", Kind: semantic.Activity.String(), Namespace: "billing", Name: "PayOrder"},
	}, Relations: []SemanticRelation{
		{Subject: "urn:gooo:activity:pay-order", Predicate: "uses", Object: "urn:gooo:entity:order"},
		{Subject: "urn:gooo:activity:pay-order", Predicate: "uses", Object: "urn:gooo:entity:payment-method"},
		{Subject: "urn:gooo:entity:payment", Predicate: "wasGeneratedBy", Object: "urn:gooo:activity:pay-order"},
	}}
	if withDelta {
		ir.Nodes = append(ir.Nodes, SemanticNode{ID: "urn:gooo:entity:receipt", Kind: semantic.Entity.String(), Namespace: "billing", Name: "Receipt"})
		ir.Relations = append(ir.Relations, SemanticRelation{Subject: "urn:gooo:activity:pay-order", Predicate: "emits", Object: "urn:gooo:entity:receipt"})
	}
	return ir
}
