package extractor

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"
)

func Extract(root, logical string) (map[string][]byte, []string, error) {
	path, err := safePath(root, logical); if err != nil { return nil, nil, fail("observe-plan", "resolve-source", "WRITE_SET_ESCAPE", "KNOWN_CONTRADICTION", "report-contradiction", nil) }
	source, err := os.ReadFile(path); if err != nil { return nil, nil, fail("observe-plan", "read-source", "SOURCE_UNAVAILABLE", "DIRECT_MISSING", "restore-source", nil) }
	fset := token.NewFileSet(); file, err := parser.ParseFile(fset, logical, source, parser.ParseComments); if err != nil { return nil, nil, fail("observe-plan", "parse-source", "PARSER_EVIDENCE_MISSING", "DIRECT_MISSING", "restore-parser-evidence", nil) }
	if err := validateBuildHeader(source[:fset.Position(file.Package).Offset]); err != nil { return nil, nil, err }
	list, err := imports(file); if err != nil { return nil, nil, err }; if err := validateImports(root, list); err != nil { return nil, nil, err }
	all, err := candidates(fset, file); if err != nil { return nil, nil, err }; if len(all) == 0 { return nil, nil, fail("derive-recipe", "select-declaration", "UNSUPPORTED_DECLARATION", "KNOWN_CONTRADICTION", "report-contradiction", nil) }
	ordered := append([]declaration(nil), all...); sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].end-ordered[i].start > ordered[j].end-ordered[j].start })
	var chosen []declaration; var output rendered; remainingLines := 0
	for _, item := range ordered { chosen = append(chosen, item); candidate, renderErr := render(fset, file, source, chosen, list); if renderErr != nil { return nil, nil, renderErr }; remainingLines = bytes.Count(candidate.source, []byte{'\n'}); if remainingLines <= 75 { output = candidate; break } }
	if len(output.source) == 0 { ids := make([]string, 0, len(all)); for _, item := range all { ids = append(ids, item.identity) }; return nil, nil, failWithDiagnostics("derive-recipe", "select-declaration", "NO_SAFE_CAPACITY", "KNOWN_CONTRADICTION", "report-contradiction", []string{fmt.Sprintf("candidate_count=%d", len(all)), fmt.Sprintf("remaining_lines=%d", remainingLines), strings.Join(ids, ",")) ) }
	helper := helperPath(logical); helperName, err := safePath(root, helper); if err != nil { return nil, nil, fail("generate-helpers", "resolve-helper", "WRITE_SET_ESCAPE", "KNOWN_CONTRADICTION", "report-contradiction", nil) }; if _, err := os.Lstat(helperName); err == nil { return nil, nil, fail("generate-helpers", "resolve-helper", "DECLARATION_IDENTITY_COLLISION", "KNOWN_CONTRADICTION", "report-contradiction", []string{helper}) }
	return map[string][]byte{logical: output.source, helper: output.helper}, []string{logical, helper}, nil
}
