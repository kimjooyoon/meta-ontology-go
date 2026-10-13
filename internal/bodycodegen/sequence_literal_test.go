package bodycodegen

import (
	"context"
	"strings"
	"testing"
)

func TestLiteralProfile9143(t *testing.T) {
	for name, body := range map[string]string{
		"bounded":     "return []int64{" + strings.Repeat("317,", 257) + "}[0]",
		"keyed":       "return []int64{4: 331}[0]",
		"fixed_array": "return [2]int64{337,347}[0]",
		"nested":      "return [][]int64{{349}}[0][0]",
		"float":       "return int64([]float64{3.5}[0])",
		"mutation":    "let values = []int64{353,359}; values[0] = 367; return values[0]",
	} {
		t.Run(name, func(t *testing.T) {
			source := []byte("package literalprofile\nnamespace literal9143\nentity Integer id \"urn:gooo:type:integer\"\nactivity Probe(Integer) -> Integer id \"urn:literal9143:" + name + "\" computes \"" + body + "\"\n")
			if _, err := GenerateWithPlanner(context.Background(), "literal9143.gooo", source, "Probe", "", ""); err == nil {
				t.Fatal("accepted unsupported literal profile", name)
			}
		})
	}
	t.Log("literal9143: bound, keyed/fixed/nested/non-scalar layouts and element mutation rejected without native execution")
}
