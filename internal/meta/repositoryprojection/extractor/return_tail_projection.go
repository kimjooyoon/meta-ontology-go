package extractor

import (
	"bytes"
	"fmt"
	"go/ast"
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
	finalCapacity, finalCapacityPayload, err := finalRenderedCapacityEvidence(generated, finalPayload)
	if err != nil {
		return nil, err
	}
	if finalCapacity.Overage > 0 {
		return nil, failWithDiagnostics("verify-result", "consume-rendered-capacity-proof", "NO_SAFE_DECLARATION_CAPACITY", "KNOWN_CONTRADICTION", "report-counterexample", []string{
			fmt.Sprintf("final_overage=%d", finalCapacity.Overage),
		})
	}
	result := make([]StrategyEvidence, 0, len(evidence))
	for _, item := range evidence {
		if item.Strategy == suffixStrategy {
			item.FinalRenderedCapacity = &finalCapacity
			item.FinalGeneratedBytes = generatedSourceBytes(generated)
			item.FinalGeneratedEvidenceBytes = len(finalPayload)
			item.FinalGeneratedUnits = len(generated)
			result = append(result, item)
			continue
		}
		if item.Strategy != returnTailStrategy {
			return nil, fail("verify-result", "consume-extraction-proof", "EXTRACTION_STRATEGY_UNSUPPORTED", "KNOWN_CONTRADICTION", "report-counterexample", nil)
		}
		if len(item.ProofStages) != len(returnTailObligations)-2 || len(item.ContractObligations) != len(returnTailObligations) {
			return nil, fail("verify-result", "consume-return-tail-proof", "RETURN_TAIL_PROOF_CHAIN_UNPROVEN", "DIRECT_MISSING", "restore-return-tail-proof", nil)
		}
		stages := append([]ProofStageEvidence{}, item.ProofStages...)
		chain := returnTailProofChain{
			contract:     append([]ContractObligationEvidence{}, item.ContractObligations...),
			sourceDigest: stages[0].SourceDigest, contractSource: item.ContractSourceDigest,
			contractSemantic: item.ContractSemanticDigest, candidateDigest: proofDigest(finalPayload), stages: stages,
		}
		if err := chain.consume(len(stages), returnTailPredicateResult{Status: "PASS", Payload: finalCapacityPayload, CandidateDigest: proofDigest(finalPayload), Detail: fmt.Sprintf("final generated capacity is within the rendered line limit (lines=%d overage=%d)", finalCapacity.Lines, finalCapacity.Overage)}); err != nil {
			return nil, err
		}
		if err := chain.consume(len(stages)+1, returnTailPredicateResult{Status: "PASS", Payload: finalPayload, CandidateDigest: proofDigest(finalPayload), Detail: fmt.Sprintf("final generated package type-check passed; runtime conformance is not asserted (units=%d)", len(generated))}); err != nil {
			return nil, err
		}
		item.ProofStages = chain.stages
		item.Obligations = obligationsFromProofStages(chain.stages)
		item.FinalRenderedCapacity = &finalCapacity
		item.FinalGeneratedBytes = generatedSourceBytes(generated)
		item.FinalGeneratedEvidenceBytes = len(finalPayload)
		item.FinalGeneratedUnits = len(generated)
		result = append(result, item)
	}
	return result, nil
}

func finalRenderedCapacityEvidence(generated map[string][]byte, finalPayload []byte) (FinalRenderedCapacityEvidence, []byte, error) {
	capacityPayload := append([]byte("final-rendered-capacity\x00"), finalPayload...)
	evidence := FinalRenderedCapacityEvidence{
		Scope:         "final-generated-functions",
		PayloadDigest: proofDigest(capacityPayload),
		Status:        "PASS",
	}
	paths := make([]string, 0, len(generated))
	for path := range generated {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		source := generated[path]
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, source, parser.ParseComments)
		if err != nil {
			return FinalRenderedCapacityEvidence{}, nil, failWithDiagnostics("verify-result", "render-final-capacity", "PREFLIGHT_RENDER_FAILED", "DIRECT_MISSING", "restore-render-evidence", []string{"logical=" + path})
		}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Name == nil {
				continue
			}
			rendered, err := renderedDeclarationHelper(fset, file, source, function)
			if err != nil {
				return FinalRenderedCapacityEvidence{}, nil, err
			}
			measurement, err := canonicalRenderedCapacity(rendered)
			if err != nil {
				return FinalRenderedCapacityEvidence{}, nil, err
			}
			evidence.Bytes += measurement.bytes
			evidence.Lines += measurement.lines
			evidence.Overage += measurement.overage
		}
	}
	if evidence.Bytes <= 0 || evidence.Lines <= 0 {
		return FinalRenderedCapacityEvidence{}, nil, failWithDiagnostics("verify-result", "render-final-capacity", "PREFLIGHT_RENDER_FAILED", "DIRECT_MISSING", "restore-render-evidence", []string{"measurement=UNMEASURED"})
	}
	return evidence, capacityPayload, nil
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
