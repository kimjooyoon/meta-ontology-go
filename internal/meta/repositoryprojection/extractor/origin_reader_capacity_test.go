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

func TestOriginReaderRenderedCapacity(t *testing.T) {
	root := os.Getenv("GOOO_ORIGIN_READER_SOURCE")
	if root == "" {
		root = filepath.Join("..", "..", "..", "bodycodegen")
	}
	// Extraction moves declarations into helper files. Receiver and method names
	// identify these routines independently of the physical file partition.
	wanted := map[string]bool{"block": false, "statement": false, "assign": false, "loop": false,
		"collectionLoop": false, "declaration": false, "branch": false}
	paths, err := filepath.Glob(filepath.Join(root, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	methods, maximumFunction, maximumHelper := 0, 0, 0
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
		for _, node := range file.Decls {
			function, ok := node.(*ast.FuncDecl)
			if !ok || function.Recv == nil {
				continue
			}
			receiver, ok := receiverBaseIdentifier(function.Recv)
			if !ok || receiver != "valueOriginReader" {
				continue
			}
			if _, included := wanted[function.Name.Name]; !included {
				continue
			}
			if wanted[function.Name.Name] {
				t.Fatalf("duplicate origin method %s", function.Name.Name)
			}
			wanted[function.Name.Name] = true
			observation := observeRenderedCapacity(set, file, source, function)
			if observation.functionStatus != renderedCapacityWithinCap || observation.helperStatus != renderedCapacityWithinCap || observation.helperLines == nil {
				t.Fatalf("%s capacity not observed: %+v", function.Name.Name, observation)
			}
			methods++
			maximumFunction = max(maximumFunction, observation.functionLines)
			maximumHelper = max(maximumHelper, *observation.helperLines)
		}
	}
	if methods != len(wanted) {
		t.Fatalf("methods=%d missing=%v", methods, wanted)
	}
	t.Logf("seven origin-reader methods observed by identity with the actual helper renderer; maximum function lines=%d/helper lines=%d; no source program execution", maximumFunction, maximumHelper)
}
