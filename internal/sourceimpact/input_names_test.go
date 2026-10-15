package sourceimpact

import (
	"context"
	"slices"
	"strings"
	"testing"
)

func TestAuthoredInputNameImpactFirst(t *testing.T) {
	const source = `package nameimpact
namespace nameimpact
entity Integer id "urn:gooo:type:integer"
activity Difference(left Integer, right Integer) -> Integer id "urn:nameimpact:difference" computes "return left - right"
activity Main(value Integer) -> Integer id "urn:nameimpact:main" computes "return Difference(value, 19)"
activity Other(value Integer) -> Integer id "urn:nameimpact:other" computes "return value"
`
	after := strings.ReplaceAll(source, "left", "minuend")
	r, err := Compare(context.Background(), "before.gooo", []byte(source), "after.gooo", []byte(after))
	if err != nil || r.Status != "COMPLETE_STATIC" || len(r.Changes) != 1 || len(r.Affected) != 2 ||
		!slices.Contains(r.Changes[0].Facets, "input_names") || r.Changes[0].ID != "urn:nameimpact:difference" {
		t.Fatal(r, err)
	}
	shadowed := strings.Replace(source, "Main(value Integer)", "Main(Difference Integer)", 1)
	shadowed = strings.Replace(shadowed, "Difference(value, 19)", "Difference(1, 19)", 1)
	r, err = Compare(context.Background(), "old.gooo", []byte(source), "new.gooo", []byte(shadowed))
	if err != nil || r.Status != "PARTIAL_STATIC" || len(r.Unknowns) != 1 {
		t.Fatal("input call claimed as an activity dependency", r, err)
	}
}
