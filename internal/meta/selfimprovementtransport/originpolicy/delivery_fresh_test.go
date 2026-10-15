package originpolicy

import (
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestDeliverySourceProjectionFresh1613(t *testing.T) {
	for _, item := range []struct{ activity, file string }{
		{"CandidateArtifactRole", "role.generated.go"}, {"BindCandidateSubject", "subject.generated.go"},
		{"ContinuationEligible", "eligibility.generated.go"}, {"ContinuationSubject", "continuation_subject.generated.go"},
		{"ContinuationArtifactName", "continuation_name.generated.go"},
	} {
		t.Run(item.activity, func(t *testing.T) {
			generated, err := bodycodegen.Generate("source.gooo", source, item.activity)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(item.file)
			if err != nil || string(raw) != generated.Source {
				t.Fatalf("source projection differs: %v", err)
			}
		})
	}
	changed := strings.Replace(string(source), `outcome == "AUTHORIZED"`, `outcome == "REJECTED"`, 1)
	before, _ := bodycodegen.Generate("source.gooo", source, "ContinuationEligible")
	after, err := bodycodegen.Generate("source.gooo", []byte(changed), "ContinuationEligible")
	if err != nil || before.Source == after.Source || before.Report.SourceDigest == after.Report.SourceDigest {
		t.Fatal("Gooo intent did not change projection", err)
	}
}

func TestDeliverySourceDecisionsFresh1613(t *testing.T) {
	name := "self-improvement-nonexecuting-candidate-" + strings.Repeat("f", 40)
	if CandidateArtifactRole(name) == "" || CandidateArtifactRole("unrelated") != "" {
		t.Fatal("role selection")
	}
	if BindCandidateSubject("s", "s", "s") != "CANDIDATE_SUBJECT_BOUND" ||
		BindCandidateSubject("s", "", "s") != "CANDIDATE_SUBJECT_UNKNOWN" ||
		BindCandidateSubject("s", "t", "s") != "CANDIDATE_SUBJECT_MISMATCH" {
		t.Fatal("subject relation")
	}
	for _, state := range []string{"UNKNOWN", "REFUTED", "CLOSED"} {
		for _, verified := range []bool{false, true} {
			if ContinuationEligible(state, "AUTHORIZED", verified) != (state == "CLOSED" && verified) {
				t.Fatal("continuation decision")
			}
		}
	}
	if ContinuationEligible("CLOSED", "REJECTED", true) {
		t.Fatal("rejected candidate continued")
	}
}
