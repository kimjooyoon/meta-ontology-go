package closure_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/closure"
	closureverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/closure/verify"
)

func TestVerifierRejectsArtifactSubstitution(t *testing.T) {
	input := fixtureInput()
	receipt, err := closure.Build(input)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(receipt)
	verifyInput := verifierFixture(input)
	verifyInput.Artifact.Digest = strings.Repeat("7", 64)
	if _, err := closureverify.Verify(verifyInput, raw); err == nil {
		t.Fatal("expected artifact substitution rejection")
	}
}
