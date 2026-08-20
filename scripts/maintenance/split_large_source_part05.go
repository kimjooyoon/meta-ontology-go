package main

import (
	"go/ast"
	"go/token"
	"strings"
)

func buildImportMap(file *ast.File) []importSpec {
	result := make([]importSpec, 0, len(file.Imports))
	for _, spec := range file.Imports {
		path := strings.Trim(spec.Path.Value, "\"")
		name := path
		if spec.Name != nil && spec.Name.Name != "" {
			if spec.Name.Name == "_" || spec.Name.Name == "." {
				name = spec.Name.Name
			} else {
				name = spec.Name.Name
			}
		}
		if spec.Name == nil {
			if slash := strings.LastIndex(path, "/"); slash >= 0 {
				name = path[slash+1:]
			}
		}
		result = append(result, importSpec{name: name, path: path})
	}
	return result
}
func buildChunks(file *ast.File, fset *token.FileSet, decls []ast.Decl, imports []importSpec, preamble []byte, maxLines int) ([]declChunk, int) {
	chunks := make([]declChunk, 0)
	current := declChunk{imports: make(map[string]struct{})}
	allImports := make(map[string]importSpec, len(imports))
	for _, imp := range imports {
		allImports[imp.key()] = imp
	}
	unsplittable := 0
	for _, decl := range decls {
		declSource := renderDecl(fset, decl)
		declImports := usedImportsForDecl(file, decl)
		trial := declChunk{
			decls:      append(current.decls[:0:0], current.decls...),
			fileBodies: append(current.fileBodies[:0:0], current.fileBodies...),
			imports:    mergeImportSet(current.imports, declImports),
		}
		trial.decls = append(trial.decls, decl)
		trial.fileBodies = append(trial.fileBodies, declSource)
		if estimatedLines(preamble, file.Name.Name, trial, allImports) <= maxLines {
			current = trial
			continue
		}
		if len(current.decls) > 0 {
			chunks = append(chunks, current)
			current = declChunk{imports: make(map[string]struct{})}
		}
		trial = declChunk{decls: []ast.Decl{decl}, fileBodies: []string{declSource}, imports: declImports}
		if estimatedLines(preamble, file.Name.Name, trial, allImports) <= maxLines {
			current = trial
			continue
		}
		unsplittable++
		chunks = append(chunks, trial)
		current = declChunk{imports: make(map[string]struct{})}
	}
	if len(current.decls) > 0 {
		chunks = append(chunks, current)
	}
	return chunks, unsplittable
}
