package receiptprojection

import (
	"bytes"
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

func fixture(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile("../completeness/receipt.gooo")
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestCompileKeepsNumericListsNullsStatesAndStableIDs(t *testing.T) {
	p, err := Compile("receipt.gooo", fixture(t), "CompletenessReceipt")
	if err != nil {
		t.Fatal(err)
	}
	goBytes, err := p.Go()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "receipt.go", goBytes, parser.AllErrors); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Numerator   int", "[]CompletenessDimension", "*UnresolvedCompletenessClaim", "*float64", "map[string]int", `json:"id" gooo:"gooo://completeness/field/dimension/id"`} {
		if !bytes.Contains(goBytes, []byte(want)) {
			t.Fatalf("typed projection lost %s", want)
		}
	}
	schema, err := p.JSONSchema()
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(schema, &doc); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"$schema", "$id", "$ref", "$defs", "x-gooo-declaration-sha256"} {
		if len(doc[key]) == 0 {
			t.Fatalf("schema lost %s", key)
		}
	}
	for _, want := range []string{`"minimum": 0`, `"anyOf"`, `"UNKNOWN"`, `"additionalProperties": false`, `"x-gooo-source-span"`} {
		if !bytes.Contains(schema, []byte(want)) {
			t.Fatalf("JSON schema lost %s", want)
		}
	}
}

func TestInvalidSourceFailsBeforeProjection(t *testing.T) {
	original := string(fixture(t))
	for name, source := range map[string]string{
		"missing-type":        strings.Replace(original, "type Count required one", "type Missing required one", 1),
		"duplicate-field-id":  strings.Replace(original, "gooo://completeness/field/dimension/status", "gooo://completeness/field/dimension/id", 1),
		"unknown-scalar":      strings.Replace(original, "receipt-structure/type/count", "receipt-structure/type/invented", 1),
		"optional-many":       strings.Replace(original, "type Text required many", "type Text optional many", 1),
		"value-recursion":     strings.Replace(original, "type CompletenessDimension required many", "type CompletenessReceipt required one", 1),
		"activity":            original + "\nactivity Execute(CompletenessReceipt) -> CompletenessReceipt\n",
		"duplicate-json-name": strings.Replace(original, "field status id", "field ID id", 1),
		"empty":               "",
		"missing-schema":      strings.Replace(original, "field schema id", "field unrelated id", 1),
		"metadata-name":       strings.Replace(original, "entity CompletenessDimension id", "entity DeclarationSHA256 id", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Compile("bad.gooo", []byte(source), "CompletenessReceipt"); err == nil {
				t.Fatal("invalid source accepted")
			}
		})
	}
	if _, err := Compile("oversized.gooo", bytes.Repeat([]byte("x"), 64<<10+1), "CompletenessReceipt"); err == nil {
		t.Fatal("source byte overflow accepted")
	}
}
