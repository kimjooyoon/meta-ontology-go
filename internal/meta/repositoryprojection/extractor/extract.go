package extractor

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"os"
)

func Extract(root, logical string) (map[string][]byte, []string, error) {
	result, err := ExtractWithResult(root, logical)
	return result.Generated, result.Paths, err
}

func ExtractWithResult(root, logical string) (Result, error) {
	path, err := safePath(root, logical)
	if err != nil {
		return Result{}, fail("observe-plan", "resolve-source", "WRITE_SET_ESCAPE", "KNOWN_CONTRADICTION", "report-contradiction", nil)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		return Result{}, fail("observe-plan", "read-source", "SOURCE_UNAVAILABLE", "DIRECT_MISSING", "restore-source", nil)
	}
	original := append([]byte(nil), source...)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, logical, source, parser.ParseComments)
	if err != nil {
		return Result{}, fail("observe-plan", "parse-source", "PARSER_EVIDENCE_MISSING", "DIRECT_MISSING", "restore-parser-evidence", nil)
	}
	source, fset, file, err = prepareParsedSource(root, logical, source, fset, file)
	if err != nil {
		return Result{}, err
	}
	if err := validateBuildHeader(source[:fset.Position(file.Package).Offset]); err != nil {
		return Result{}, err
	}
	list, err := imports(file)
	if err != nil {
		return Result{}, err
	}
	if err := validateImports(root, list); err != nil {
		return Result{}, err
	}
	all, _, err := candidates(fset, file)
	if err != nil {
		return Result{}, err
	}
	if len(all) == 0 {
		return Result{}, fail("derive-recipe", "select-declaration", "UNSUPPORTED_DECLARATION", "KNOWN_CONTRADICTION", "report-contradiction", nil)
	}
	output, partitions, err := capacityRender(fset, file, source, all, list, 75)
	if err != nil {
		return Result{}, err
	}
	generated := map[string][]byte{logical: output.source}
	paths := []string{logical}
	operations := make([]string, 0, 2)
	if !bytes.Equal(original, source) {
		operations = append(operations, "extract-function")
	}
	if len(partitions) > 0 {
		operations = append(operations, "move-complete-declarations")
	}
	for index, partition := range partitions {
		renderedHelper, renderErr := render(fset, file, source, partition, list)
		if renderErr != nil {
			return Result{}, renderErr
		}
		helper := helperPath(logical, index, len(partitions))
		helperName, pathErr := safePath(root, helper)
		if pathErr != nil {
			return Result{}, fail("generate-helpers", "resolve-helper", "WRITE_SET_ESCAPE", "KNOWN_CONTRADICTION", "report-contradiction", nil)
		}
		if _, statErr := os.Lstat(helperName); statErr == nil {
			return Result{}, fail("generate-helpers", "resolve-helper", "DECLARATION_IDENTITY_COLLISION", "KNOWN_CONTRADICTION", "report-contradiction", []string{helper})
		}
		if physicalLines(renderedHelper.helper) > 75 {
			return Result{}, failWithDiagnostics("generate-helpers", "render-helper", "NO_SAFE_DECLARATION_CAPACITY", "KNOWN_CONTRADICTION", "report-contradiction", []string{fmt.Sprintf("helper=%s", helper), fmt.Sprintf("helper_lines=%d", physicalLines(renderedHelper.helper))})
		}
		generated[helper] = renderedHelper.helper
		paths = append(paths, helper)
	}
	return Result{Generated: generated, Paths: paths, Operations: operations}, nil
}
