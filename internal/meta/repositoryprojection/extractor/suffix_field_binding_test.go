package extractor

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"reflect"
	"testing"
)

func TestSuffixFieldSelectorsAreNotLexicalCaptures(t *testing.T) {
	cases := []struct {
		name, source string
		want         []string
	}{
		{"anonymous-record", `package p
func probe() {
 first := struct{ Digest string }{Digest: "original"}
 replay := first
 if replay.Digest != first.Digest { panic("changed") }
}`, []string{"first", "replay"}},
		{"same-named-local", `package p
func probe() {
 Digest := "original"
 record := struct{ Digest string }{Digest: Digest}
 if record.Digest != Digest { panic("changed") }
}`, []string{"Digest", "record"}},
		{"nested-record", `package p
func probe() {
 first := struct{ Payload struct{ Digest string } }{}
 replay := first
 if replay.Payload.Digest != first.Payload.Digest { panic("changed") }
}`, []string{"first", "replay"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "fixture.go", tc.source, 0)
			if err != nil {
				t.Fatal(err)
			}
			info := &types.Info{Defs: map[*ast.Ident]types.Object{}, Uses: map[*ast.Ident]types.Object{}}
			var config types.Config
			pkg, err := config.Check("fixture", fset, []*ast.File{file}, info)
			if err != nil {
				t.Fatal(err)
			}
			function := file.Decls[0].(*ast.FuncDecl)
			statements := function.Body.List[2:]
			bindings, err := suffixBindings(statements, function, fset, typeEvidence{info: info, pkg: pkg})
			if err != nil {
				t.Fatal(err)
			}
			names := make([]string, 0, len(bindings))
			for _, b := range bindings {
				names = append(names, b.name)
			}
			if !reflect.DeepEqual(names, tc.want) {
				t.Fatalf("field became a free variable: got %v want %v", names, tc.want)
			}
			start := fset.Position(statements[0].Pos()).Offset
			end := fset.Position(statements[len(statements)-1].End()).Offset
			helper, err := renderSuffixHelper(fset, "extracted", bindings, []byte(tc.source[start:end]))
			if err != nil {
				t.Fatal(err)
			}
			rewritten := tc.source[:start] + "extracted(" + bindingArguments(bindings) + ")" + tc.source[end:] + "\n" + string(helper)
			newSet := token.NewFileSet()
			newFile, err := parser.ParseFile(newSet, "projected.go", rewritten, 0)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = config.Check("fixture", newSet, []*ast.File{newFile}, nil); err != nil {
				t.Fatalf("projected binding no longer typechecks: %v\n%s", err, rewritten)
			}
		})
	}
}
