package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/analysisprovenance"

// ObservationIdentity binds the generated semantic observation to the exact
// lowered IR and analysis provenance. It does not add execution, promotion,
// or repository-write authority to semantic adoption.
func (provenance SemanticAdoptionProvenance) ObservationIdentity(semanticDigest string) analysisprovenance.ObservationIdentity {
	return analysisprovenance.NewObservationIdentity(
		provenance.SourceDigest,
		semanticDigest,
		provenance.ProfileDigest,
		provenance.ToolchainDigest,
		provenance.ContractDigest,
	)
}
