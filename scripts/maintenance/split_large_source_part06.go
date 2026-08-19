package main

import (
	"fmt"
	"go/format"
	"path/filepath"
	"sort"
	"strings"
)

func mergeImportSet(base map[string]struct{}, extra map[string]struct{}) map[string]struct{} {
	merged := make(map[string]struct{}, len(base)+len(extra))
	for value := range base {
		merged[value] = struct{}{}
	}
	for value := range extra {
		merged[value] = struct{}{}
	}
	return merged
}
func estimatedLines(packageName string, chunk declChunk, allImports map[string]importSpec) int {
	parts := make([]byte, 0, 256)
	parts = append(parts, []byte("package "+packageName+"\n\n")...)
	useImports := make([]importSpec, 0, len(chunk.imports))
	for name := range chunk.imports {
		if imp, ok := allImports[name]; ok {
			useImports = append(useImports, imp)
		}
	}
	if len(useImports) > 0 {
		sort.Slice(useImports, func(i, j int) bool {
			if useImports[i].path == useImports[j].path {
				return useImports[i].name < useImports[j].name
			}
			return useImports[i].path < useImports[j].path
		})
		parts = append(parts, []byte("import (\n")...)
		for _, imp := range useImports {
			if imp.name == "_" || imp.name == "." {
				parts = append(parts, []byte(fmt.Sprintf("\t%s \"%s\"\n", imp.name, imp.path))...)
			} else if imp.name == filepath.Base(imp.path) {
				parts = append(parts, []byte(fmt.Sprintf("\t\"%s\"\n", imp.path))...)
			} else {
				parts = append(parts, []byte(fmt.Sprintf("\t%s \"%s\"\n", imp.name, imp.path))...)
			}
		}
		parts = append(parts, []byte(")\n\n")...)
	}
	parts = append(parts, []byte(strings.Join(chunk.fileBodies, "\n"))...)
	parts = append(parts, '\n')
	formatted, err := format.Source(parts)
	if err != nil {
		formatted = parts
	}
	return lineCount(formatted)
}
