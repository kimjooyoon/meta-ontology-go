package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/provenance"

const (
	SelfImprovementProvenanceChainBoundPart01          = provenance.SelfImprovementProvenanceChainBoundPart01
	SelfImprovementProvenanceChainUnknownPart01        = provenance.SelfImprovementProvenanceChainUnknownPart01
	SelfImprovementProvenanceChainCompleteReasonPart01 = provenance.SelfImprovementProvenanceChainCompleteReasonPart01
	SelfImprovementProvenanceChainNonAuthorizingPart01 = provenance.SelfImprovementProvenanceChainNonAuthorizingPart01
)

type SelfImprovementProvenanceStagePart01 = provenance.SelfImprovementProvenanceStagePart01
type SelfImprovementProvenanceChainPart01 = provenance.SelfImprovementProvenanceChainPart01

func BuildSelfImprovementProvenanceChainPart01(
	declarationDigest,
	sourceDigest,
	semanticIRDigest,
	graphDigest,
	generatedDigest,
	reverseObservationDigest string,
) SelfImprovementProvenanceChainPart01 {
	return provenance.BuildSelfImprovementProvenanceChainPart01(
		declarationDigest,
		sourceDigest,
		semanticIRDigest,
		graphDigest,
		generatedDigest,
		reverseObservationDigest,
	)
}
