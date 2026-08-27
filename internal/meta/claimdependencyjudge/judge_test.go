package claimdependencyjudge

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/claimdependency"
)

func fixtureSource(marker string) []byte {
	return []byte("package claimdependency\nnamespace claimdependency\nentity Integer id \"gooo://claim-dependency/entity/integer\"\nactivity Root(Integer) -> Integer computes \"" + marker + "\"\nactivity Derived(Integer) -> Integer computes \"int.add:2\"\n")
}

func TestIndependentJudgeAcceptsAllThreeStates(t *testing.T) {
	cases := []struct {
		name   string
		marker string
	}{
		{claimdependency.CaseDirectUnknown, "int.unknown:1"},
		{claimdependency.CaseRefuted, "int.add:-1"},
		{claimdependency.CaseRecovered, "int.add:1"},
	}
	for _, testCase := range cases {
		receipt, err := claimdependency.Evaluate(fixtureSource(testCase.marker), "fixture.gooo", testCase.name)
		if err != nil {
			t.Fatal(err)
		}
		judgment, err := Judge(receipt)
		if err != nil {
			t.Fatalf("%s: %v", testCase.name, err)
		}
		if !judgment.Accepted || !judgment.ReadOnly || judgment.SemanticPromotionAuthorized {
			t.Fatalf("%s: unexpected judgment %+v", testCase.name, judgment)
		}
	}
}

func TestIndependentJudgeRejectsResealedDecision(t *testing.T) {
	receipt, err := claimdependency.Evaluate(fixtureSource("int.unknown:1"), "fixture.gooo", claimdependency.CaseDirectUnknown)
	if err != nil {
		t.Fatal(err)
	}
	receipt.Decision.Value = "PASS"
	receipt.Digest = digestReceipt(receipt)
	if _, err := Judge(receipt); err == nil {
		t.Fatal("resealed contradictory decision was accepted")
	}
}

func TestIndependentJudgeRejectsNonMinimumPath(t *testing.T) {
	receipt, err := claimdependency.Evaluate(fixtureSource("int.unknown:1"), "fixture.gooo", claimdependency.CaseDirectUnknown)
	if err != nil {
		t.Fatal(err)
	}
	receipt.Resolutions[5].CausePath = []string{expectedClaims[0].ClaimID, expectedClaims[1].ClaimID, expectedClaims[2].ClaimID, expectedClaims[5].ClaimID}
	receipt.Digest = digestReceipt(receipt)
	if _, err := Judge(receipt); err == nil {
		t.Fatal("non-minimum cause path was accepted")
	}
}
