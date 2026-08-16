package coupling

import (
	"testing"

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

func makePathAndReceipt(input Input, registry registryView, beforeDigest, afterDigest, deltaText string, claim ChangeClaim, candidate bool) ([]CouplingReceipt, semantic.InferencePathV1) {
	binding := registry.bySurface["urn:gooo:surface:billing/pay-order"]
	root := semantic.MustIdentity("urn:gooo:source:billing")
	owner := semantic.MustIdentity(binding.SemanticOwnerID)
	term := semantic.MustIdentity("urn:gooo:term:billing/pay-order")
	code := semantic.MustIdentity(binding.CodeSymbolID)
	receiptID := semantic.MustIdentity("urn:gooo:receipt:billing/pay-order")
	controls := semantic.InferenceControls{CatalogDigest: digestText("catalog-v1"), PolicyDigest: digestText("policy-v1"), Profile: semantic.ProfileBinding{ID: input.Config.Profile.ID, Version: input.Config.Profile.Version, Digest: input.Config.Profile.Digest}}
	snapshotBefore, snapshotAfter := semantic.SnapshotDigests{Source: sourceDigest(input.AuthoritySourceBefore), Semantic: beforeDigest}, semantic.SnapshotDigests{Source: sourceDigest(input.AuthoritySourceAfter), Semantic: afterDigest}
	evidence := []semantic.InferenceEvidence{}
	addEvidence := func(id string, independent, sourceBacked bool) semantic.EvidenceReference {
		parsed := semantic.MustIdentity(id)
		digest := digestText("evidence:" + id)
		evidence = append(evidence, semantic.InferenceEvidence{ID: parsed, Digest: digest, Before: snapshotBefore, After: snapshotAfter, SourceBacked: sourceBacked, Independent: independent, Controls: controls})
		return semantic.EvidenceReference{ID: parsed, Digest: digest}
	}
	declEvidence := addEvidence("urn:gooo:evidence:declaration", false, false)
	deriveEvidence := addEvidence("urn:gooo:evidence:derivation", false, false)
	projectionEvidence := addEvidence("urn:gooo:evidence:projection", false, false)
	verificationEvidence := addEvidence("urn:gooo:evidence:verification", true, true)
	record := func(id string, subject, object semantic.ID, rule string, phase semantic.InferencePhase, ordinal uint64, layer semantic.AuthorityLayer, effect semantic.AuthorityEffect, refs []semantic.EvidenceReference) semantic.InferenceRecord {
		return semantic.InferenceRecord{RecordID: semantic.MustIdentity(id), SubjectID: subject, ObjectID: object, Rule: semantic.RuleBinding{ID: semantic.MustIdentity("urn:gooo:rule:" + rule), Version: "v1", Digest: digestText("rule:" + rule)}, Phase: semantic.PhasePlacement{Phase: phase, Ordinal: ordinal}, Before: snapshotBefore, After: snapshotAfter, Authority: semantic.AuthorityBinding{Layer: layer, Effect: effect}, Evidence: refs, Controls: controls}
	}
	edges := []semantic.InferenceEdge{
		{InferenceRecord: record("urn:gooo:path:declaration", root, owner, "declaration", semantic.PhaseDeclaration, 1, semantic.AuthoritySource, semantic.AuthorityDeclare, []semantic.EvidenceReference{declEvidence}), Kind: semantic.InferenceAuthoritativeDeclaration, SourceRoots: []semantic.ID{root}},
		{InferenceRecord: record("urn:gooo:path:derivation", owner, term, "derivation", semantic.PhaseDerivation, 2, semantic.AuthoritySemantic, semantic.AuthorityDerive, []semantic.EvidenceReference{deriveEvidence}), Kind: semantic.InferenceDeterministicDerivation},
	}
	if candidate {
		candidateID := semantic.MustIdentity("urn:gooo:candidate:billing/pay-order")
		candidateRecord := record("urn:gooo:path:candidate", term, candidateID, "candidate", semantic.PhaseObservation, 3, semantic.AuthorityCandidate, semantic.AuthorityObserve, []semantic.EvidenceReference{projectionEvidence})
		candidateRecord.Before.Semantic, candidateRecord.After.Semantic = beforeDigest, beforeDigest
		edges = append(edges, semantic.InferenceEdge{InferenceRecord: candidateRecord, Kind: semantic.InferenceObservationCandidate})
		edges = append(edges, semantic.InferenceEdge{InferenceRecord: record("urn:gooo:path:projection", candidateID, code, "projection", semantic.PhaseProjection, 4, semantic.AuthorityDerived, semantic.AuthorityProject, []semantic.EvidenceReference{projectionEvidence}), Kind: semantic.InferenceDerivedProjection})
	} else {
		edges = append(edges, semantic.InferenceEdge{InferenceRecord: record("urn:gooo:path:projection", term, code, "projection", semantic.PhaseProjection, 3, semantic.AuthorityDerived, semantic.AuthorityProject, []semantic.EvidenceReference{projectionEvidence}), Kind: semantic.InferenceDerivedProjection})
	}
	verification := record("urn:gooo:path:verification", code, receiptID, "verification", semantic.PhaseVerification, 5, semantic.AuthorityVerification, semantic.AuthorityVerify, []semantic.EvidenceReference{verificationEvidence})
	edges = append(edges, semantic.InferenceEdge{InferenceRecord: verification, Kind: semantic.InferenceIndependentVerification})
	claimRecord := record("urn:gooo:claim:billing/pay-order", owner, receiptID, "claim", semantic.PhaseVerification, 6, semantic.AuthoritySemantic, semantic.AuthorityNoChange, []semantic.EvidenceReference{verificationEvidence})
	semanticClaimKind := semantic.NoSemanticDelta
	if claim == ClaimDelta {
		claimRecord.Authority.Effect = semantic.AuthorityDelta
		semanticClaimKind = semantic.SemanticDelta
	}
	claimRecord.Before.Semantic, claimRecord.After.Semantic = beforeDigest, afterDigest
	claimRecord.Before.Source, claimRecord.After.Source = snapshotBefore.Source, snapshotAfter.Source
	claims := []semantic.SemanticChangeClaim{{InferenceRecord: claimRecord, Kind: semanticClaimKind}}
	if claim == ClaimDelta {
		claims[0].CanonicalDelta, claims[0].DeltaDigest = deltaText, digestText(deltaText)
	}
	path := semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion, Edges: edges, Claims: claims, Evidence: evidence}
	originPathIDs := []string{"urn:gooo:path:verification", "urn:gooo:path:projection", "urn:gooo:path:derivation", "urn:gooo:path:declaration"}
	if candidate {
		originPathIDs = []string{"urn:gooo:path:verification", "urn:gooo:path:projection", "urn:gooo:path:candidate", "urn:gooo:path:derivation", "urn:gooo:path:declaration"}
	}
	receipt := CouplingReceipt{ReceiptID: receiptID.String(), SurfaceID: binding.RegisteredSurfaceID, SemanticOwnerID: binding.SemanticOwnerID, CodeSymbolID: binding.CodeSymbolID, SourceMapBindingDigest: binding.BindingDigest, SnapshotDigest: snapshotDigest(input, beforeDigest, afterDigest, registry.digest), RegistryDigest: registry.digest, ToolchainDigest: input.Config.ToolchainDigest, ProfileDigest: input.Config.Profile.Digest, BeforeIRDigest: beforeDigest, AfterIRDigest: afterDigest, AuthoritySourceBeforeDigest: sourceDigest(input.AuthoritySourceBefore), AuthoritySourceAfterDigest: sourceDigest(input.AuthoritySourceAfter), ChangeClaim: claim, ReceiptKind: ReceiptNoSemanticDelta, OriginPathIDs: originPathIDs, ClaimRecordID: claimRecord.RecordID.String(), EvidenceRefs: []string{declEvidence.ID.String(), deriveEvidence.ID.String(), projectionEvidence.ID.String(), verificationEvidence.ID.String()}, State: "CURRENT"}
	if claim == ClaimDelta {
		receipt.ReceiptKind = ReceiptSemanticDelta
		receipt.SemanticDelta, receipt.SemanticDeltaDigest, receipt.AuthoritativeSourceRef = deltaText, digestText(deltaText), "gooo://billing/source#authorize-payment"
	}
	return []CouplingReceipt{receipt}, path
}

func makeResourceReceipts(binding ResourceBindingConfig) []ExternalResourceReceipt {
	values := []ExternalResourceReceipt{{ReceiptID: "urn:gooo:resource:cpu", Metric: "cpu-core-ns", Value: 10, Unit: "ns"}, {ReceiptID: "urn:gooo:resource:memory", Metric: "peak-memory-bytes", Value: 20, Unit: "bytes"}, {ReceiptID: "urn:gooo:resource:work", Metric: "work-units", Value: 30, Unit: "units"}}
	for i := range values {
		values[i].ProviderDigest, values[i].ObserverDigest, values[i].SnapshotDigest, values[i].SourceDigest = binding.ProviderDigest, binding.ObserverDigest, binding.SnapshotDigest, binding.SourceDigest
		values[i].Present, values[i].Independent, values[i].State = true, true, "CURRENT"
		values[i].BindingDigest = resourceBindingDigest(values[i])
	}
	return values
}

func cloneInput(input Input) Input {
	output := input
	output.SemanticBefore = cloneSemanticIR(input.SemanticBefore)
	output.SemanticAfter = cloneSemanticIR(input.SemanticAfter)
	output.Registry = append([]CodeBinding(nil), input.Registry...)
	output.Changes = append([]CodeChange(nil), input.Changes...)
	output.Receipts = append([]CouplingReceipt(nil), input.Receipts...)
	for i := range output.Receipts {
		output.Receipts[i].EvidenceRefs = append([]string(nil), input.Receipts[i].EvidenceRefs...)
	}
	output.ResourceReceipts = append([]ExternalResourceReceipt(nil), input.ResourceReceipts...)
	output.Roots = append([]string(nil), input.Roots...)
	output.Path.Edges = append([]semantic.InferenceEdge(nil), input.Path.Edges...)
	output.Path.Claims = append([]semantic.SemanticChangeClaim(nil), input.Path.Claims...)
	output.Path.Evidence = append([]semantic.InferenceEvidence(nil), input.Path.Evidence...)
	for i := range output.Path.Edges {
		output.Path.Edges[i].SourceRoots = append([]semantic.ID(nil), input.Path.Edges[i].SourceRoots...)
		output.Path.Edges[i].Evidence = append([]semantic.EvidenceReference(nil), input.Path.Edges[i].Evidence...)
	}
	for i := range output.Path.Claims {
		output.Path.Claims[i].Evidence = append([]semantic.EvidenceReference(nil), input.Path.Claims[i].Evidence...)
	}
	return output
}

func digestText(value string) string { return digestBytes([]byte(value)) }

func TestFixtureBuilderSanity(t *testing.T) {
	for _, row := range testCorpus() {
		if row.Name == "" || row.Input.FixtureID == "" || row.Expected.Decision == "" || row.Expected.Reason == "" {
			t.Fatal("fixture metadata missing")
		}
	}
}
