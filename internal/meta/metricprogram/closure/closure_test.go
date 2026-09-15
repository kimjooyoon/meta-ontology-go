package closure_test

import (
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/closure"
	closureverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/closure/verify"
)

func TestBuildAndIndependentlyVerify(t *testing.T) {
	input := fixtureInput()
	receipt, err := closure.Build(input)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Schema != closure.Schema || len(receipt.Indicators) != 4 {
		t.Fatalf("unexpected receipt: %#v", receipt)
	}
	raw, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	report, err := closureverify.Verify(verifierFixture(input), raw)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != closure.StatusVerified || report.ArtifactID != input.Artifact.ID {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func verifierFixture(input closure.Input) closureverify.Input {
	return closureverify.Input{
		Repository: input.Repository, SubjectSHA: input.SubjectSHA,
		RunID: input.RunID, RunAttempt: input.RunAttempt,
		Artifact: closureverify.ArtifactIdentity{
			ID: input.Artifact.ID, Name: input.Artifact.Name,
			Digest: input.Artifact.Digest, URL: input.Artifact.URL,
		},
		ProgramJSON: input.ProgramJSON, Source: input.Source,
		VerificationJSON: input.VerificationJSON,
	}
}
