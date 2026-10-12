package sourceimpact

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const sourceFixture = `package changed
namespace changed
entity Integer id "changed://integer"
activity Reference(Integer) -> Integer id "changed://reference" computes "return 31 - input"
activity Fee(Integer) -> Integer id "changed://fee" computes "return input - 31" assembling {
 choice "direction" operand_order at "0" intent "31에서 입력을 뺀다. Subtract input from thirty one."
 case "31" -> "0"
 observe { input "43" oracle "Reference" max_candidates "2" max_rounds "1" resolve_unique "true" }
 holdout_case "59" -> "-28"
 attempts "2"
}
activity Main(Integer) -> Integer id "changed://main" computes "return Fee(input + 5)"
activity Unrelated(Integer) -> Integer id "changed://unrelated" computes "return input * 2"
`

func TestOracleChangeExplainsDependentObligations(t *testing.T) {
	before := []byte(sourceFixture)
	after := []byte(strings.Replace(sourceFixture, "return 31 - input", "return 32 - input", 1))
	r, err := Compare(context.Background(), "before.gooo", before, "after.gooo", after)
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != "COMPLETE_STATIC" || len(r.Changes) != 1 || len(r.Affected) != 3 || r.ModelCalls != 0 || r.CandidateExecutions != 0 {
		t.Fatal(r)
	}
	want := []string{"changed://fee", "changed://main", "changed://reference"}
	for i, row := range r.Affected {
		if row.ID != want[i] || len(row.Causes) != 2 {
			t.Fatal(row)
		}
		for _, cause := range row.Causes {
			if cause.ChangedID != "changed://reference" {
				t.Fatal(cause)
			}
		}
	}
	fee := r.Affected[0]
	if fee.Causes[0].Via.Kind != "observation_oracle" || fee.AfterObligations.Cases != 1 || fee.AfterObligations.Holdouts != 1 || fee.AfterObligations.ObservationInputs != 1 {
		t.Fatal(fee)
	}
	main := r.Affected[1]
	if main.Causes[1].Depth != 2 || main.Causes[1].Via.Dependency != fee.ID {
		t.Fatal(main)
	}
	again, err := Compare(context.Background(), "before.gooo", before, "after.gooo", after)
	if err != nil || !reflect.DeepEqual(r, again) {
		t.Fatal("non-deterministic impact", err)
	}
	digest := r.ReportSHA256
	r.ReportSHA256 = ""
	raw, _ := json.Marshal(r)
	if semantic.StableHash(raw) != digest {
		t.Fatal("unbound report")
	}
}

func TestCommentsSpansAndDeclarationOrderDoNotBecomeChanges(t *testing.T) {
	lines := strings.Split(sourceFixture, "\n")
	// Move the unrelated declaration before Reference; stable IDs own identity.
	unrelated := lines[len(lines)-2]
	lines = append(lines[:len(lines)-2], lines[len(lines)-1:]...)
	after := "// edited description\n" + strings.Join(append(append([]string{}, lines[:3]...), append([]string{unrelated}, lines[3:]...)...), "\n")
	r, err := Compare(context.Background(), "old.gooo", []byte(sourceFixture), "new.gooo", []byte(after))
	if err != nil || len(r.Changes) != 0 || len(r.Affected) != 0 || r.BeforeSourceSHA256 == r.AfterSourceSHA256 {
		t.Fatal(r, err)
	}
}

func TestUnresolvedCallsRemainVisible(t *testing.T) {
	for _, body := range []string{"return Missing(input)", "return external.Fee(input)", "let Fee = input; return Fee(input)", "return ("} {
		source := strings.Replace(sourceFixture, "return Fee(input + 5)", body, 1)
		r, err := Compare(context.Background(), "a.gooo", []byte(source), "b.gooo", []byte(source))
		if err != nil || r.Status != "PARTIAL_STATIC" || len(r.Unknowns) != 2 || len(r.Changes) != 0 {
			t.Fatalf("body %q: %+v %v", body, r, err)
		}
	}
}

func TestHeldAnswerChangeRetainsOwnDenominator(t *testing.T) {
	after := strings.Replace(sourceFixture, `"59" -> "-28"`, `"59" -> "28"`, 1)
	r, err := Compare(context.Background(), "a.gooo", []byte(sourceFixture), "b.gooo", []byte(after))
	if err != nil || len(r.Changes) != 1 || r.Changes[0].ID != "changed://fee" || len(r.Affected) != 2 {
		t.Fatal(r, err)
	}
	if !reflect.DeepEqual(r.Changes[0].Facets, []string{"assembly"}) {
		t.Fatal(r.Changes)
	}
}
