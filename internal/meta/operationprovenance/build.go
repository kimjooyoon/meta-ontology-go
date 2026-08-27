package operationprovenance

import "fmt"

func Build(source []byte, observation WorkspaceObservation) (Receipt, error) {
	ir, err := lowerSource(source)
	if err != nil {
		return Receipt{}, err
	}
	metrics, scenarios, reconstruction, err := reconstructSemanticData(ir)
	if err != nil {
		return Receipt{}, err
	}
	receipt := Receipt{
		Schema: ReceiptSchema, Toolchain: Toolchain,
		SourceDigest:            digestBytes(source),
		CanonicalSemanticDigest: "sha256:" + ir.StableHash(),
		SourceReconstruction:    reconstruction, WorkspaceObservation: observation,
		Scenarios: make([]ScenarioResult, 0, len(scenarios)),
	}
	for _, scenario := range scenarios {
		result, err := evaluateScenario(metrics, scenario, receipt.SourceDigest, receipt.CanonicalSemanticDigest)
		if err != nil {
			return Receipt{}, err
		}
		receipt.Scenarios = append(receipt.Scenarios, result)
	}
	return sealReceipt(receipt)
}

// BuildObserved binds the receipt to isolated repository before/after status.
func BuildObserved(source []byte, repositoryRoot string) (Receipt, error) {
	before, err := readRepositorySnapshot(repositoryRoot)
	if err != nil {
		return Receipt{}, fmt.Errorf("observe repository before producer: %w", err)
	}
	receipt, buildErr := Build(source, WorkspaceObservation{})
	after, afterErr := readRepositorySnapshot(repositoryRoot)
	if buildErr != nil {
		return Receipt{}, buildErr
	}
	if afterErr != nil {
		return Receipt{}, fmt.Errorf("observe repository after producer: %w", afterErr)
	}
	receipt.WorkspaceObservation = deriveObservation(before, after)
	return sealReceipt(receipt)
}
