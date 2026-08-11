package performance

import (
	"strings"
	"testing"
)

func TestVerifiedHostsAreComparable(t *testing.T) {
	contract := parserContract()
	goEvidence := verifiedEvidence(GoHosted, 10, 2)
	goooEvidence := verifiedEvidence(GoooHosted, 12, 3)
	comparison, err := Compare(contract, goEvidence, goooEvidence)
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if !comparison.Comparable {
		t.Fatalf("comparison is not comparable: %s", comparison.Reason)
	}
	if comparison.OperationDelta != 2 || comparison.AllocationDelta != 1 {
		t.Fatalf("comparison deltas = (%v, %v), want (2, 1)", comparison.OperationDelta, comparison.AllocationDelta)
	}
}

func TestStandardContractsCoverAllCompilerStages(t *testing.T) {
	contracts := StandardContracts()
	if len(contracts) != 5 {
		t.Fatalf("StandardContracts() length = %d, want 5", len(contracts))
	}
	for i, contract := range contracts {
		if err := contract.Validate(); err != nil {
			t.Fatalf("contract %d invalid: %v", i, err)
		}
	}
}

func TestPlannedFutureHostIsNotSuccess(t *testing.T) {
	contract := parserContract()
	goEvidence := verifiedEvidence(GoHosted, 10, 2)
	goooEvidence := Evidence{
		ContractID: contract.ID,
		Host:       GoooHosted,
		Status:     StatusPlanned,
		Source:     "self-hosting roadmap",
	}
	comparison, err := Compare(contract, goEvidence, goooEvidence)
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if comparison.Comparable {
		t.Fatal("planned gooo-hosted evidence was marked comparable")
	}
	if !strings.Contains(comparison.Reason, "planned") {
		t.Fatalf("comparison reason = %q, want planned state", comparison.Reason)
	}
}

func TestPlannedEvidenceCannotCarryMetrics(t *testing.T) {
	contract := parserContract()
	evidence := Evidence{
		ContractID:             contract.ID,
		Host:                   GoooHosted,
		Status:                 StatusPlanned,
		OperationsPerIteration: 1,
		Source:                 "self-hosting roadmap",
	}
	if err := evidence.Validate(contract); err == nil {
		t.Fatal("planned evidence with metrics was accepted")
	}
}

func TestEvidenceRequiresStableContractAndSource(t *testing.T) {
	contract := parserContract()
	missingSource := verifiedEvidence(GoHosted, 10, 2)
	missingSource.Source = ""
	if err := missingSource.Validate(contract); err == nil {
		t.Fatal("evidence without source was accepted")
	}
	wrongContract := verifiedEvidence(GoHosted, 10, 2)
	wrongContract.ContractID = "performance://stage/semantic"
	if err := wrongContract.Validate(contract); err == nil {
		t.Fatal("evidence with a different contract ID was accepted")
	}
}

func parserContract() Contract {
	return Contract{
		ID:             "performance://stage/parser",
		Stage:          "parser",
		OperationUnit:  OperationsPerIterationUnit,
		AllocationUnit: AllocationsPerIterationUnit,
	}
}

func verifiedEvidence(host Host, operations, allocations uint64) Evidence {
	return Evidence{
		ContractID:              parserContract().ID,
		Host:                    host,
		Status:                  StatusVerified,
		OperationsPerIteration:  operations,
		AllocationsPerIteration: allocations,
		Source:                  "benchmark fixture",
	}
}
