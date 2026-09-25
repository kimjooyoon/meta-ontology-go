package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
)

func refactor(path string, opts options) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, data, parser.ParseComments)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", path, err)
	}
	allLines := lineCount(data)
	if allLines <= opts.maxLines {
		return 0, nil
	}
	if isGeneratedPart(path) {
		return 0, fmt.Errorf("generated source exceeds cap and is not idempotent: %s", path)
	}
	importMap := buildImportMap(file)
	declarations := filterTopLevelDecls(file.Decls)
	if len(declarations) == 0 {
		return 0, fmt.Errorf("no splittable declarations in %s", path)
	}
	if err := validateSplitSafety(file, declarations, importMap); err != nil {
		return 0, fmt.Errorf("unsafe split %s: %w", path, err)
	}
	preamble := sourcePreamble(fset, file, data)
	chunks, unsplittable := buildChunks(file, fset, declarations, importMap, preamble, opts.maxLines)
	if unsplittable > 0 {
		return 0, fmt.Errorf("unsplittable declarations in %s: %d", path, unsplittable)
	}
	if len(chunks) <= 1 {
		return 0, fmt.Errorf("source exceeds cap but cannot be partitioned: %s", path)
	}
	generated, err := planGenerated(path, file.Name.Name, preamble, chunks, importMap, opts.maxLines)
	if err != nil {
		return 0, err
	}
	if err := validateDirectoryProjection(path, generated, opts.maxEntries); err != nil {
		return 0, err
	}
	if opts.write {
		if err := commitGenerated(path, generated); err != nil {
			return 0, err
		}
		return 1, nil
	}
	fmt.Printf("splitter: would split %s into %d files (orig=%d)\n", path, len(chunks), allLines)
	return 1, nil
}
