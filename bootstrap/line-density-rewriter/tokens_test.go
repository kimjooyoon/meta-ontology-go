package main

import (
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"
)

func TestOneLineTokensPreservesFragmentBoundaries(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   string
		ok     bool
	}{
		{"string EOF", "\"a\"", "\"a\"", true},
		{"continued sum", "\"a\" +\n\"b\"", "\"a\" + \"b\"", true},
		{"internal newline", "left := 1\nright := 2", "left := 1 ; right := 2", true},
		{"explicit terminator", "left := 1;", "left := 1 ;", true},
		{"raw literal", "`first\nsecond`", "`first\nsecond`", true},
		{"comment", "left := 1 // observed\n", "", false},
		{"invalid literal", "\"unfinished", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := oneLineTokens([]byte(tc.source))
			if ok != tc.ok || ok && got != tc.want {
				t.Fatalf("tokens=%q accepted=%t, want %q accepted=%t", got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestCompactDensityPreservesContinuedStringValue(t *testing.T) {
	source := []byte("package p\n\n" + strings.Repeat("// retained context\n", 68) +
		"func F() string {\n\treturn \"first\" +\n\t\t\"second\" +\n\t\t\"third\" +\n\t\t\"fourth\"\n}\n")
	before := densityStringResult(t, source)
	rewritten, operations, err := compactDensity("string_fixture.go", source)
	if err != nil {
		t.Fatal(err)
	}
	if densityLines(source) <= 75 || densityLines(rewritten) > 75 || operations == 0 {
		t.Fatalf("fixture did not exercise bounded compaction: before=%d after=%d operations=%d",
			densityLines(source), densityLines(rewritten), operations)
	}
	if after := densityStringResult(t, rewritten); before != "firstsecondthirdfourth" || after != before {
		t.Fatalf("compaction changed the typed string value: before=%q after=%q", before, after)
	}
}

func densityStringResult(t *testing.T, source []byte) string {
	t.Helper()
	files := token.NewFileSet()
	parsed, err := parser.ParseFile(files, "string_fixture.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	info := &types.Info{Types: make(map[ast.Expr]types.TypeAndValue)}
	configuration := types.Config{}
	if _, err := configuration.Check("example.test", files, []*ast.File{parsed}, info); err != nil {
		t.Fatalf("compaction produced a non-typechecking string expression: %v", err)
	}
	if len(parsed.Decls) != 1 {
		t.Fatal("string fixture has unexpected declarations")
	}
	function, ok := parsed.Decls[0].(*ast.FuncDecl)
	if !ok || function.Body == nil || len(function.Body.List) != 1 {
		t.Fatal("string fixture no longer has a single return statement")
	}
	statement, ok := function.Body.List[0].(*ast.ReturnStmt)
	if !ok || len(statement.Results) != 1 {
		t.Fatal("string fixture lost its return value")
	}
	value := info.Types[statement.Results[0]].Value
	if value == nil || value.Kind() != constant.String {
		t.Fatal("string fixture no longer returns a typed string constant")
	}
	return constant.StringVal(value)
}
