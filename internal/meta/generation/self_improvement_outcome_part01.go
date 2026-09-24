package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/provenance"

type SelfImprovementOutcomeReceipt = provenance.SelfImprovementOutcomeReceipt
type SelfImprovementOutcomeStatus = provenance.SelfImprovementOutcomeStatus

const (
	SelfImprovementOutcomeSchema  = provenance.SelfImprovementOutcomeSchema
	SelfImprovementOutcomeClosed  = provenance.SelfImprovementOutcomeClosed
	SelfImprovementOutcomeUnknown = provenance.SelfImprovementOutcomeUnknown
	SelfImprovementOutcomeRefuted = provenance.SelfImprovementOutcomeRefuted
)

func ValidateSelfImprovementOutcome(receipt SelfImprovementOutcomeReceipt) error {
	return receipt.Validate()
}
