package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func baseObservation(changed, receipts int) ObservationVector {
	return ObservationVector{
		ChangedSurfaces: knownDimension(uint64(changed)),
		Receipts:        knownDimension(uint64(receipts)),
	}
}
func knownDimension(value uint64) CountDimension { return CountDimension{Known: true, Value: value} }
func indexReceipts(receipts []CouplingReceipt, changed []ManifestEntry) (map[semantic.ID]CouplingReceipt, *evaluationIssue) {
	changedIDs := make(map[semantic.ID]struct{}, len(changed))
	for _, entry := range changed {
		changedIDs[entry.SurfaceID] = struct{}{}
	}
	indexed := make(map[semantic.ID]CouplingReceipt, len(receipts))
	seenReceiptIDs := make(map[semantic.ID]struct{}, len(receipts))
	for _, receipt := range receipts {
		if _, duplicate := seenReceiptIDs[receipt.ReceiptID]; duplicate {
			return nil, failIssue(ReasonDuplicateReceipt, receipt.SurfaceID.String())
		}
		seenReceiptIDs[receipt.ReceiptID] = struct{}{}
		if _, expected := changedIDs[receipt.SurfaceID]; !expected {
			return nil, failIssue(ReasonOrphanReceipt, receipt.SurfaceID.String())
		}
		if _, duplicate := indexed[receipt.SurfaceID]; duplicate {
			return nil, failIssue(ReasonDuplicateReceipt, receipt.SurfaceID.String())
		}
		indexed[receipt.SurfaceID] = receipt
	}
	return indexed, nil
}
func validateReceipt(receipt CouplingReceipt, entry ManifestEntry, config Config, surface Surface) *evaluationIssue {
	if issue := validateReceiptIdentity(receipt, entry, config, surface); issue != nil {
		return issue
	}
	if issue := validateReceiptClaim(receipt); issue != nil {
		return issue
	}
	return validateReceiptReferences(receipt)
}
