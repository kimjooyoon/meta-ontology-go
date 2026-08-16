package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func testCorpus() []CorpusCase {
	return append(testCorpusPositive(), testCorpusNegative()...)
}

func testCorpusPositive() []CorpusCase {
	base := makeCouplingInput(false, false)
	positive := FixtureExpectation{Decision: DecisionPass, Reason: ReasonNone, AcceptedSurfaces: []string{"urn:gooo:surface:billing/pay-order"}, ChangedSurfaces: []string{"urn:gooo:surface:billing/pay-order"}, ReceiptSurfaces: []string{"urn:gooo:surface:billing/pay-order"}, ObservationCounts: ObservationCounts{RegistryBindings: 2, ChangedCodeSurfaces: 1, ChangedRegistered: 1, ReceiptRecords: 1, ValidReceipts: 1, PathEdges: 4, PathClaims: 1, PathEvidence: 4, ResourceReceipts: 3}, Resources: ResourceObservation{CPUCoreNS: 10, PeakMemoryBytes: 20, WorkUnits: 30}}
	delta := makeCouplingInput(true, false)
	deltaExpected := positive
	deltaExpected.ObservationCounts.AddedSemanticFacts = 2
	candidate := makeCouplingInput(false, true)
	candidateExpected := positive
	candidateExpected.ObservationCounts.PathEdges = 5
	candidateExpected.ObservationCounts.CandidateObservations = 1
	noWrite := makeCouplingInput(false, false)
	noWrite.Changes = nil
	noWrite.Manifest.ZeroChange = true
	noWrite.Receipts = nil
	noWrite.Path = semantic.InferencePathV1{}
	noWrite.Roots = nil
	noWriteExpected := FixtureExpectation{Decision: DecisionPass, Reason: ReasonNone, AcceptedSurfaces: []string{}, ChangedSurfaces: []string{}, ReceiptSurfaces: []string{}, ObservationCounts: ObservationCounts{RegistryBindings: 2, ResourceReceipts: 3}, Resources: ResourceObservation{CPUCoreNS: 10, PeakMemoryBytes: 20, WorkUnits: 30}}
	return []CorpusCase{{Name: "positive-no-delta", Input: base, Expected: positive}, {Name: "positive-semantic-delta", Input: delta, Expected: deltaExpected}, {Name: "positive-candidate-observation", Input: candidate, Expected: candidateExpected}, {Name: "positive-no-write", Input: noWrite, Expected: noWriteExpected}}
}

func testCorpusNegative() []CorpusCase {
	return negativeCorpusRows(negativeCorpusInputs())
}

func negativeCorpusInputs() map[string]Input {
	base := makeCouplingInput(false, false)
	delta := makeCouplingInput(true, false)
	inputs := map[string]Input{}
	missing := cloneInput(base)
	missing.Receipts, missing.Path, missing.Roots = nil, semantic.InferencePathV1{}, nil
	inputs["missing-receipt"] = missing
	duplicate := cloneInput(base)
	duplicate.Receipts = append(duplicate.Receipts, duplicate.Receipts[0])
	inputs["duplicate-receipt"] = duplicate
	orphan := cloneInput(base)
	orphan.Receipts[0].SurfaceID = "urn:gooo:surface:orphan"
	inputs["orphan-receipt"] = orphan
	stale := cloneInput(base)
	stale.Receipts[0].SnapshotDigest = digestText("stale snapshot")
	inputs["stale-receipt"] = stale
	unregistered := cloneInput(base)
	unregistered.Changes[0].CodeSymbolID = "urn:gooo:code:unregistered"
	inputs["unregistered-surface"] = unregistered
	deltaNoSource := cloneInput(delta)
	deltaNoSource.Receipts[0].AuthoritativeSourceRef = ""
	inputs["delta-without-source"] = deltaNoSource
	inputs["no-delta-without-equality"] = makeNoDeltaWithoutEquality(base, delta)
	missingRoot := cloneInput(base)
	missingRoot.Roots = []string{"urn:gooo:source:wrong"}
	inputs["missing-root"] = missingRoot
	unproven := positiveNoWriteInput()
	unproven.Manifest.Complete = false
	inputs["unproven-zero-change"] = unproven
	return inputs
}

func positiveNoWriteInput() Input {
	input := makeCouplingInput(false, false)
	input.Changes = nil
	input.Manifest.ZeroChange = true
	input.Receipts = nil
	input.Path = semantic.InferencePathV1{}
	input.Roots = nil
	return input
}

func makeNoDeltaWithoutEquality(base, delta Input) Input {
	input := cloneInput(base)
	input.SemanticAfter = delta.SemanticAfter
	input.AuthoritySourceAfter = delta.AuthoritySourceAfter
	before, _ := normalizeSemantic(input.SemanticBefore)
	after, _ := normalizeSemantic(input.SemanticAfter)
	registry, _ := normalizeRegistry(input.Registry)
	input.Manifest.AfterSnapshotDigest = stateSnapshotDigest(input.AuthoritySourceAfter, after.digest, registry.digest, input.Config)
	snapshot := snapshotDigest(input, before.digest, after.digest, registry.digest)
	input.Config.ResourceBinding.SnapshotDigest = snapshot
	input.Config.ResourceBinding.SourceDigest = resourceSourceDigest(input.Config.ResourceBinding.ProviderID, input.Config.ResourceBinding.ObserverID, snapshot)
	input.ResourceRegistry.SnapshotDigest = snapshot
	input.ResourceRegistry.SourceDigest = input.Config.ResourceBinding.SourceDigest
	for i := range input.ResourceReceipts {
		input.ResourceReceipts[i].SnapshotDigest = snapshot
		input.ResourceReceipts[i].SourceDigest = input.Config.ResourceBinding.SourceDigest
		input.ResourceReceipts[i].BindingDigest = resourceBindingDigest(input.ResourceReceipts[i])
	}
	input.Receipts[0].AfterIRDigest = after.digest
	input.Receipts[0].AuthoritySourceAfterDigest = sourceDigest(input.AuthoritySourceAfter)
	input.Receipts[0].SnapshotDigest = snapshot
	updatePathAfter(input.Path.Edges, input.Path.Claims, input.Path.Evidence, after.digest, sourceDigest(input.AuthoritySourceAfter))
	return input
}

func updatePathAfter(edges []semantic.InferenceEdge, claims []semantic.SemanticChangeClaim, evidence []semantic.InferenceEvidence, digest, source string) {
	for i := range edges {
		edges[i].After.Semantic, edges[i].After.Source = digest, source
	}
	for i := range claims {
		claims[i].After.Semantic, claims[i].After.Source = digest, source
	}
	for i := range evidence {
		evidence[i].After.Semantic, evidence[i].After.Source = digest, source
	}
}

func negativeCorpusRows(inputs map[string]Input) []CorpusCase {
	resource := ResourceObservation{CPUCoreNS: 10, PeakMemoryBytes: 20, WorkUnits: 30}
	changed := []string{"urn:gooo:surface:billing/pay-order"}
	counts := func(extra ObservationCounts) ObservationCounts {
		extra.RegistryBindings, extra.ChangedCodeSurfaces, extra.ChangedRegistered, extra.ResourceReceipts = 2, 1, 1, 3
		return extra
	}
	unregisteredCounts := counts(ObservationCounts{ReceiptRecords: 1, PathEdges: 4, PathClaims: 1, PathEvidence: 4})
	unregisteredCounts.ChangedRegistered = 0
	return []CorpusCase{
		{Name: "negative-missing-receipt", Input: inputs["missing-receipt"], Expected: FixtureExpectation{Decision: DecisionUnknown, Reason: ReasonMissingReceipt, AcceptedSurfaces: []string{}, ChangedSurfaces: changed, ReceiptSurfaces: []string{}, ObservationCounts: counts(ObservationCounts{}), Resources: resource}},
		{Name: "negative-duplicate-receipt", Input: inputs["duplicate-receipt"], Expected: FixtureExpectation{Decision: DecisionFailClosed, Reason: ReasonDuplicateReceipt, AcceptedSurfaces: []string{}, ChangedSurfaces: changed, ReceiptSurfaces: changed, ObservationCounts: counts(ObservationCounts{ReceiptRecords: 2, ValidReceipts: 1, PathEdges: 4, PathClaims: 1, PathEvidence: 4}), Resources: resource}},
		{Name: "negative-orphan-receipt", Input: inputs["orphan-receipt"], Expected: FixtureExpectation{Decision: DecisionFailClosed, Reason: ReasonOrphanReceipt, AcceptedSurfaces: []string{}, ChangedSurfaces: changed, ReceiptSurfaces: []string{}, ObservationCounts: counts(ObservationCounts{ReceiptRecords: 1, PathEdges: 4, PathClaims: 1, PathEvidence: 4}), Resources: resource}},
		{Name: "negative-stale-receipt", Input: inputs["stale-receipt"], Expected: FixtureExpectation{Decision: DecisionUnknown, Reason: ReasonStaleReceipt, AcceptedSurfaces: []string{}, ChangedSurfaces: changed, ReceiptSurfaces: []string{}, ObservationCounts: counts(ObservationCounts{ReceiptRecords: 1, PathEdges: 4, PathClaims: 1, PathEvidence: 4}), Resources: resource}},
		{Name: "negative-unregistered-surface", Input: inputs["unregistered-surface"], Expected: FixtureExpectation{Decision: DecisionFailClosed, Reason: ReasonSurfaceUnregistered, AcceptedSurfaces: []string{}, ChangedSurfaces: []string{}, ReceiptSurfaces: []string{}, ObservationCounts: unregisteredCounts, Resources: resource}},
		{Name: "negative-delta-without-source", Input: inputs["delta-without-source"], Expected: FixtureExpectation{Decision: DecisionFailClosed, Reason: ReasonDeltaWithoutSource, AcceptedSurfaces: []string{}, ChangedSurfaces: changed, ReceiptSurfaces: []string{}, ObservationCounts: counts(ObservationCounts{ReceiptRecords: 1, PathEdges: 4, PathClaims: 1, PathEvidence: 4, AddedSemanticFacts: 2}), Resources: resource}},
		{Name: "negative-no-delta-without-equality", Input: inputs["no-delta-without-equality"], Expected: FixtureExpectation{Decision: DecisionFailClosed, Reason: ReasonNoDeltaWithoutEquality, AcceptedSurfaces: []string{}, ChangedSurfaces: changed, ReceiptSurfaces: []string{}, ObservationCounts: counts(ObservationCounts{ReceiptRecords: 1, PathEdges: 4, PathClaims: 1, PathEvidence: 4, AddedSemanticFacts: 2}), Resources: resource}},
		{Name: "negative-missing-root", Input: inputs["missing-root"], Expected: FixtureExpectation{Decision: DecisionFailClosed, Reason: ReasonPathMissing, AcceptedSurfaces: []string{}, ChangedSurfaces: changed, ReceiptSurfaces: changed, ObservationCounts: counts(ObservationCounts{ReceiptRecords: 1, ValidReceipts: 1, PathEdges: 4, PathClaims: 1, PathEvidence: 4}), Resources: resource}},
		{Name: "negative-unproven-zero-change", Input: inputs["unproven-zero-change"], Expected: FixtureExpectation{Decision: DecisionUnknown, Reason: ReasonRequiredInputMissing, AcceptedSurfaces: []string{}, ChangedSurfaces: []string{}, ReceiptSurfaces: []string{}, ObservationCounts: ObservationCounts{RegistryBindings: 2}, Resources: ResourceObservation{}}},
	}
}
