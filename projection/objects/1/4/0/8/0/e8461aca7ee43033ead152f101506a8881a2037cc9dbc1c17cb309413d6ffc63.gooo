package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func requirements(input Input) []pathclosure.Requirement {
	requirements := make([]pathclosure.Requirement, 0, len(input.Paths))
	for _, path := range input.Paths {
		requirements = append(requirements, pathclosure.Requirement{
			PathID: path.PathID, RecordIDs: path.RecordIDs, ExpectedKinds: path.ExpectedKinds,
			StartID: path.RootID, EndID: path.ReceiptID,
		})
	}
	return requirements
}
func makeReceipt(input Input, status DecisionStatus, fallback FallbackMode, code string) Receipt {
	paths := make([]semantic.ID, 0, len(input.Paths))
	for _, path := range input.Paths {
		paths = append(paths, path.PathID)
	}
	receipt := Receipt{
		Schema: SchemaVersion, Status: status, Fallback: fallback, Code: code,
		Snapshots: input.Snapshots, RegistryDigest: input.RegistryDigest, PlanDigest: input.PlanDigest,
		SelectedCommandIDs: append([]semantic.ID(nil), input.SelectedCommandIDs...),
		ObligationIDs:      append([]semantic.ID(nil), input.ObligationIDs...), PathIDs: paths,
		RequiredCommandCount: len(input.SelectedCommandIDs), RequiredObligationCount: len(input.ObligationIDs),
	}
	if status == Verified {
		receipt.VerifiedCommandCount = len(input.SelectedCommandIDs)
		receipt.VerifiedObligationCount = len(input.ObligationIDs)
		receipt.VerifiedPathCount = len(input.Paths)
		receipt.VerifiedCommandIDs = append([]semantic.ID(nil), input.SelectedCommandIDs...)
		receipt.VerifiedObligationIDs = append([]semantic.ID(nil), input.ObligationIDs...)
		receipt.VerifiedPathIDs = append([]semantic.ID(nil), paths...)
	}
	receipt.Digest = receipt.expectedDigest()
	return receipt
}
