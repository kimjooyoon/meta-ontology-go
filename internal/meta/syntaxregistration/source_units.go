package syntaxregistration

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"slices"
	"strconv"
	"strings"
)

// Package namespaces and native symbols define roles; filenames do not.
// Every observed unit is pinned, including units that contribute no edit.
func sourceInputPaths(repository fs.FS) ([]string, error) {
	var paths []string
	for _, root := range []string{syntaxRoot, syntaxRoot + "conformance/", closureRoot} {
		found, err := fs.Glob(repository, root+"*.go")
		if err != nil {
			return nil, err
		}
		paths = append(paths, found...)
	}
	slices.Sort(paths)
	return paths, nil
}

func parseSourceUnits(inputs map[string][]byte, root, packageName string, tests bool) (*goSource, error) {
	source := &goSource{set: token.NewFileSet(), file: &ast.File{Name: &ast.Ident{Name: packageName}},
		units: map[string][]byte{}}
	for _, name := range sortedPaths(inputs) {
		if path.Dir(name)+"/" != root || !strings.HasSuffix(name, ".go") ||
			(!tests && strings.HasSuffix(name, "_test.go")) {
			continue
		}
		unit, err := parser.ParseFile(source.set, name, inputs[name], parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", failure("REFUTED", "parse-source:"+name,
				"REGISTRATION_SOURCE_UNIT_INVALID", "", "repair-source-unit"), err)
		}
		if unit.Name.Name != packageName {
			continue
		}
		source.units[name] = inputs[name]
		source.file.Decls = append(source.file.Decls, unit.Decls...)
	}
	if len(source.units) == 0 {
		return nil, failure("UNKNOWN", "bind-package:"+packageName, "REGISTRATION_SOURCE_UNITS_MISSING",
			"DIRECT_MISSING", "restore-source-units")
	}
	return source, nil
}

func sortedPaths[V any](values map[string]V) []string {
	paths := make([]string, 0, len(values))
	for name := range values {
		paths = append(paths, name)
	}
	slices.Sort(paths)
	return paths
}

func (source *goSource) requireImport(anchor ast.Node, imported, alias string) error {
	name, _ := source.location(anchor.Pos())
	unit, err := parser.ParseFile(token.NewFileSet(), name, source.units[name], parser.ImportsOnly)
	if err != nil {
		return err
	}
	for _, item := range unit.Imports {
		value, err := strconv.Unquote(item.Path.Value)
		if err != nil || value != imported {
			continue
		}
		bound := path.Base(imported)
		if item.Name != nil {
			bound = item.Name.Name
		}
		if bound == alias {
			return nil
		}
	}
	return failure("UNKNOWN", "bind-import:"+name+":"+imported, "REGISTRATION_IMPORT_BINDING_MISSING",
		"DIRECT_MISSING", "restore-native-import-binding")
}
