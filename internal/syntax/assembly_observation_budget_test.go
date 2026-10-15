package syntax

import (
	"fmt"
	"testing"
)

func observationBudgetSource(candidates, rounds string) string {
	return fmt.Sprintf(`package bounds
namespace bounds
entity Integer id "bounds://integer"
entity Box id "bounds://box" fields { field value id "bounds://value" type integer required one }
activity Select(Integer) -> Box computes "return Box{value:input}" assembling {
 choice "value" field_value at "0" alternative "0-input" intent "Negate"
 value_case "[0]" -> "{\"value\":0}"
 observe { value_input "[7]" max_candidates "%s" max_rounds "%s" }
 attempts "2"
}`, candidates, rounds)
}

func TestObservationBudgetConversionsHaveExplicitLanguageBounds(t *testing.T) {
	for _, tc := range []struct {
		name, candidates, rounds   string
		wantCandidates, wantRounds int
	}{
		{"minimum", "1", "1", 1, 1},
		{"maximum", "64", "8", 64, 8},
		{"negative", "-1", "1", 0, 0},
		{"candidate zero", "0", "1", 0, 0},
		{"candidate above", "65", "1", 0, 0},
		{"candidate int32 overflow", "2147483648", "1", 0, 0},
		{"candidate int64 maximum", "9223372036854775807", "1", 0, 0},
		{"round zero", "2", "0", 0, 0},
		{"round above", "64", "9", 0, 0},
		{"round int64 maximum", "2", "9223372036854775807", 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file, diagnostics := ParseFileWithEntityFieldsSupport("bounds.gooo", observationBudgetSource(tc.candidates, tc.rounds), EntityFieldsV1Support())
			if tc.wantCandidates == 0 {
				if !diagnostics.HasErrors() {
					t.Fatal("out-of-range observation budget accepted")
				}
				return
			}
			if diagnostics.HasErrors() {
				t.Fatal(diagnostics)
			}
			o := file.Declarations[2].(*ActivityDecl).Assembly.Spec.Observation
			if o.MaxCandidates != tc.wantCandidates || o.MaxRounds != tc.wantRounds {
				t.Fatal("language bounds changed after integer conversion", o)
			}
		})
	}
}
