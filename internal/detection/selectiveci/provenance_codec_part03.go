package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func pathWireToSemantic(path pathWire) (semantic.InferencePathV1, error) {
	result := semantic.InferencePathV1{Version: path.Version}
	for _, edge := range path.Edges {
		converted, err := edgeWireToSemantic(edge)
		if err != nil {
			return semantic.InferencePathV1{}, err
		}
		result.Edges = append(result.Edges, converted)
	}
	for _, claim := range path.Claims {
		converted, err := claimWireToSemantic(claim)
		if err != nil {
			return semantic.InferencePathV1{}, err
		}
		result.Claims = append(result.Claims, converted)
	}
	for _, evidence := range path.Evidence {
		converted, err := evidenceWireToSemantic(evidence)
		if err != nil {
			return semantic.InferencePathV1{}, err
		}
		result.Evidence = append(result.Evidence, converted)
	}
	return result, nil
}
func edgeWireFromSemantic(edge semantic.InferenceEdge) (edgeWire, error) {
	record, err := recordWireFromSemantic(edge.InferenceRecord)
	if err != nil {
		return edgeWire{}, err
	}
	roots := make([]string, len(edge.SourceRoots))
	for i, root := range edge.SourceRoots {
		roots[i] = root.String()
	}
	return edgeWire{Record: record, Kind: string(edge.Kind), SourceRoots: roots, AcceptanceReceipt: edge.AcceptanceReceipt.String()}, nil
}
func edgeWireToSemantic(edge edgeWire) (semantic.InferenceEdge, error) {
	record, err := recordWireToSemantic(edge.Record)
	if err != nil {
		return semantic.InferenceEdge{}, err
	}
	roots := make([]semantic.ID, len(edge.SourceRoots))
	for i, root := range edge.SourceRoots {
		roots[i], err = semantic.ParseIdentity(root)
		if err != nil {
			return semantic.InferenceEdge{}, err
		}
	}
	var receipt semantic.ID
	if edge.AcceptanceReceipt != "" {
		receipt, err = semantic.ParseIdentity(edge.AcceptanceReceipt)
		if err != nil {
			return semantic.InferenceEdge{}, err
		}
	}
	return semantic.InferenceEdge{InferenceRecord: record, Kind: semantic.InferenceKind(edge.Kind), SourceRoots: roots, AcceptanceReceipt: receipt}, nil
}
func claimWireFromSemantic(claim semantic.SemanticChangeClaim) (claimWire, error) {
	record, err := recordWireFromSemantic(claim.InferenceRecord)
	if err != nil {
		return claimWire{}, err
	}
	return claimWire{Record: record, Kind: string(claim.Kind), CanonicalDelta: claim.CanonicalDelta, DeltaDigest: claim.DeltaDigest}, nil
}
