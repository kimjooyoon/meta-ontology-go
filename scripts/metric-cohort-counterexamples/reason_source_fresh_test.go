package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestCountercaseKernelReasonsFresh1689(t *testing.T) {
	for _, row := range []struct{ name, directory, leaf, reason string }{
		{"source-rebuild", "metricprogram/conceptoperation", "concept-operation receipt does not match independently rebuilt evidence", sourceRebuildReason},
		{"source-replay", "metricstrategy/verify", "metric strategy independent synthesis diverged", sourceReplayReason},
		{"program-replay", "metricprogram/verify", "program does not match independent reconstruction", programReplayReason},
		{"program-source", "metricprogram/verify", "Gooo meta program source is not canonical", programSourceReason},
	} {
		t.Run(row.name, func(t *testing.T) {
			if !strings.HasPrefix(row.reason, sourceReasonPrefix) || !strings.HasSuffix(row.reason, row.leaf) {
				t.Fatal("countercase expects an incomplete refusal stage")
			}
			packages, err := parser.ParseDir(token.NewFileSet(), filepath.Join("../../internal/meta", row.directory),
				func(file os.FileInfo) bool { return !strings.HasSuffix(file.Name(), "_test.go") }, 0)
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, pkg := range packages {
				for _, file := range pkg.Files {
					ast.Inspect(file, func(node ast.Node) bool {
						if literal, ok := node.(*ast.BasicLit); ok && literal.Kind == token.STRING {
							value, err := strconv.Unquote(literal.Value)
							found = found || err == nil && value == row.leaf
						}
						return true
					})
				}
			}
			if !found {
				t.Fatal("countercase invented a refusal outside the inspected runtime kernel")
			}
		})
	}
}
