package languagetestexperiment

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/languagetest"
)

func TestEvaluatePassesFixedContractAndLowersUnknown(t *testing.T) {
	positive := languagetest.Observe(languagetest.Request{Filename: "main.gooo", Source: positiveFixture})
	failure := languagetest.Observe(languagetest.Request{Filename: "wrong.gooo", Source: failureFixture})
	missing := languagetest.Observe(languagetest.Request{Filename: "missing.gooo", Source: missingFixture})
	input := Input{
		SubjectSHA: "subject", ExecutableDigest: "sha256:binary", Contract: fixedContract(),
		First: Observation{Runtime: "go1.27.0", Receipt: positive},
		Replay: Observation{Runtime: "go1.27.0", Receipt: positive},
		AssertionFailure: failure, Missing: missing,
	}
	report, err := Evaluate(input)
	if err != nil || report.Decision != "PASS" || report.Summary.Coordinates.Satisfied != 12 {
		t.Fatalf("report=%+v err=%v", report, err)
	}
	input.First.Receipt.Decision = "UNKNOWN"
	unknown, err := Evaluate(input)
	if err != nil || unknown.Decision != "FAIL_CLOSED" || unknown.Resolution != "LOWER_RESOLUTION" || unknown.Summary.Coordinates.Satisfied != 0 {
		t.Fatalf("unknown=%+v err=%v", unknown, err)
	}
}

func fixedContract() Contract {
	return Contract{ContractSchema, "fixed", 2, 2, 2, 2, 2, 1, 1, 2, 1, 1, 3, 0, false}
}

const positiveFixture = `package p
namespace p
entity Input id "p://input"
entity Output id "p://output"
entity TestBuild id "gooo://test/activity/Build/output/Output"
activity Build(Input) -> Output
`

const failureFixture = `package p
namespace p
entity Input id "p://input"
entity Output id "p://output"
entity Other id "p://other"
entity TestBuild id "gooo://test/activity/Build/output/Other"
activity Build(Input) -> Output
`

const missingFixture = `package p
namespace p
entity Output id "p://output"
activity Build() -> Output
`
