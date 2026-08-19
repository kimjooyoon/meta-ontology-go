package shadow

import (
	"reflect"
	"strings"
	"testing"
)

func TestExpectedLabelsDoNotAffectOracle(t *testing.T) {
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	fixture := corpus.Cases[0]
	want := Evaluate(fixture)
	mutations := []struct {
		name   string
		mutate func(*Result)
	}{
		{"status", func(result *Result) { result.Status = "changed-status" }},
		{"stage", func(result *Result) { result.Stage = "changed-stage" }},
		{"reason", func(result *Result) { result.Reason = "changed-reason" }},
		{"selected command IDs", func(result *Result) { result.SelectedCommandIDs = []string{"changed-command"} }},
		{"selected guard IDs", func(result *Result) { result.SelectedGuardIDs = []string{"changed-guard"} }},
		{"selected work IDs", func(result *Result) { result.SelectedWorkIDs = []string{"changed-work"} }},
		{"selected argv", func(result *Result) { result.SelectedArgv = map[string][]string{"changed-command": {"changed-argv"}} }},
		{"execution authorization", func(result *Result) { result.ExecutionAuthorized = true }},
		{"canonical digest", func(result *Result) { result.CanonicalDigest = "changed-digest" }},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			mutated := fixture
			mutated.Expected = fixture.Expected
			mutation.mutate(&mutated.Expected)
			if got := Evaluate(mutated); !reflect.DeepEqual(got, want) {
				t.Fatalf("oracle changed after expected-label mutation: got %#v want %#v", got, want)
			}
		})
	}
}
func TestGovernedInputMutationChangesOracleDigest(t *testing.T) {
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	fixture := corpus.Cases[0]
	baseline := Evaluate(fixture)
	mutated := fixture
	mutated.Files.AnalyzerHead = strings.Replace(mutated.Files.AnalyzerHead, "blob-head", "blob-head-mutated", 1)
	if mutated.Files.AnalyzerHead == fixture.Files.AnalyzerHead {
		t.Fatal("governed-input mutation did not change the input")
	}
	if got := Evaluate(mutated); got.CanonicalDigest == baseline.CanonicalDigest {
		t.Fatalf("governed-input mutation preserved oracle digest %q", got.CanonicalDigest)
	}
}
func TestCorrectionRecordBindsSupersededAndCorrectedEvidence(t *testing.T) {
	record, err := LoadCorrection()
	if err != nil {
		t.Fatal(err)
	}
	if record.ReasonCode != "EXECUTION_AUTHORITY_CONTRADICTION" {
		t.Fatalf("correction reason = %q", record.ReasonCode)
	}
	if record.Superseded.CorpusDigest != "e79ba3696eec2bb67c915398a1f652f523b9b98a3227bee2e4e2c4b9f2f8120e" || record.Superseded.ExpectedVectorDigest != "a9661672d1dcf30df297b8aae90d2b7138ef7126dccf9f45b495a5399dd82c58" {
		t.Fatalf("superseded correction evidence = %#v", record.Superseded)
	}
	if record.Corrected.CorpusDigest != "36359077392431f4e4136baeb022b78f87fdf7c69a0dbab18ca38e3e92ae6954" || record.Corrected.ExpectedVectorDigest != "fe260bba00c58fb3ab761910c253905dfb749be60ed655b03810bebddc2b3ef5" {
		t.Fatalf("corrected correction evidence = %#v", record.Corrected)
	}
}
