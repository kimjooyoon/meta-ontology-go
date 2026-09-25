package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/resourceenvelope"
)

func validateSelectedEvidence(input Input, selected []selectedPath) ([]string, []string, error) {
	receiptDigests := make([]string, 0, len(selected))
	pathIDs := make([]string, 0, len(selected))
	receipts := map[string]Receipt{}
	for _, receipt := range input.Receipts {
		receipts[receipt.CommandID] = receipt
	}
	paths := map[string]ProvenancePath{}
	for _, path := range input.ProvenancePaths {
		paths[path.CommandID] = path
	}
	for _, entry := range selected {
		receipt, ok := receipts[entry.command.ID]
		if !ok {
			return nil, nil, failure(ReasonResourceReceipt, "selected command has no resource receipt")
		}
		if receipt.SnapshotDigest != input.Head.Digest {
			return nil, nil, failure(ReasonMismatchedDigest, "resource receipt snapshot does not match head")
		}
		resource := resourceenvelope.Evaluate(receipt.Envelope)
		if resource.Status != resourceenvelope.PASS {
			return nil, nil, resourceFailure(resource.ReasonCode)
		}
		if resource.CPUCoreNS > entry.command.CPUWorkUnits || resource.PeakRSSBytes > entry.command.MemoryBytes {
			return nil, nil, failure(ReasonResourceLimit, "resource receipt exceeds command ceiling")
		}
		path, ok := paths[entry.command.ID]
		if !ok {
			return nil, nil, failure(ReasonAmbiguousPath, "selected command has no provenance path")
		}
		if err := evaluatePath(path); err != nil {
			return nil, nil, err
		}
		receiptDigests = append(receiptDigests, digestBytes([]byte(entry.command.ID+"\x00"+receipt.SnapshotDigest+"\x00"+resource.Canonical())))
		pathIDs = append(pathIDs, path.Requirement.PathID)
	}
	return sortedUnique(receiptDigests), sortedUnique(pathIDs), nil
}
func resourceFailure(reason string) error {
	if reason == "cpu-arithmetic" {
		return failure(ReasonResourceArithmetic, reason)
	}
	return failure(ReasonResourceReceipt, reason)
}
