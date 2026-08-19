package coupling

import (
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func literalInputMutationCases() []productionInputMutationCase {
	result := literalReceiptMutationCases()
	result = append(result, literalPathMutationCases()...)
	return append(result, literalAuthorityMutationCases()...)
}
func literalReceiptMutationCases() []productionInputMutationCase {
	return []productionInputMutationCase{
		{name: "input-stale-receipt", mutate: func(input *production.Input) { input.Receipts[0].SnapshotDigest = bridgeHash("stale-receipt") }},
		{name: "input-missing-receipt", mutate: func(input *production.Input) { input.Receipts = nil }},
		{name: "input-arbitrary-provider", mutate: func(input *production.Input) { input.ExternalReceipt.ProviderDigest = bridgeHash("arbitrary-provider") }},
		{name: "input-arbitrary-observer", mutate: func(input *production.Input) { input.ExternalReceipt.ObserverDigest = bridgeHash("arbitrary-observer") }},
		{name: "input-wrong-path-endpoint", mutate: mutateWrongPathEndpoint},
		{name: "input-omitted-evidence", mutate: func(input *production.Input) { input.Receipts[0].EvidenceRefs = nil }},
		{name: "input-extra-unrelated-evidence", mutate: mutateExtraEvidence},
		{name: "input-duplicate-evidence", mutate: func(input *production.Input) {
			input.Receipts[0].EvidenceRefs = append(input.Receipts[0].EvidenceRefs, input.Receipts[0].EvidenceRefs[0])
		}},
		{name: "input-reordered-path-id-presentation", mutate: func(input *production.Input) { reverseProductionIDs(input.Receipts[0].OriginPathIDs) }},
	}
}
func mutateWrongPathEndpoint(input *production.Input) {
	last := len(input.InferencePath.Edges) - 1
	input.InferencePath.Edges[last].ObjectID = bridgeID("urn:gooo:evidence:wrong-endpoint")
	input.InferencePath.Edges[last].InferenceRecord.ObjectID = input.InferencePath.Edges[last].ObjectID
}
func mutateExtraEvidence(input *production.Input) {
	input.Receipts[0].EvidenceRefs = append(input.Receipts[0].EvidenceRefs, semantic.EvidenceReference{ID: bridgeID("urn:gooo:evidence:unrelated"), Digest: bridgeHash("unrelated-evidence")})
}
func literalPathMutationCases() []productionInputMutationCase {
	return []productionInputMutationCase{
		{name: "input-disconnected-selected-edge", mutate: mutateDisconnectedPath},
		{name: "input-forked-selected-path", mutate: mutateForkedPath},
		{name: "input-cyclic-selected-path", mutate: mutateCyclicPath},
		{name: "input-wrong-start-end", mutate: mutateWrongPathStartEnd},
	}
}
func mutateDisconnectedPath(input *production.Input) {
	edge := input.InferencePath.Edges[1]
	edge.RecordID = bridgeID("urn:gooo:path:disconnected")
	edge.SubjectID, edge.ObjectID = bridgeID("urn:gooo:term:disconnected"), bridgeID("urn:gooo:code:disconnected")
	edge.InferenceRecord.RecordID, edge.InferenceRecord.SubjectID, edge.InferenceRecord.ObjectID = edge.RecordID, edge.SubjectID, edge.ObjectID
	input.InferencePath.Edges = append(input.InferencePath.Edges, edge)
	input.Receipts[0].OriginPathIDs = append(input.Receipts[0].OriginPathIDs, edge.RecordID)
}
func mutateForkedPath(input *production.Input) {
	edge := input.InferencePath.Edges[1]
	edge.RecordID, edge.ObjectID = bridgeID("urn:gooo:path:fork"), bridgeID("urn:gooo:surface:fork")
	edge.InferenceRecord.RecordID, edge.InferenceRecord.ObjectID = edge.RecordID, edge.ObjectID
	input.InferencePath.Edges = append(input.InferencePath.Edges, edge)
	input.Receipts[0].OriginPathIDs = append(input.Receipts[0].OriginPathIDs, edge.RecordID)
}
func mutateCyclicPath(input *production.Input) {
	edge := input.InferencePath.Edges[1]
	edge.RecordID, edge.SubjectID, edge.ObjectID = bridgeID("urn:gooo:path:cycle"), bridgeID("urn:gooo:evidence:verification"), bridgeID("urn:gooo:code:billing/pay-order")
	edge.InferenceRecord.RecordID, edge.InferenceRecord.SubjectID, edge.InferenceRecord.ObjectID = edge.RecordID, edge.SubjectID, edge.ObjectID
	input.InferencePath.Edges = append(input.InferencePath.Edges, edge)
	input.Receipts[0].OriginPathIDs = append(input.Receipts[0].OriginPathIDs, edge.RecordID)
}
func mutateWrongPathStartEnd(input *production.Input) {
	input.InferencePath.Edges[0].SubjectID = bridgeID("urn:gooo:source:wrong")
	input.InferencePath.Edges[0].InferenceRecord.SubjectID = input.InferencePath.Edges[0].SubjectID
	last := len(input.InferencePath.Edges) - 1
	input.InferencePath.Edges[last].SubjectID = bridgeID("urn:gooo:surface:wrong")
	input.InferencePath.Edges[last].InferenceRecord.SubjectID = input.InferencePath.Edges[last].SubjectID
}
