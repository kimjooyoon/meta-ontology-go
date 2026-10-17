package valuejson_test

import (
	"bytes"
	"go/scanner"
	"go/token"
	"os"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestSharedEscapeProjectionSourceFresh1905(t *testing.T) {
	source, err := os.ReadFile("../meta/textinput/escapes.gooo")
	if err != nil {
		t.Fatal(err)
	}
	projected, err := bodycodegen.Generate("escapes.gooo", source, "TextEscapePolicy")
	if err != nil || !projected.Report.TypecheckPassed || projected.Report.ActivityID != "urn:gooo:text-input:escape-policy" {
		t.Fatal("current Gooo source did not project", err, projected.Report)
	}
	retained, err := os.ReadFile("escape_policy_generated.go")
	if err != nil {
		t.Fatal(err)
	}
	// The Go host package is a transport choice outside the generated region.
	// Compare function tokens independently of layout and package spelling.
	retained = bytes.Replace(retained, []byte("package valuejson\n"), []byte("package bodyexecution\n"), 1)
	if !reflect.DeepEqual(projectionTokens1905(t, retained), projectionTokens1905(t, []byte(projected.Source))) {
		t.Fatal("retained Gooo projection differs from current source generation")
	}
}

func projectionTokens1905(t *testing.T, raw []byte) []string {
	t.Helper()
	var reader scanner.Scanner
	fset := token.NewFileSet()
	reader.Init(fset.AddFile("projection.go", -1, len(raw)), raw, func(_ token.Position, message string) {
		t.Error(message)
	}, 0)
	var tokens []string
	for {
		_, kind, spelling := reader.Scan()
		if kind == token.EOF {
			return tokens
		}
		tokens = append(tokens, kind.String()+":"+spelling)
	}
}
