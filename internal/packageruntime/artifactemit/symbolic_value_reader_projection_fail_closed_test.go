package artifactemit

import (
	"strings"
	"testing"
)

func TestReaderProjectionRejectsUnknownDecision(t *testing.T) {
	payload := symbolicReaderMutate(func(source *SymbolicValueReachability) {
		source.Decision = "UNKNOWN"
	})
	assertSymbolicReaderFailClosed(t, payload)
}

func TestReaderProjectionRejectsUnknownBranches(t *testing.T) {
	payload := symbolicReaderMutate(func(source *SymbolicValueReachability) {
		source.Summary.UnknownPolicyBranches = 1
	})
	assertSymbolicReaderFailClosed(t, payload)
}

func TestReaderProjectionRejectsUnknownReaderResolution(t *testing.T) {
	payload := symbolicReaderMutate(func(source *SymbolicValueReachability) {
		source.Views[0].Resolution = "UNKNOWN"
	})
	assertSymbolicReaderFailClosed(t, payload)
}

func assertSymbolicReaderFailClosed(t *testing.T, payload []byte) {
	t.Helper()
	result, err := CompileSymbolicValueReaderProjection(payload, strings.Repeat("a", 40))
	if err != nil {
		t.Fatal(err)
	}
	if result.Decision != "FAIL_CLOSED" || result.Resolution != "INVARIANT_ONLY" {
		t.Fatalf("decision=%s resolution=%s", result.Decision, result.Resolution)
	}
	for _, reader := range result.Readers {
		if reader.EffectiveResolution != "INVARIANT_ONLY" {
			t.Fatalf("%s resolution=%s", reader.Audience, reader.EffectiveResolution)
		}
	}
}
