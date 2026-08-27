package experimentportfolio

import (
	"reflect"
)

const (
	ContractSchemaV1 = "gooo/meta-experiment-portfolio-contract/v1"
	ReceiptSchemaV1  = "gooo/meta-experiment-portfolio-receipt/v1"
	ReportSchemaV1   = "gooo/meta-experiment-portfolio-report/v1"

	ContractSchema            = "gooo/meta-experiment-portfolio-contract/v2"
	ReceiptSchema             = "gooo/meta-experiment-portfolio-receipt/v2"
	ReportSchema              = "gooo/meta-experiment-portfolio-report/v2"
	CausalityManifestSchema   = "gooo/meta-source-semantic-causality-manifest/v1"
	CausalityReportSchema     = "gooo/meta-source-semantic-causality-report/v1"
	ExpectedCandidates        = 3
	ExpectedCoordinatesV1     = 6
	ExpectedCoordinates       = 7
	ExpectedCausalCases       = 3
	ExpectedCausalTransitions = 9
)

const predecessorContractDigest = "sha256:889c669db94229c8391446533dfffb51f7e53d165e5afdae8d3a5a6878751981"
const causalityTransitionDenominatorReason = "three operations x three intervention claims"

type CandidateContract struct {
	ID            string `json:"id"`
	SourcePath    string `json:"source_path"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
}

type Contract struct {
	Schema                               string              `json:"schema"`
	ID                                   string              `json:"contract_id"`
	ReceiptSchema                        string              `json:"receipt_schema"`
	ReportSchema                         string              `json:"report_schema"`
	Version                              int                 `json:"version"`
	PredecessorContractID                string              `json:"predecessor_contract_id"`
	PredecessorContractPath              string              `json:"predecessor_contract_path"`
	PredecessorContractDigest            string              `json:"predecessor_contract_digest"`
	UpgradeReason                        string              `json:"upgrade_reason"`
	CausalityManifestSchema              string              `json:"causality_manifest_schema"`
	CausalityManifestPath                string              `json:"causality_manifest_path"`
	CausalityCoordinateID                string              `json:"causality_coordinate_id"`
	CausalityDenominator                 int                 `json:"causality_denominator"`
	CausalityTransitionDenominator       int                 `json:"causality_transition_denominator"`
	CausalityTransitionDenominatorReason string              `json:"causality_transition_denominator_reason"`
	Candidates                           []CandidateContract `json:"candidates"`
	CoordinateIDs                        []string            `json:"coordinate_ids"`
	CoordinateDenominators               map[string]int      `json:"coordinate_denominators"`
	CounterexampleSlots                  int                 `json:"counterexample_slots"`
	UnknownLocationSlots                 int                 `json:"unknown_location_slots"`
	NotClaimed                           []string            `json:"not_claimed"`
}

func ExpectedContractV1() Contract {
	return Contract{
		Schema:        ContractSchemaV1,
		ID:            "meta-ontology-experiment-portfolio-v1",
		ReceiptSchema: ReceiptSchemaV1,
		ReportSchema:  ReportSchemaV1,
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

func ExpectedContract() Contract {
	contract := ExpectedContractV1()
	contract.Schema = ContractSchema
	contract.ID = "meta-ontology-experiment-portfolio-v2"
	contract.ReceiptSchema = ReceiptSchema
	contract.ReportSchema = ReportSchema
	contract.Version = 2
	contract.PredecessorContractID = ExpectedContractV1().ID
	contract.PredecessorContractPath = "examples/experiment-portfolio/contract-v1.json"
	contract.PredecessorContractDigest = predecessorContractDigest
	contract.UpgradeReason = "add source-semantic-causality without changing the six v1 denominators"
	contract.CausalityManifestSchema = CausalityManifestSchema
	contract.CausalityManifestPath = "examples/experiment-portfolio/causality-manifest.json"
	contract.CausalityCoordinateID = "source-semantic-causality"
	contract.CausalityDenominator = ExpectedCausalCases
	contract.CausalityTransitionDenominator = ExpectedCausalTransitions
	contract.CausalityTransitionDenominatorReason = causalityTransitionDenominatorReason
	contract.CoordinateIDs = append(contract.CoordinateIDs, contract.CausalityCoordinateID)
	contract.CoordinateDenominators[contract.CausalityCoordinateID] = contract.CausalityDenominator
	return contract
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
