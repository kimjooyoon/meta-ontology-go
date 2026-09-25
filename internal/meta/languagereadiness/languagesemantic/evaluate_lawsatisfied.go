package languagesemantic

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesemantic/replay"
)

func lawSatisfied(law string, observation replay.LawObservation) bool {
	switch law {
	case "PRESENTATION_INVARIANCE":
		return observation.PresentationChanged && observation.PresentationInvariant
	case "CANDIDATE_NON_AUTHORITY":
		return observation.CandidateRecorded && observation.CandidateCanonicalChanged && observation.CandidateNonAuthoritative
	case "DETERMINISTIC_AUTHORITY":
		return observation.DeterministicRecorded && observation.DeterministicCanonicalChanged && observation.DeterministicAuthoritative
	default:
		return false
	}
}
