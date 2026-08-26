package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
)

func transformCandidate(root string, candidate candidate, write bool) error {
	if filepath.IsAbs(candidate.source.Path) {
		return fmt.Errorf("absolute candidate path %q", candidate.source.Path)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	path := filepath.Join(absRoot, filepath.FromSlash(candidate.source.Path))
	relative, err := filepath.Rel(absRoot, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("candidate escapes root: %q", candidate.source.Path)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, source, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return err
	}
	matches, collapsible := 0, true
	ast.Inspect(file, func(node ast.Node) bool {
		name, line, ok := linecaps.FunctionIdentity(fset, node)
		if !ok || name != candidate.source.Name || line != candidate.source.Line {
			return true
		}
		matches++
		collapsible = linecaps.CollapseAssignReturn(fset, node, file.Comments) && collapsible
		return false
	})
	if matches != 1 || !collapsible {
		return fmt.Errorf("candidate %q matches=%d collapsible=%t", candidate.raw, matches, collapsible)
	}
	var output bytes.Buffer
	if err := format.Node(&output, fset, file); err != nil {
		return err
	}
	if bytes.Equal(source, output.Bytes()) {
		return fmt.Errorf("candidate %q produced no change", candidate.raw)
	}
	if !write {
		return nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return writeAtomic(path, output.Bytes(), info.Mode())
}
