package protectedregions

import (
	"reflect"
	"strings"
	"testing"
)

// TestContractGeneratedBodyOnlyChangesAreLocal is the prototype's falsifiable
// hypothesis: a refresh that changes only generated bodies passes, while the
// same refresh crossing a protected or unmarked boundary fails.
func TestContractGeneratedBodyOnlyChangesAreLocal(t *testing.T) {
	before := readFixture(t, "before.go")
	cases := []struct {
		name       string
		after      string
		wantValid  bool
		wantIssue  LocalityIssueKind
		wantStruct IssueKind
	}{
		{
			name:      "generated-body-only",
			after:     strings.Replace(string(before), "func Activity() int {", "func Activity() int {\n\t// regenerated", 1),
			wantValid: true,
		},
		{
			name:      "slot-body-overwrite",
			after:     strings.Replace(string(before), "return 7", "return 8", 1),
			wantIssue: LocalityProtectedChange,
		},
		{
			name:      "unmarked-handwritten-overwrite",
			after:     strings.Replace(string(before), "var Keep = 7", "var Keep = 8", 1),
			wantIssue: LocalityUnownedChange,
		},
		{
			name:       "missing-slot-end",
			after:      strings.Replace(string(before), "\t//gooo:slot:end id=\"fixture/activity/implementation\"\n", "", 1),
			wantStruct: IssueMissingEnd,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			report := ValidateLocality(before, []byte(testCase.after))
			if report.Valid() != testCase.wantValid {
				t.Fatalf("valid=%v, want %v: %v", report.Valid(), testCase.wantValid, report.Err())
			}
			if testCase.wantIssue != "" && !hasLocalityIssue(report.Violations, testCase.wantIssue) {
				t.Fatalf("missing locality issue %q: %#v", testCase.wantIssue, report.Violations)
			}
			if testCase.wantStruct != "" && !hasIssue(report.After.Issues, testCase.wantStruct) {
				t.Fatalf("missing structural issue %q: %#v", testCase.wantStruct, report.After.Issues)
			}
		})
	}
}

func TestContractOutputIsDeterministic(t *testing.T) {
	before := readFixture(t, "before.go")
	after := readFixture(t, "after.go")
	want := ValidateLocality(before, after)
	for attempt := 0; attempt < 20; attempt++ {
		got := ValidateLocality(before, after)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("attempt %d changed output:\nwant=%#v\ngot=%#v", attempt, want, got)
		}
	}
}
