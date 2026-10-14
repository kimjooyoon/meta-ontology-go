package extractor

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalOriginRenderedCapacity10921(t *testing.T) {
	root := os.Getenv("GOOO_LOCAL_CAPACITY_SOURCE")
	if root == "" {
		root = filepath.Join("..", "..", "..", "..")
	}
	wanted := map[string]bool{}
	for _, name := range []string{"bindLocalProjection", "bindLocalFunctionWrites", "bindLocalFunctionScope", "localFunctionSourceUses", "bindLocalFunctionUses", "compositionLocalSites", "localSiteBuilder.visit", "validateLocalTrace", "validateLocalWrites", "validateLocalUses", "observedReadProjectionWithLocals", "readProjection.collectLocalReadNodes"} {
		wanted[name] = false
	}
	paths, err := filepath.Glob(filepath.Join(root, "internal", "bodyexecution", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	maximum := 0
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		set := token.NewFileSet()
		file, err := parser.ParseFile(set, path, source, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
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
			seen, found := wanted[name]
			if !found {
				continue
			}
			if seen {
				t.Fatal("duplicate declaration", name)
			}
			wanted[name] = true
			observed := observeRenderedCapacity(set, file, source, function)
			if observed.functionStatus != renderedCapacityWithinCap || observed.helperStatus != renderedCapacityWithinCap || observed.helperLines == nil {
				t.Fatalf("%s: %+v", name, observed)
			}
			maximum = max(maximum, *observed.helperLines)
		}
	}
	for name, found := range wanted {
		if !found {
			t.Fatal("missing declaration", name)
		}
	}
	t.Logf("12 local-origin routines observed by declaration identity;maximum rendered helper%d lines;no source program executions", maximum)
}
