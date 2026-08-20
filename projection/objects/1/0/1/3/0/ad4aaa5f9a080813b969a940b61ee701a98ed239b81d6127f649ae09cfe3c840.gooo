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
