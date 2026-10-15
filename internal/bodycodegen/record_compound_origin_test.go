package bodycodegen

import (
	"strconv"
	"strings"
	"testing"
)

func assertRecordCompoundFlow(t *testing.T) {
	t.Helper()
	r, err := InspectTypedValueOrigins("field-update-flow.gooo", []byte(recordCompoundSource), "Update")
	if err != nil || r.Status != "RESOLVED" {
		t.Fatal(err, r.Reason)
	}
	for _, output := range r.Outputs {
		ancestors := compoundAncestors78611(r.Flow.Nodes, output.Root)
		wanted := map[string]bool{"urn:field-update:total": true, "": true}
		if output.FieldID != "urn:field-update:total" {
			wanted = map[string]bool{output.FieldID: true}
		}
		observed := map[string]bool{}
		for _, input := range r.Inputs {
			if ancestors[input.Node] {
				observed[input.FieldID] = true
			}
		}
		if len(observed) != len(wanted) {
			t.Fatal("field ancestry population", output.FieldID, observed)
		}
		for field := range wanted {
			if !observed[field] {
				t.Fatal("missing original input", output.FieldID, field)
			}
		}
	}
}

func TestRecordCompoundRefusesInvalidTargets(t *testing.T) {
	prefix := recordCompoundSource[:strings.Index(recordCompoundSource, " computes ")]
	for _, body := range []string{
		"기록.total += 증가; return 기록",
		"missing.total += 증가; return 기록",
		"let result = 기록; result.missing += 증가; return result",
		"let result = 기록; result.total += true; return result",
		"let result = 기록; result.label -= 1; return result",
		"let result = 기록; result.total %= 3; return result",
		"let result = 기록; result.total /= 0; return result",
		"let result = 기록; result.total[0] += 1; return result",
	} {
		if _, err := Generate("invalid-field-update.gooo", []byte(prefix+" computes "+strconv.Quote(body)+"\n"), "Update"); err == nil {
			t.Fatal("invalid field update accepted", body)
		}
	}
}
