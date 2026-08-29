package extractor

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
)

func Extract(root, logical string) (map[string][]byte, []string, error) {
	path, err := safePath(root, logical)
	if err != nil {
		return nil, nil, fail("observe-plan", "resolve-source", "WRITE_SET_ESCAPE", "KNOWN_CONTRADICTION", "report-contradiction", nil)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fail("observe-plan", "read-source", "SOURCE_UNAVAILABLE", "DIRECT_MISSING", "restore-source", nil)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, logical, source, parser.ParseComments)
	if err != nil {
		return nil, nil, fail("observe-plan", "parse-source", "PARSER_EVIDENCE_MISSING", "DIRECT_MISSING", "restore-parser-evidence", nil)
	}
	source, fset, file, err = prepareParsedSource(root, logical, source, fset, file)
	if err != nil {
		return nil, nil, err
	}
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
	output, partitions, err := capacityRender(fset, file, source, all, list, 75)
	if err != nil {
		return nil, nil, err
	}
	generated := map[string][]byte{logical: output.source}
	paths := []string{logical}
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
			return nil, nil, failWithDiagnostics("generate-helpers", "render-helper", "NO_SAFE_DECLARATION_CAPACITY", "KNOWN_CONTRADICTION", "report-contradiction", []string{fmt.Sprintf("helper=%s", helper), fmt.Sprintf("helper_lines=%d", physicalLines(renderedHelper.helper))})
		}
		generated[helper] = renderedHelper.helper
		paths = append(paths, helper)
	}
	return generated, paths, nil
}
