package valueexecution

import (
	"bytes"
	"strings"
	"testing"
)

func TestMeasurementEditsSelectedLiteralWithRepeatedAndEscapedText(t *testing.T) {
	for _, literal := range []string{`"int.add:1"`, `"int.add:\u0031"`, "`int.add:1`"} {
		source := []byte(`package measured
namespace measured
entity Integer id "gooo://int.add:1"
activity Other(Integer) -> Integer computes "int.add:1"
activity Selected(Integer) -> Integer computes ` + literal + "\n")
		program, err := Compile("measured.gooo", source, "Selected")
		if err != nil {
			t.Fatal(err)
		}
		baseline, changed := measurementSources("measured.gooo", source, program)
		prefix := string(source[:bytes.LastIndex(source, []byte(literal))])
		if string(baseline) != prefix+`""`+"\n" || string(changed) != prefix+`"int.add:2"`+"\n" {
			t.Fatalf("counterfactual changed unrelated source: %q / %q", baseline, changed)
		}
		if reason := compileReason("measured.gooo", baseline, "Selected"); reason != ReasonProgramMissing {
			t.Fatalf("baseline reason = %s", reason)
		}
		candidate, err := Compile("measured.gooo", changed, "Selected")
		if err != nil || candidate.SemanticFingerprint == program.SemanticFingerprint {
			t.Fatalf("selected program change was not measured: %v", err)
		}
	}
}

func TestMeasurementRejectsStaleAmbiguousAndUnmatchedSources(t *testing.T) {
	source := []byte("package scope\nnamespace scope\nentity Integer\n" +
		`activity Selected(Integer) -> Integer computes "int.add:1"` + "\n")
	for _, text := range []string{string(source), string(source) + string(source), "activity"} {
		program := Program{Activity: "Selected", Text: "int.add:1", SourceDigest: digestBytes([]byte(text))}
		if text == string(source) {
			program.SourceDigest = strings.Repeat("0", 64)
		}
		if baseline, changed := measurementSources("scope.gooo", []byte(text), program); baseline != nil || changed != nil {
			t.Fatalf("invalid source produced counterfactuals: %q / %q", baseline, changed)
		}
	}
	for _, program := range []Program{
		{Activity: "Missing", Text: "int.add:1", SourceDigest: digestBytes(source)},
		{Activity: "Selected", Text: "int.add:2", SourceDigest: digestBytes(source)},
	} {
		if baseline, changed := measurementSources("scope.gooo", source, program); baseline != nil || changed != nil {
			t.Fatal("unmatched activity produced counterfactuals")
		}
	}
}
