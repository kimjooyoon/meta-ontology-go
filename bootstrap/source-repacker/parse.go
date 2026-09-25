package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type parsedSource struct {
	Path      string
	Subject   string
	Source    []byte
	Mode      os.FileMode
	File      *ast.File
	Fset      *token.FileSet
	Domain    buildDomain
	Imports   map[string]string
	DotImport bool
}

func loadSource(root, subject string) (parsedSource, error) {
	path, err := secureSourcePath(root, subject)
	if err != nil {
		return parsedSource{}, err
	}
	return parseSource(path, subject)
}

func parseSource(path, subject string) (parsedSource, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return parsedSource{}, fmt.Errorf("read %s: %w", subject, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return parsedSource{}, fmt.Errorf("stat %s: %w", subject, err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, subject, data, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return parsedSource{}, fmt.Errorf("parse %s: %w", subject, err)
	}
	imports, dot := importBindings(file)
	return parsedSource{Path: path, Subject: subject, Source: data, Mode: info.Mode(), File: file,
		Fset: fset, Domain: domainFor(subject, data, file), Imports: imports, DotImport: dot}, nil
}

func destinationSources(source parsedSource, limit int) ([]parsedSource, error) {
	entries, err := os.ReadDir(filepath.Dir(source.Path))
	if err != nil {
		return nil, err
	}
	destinations := make([]parsedSource, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || entry.Name() == filepath.Base(source.Path) {
			continue
		}
		subject := filepath.ToSlash(filepath.Join(filepath.Dir(source.Subject), entry.Name()))
		candidate, parseErr := parseSource(filepath.Join(filepath.Dir(source.Path), entry.Name()), subject)
		if parseErr != nil {
			return nil, parseErr
		}
		if candidate.Domain == source.Domain && physicalLines(candidate.Source) < limit {
			destinations = append(destinations, candidate)
		}
	}
	sort.Slice(destinations, func(i, j int) bool { return destinations[i].Subject < destinations[j].Subject })
	return destinations, nil
}
