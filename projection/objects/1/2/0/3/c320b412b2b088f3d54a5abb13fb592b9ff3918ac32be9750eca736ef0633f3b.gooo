package main

import (
	"fmt"
	"go/format"
	"path/filepath"
	"sort"
	"strings"
)

func renderChunk(preamble []byte, packageName string, chunk declChunk, imports map[string]importSpec) ([]byte, error) {
	include := make([]importSpec, 0)
	for key := range chunk.imports {
		if spec, ok := imports[key]; ok {
			include = append(include, spec)
		}
	}
	sort.Slice(include, func(i, j int) bool {
		if include[i].path == include[j].path {
			return include[i].name < include[j].name
		}
		return include[i].path < include[j].path
	})
	head := append([]byte(nil), preamble...)
	if len(head) > 0 && head[len(head)-1] != '\n' {
		head = append(head, '\n')
	}
	head = append(head, []byte("package "+packageName+"\n\n")...)
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
		return nil, fmt.Errorf("format generated source: %w", err)
	}
	return formatted, nil
}
