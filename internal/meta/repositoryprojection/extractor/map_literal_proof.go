package extractor

import (
	"encoding/json"
	"go/ast"
	"go/token"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

func mapLiteralStrategyEvidence(root, logical string, source []byte, fset *token.FileSet, file *ast.File, function *ast.FuncDecl, candidate *suffixCandidate, preflight []renderedCapacityObservation, previous Failure) (*StrategyEvidence, error) {
	item, err := suffixStrategyEvidence(root, logical, source, fset, file, function, candidate, preflight)
	if err != nil {
		return nil, err
	}
	contract, err := generation.ExtractFunctionInputContractEvidence()
	if err != nil {
		return nil, err
	}
	obligations, ok := returnTailContractObligations(contract.Obligations)
	if !ok {
		return nil, fail("derive-recipe", "admit-map-construction", "RETURN_TAIL_OBLIGATIONS_UNPROVEN", "DIRECT_MISSING", "restore-operation-input-contract", nil)
	}
	previous.BlockedBy = append([]string{}, previous.BlockedBy...)
	retained, err := json.Marshal(previous)
	if err != nil {
		return nil, err
	}
	chain := newReturnTailProofChain(obligations, source, candidate.result, contract.SourceDigest, contract.SemanticDigest)
	details := []string{
		"helper returns the same unnamed map[string]any type; original function signature and returns are unchanged",
		"one direct short-declaration literal is replaced; no statements, returns, defer, or go operations move",
		"every value expression remains in caller scope as one positional any argument; unique constant string keys are retained",
		"constructor contains only a fresh map and parameter reads; original calls and lexical evaluation order remain in caller; allocation timing and performance are not asserted; previous_return_tail_failure=" + string(retained),
	}
	for index, detail := range details {
		payload := proofCanonical(mapLiteralStrategy, string(source[candidate.start:candidate.end]), string(candidate.helper), detail)
		if err := chain.consume(index, returnTailPredicateResult{Status: "PASS", Payload: payload, Detail: detail}); err != nil {
			return nil, err
		}
	}
	item.Strategy = mapLiteralStrategy
	item.ContractObligations = obligations
	item.ProofStages = chain.stages
	item.Obligations = obligationsFromProofStages(chain.stages)
	return item, nil
}
