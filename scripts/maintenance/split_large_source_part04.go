package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func refactor(path string, opts options) (int, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, data, parser.ParseComments)
	if err != nil {
		return 0, 0, nil
	}
	allLines := lineCount(data)
	if allLines <= opts.maxLines {
		return 0, 0, nil
	}
	importMap := buildImportMap(file)
	declarations := filterTopLevelDecls(file.Decls)
	if len(declarations) == 0 {
		return 0, 0, nil
	}
	chunks, unsplittable := buildChunks(file, fset, declarations, importMap, opts.maxLines)
	if len(chunks) <= 1 {
		return 0, unsplittable, nil
	}
	if opts.write {
		dir := filepath.Dir(path)
		base := filepath.Base(path)
		ext := filepath.Ext(base)
		stem := strings.TrimSuffix(base, ext)
		generated := make([]string, 0, len(chunks))
		for i, chunk := range chunks {
			next := filepath.Join(dir, fmt.Sprintf("%s_part%02d%s", stem, i+1, ext))
			generated = append(generated, next)
			err := writeChunk(next, path, file.Name.Name, chunk, importMap)
			if err != nil {
				return 0, unsplittable, err
			}
		}
		if err := os.Remove(path); err != nil {
			return 0, unsplittable, err
		}
		return 1, unsplittable, nil
	}
	fmt.Printf("splitter: would split %s into %d files (orig=%d)\n", path, len(chunks), allLines)
	return 1, unsplittable, nil
}
func filterTopLevelDecls(decls []ast.Decl) []ast.Decl {
	result := make([]ast.Decl, 0, len(decls))
	for _, decl := range decls {
		if gen, ok := decl.(*ast.GenDecl); ok {
			if gen.Tok == token.IMPORT {
				continue
			}
		}
		result = append(result, decl)
	}
	return result
}
