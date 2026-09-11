package selfimprovementvaluewitnessinput

import (
	"os"
	"strings"
	"testing"
)

func TestBuildBindsExactValueWitnessInput(t *testing.T) {
	input, err := Build(os.DirFS("../../.."), SourcePath, ActivityName,
		digestBytes([]byte("candidate")), digestBytes([]byte("candidate-digest")), strings.Repeat("a", 40), digestBytes([]byte("observation")))
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(input); err != nil {
		t.Fatal(err)
	}
	if input.Source.Path != SourcePath || input.Activity.QualifiedName != "valuewitness.Increment" ||
		len(input.Corpus) != 5 || len(input.AllowedEffects) != 0 || input.MaxExecutions != 1 || input.RepositoryWritesAllowed {
		t.Fatalf("input binding drifted: %#v", input)
	}
}

func TestValidateRejectsSnapshotMutation(t *testing.T) {
	input, err := BuildFromBytes(SourcePath, []byte(CanonicalSource), ActivityName,
		digestBytes([]byte("candidate")), digestBytes([]byte("candidate-digest")), strings.Repeat("b", 40), digestBytes([]byte("observation")))
	if err != nil {
		t.Fatal(err)
	}
	input.Source.Bytes = strings.Replace(input.Source.Bytes, "int.add:1", "int.add:2", 1)
	if err := Validate(input); err == nil {
		t.Fatal("mutated source snapshot passed validation")
	}
}
