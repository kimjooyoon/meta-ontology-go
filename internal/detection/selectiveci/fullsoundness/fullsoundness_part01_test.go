package fullsoundness

import (
	"strings"
	"testing"
)

func TestSoundFixture(t *testing.T) {
	input := soundInput()
	got := Evaluate(input)
	if got.Decision != DecisionSound || got.Reason != ReasonSound {
		t.Fatalf("got %s/%s, want SOUND/SOUND", got.Decision, got.Reason)
	}
	if got.CommandCount != 3 || got.SelectedCommandCount != 2 || got.ObligationCount != 2 || got.AuthoritativeImpactedObligationCount != 1 {
		t.Fatalf("semantic counts = %#v", got)
	}
	if got.ResourceVector == nil || got.ResourceVector.Class != ResourceImproved {
		t.Fatalf("resource vector = %#v, want IMPROVED", got.ResourceVector)
	}
	if !got.SemanticEvaluated {
		t.Fatal("sound result did not evaluate semantics")
	}
	if got.ExecutionAuthorized || got.CIAuthorized || !got.ValidDigests() {
		t.Fatalf("output flags or digest invalid: %#v", got)
	}
	t.Logf("direct decision=%s envelope=%s", got.DecisionDigest, got.CanonicalDigest)
}
func TestClosedIDs(t *testing.T) {
	valid := []string{"c1", "o1", "a", "z_9", "command-guard"}
	for _, value := range valid {
		if !validID(value) {
			t.Errorf("validID(%q) = false", value)
		}
	}
	invalid := []string{"", strings.Repeat("a", 65), "C1", "c/1", "c:1", "c 1", "1command"}
	for _, value := range invalid {
		if validID(value) {
			t.Errorf("validID(%q) = true", value)
		}
	}
}
