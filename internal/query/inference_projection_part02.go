package query

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func cloneInferencePath(path semantic.InferencePathV1) semantic.InferencePathV1 {
	out := semantic.InferencePathV1{Version: path.Version,
		Edges:    append([]semantic.InferenceEdge(nil), path.Edges...),
		Claims:   append([]semantic.SemanticChangeClaim(nil), path.Claims...),
		Evidence: append([]semantic.InferenceEvidence(nil), path.Evidence...)}
	for i := range out.Edges {
		out.Edges[i].SourceRoots = append([]semantic.ID(nil), path.Edges[i].SourceRoots...)
		out.Edges[i].Evidence = append([]semantic.EvidenceReference(nil), path.Edges[i].Evidence...)
	}
	for i := range out.Claims {
		out.Claims[i].Evidence = append([]semantic.EvidenceReference(nil), path.Claims[i].Evidence...)
	}
	return out
}
func (query InferenceQuery) hasSelectors() bool {
	return query.RecordID != "" || query.SubjectID != "" || query.ObjectID != "" ||
		query.EvidenceID != "" || query.Kind != "" || query.Phase != "" ||
		query.Layer != "" || query.Effect != "" || query.ClaimKind != "" ||
		query.hasSnapshotOrControlSelectors()
}
