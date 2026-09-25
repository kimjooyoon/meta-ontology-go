package predecessorbinding

import "testing"

func TestCompareProvesDynamicBindingCutover(t *testing.T) {
	before := Evaluate(testHead, observations(StateStaticLiteral), 0)
	after := Evaluate("2222222222222222222222222222222222222222",
		observations(StateDynamicInput), 0)
	transition, err := Compare(before, after)
	if err != nil {
		t.Fatal(err)
	}
	if transition.Decision != "IMPROVED" || transition.StaticDelta != -8 ||
		transition.DynamicDelta != 8 || transition.BPSDelta != 10_000 {
		t.Fatalf("dynamic cutover not proven: %+v", transition)
	}
}
