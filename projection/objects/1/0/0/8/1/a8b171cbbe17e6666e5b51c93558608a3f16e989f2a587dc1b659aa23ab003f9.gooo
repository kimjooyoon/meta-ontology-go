package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

type generatedFile struct {
	path     string
	contents []byte
}

func sourcePreamble(fset *token.FileSet, file *ast.File, source []byte) []byte {
	offset := fset.Position(file.Package).Offset
	if offset <= 0 || offset > len(source) {
		return nil
	}
	return append([]byte(nil), source[:offset]...)
}

func planGenerated(path, packageName string, preamble []byte, chunks []declChunk, imports []importSpec, maxLines int) ([]generatedFile, error) {
	indexed := make(map[string]importSpec, len(imports))
	for _, imp := range imports {
		indexed[imp.key()] = imp
	}
	generated := make([]generatedFile, 0, len(chunks))
	seen := make(map[string]struct{}, len(chunks))
	for index, chunk := range chunks {
		next, err := generatedPath(path, index+1)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[next]; ok {
			return nil, fmt.Errorf("duplicate generated path: %s", next)
		}
		seen[next] = struct{}{}
		contents, err := renderChunk(preamble, packageName, chunk, indexed)
		if err != nil {
			return nil, err
		}
		if lines := lineCount(contents); lines > maxLines {
			return nil, fmt.Errorf("generated file exceeds cap: %s: %d", next, lines)
		}
		if _, err := parser.ParseFile(token.NewFileSet(), next, contents, parser.ParseComments); err != nil {
			return nil, fmt.Errorf("validate generated file %s: %w", next, err)
		}
		generated = append(generated, generatedFile{path: next, contents: contents})
	}
	return generated, nil
}
