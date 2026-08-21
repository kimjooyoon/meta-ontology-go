package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func cloneFixture(fixture testFixture) testFixture {
	copy := fixture
	copy.input.ChangedRootIDs = append([]semantic.ID(nil), fixture.input.ChangedRootIDs...)
	copy.input.SelectedCommandIDs = append([]semantic.ID(nil), fixture.input.SelectedCommandIDs...)
	copy.input.ObligationIDs = append([]semantic.ID(nil), fixture.input.ObligationIDs...)
	copy.input.Paths = append([]Path(nil), fixture.input.Paths...)
	copy.input.CommandReceipts = append([]CommandReceipt(nil), fixture.input.CommandReceipts...)
	copy.input.EvidenceIDs = append([]semantic.ID(nil), fixture.input.EvidenceIDs...)
	copy.input.InferencePath.Edges = append([]semantic.InferenceEdge(nil), fixture.input.InferencePath.Edges...)
	copy.input.InferencePath.Claims = append([]semantic.SemanticChangeClaim(nil), fixture.input.InferencePath.Claims...)
	copy.input.InferencePath.Evidence = append([]semantic.InferenceEvidence(nil), fixture.input.InferencePath.Evidence...)
	for index := range copy.input.InferencePath.Edges {
		copy.input.InferencePath.Edges[index].Evidence = append([]semantic.EvidenceReference(nil), fixture.input.InferencePath.Edges[index].Evidence...)
		copy.input.InferencePath.Edges[index].SourceRoots = append([]semantic.ID(nil), fixture.input.InferencePath.Edges[index].SourceRoots...)
	}
	return copy
}
