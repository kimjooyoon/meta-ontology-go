package bodycodegen

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func singleHoleByteFixture(t *testing.T) ([]byte, *syntax.ActivityDecl) {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/source-ir-fill-single-byte.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := ParseBodyFile("single-byte-fill.gooo", source)
	if file == nil || diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if ok && activity.Name == "IndexByte" {
			checkSingleHoleByteOracle(t, activity)
			return source, activity
		}
	}
	t.Fatal("single-hole byte search activity missing")
	return nil, nil
}

func checkSingleHoleByteOracle(t *testing.T, activity *syntax.ActivityDecl) {
	t.Helper()
	spec := &activity.Assembly.Spec
	if len(spec.ValueCases) != 6 || len(spec.ValueHoldoutCases) != 4 {
		t.Fatal("single-hole fixture case counts changed")
	}
	for _, row := range append(spec.ValueCases, spec.ValueHoldoutCases...) {
		var inputs []struct {
			Text   string `json:"text"`
			Needle int    `json:"needle"`
		}
		var expected map[string]int
		if err := json.Unmarshal([]byte(row.Inputs), &inputs); err != nil || len(inputs) != 1 {
			t.Fatalf("decode single-hole inputs: %v", err)
		}
		if err := json.Unmarshal([]byte(row.Expected), &expected); err != nil || len(expected) != 1 {
			t.Fatalf("decode single-hole expected output: %v", err)
		}
		want := -1
		if inputs[0].Needle >= 0 && inputs[0].Needle <= 255 {
			want = strings.IndexByte(inputs[0].Text, byte(inputs[0].Needle))
		}
		if offset, present := expected["offset"]; !present || offset != want {
			t.Fatal("stored byte offset differs from Go reference", expected, want)
		}
	}
}
