package bodycodegen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestElseSurfaceRetainsOriginalExpressionAndLiteralRanges(t *testing.T) {
	body := "let 문구 = `원문 }\nelse`\nlet total = 0\n" +
		"for let index = 0; index < 2; index = index + 1 {\n" +
		"if input && 문구 != \"\" { total = total + 17 }\nelse { total = total + 19 }\n}\n" +
		"if !input { return total + 23 }\nelse { return total + 29 }"
	normalized, mapping, err := NormalizeBodySource(body)
	if err != nil || !strings.Contains(normalized, "`원문 }\nelse`") {
		t.Fatal("authored literal changed", err)
	}
	header := "package p\nfunc Pick(input bool) int64 {\n"
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, "projection.go", header+normalized+"\n}", 0)
	if err != nil {
		t.Fatal("joined projection did not parse", err)
	}
	conditions, literals := 0, 0
	ast.Inspect(file, func(node ast.Node) bool {
		var expression ast.Node
		switch n := node.(type) {
		case *ast.IfStmt:
			expression = n.Cond
			conditions++
		case *ast.BasicLit:
			if n.Kind == token.STRING {
				expression = n
				literals++
			}
		}
		if expression == nil {
			return true
		}
		start, end := set.Position(expression.Pos()).Offset-len(header), set.Position(expression.End()).Offset-len(header)
		original := mapping.OriginalRange(start, end)
		if original.Start < 0 || original.End > len(body) ||
			body[original.Start:original.End] != normalized[start:end] {
			t.Fatal("expression lost its original bytes", original)
		}
		if roundtrip := mapping.NormalizedRange(original.Start, original.End); roundtrip.Start != start || roundtrip.End != end {
			t.Fatal("source range did not round-trip", original, roundtrip)
		}
		return true
	})
	if conditions != 2 || literals != 2 {
		t.Fatal("new source probe did not inspect the intended expressions", conditions, literals)
	}
}

func TestElseSurfaceKeepsCommentsAndAuthoredSeparators(t *testing.T) {
	for _, gap := range []string{" // original note\n", "\n/* original note */\n", ";\n"} {
		body := "if input { return 233 }" + gap + "else { return 239 }"
		actual, _, err := NormalizeBodySource(body)
		if err != nil || actual != body {
			t.Fatal("comment or separator was reinterpreted", gap, err)
		}
		if _, err := parser.ParseFile(token.NewFileSet(), "rejected.go",
			"package p\nfunc Pick(input bool) int64 {\n"+actual+"\n}", 0); err == nil {
			t.Fatal("unsupported layout was silently accepted", gap)
		}
	}
}
