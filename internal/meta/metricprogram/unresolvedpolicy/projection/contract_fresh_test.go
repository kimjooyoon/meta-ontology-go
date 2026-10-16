package policyprojection

import (
	"bytes"
	"os"
	"testing"
)

func TestDispositionContractBoundaryFresh1869(t *testing.T) {
	source, err := os.ReadFile("../policy.gooo")
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		name, before, after string
	}{
		{"activity-ID", ActivityID, ActivityID + "-other"},
		{"null-is-not-false", "[true,true]", "[null,true]"},
		{"duplicate-case", "[false,false]", "[false,true]"},
		{"single-attempt", `attempts "1"`, `attempts "2"`},
	} {
		t.Run(item.name, func(t *testing.T) {
			changed := bytes.Replace(source, []byte(item.before), []byte(item.after), 1)
			if bytes.Equal(changed, source) {
				t.Fatal("counterexample did not change its target")
			}
			if _, err := sourceActivity("contract1869.gooo", changed); err == nil {
				t.Fatal("ambiguous or incomplete fixed source contract accepted")
			}
		})
	}
}
