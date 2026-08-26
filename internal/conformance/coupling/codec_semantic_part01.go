package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

func cloneSemanticIR(input SemanticIR) SemanticIR {
	input.Nodes = append([]SemanticNode(nil), input.Nodes...)
	input.Relations = append([]SemanticRelation(nil), input.Relations...)
	for i := range input.Nodes {
		input.Nodes[i].Aliases = append([]string(nil), input.Nodes[i].Aliases...)
	}
	return input
}
func normalizeSemanticIR(input *SemanticIR) {
	sort.Slice(input.Nodes, func(i, j int) bool { return input.Nodes[i].ID < input.Nodes[j].ID })
	sort.Slice(input.Relations, func(i, j int) bool {
		left := input.Relations[i].Subject + "\x00" + input.Relations[i].Predicate + "\x00" + input.Relations[i].Object
		right := input.Relations[j].Subject + "\x00" + input.Relations[j].Predicate + "\x00" + input.Relations[j].Object
		return left < right
	})
	for i := range input.Nodes {
		sort.Strings(input.Nodes[i].Aliases)
	}
}
func pathToWire(path semantic.InferencePathV1) wirePath {
	out := wirePath{Version: path.Version, Edges: make([]wireEdge, 0, len(path.Edges)), Claims: make([]wireClaim, 0, len(path.Claims)), Evidence: make([]wireEvidence, 0, len(path.Evidence))}
	for _, edge := range path.Edges {
		out.Edges = append(out.Edges, wireEdgeFromSemantic(edge))
	}
	for _, claim := range path.Claims {
		out.Claims = append(out.Claims, wireClaimFromSemantic(claim))
	}
	for _, evidence := range path.Evidence {
		out.Evidence = append(out.Evidence, wireEvidenceFromSemantic(evidence))
	}
	return out
}
func pathFromWire(raw wirePath) (semantic.InferencePathV1, error) {
	out := semantic.InferencePathV1{Version: raw.Version, Edges: make([]semantic.InferenceEdge, 0, len(raw.Edges)), Claims: make([]semantic.SemanticChangeClaim, 0, len(raw.Claims)), Evidence: make([]semantic.InferenceEvidence, 0, len(raw.Evidence))}
	for _, edge := range raw.Edges {
		value, err := semanticEdgeFromWire(edge)
		if err != nil {
			return semantic.InferencePathV1{}, err
		}
		out.Edges = append(out.Edges, value)
	}
	for _, claim := range raw.Claims {
		value, err := semanticClaimFromWire(claim)
		if err != nil {
			return semantic.InferencePathV1{}, err
		}
		out.Claims = append(out.Claims, value)
	}
	for _, evidence := range raw.Evidence {
		value, err := semanticEvidenceFromWire(evidence)
		if err != nil {
			return semantic.InferencePathV1{}, err
		}
		out.Evidence = append(out.Evidence, value)
	}
	return out, nil
}
func wireEdgeFromSemantic(edge semantic.InferenceEdge) wireEdge {
	record := wireRecordFromSemantic(edge.InferenceRecord)
	return wireEdge{RecordID: record.RecordID, SubjectID: record.SubjectID, ObjectID: record.ObjectID, Rule: record.Rule, Phase: record.Phase, Ordinal: record.Ordinal, Before: record.Before, After: record.After, Authority: record.Authority, Evidence: record.Evidence, Controls: record.Controls, Kind: edge.Kind.String(), SourceRoots: idsToStrings(edge.SourceRoots), AcceptanceReceipt: edge.AcceptanceReceipt.String()}
}
func wireClaimFromSemantic(claim semantic.SemanticChangeClaim) wireClaim {
	record := wireRecordFromSemantic(claim.InferenceRecord)
	return wireClaim{RecordID: record.RecordID, SubjectID: record.SubjectID, ObjectID: record.ObjectID, Rule: record.Rule, Phase: record.Phase, Ordinal: record.Ordinal, Before: record.Before, After: record.After, Authority: record.Authority, Evidence: record.Evidence, Controls: record.Controls, Kind: claim.Kind.String(), CanonicalDelta: claim.CanonicalDelta, DeltaDigest: claim.DeltaDigest}
}
