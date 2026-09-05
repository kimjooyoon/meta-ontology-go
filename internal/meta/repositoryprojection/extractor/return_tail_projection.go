package extractor

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"sort"
)

func finalizeReturnTailEvidence(root, logical string, generated map[string][]byte, evidence []StrategyEvidence) ([]StrategyEvidence, error) {
	if len(evidence) == 0 {
		return evidence, nil
	}
	if err := projectedFinalConformance(root, logical, generated); err != nil {
		return nil, err
	}
	finalPayload := generatedPackagePayload(generated)
	result := make([]StrategyEvidence, 0, len(evidence))
	for _, item := range evidence {
		if len(item.ProofStages) != len(returnTailObligations)-1 || len(item.ContractObligations) != len(returnTailObligations) {
			return nil, fail("verify-result", "consume-return-tail-proof", "RETURN_TAIL_PROOF_CHAIN_UNPROVEN", "DIRECT_MISSING", "restore-return-tail-proof", nil)
		}
		stages := append([]ProofStageEvidence{}, item.ProofStages...)
		chain := returnTailProofChain{
			contract: append([]ContractObligationEvidence{}, item.ContractObligations...),
			sourceDigest: stages[0].SourceDigest, contractSource: item.ContractSourceDigest,
			contractSemantic: item.ContractSemanticDigest, candidateDigest: proofDigest(finalPayload), stages: stages,
		}
		if err := chain.consume(len(stages), returnTailPredicateResult{Status: "PASS", Payload: finalPayload, CandidateDigest: proofDigest(finalPayload), Detail: fmt.Sprintf("final generated package type-check passed; runtime conformance is not asserted (units=%d)", len(generated))}); err != nil {
			return nil, err
		}
		item.ProofStages = chain.stages
		item.Obligations = obligationsFromProofStages(chain.stages)
		item.FinalGeneratedBytes = generatedSourceBytes(generated)
		item.FinalGeneratedEvidenceBytes = len(finalPayload)
		item.FinalGeneratedUnits = len(generated)
		result = append(result, item)
	}
	return result, nil
}

func projectedFinalConformance(root, logical string, generated map[string][]byte) error {
	targetSource, ok := generated[logical]
	if !ok {
		return fail("verify-result", "projected-conformance", "PROJECTED_RESULT_MISSING", "DIRECT_MISSING", "restore-generated-result", nil)
	}
	fset := token.NewFileSet()
	target, err := parser.ParseFile(fset, logical, targetSource, parser.ParseComments)
	if err != nil {
		return failWithDiagnostics("verify-result", "projected-conformance", "PROJECTED_CONFORMANCE_FAILED", "KNOWN_CONTRADICTION", "report-counterexample", []string{"logical=" + logical})
	}
	files, err := packageTypeFiles(root, logical, fset, target)
	if err != nil {
		return err
	}
	paths := make([]string, 0, len(generated))
	for path := range generated {
		if path != logical {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	for _, path := range paths {
		file, parseErr := parser.ParseFile(fset, path, generated[path], parser.ParseComments)
		if parseErr != nil || file.Name.Name != target.Name.Name {
			return failWithDiagnostics("verify-result", "projected-conformance", "PROJECTED_CONFORMANCE_FAILED", "KNOWN_CONTRADICTION", "report-counterexample", []string{"generated=" + path})
		}
		files = append(files, file)
	}
	configuration := types.Config{Importer: newModuleImporter(root), Error: func(error) {}}
	if _, err := configuration.Check(filepath.ToSlash(filepath.Dir(logical)), fset, files, nil); err != nil {
		return failWithDiagnostics("verify-result", "projected-conformance", "PROJECTED_CONFORMANCE_FAILED", "KNOWN_CONTRADICTION", "report-counterexample", []string{"logical=" + logical})
	}
	return nil
}

func generatedSourceBytes(generated map[string][]byte) int {
	total := 0
	for _, source := range generated {
		total += len(source)
	}
	return total
}

func generatedPackagePayload(generated map[string][]byte) []byte {
	paths := make([]string, 0, len(generated))
	for path := range generated {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	var payload bytes.Buffer
	for _, path := range paths {
		payload.WriteString(path)
		payload.WriteByte(0)
		payload.Write(generated[path])
		payload.WriteByte(0)
	}
	return payload.Bytes()
}
