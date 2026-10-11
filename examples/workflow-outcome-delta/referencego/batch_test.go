package referencego

import (
	"encoding/json"
	"os"
	"testing"
)

type batchCases struct {
	Schema string `json:"schema"`
	Cases  []struct {
		Inputs   map[string]int64 `json:"inputs"`
		Expected map[string]int64 `json:"expected"`
	} `json:"cases"`
}

func readBatchCases(t *testing.T, filename string) batchCases {
	t.Helper()
	raw, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var cases batchCases
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	if cases.Schema != "gooo/body-composition-cases/v1" || len(cases.Cases) != 10 {
		t.Fatal("expected the ten original caller cases")
	}
	return cases
}

func TestBatchRulesUseOriginalGoooCallerCases(t *testing.T) {
	for _, rule := range []struct {
		filename string
		apply    func(int64) int64
	}{
		{"../evaluation16.json", Batch16},
		{"../evaluation32.json", Batch32},
	} {
		t.Run(rule.filename, func(t *testing.T) {
			for i, row := range readBatchCases(t, rule.filename).Cases {
				input, hasInput := row.Inputs["Main"]
				expected, hasExpected := row.Expected["Main"]
				if !hasInput || !hasExpected || len(row.Inputs) != 1 || len(row.Expected) != 1 {
					t.Fatalf("case %d: expected one Main input and output", i)
				}
				if actual := rule.apply(input); actual != expected {
					t.Errorf("case %d input %d: got %d, want %d", i, input, actual, expected)
				}
			}
		})
	}
}
