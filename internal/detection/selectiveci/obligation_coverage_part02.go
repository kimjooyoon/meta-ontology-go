package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
)

func ObserveCoverage(input ObligationCoverageInput) ObligationCoverageResult {
	return ObserveObligationCoverage(input)
}
func coverageResultFor(input ObligationCoverageInput) ObligationCoverageResult {
	result := ObligationCoverageResult{
		SchemaVersion:         ObligationCoverageSchemaVersion,
		UncoveredRootIDs:      []string{},
		RequiredObligationIDs: []string{},

		SnapshotDigest: input.SnapshotDigest,
		RegistryDigest: input.Registry.Digest,
		GraphDigest:    input.Graph.Digest()}
	return result
}
func coverageInputReason(input ObligationCoverageInput) CoverageReason {
	if input.SchemaVersion != ObligationCoverageSchemaVersion {
		return CoverageReasonUnsupportedSchema
	}
	if input.ChangedRootIDs == nil {
		return CoverageReasonMissingInput
	}
	if !validDigest(input.SnapshotDigest) {
		return CoverageReasonInvalidSnapshot
	}
	if input.Registry.SchemaVersion != RegistrySchemaVersion {
		return CoverageReasonInvalidRegistry
	}
	if input.Graph.Version != impactgraph.SchemaVersion {
		return CoverageReasonInvalidGraph
	}
	return CoverageReasonInvalidInput
}
func validateCoverageRegistry(registry Registry) CoverageReason {
	if err := registry.Validate(); err != nil {
		reason := reasonFor(err)
		switch reason {
		case ReasonUnsupportedSchema:
			return CoverageReasonUnsupportedSchema
		case ReasonMismatchedDigest:
			return CoverageReasonStaleRegistry
		case ReasonMissingBinding, ReasonMissingCommand:
			return CoverageReasonMissingCommand
		case ReasonDanglingReference:
			return CoverageReasonDanglingCommand
		default:
			return CoverageReasonInvalidRegistry
		}
	}
	return ""
}
