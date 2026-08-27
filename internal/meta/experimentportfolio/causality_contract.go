package experimentportfolio

import "reflect"

var requiredCausalityReceiptFields = []string{"semantic_value", "decision", "claim_transitions"}

func ExpectedCausalityManifest() CausalityManifest {
	return CausalityManifest{
		Schema:                    CausalityManifestSchema,
		ManifestID:                "meta-ontology-source-semantic-causality-v1",
		Version:                   1,
		PredecessorContractID:     ExpectedContractV1().ID,
		PredecessorContractDigest: predecessorContractDigest,
		CoordinateID:              "source-semantic-causality",
		CasesPerCandidate:         ExpectedCausalCases,
		RequiredReceiptFields:     append([]string(nil), requiredCausalityReceiptFields...),
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
			candidateCase.NonSemanticComment == "" ||
			!reflect.DeepEqual(candidateCase.RequiredChangeFields, requiredCausalityReceiptFields) {
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

func causalityCaseContract(manifest CausalityManifest, candidateID string) (CausalityCaseContract, bool) {
	for _, candidateCase := range manifest.Cases {
		if candidateCase.CandidateID == candidateID {
			return candidateCase, true
		}
	}
	return CausalityCaseContract{}, false
}
