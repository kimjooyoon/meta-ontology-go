package extractor

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"strings"
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
	all, fallbackUsed, err := candidates(fset, file)
	if err != nil {
		return nil, nil, err
	}
	if len(all) == 0 {
		return nil, nil, fail("derive-recipe", "select-declaration", "UNSUPPORTED_DECLARATION", "KNOWN_CONTRADICTION", "report-contradiction", nil)
	}
	output, err := render(fset, file, source, all, list)
	if err != nil {
		return nil, nil, err
	}
	remainingLines := bytes.Count(output.source, []byte{'\n'})
	if remainingLines > 75 {
		ids := make([]string, 0, len(all))
		for _, item := range all {
			ids = append(ids, item.identity)
		}
		return nil, nil, failWithDiagnostics("derive-recipe", "select-declaration", "NO_SAFE_CAPACITY", "KNOWN_CONTRADICTION", "report-contradiction", []string{fmt.Sprintf("candidate_count=%d", len(all)), fmt.Sprintf("remaining_lines=%d", remainingLines), strings.Join(ids, ",")})
	}
	partitions := [][]declaration{all}
	if fallbackUsed && len(all) > 1 {
		midpoint := len(all) / 2
		partitions = [][]declaration{all[:midpoint], all[midpoint:]}
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
		generated[helper] = renderedHelper.helper
		paths = append(paths, helper)
	}
	return generated, paths, nil
}
