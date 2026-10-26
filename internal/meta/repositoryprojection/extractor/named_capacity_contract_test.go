package extractor

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type namedCapacityObservation struct {
	path     string
	capacity renderedCapacityObservation
}

// A logical source split moves declarations between files in one package.
// Package/receiver/declaration identity keeps the required subject set intact.
func collectNamedCapacities(root string, targets []string) (map[string]namedCapacityObservation, error) {
	return collectNamedCapacityFiles(root, targets, false)
}

func collectNamedCapacityFiles(root string, targets []string, includeTests bool) (map[string]namedCapacityObservation, error) {
	wanted, packages := map[string]bool{}, map[string]bool{}
	for _, target := range targets {
		pkg, _, ok := strings.Cut(target, "#")
		if !ok || wanted[target] {
			return nil, fmt.Errorf("invalid or repeated capacity subject %q", target)
		}
		wanted[target], packages[pkg] = true, true
	}
	found := map[string]namedCapacityObservation{}
	for pkg := range packages {
		paths, err := filepath.Glob(filepath.Join(root, filepath.FromSlash(pkg), "*.go"))
		if err != nil {
			return nil, err
		}
		for _, path := range paths {
			if !includeTests && strings.HasSuffix(path, "_test.go") {
				continue
			}
			if err := collectNamedCapacityFile(path, pkg, wanted, found); err != nil {
				return nil, err
			}
		}
	}
	for target := range wanted {
		if _, ok := found[target]; !ok {
			return nil, fmt.Errorf("required declaration missing: %s", target)
		}
	}
	return found, nil
}

func collectNamedCapacityFile(path, pkg string, wanted map[string]bool, found map[string]namedCapacityObservation) error {
	source, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, path, source, parser.ParseComments)
	if err != nil {
		return err
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		name := function.Name.Name
		if function.Recv != nil {
			receiver, ok := receiverBaseIdentifier(function.Recv)
			if !ok {
				continue
			}
			name = receiver + "." + name
		}
		key := pkg + "#" + name
		if !wanted[key] {
			continue
		}
		if _, duplicate := found[key]; duplicate {
			return fmt.Errorf("duplicate declaration: %s", key)
		}
		found[key] = namedCapacityObservation{path, observeRenderedCapacity(set, file, source, function)}
	}
	return nil
}

func requireNamedCapacities(t *testing.T, root string, targets []string) {
	t.Helper()
	observed, err := collectNamedCapacities(root, targets)
	if err != nil {
		t.Fatal(err)
	}
	requireObservedCapacities(t, targets, observed)
}

func requireObservedCapacities(t *testing.T, targets []string, observed map[string]namedCapacityObservation) {
	t.Helper()
	for _, target := range targets {
		value := observed[target]
		result := value.capacity
		if result.functionStatus != renderedCapacityWithinCap || result.helperStatus != renderedCapacityWithinCap || result.helperLines == nil {
			t.Fatalf("%s in %s: %+v", target, value.path, result)
		}
		t.Log(target, "rendered", *result.helperLines)
	}
	if len(observed) != len(targets) {
		t.Fatal("declaration population", len(observed), len(targets))
	}
}
