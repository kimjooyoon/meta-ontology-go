package bodycodegen

import (
	"strings"
	"testing"
)

func TestZeroInputRootEquivalenceRetainsArity(t *testing.T) {
	source := []byte(nullaryCallsSource)
	generated, err := Generate("zero.gooo", source, "Seed")
	if err != nil {
		t.Fatal(err)
	}
	p, err := prepareActivityBodySource("zero.gooo", source, "Seed")
	if err != nil || len(p.parameters) != 0 {
		t.Fatal("constant root acquired parameters", err)
	}
	changed := strings.Replace(generated.Source, "func Seed()", "func Seed(unused int64)", 1)
	if changed == generated.Source {
		t.Fatal("signature counterexample changed no source")
	}
	proof, err := routeEquivalenceParameters(p.packageName, p.activity.Name, p.parameters, p.outputType, p.body,
		[]byte(changed), preserveRoute)
	if err == nil && proof.Equivalent {
		t.Fatal("a new caller input survived zero-input source equivalence")
	}
}
