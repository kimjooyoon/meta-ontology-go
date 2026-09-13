package extractor

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const returnTailValueRuntimeTypes = "package main\n\ntype Leaf struct { Label string }\ntype Report struct { Label string; Nested Leaf; Values []string }\n"

func TestReturnTailReturnedValueFieldExtraction(t *testing.T) {
	cases := []struct {
		name    string
		padding int
		docs    string
	}{
		{"rendered-only-overflow", 65, strings.Repeat("// value copy boundary\n", 6)},
		{"function-overflow", 80, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := returnTailValueRuntimeSource("F", tc.padding, tc.docs, "\tr.Label = \"new\"\n")
			beforeFunction, beforeRendered, err := runtimeWitnessInputLines(source, "F")
			if err != nil || beforeRendered <= runtimeWitnessLineLimit {
				t.Fatalf("missing rendered overflow: function=%d rendered=%d error=%v", beforeFunction, beforeRendered, err)
			}
			if tc.name == "rendered-only-overflow" && beforeFunction > runtimeWitnessLineLimit {
				t.Fatalf("body-only boundary was not preserved: function=%d", beforeFunction)
			}
			root := t.TempDir()
			if err := runtimeWitnessWriteModule(root, source, map[string]string{"types.go": returnTailValueRuntimeTypes}); err != nil {
				t.Fatal(err)
			}
			result, err := ExtractWithResult(root, "x.go")
			if err != nil {
				t.Fatal(err)
			}
			assertReturnTailTupleProofChain(t, result)
			replay, err := ExtractWithResult(root, "x.go")
			if err != nil || !reflect.DeepEqual(result.Generated, replay.Generated) {
				t.Fatalf("value-field replay differs: %v", err)
			}
			unchanged, err := os.ReadFile(filepath.Join(root, "x.go"))
			if err != nil || string(unchanged) != source {
				t.Fatalf("original source changed: %v", err)
			}
			t.Logf("VALUE_FIELD_EXTRACTION=%s", runtimeWitnessJSON(map[string]any{
				"case": tc.name, "before_function_lines": beforeFunction, "before_rendered_lines": beforeRendered,
				"generated_files": len(result.Generated), "replay_equal": true, "source_unchanged": true,
				"strategies": result.Evidence,
			}))
		})
	}
}

func TestReturnTailReturnedValueFieldRuntimeWitness(t *testing.T) {
	cases := []runtimeWitnessCase{
		{name: "VF1_returned_record_field", functionName: "VF1",
			source:   returnTailValueRuntimeSource("VF1", 80, "", "\tr.Label = \"changed\"\n"),
			support:  map[string]string{"types.go": returnTailValueRuntimeTypes, "main.go": returnTailValueRuntimeHarness("VF1")},
			expected: "original:changed:old:old:true\n"},
		{name: "VF2_nested_value_and_slice_header", functionName: "VF2",
			source:   returnTailValueRuntimeSource("VF2", 80, "", "\tr.Nested.Label = \"changed\"\n\tr.Values = []string{\"new\"}\n"),
			support:  map[string]string{"types.go": returnTailValueRuntimeTypes, "main.go": returnTailValueRuntimeHarness("VF2")},
			expected: "original:changed:old:new:true\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runReturnTailRuntimeWitness(t, tc)
		})
	}
}

func returnTailValueRuntimeSource(name string, padding int, docs, tail string) string {
	return "package main\n\n" + docs + "func " + name + "(r Report) (Report, error) {\n" +
		strings.Repeat("\t_ = 1\n", padding) + tail + "\treturn r, nil\n}\n"
}

func returnTailValueRuntimeHarness(name string) string {
	observed := "after.Label"
	if name == "VF2" {
		observed = "after.Nested.Label"
	}
	return "package main\n\nimport \"fmt\"\n\nfunc main() {\n" +
		"\tinput := Report{Label: \"original\", Nested: Leaf{Label: \"original\"}, Values: []string{\"old\"}}\n" +
		"\tafter, err := " + name + "(input)\n" +
		"\tfmt.Printf(\"%s:%s:%s:%s:%t\\n\", input.Label, " + observed + ", input.Values[0], after.Values[0], err == nil)\n}\n"
}
