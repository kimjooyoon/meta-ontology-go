package main

import (
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func writeChunk(path, sourcePath, packageName string, chunk declChunk, allImports []importSpec) error {
	imports := make(map[string]importSpec, len(allImports))
	for _, imp := range allImports {
		imports[imp.name] = imp
	}
	include := make([]importSpec, 0)
	for imp := range chunk.imports {
		if spec, ok := imports[imp]; ok {
			include = append(include, spec)
		}
	}
	sort.Slice(include, func(i, j int) bool {
		if include[i].path == include[j].path {
			return include[i].name < include[j].name
		}
		return include[i].path < include[j].path
	})
	head := []byte("package " + packageName + "\n\n")
	if len(include) > 0 {
		builder := &strings.Builder{}
		builder.WriteString("import (\n")
		for _, imp := range include {
			if imp.name == "_" || imp.name == "." {
				builder.WriteString("\t" + imp.name + " \"" + imp.path + "\"\n")
			} else if imp.name == filepath.Base(imp.path) {
				builder.WriteString("\t\"" + imp.path + "\"\n")
			} else {
				builder.WriteString("\t" + imp.name + " \"" + imp.path + "\"\n")
			}
		}
		builder.WriteString(")\n\n")
		head = append(head, []byte(builder.String())...)
	}
	bodies := strings.Join(chunk.fileBodies, "\n")
	contents := append(head, []byte(bodies)...)
	contents = append(contents, '\n')
	formatted, err := format.Source(contents)
	if err != nil {
		formatted = contents
	}
	return os.WriteFile(path, formatted, 0o644)
}
func renderDecl(fset *token.FileSet, decl ast.Decl) string {
	buf := &strings.Builder{}
	_ = format.Node(buf, fset, decl)
	return buf.String()
}
func lineCount(buf []byte) int {
	if len(buf) == 0 {
		return 0
	}
	lines := 1 + bytesCount(buf, '\n')
	return lines
}
