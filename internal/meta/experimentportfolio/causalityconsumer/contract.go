package causalityconsumer

import "reflect"

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
	ExpectedCausalCases       = 3
	ExpectedCausalTransitions = 9
)

const (
	predecessorContractDigest            = "sha256:889c669db94229c8391446533dfffb51f7e53d165e5afdae8d3a5a6878751981"
	causalityTransitionDenominatorReason = "three operations x three intervention claims"
	receiptProducer                      = "portfolio-receipt-producer"
	receiptConsumer                      = "portfolio-adjudicator"
)

var requiredCausalityReceiptFields = []string{"semantic_value", "decision", "claim_transitions"}

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

func ExpectedCausalityManifest() CausalityManifest {
	return CausalityManifest{
		Schema:                      CausalityManifestSchema,
		ManifestID:                  "meta-ontology-source-semantic-causality-v1",
		Version:                     1,
		PredecessorContractID:       ExpectedContractV1().ID,
		PredecessorContractDigest:   predecessorContractDigest,
		CoordinateID:                "source-semantic-causality",
		CasesPerCandidate:           ExpectedCausalCases,
		TransitionDenominator:       ExpectedCausalTransitions,
		TransitionDenominatorReason: causalityTransitionDenominatorReason,
		RequiredReceiptFields:       append([]string(nil), requiredCausalityReceiptFields...),
		Cases: []CausalityCaseContract{
			{CandidateID: "derive", SourcePath: "examples/experiment-portfolio/alternatives/derive.gooo", OperationValueBefore: "meta.portfolio.derive-coordinate", OperationValueAfter: "meta.portfolio.derive-coordinate:semantic-intervention", NonSemanticComment: "source-semantic-causality non-semantic derive", RequiredChangeFields: append([]string(nil), requiredCausalityReceiptFields...)},
			{CandidateID: "replay", SourcePath: "examples/experiment-portfolio/alternatives/replay.gooo", OperationValueBefore: "meta.portfolio.replay-independent", OperationValueAfter: "meta.portfolio.replay-independent:semantic-intervention", NonSemanticComment: "source-semantic-causality non-semantic replay", RequiredChangeFields: append([]string(nil), requiredCausalityReceiptFields...)},
			{CandidateID: "reflect", SourcePath: "examples/experiment-portfolio/alternatives/reflect.gooo", OperationValueBefore: "meta.portfolio.reflect-counterexample", OperationValueAfter: "meta.portfolio.reflect-counterexample:semantic-intervention", NonSemanticComment: "source-semantic-causality non-semantic reflect", RequiredChangeFields: append([]string(nil), requiredCausalityReceiptFields...)},
		},
	}
}

func causalityManifestReason(contract Contract, manifest CausalityManifest) string {
	if manifest.Schema != contract.CausalityManifestSchema || manifest.Schema != CausalityManifestSchema ||
		manifest.ManifestID == "" || manifest.Version != 1 ||
		manifest.PredecessorContractID != contract.PredecessorContractID ||
		manifest.PredecessorContractDigest != contract.PredecessorContractDigest ||
		manifest.CoordinateID != contract.CausalityCoordinateID ||
		manifest.CasesPerCandidate != ExpectedCausalCases ||
		manifest.TransitionDenominator != contract.CausalityTransitionDenominator ||
		manifest.TransitionDenominatorReason != contract.CausalityTransitionDenominatorReason ||
		!reflect.DeepEqual(manifest.RequiredReceiptFields, requiredCausalityReceiptFields) {
		return "CAUSALITY_MANIFEST_IDENTITY_INVALID"
	}
	if len(manifest.Cases) != len(contract.Candidates) {
		return "CAUSALITY_MANIFEST_CASE_COUNT_INVALID"
	}
	seen := map[string]bool{}
	for _, candidateCase := range manifest.Cases {
		candidate, ok := candidateContract(contract, candidateCase.CandidateID)
		if !ok || seen[candidateCase.CandidateID] {
			return "CAUSALITY_MANIFEST_CANDIDATE_INVALID"
		}
		seen[candidateCase.CandidateID] = true
		if candidateCase.SourcePath != candidate.SourcePath || candidateCase.OperationValueBefore == "" ||
			candidateCase.OperationValueAfter == "" || candidateCase.OperationValueBefore == candidateCase.OperationValueAfter ||
			candidateCase.NonSemanticComment == "" || !validRequiredChangeFields(candidateCase.RequiredChangeFields) {
			return "CAUSALITY_MANIFEST_CASE_INVALID"
		}
	}
	for _, candidate := range contract.Candidates {
		if !seen[candidate.ID] {
			return "CAUSALITY_MANIFEST_CANDIDATE_MISSING"
		}
	}
	return ""
}

func validRequiredChangeFields(fields []string) bool {
	if len(fields) == 0 {
		return false
	}
	allowed := map[string]bool{}
	for _, field := range requiredCausalityReceiptFields {
		allowed[field] = true
	}
	seen := map[string]bool{}
	for _, field := range fields {
		if !allowed[field] || seen[field] {
			return false
		}
		seen[field] = true
	}
	return true
}

func causalityCaseContract(manifest CausalityManifest, candidateID string) (CausalityCaseContract, bool) {
	for _, candidateCase := range manifest.Cases {
		if candidateCase.CandidateID == candidateID {
			return candidateCase, true
		}
	}
	return CausalityCaseContract{}, false
}
