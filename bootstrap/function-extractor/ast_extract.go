package main

import (
	"go/parser"
	"go/token"
	"os"
	"sort"
)

func genericASTExtraction(root, logical string) (map[string][]byte, []string, error) {
	path, err := extractionPath(root, logical)
	if err != nil {
		return nil, nil, extractionError("observe-plan", "resolve-source", "WRITE_SET_ESCAPE", "KNOWN_CONTRADICTION", "report-contradiction", []string{})
	}
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, extractionError("observe-plan", "read-source", "SOURCE_UNAVAILABLE", "DIRECT_MISSING", "restore-source", []string{})
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, logical, source, parser.ParseComments)
	if err != nil {
		return nil, nil, extractionError("observe-plan", "parse-source", "PARSER_EVIDENCE_MISSING", "DIRECT_MISSING", "restore-parser-evidence", []string{})
	}
	if err := validateGenericBuildHeader(source[:fset.Position(file.Package).Offset]); err != nil {
		return nil, nil, err
	}
	imports, err := genericImports(file)
	if err != nil {
		return nil, nil, err
	}
	if err := validateGenericImports(root, imports); err != nil {
		return nil, nil, err
	}
	candidates, err := genericCandidates(fset, file)
	if err != nil || len(candidates) == 0 {
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, extractionError("derive-recipe", "select-declaration", "UNSUPPORTED_DECLARATION", "KNOWN_CONTRADICTION", "report-contradiction", []string{})
	}
	ordered := append([]astDecl(nil), candidates...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].end-ordered[i].start > ordered[j].end-ordered[j].start })
	var chosen []astDecl
	var rendered astRendered
	for _, candidate := range ordered {
		chosen = append(chosen, candidate)
		candidateRender, renderErr := renderGeneric(fset, file, source, candidates, chosen, imports)
		if renderErr != nil {
			return nil, nil, renderErr
		}
		if countLines(candidateRender.source) <= 75 {
			rendered = candidateRender
			break
		}
	}
	if len(rendered.source) == 0 {
		return nil, nil, extractionError("derive-recipe", "select-declaration", "NO_SAFE_CAPACITY", "KNOWN_CONTRADICTION", "report-contradiction", []string{})
	}
	helperLogical := genericHelperPath(logical)
	helperPath, err := extractionPath(root, helperLogical)
	if err != nil {
		return nil, nil, extractionError("generate-helpers", "resolve-helper", "WRITE_SET_ESCAPE", "KNOWN_CONTRADICTION", "report-contradiction", []string{})
	}
	if _, err := os.Lstat(helperPath); err == nil {
		return nil, nil, extractionError("generate-helpers", "resolve-helper", "DECLARATION_IDENTITY_COLLISION", "KNOWN_CONTRADICTION", "report-contradiction", []string{helperLogical})
	}
	return map[string][]byte{logical: rendered.source, helperLogical: rendered.helper}, []string{logical, helperLogical}, nil
}
