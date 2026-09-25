package extractor

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"os"
)

func Extract(root, logical string) (map[string][]byte, []string, error) {
	result, err := ExtractWithResult(root, logical)
	return result.Generated, result.Paths, err
}

func ExtractWithResult(root, logical string) (Result, error) {
	source, fset, file, err := readParsedSource(root, logical)
	if err != nil {
		return Result{}, err
	}
	original := append([]byte(nil), source...)
	var strategyEvidence []StrategyEvidence
	source, fset, file, strategyEvidence, err = prepareParsedSource(root, logical, source, fset, file)
	if err != nil {
		return Result{}, withCallbackExtractionProposal(root, logical, original, err)
	}
	list, all, err := extractionInputs(root, source, fset, file)
	if err != nil {
		return Result{}, err
	}
	return buildExtractionResult(root, logical, original, source, fset, file, list, all, strategyEvidence)
}

func readParsedSource(root, logical string) ([]byte, *token.FileSet, *ast.File, error) {
	path, err := safePath(root, logical)
	if err != nil {
		return nil, nil, nil, fail("observe-plan", "resolve-source", "WRITE_SET_ESCAPE", "KNOWN_CONTRADICTION", "report-contradiction", nil)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, fail("observe-plan", "read-source", "SOURCE_UNAVAILABLE", "DIRECT_MISSING", "restore-source", nil)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, logical, source, parser.ParseComments)
	if err != nil {
		return nil, nil, nil, fail("observe-plan", "parse-source", "PARSER_EVIDENCE_MISSING", "DIRECT_MISSING", "restore-parser-evidence", nil)
	}
	return source, fset, file, nil
}

func extractionInputs(root string, source []byte, fset *token.FileSet, file *ast.File) ([]importSpec, []declaration, error) {
	if err := validateBuildHeader(source[:fset.Position(file.Package).Offset]); err != nil {
		return nil, nil, err
	}
	list, err := imports(file)
	if err != nil {
		return nil, nil, err
	}
	if err := validateImports(root, list); err != nil {
		return nil, nil, err
	}
	all, _, err := candidates(fset, file)
	if err != nil {
		return nil, nil, err
	}
	if len(all) == 0 {
		return nil, nil, fail("derive-recipe", "select-declaration", "UNSUPPORTED_DECLARATION", "KNOWN_CONTRADICTION", "report-contradiction", nil)
	}
	return list, all, nil
}

func buildExtractionResult(root, logical string, original, source []byte, fset *token.FileSet, file *ast.File, list []importSpec, all []declaration, strategyEvidence []StrategyEvidence) (Result, error) {
	output, partitions, err := capacityRender(fset, file, source, all, list, 75)
	if err != nil {
		return Result{}, err
	}
	helpers, paths, err := renderHelperFiles(root, logical, source, fset, file, list, partitions)
	if err != nil {
		return Result{}, err
	}
	generated := map[string][]byte{logical: output.source}
	maps.Copy(generated, helpers)
	paths = append([]string{logical}, paths...)
	strategyEvidence, err = finalizeReturnTailEvidence(root, logical, generated, strategyEvidence)
	if err != nil {
		return Result{}, err
	}
	return Result{Generated: generated, Paths: paths, Operations: extractionOperations(original, source, partitions), Evidence: strategyEvidence}, nil
}

func renderHelperFiles(root, logical string, source []byte, fset *token.FileSet, file *ast.File, list []importSpec, partitions [][]declaration) (map[string][]byte, []string, error) {
	generated := make(map[string][]byte, len(partitions))
	paths := make([]string, 0, len(partitions))
	for index, partition := range partitions {
		renderedHelper, renderErr := render(fset, file, source, partition, list)
		if renderErr != nil {
			return nil, nil, renderErr
		}
		helper := helperPath(logical, index, len(partitions))
		helperName, pathErr := safePath(root, helper)
		if pathErr != nil {
			return nil, nil, fail("generate-helpers", "resolve-helper", "WRITE_SET_ESCAPE", "KNOWN_CONTRADICTION", "report-contradiction", nil)
		}
		if _, statErr := os.Lstat(helperName); statErr == nil {
			return nil, nil, fail("generate-helpers", "resolve-helper", "DECLARATION_IDENTITY_COLLISION", "KNOWN_CONTRADICTION", "report-contradiction", []string{helper})
		}
		if physicalLines(renderedHelper.helper) > 75 {
			return nil, nil, failWithDiagnostics("generate-helpers", "render-helper", "NO_SAFE_DECLARATION_CAPACITY", "KNOWN_CONTRADICTION", "report-counterexample", []string{fmt.Sprintf("helper=%s", helper), fmt.Sprintf("helper_lines=%d", physicalLines(renderedHelper.helper))})
		}
		generated[helper] = renderedHelper.helper
		paths = append(paths, helper)
	}
	return generated, paths, nil
}

func extractionOperations(original, source []byte, partitions [][]declaration) []string {
	operations := make([]string, 0, 2)
	if !bytes.Equal(original, source) {
		operations = append(operations, "extract-function")
	}
	if len(partitions) > 0 {
		operations = append(operations, "move-complete-declarations")
	}
	return operations
}
