package experimentportfolio

import (
	"reflect"
)

const (
	ContractSchema      = "gooo/meta-experiment-portfolio-contract/v1"
	ReceiptSchema       = "gooo/meta-experiment-portfolio-receipt/v1"
	ReportSchema        = "gooo/meta-experiment-portfolio-report/v1"
	ExpectedCandidates  = 3
	ExpectedCoordinates = 6
)

type CandidateContract struct {
	ID            string `json:"id"`
	SourcePath    string `json:"source_path"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
}

type Contract struct {
	Schema                 string              `json:"schema"`
	ID                     string              `json:"contract_id"`
	ReceiptSchema          string              `json:"receipt_schema"`
	ReportSchema           string              `json:"report_schema"`
	Candidates             []CandidateContract `json:"candidates"`
	CoordinateIDs          []string            `json:"coordinate_ids"`
	CoordinateDenominators map[string]int      `json:"coordinate_denominators"`
	CounterexampleSlots    int                 `json:"counterexample_slots"`
	UnknownLocationSlots   int                 `json:"unknown_location_slots"`
	NotClaimed             []string            `json:"not_claimed"`
}

func ExpectedContract() Contract {
	return Contract{
		Schema:        ContractSchema,
		ID:            "meta-ontology-experiment-portfolio-v1",
		ReceiptSchema: ReceiptSchema,
		ReportSchema:  ReportSchema,
		Candidates: []CandidateContract{
			{ID: "derive", SourcePath: "examples/experiment-portfolio/alternatives/derive.gooo", MetaOperation: "derive-coordinate-vector", ProofChoice: "source-digest"},
			{ID: "replay", SourcePath: "examples/experiment-portfolio/alternatives/replay.gooo", MetaOperation: "replay-independent-receipt", ProofChoice: "independent-receipt"},
			{ID: "reflect", SourcePath: "examples/experiment-portfolio/alternatives/reflect.gooo", MetaOperation: "reflect-counterexample-boundary", ProofChoice: "counterexample-replay"},
		},
		CoordinateIDs: []string{
			"source-replay",
			"receipt-independence",
			"counterexample-boundary",
			"unknown-localization",
			"extension-evidence",
			"read-only-effects",
		},
		CoordinateDenominators: map[string]int{
			"source-replay":           1,
			"receipt-independence":    1,
			"counterexample-boundary": 2,
			"unknown-localization":    2,
			"extension-evidence":      1,
			"read-only-effects":       1,
		},
		CounterexampleSlots:  2,
		UnknownLocationSlots: 2,
		NotClaimed: []string{
			"language-quality",
			"production-readiness",
			"semantic-equivalence",
			"winner-or-rank",
			"estimated-improvement",
			"weighted-average",
		},
	}
}

func contractReason(contract Contract) string {
	if !reflect.DeepEqual(contract, ExpectedContract()) {
		return "PORTFOLIO_CONTRACT_DRIFT"
	}
	return ""
}

func candidateContract(contract Contract, id string) (CandidateContract, bool) {
	for _, candidate := range contract.Candidates {
		if candidate.ID == id {
			return candidate, true
		}
	}
	return CandidateContract{}, false
}
