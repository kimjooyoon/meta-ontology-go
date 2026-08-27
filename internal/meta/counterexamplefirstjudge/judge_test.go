package counterexamplefirstjudge

import (
	"testing"

	cf "github.com/kimjooyoon/meta-ontology-go/internal/meta/counterexamplefirst"
)

func TestIndependentJudgeRejectsMalformedContract(t *testing.T) {
	report := Evaluate(cf.JudgeInput{Contract: cf.Contract{}, HeadSHA: "head", Corpus: cf.ScenarioCorpus{Schema: cf.CorpusSchema, Version: 1}, Receipts: nil})
	if report.Decision != "FAIL_CLOSED" || report.Reason != "COUNTEREXAMPLE_INPUT_CONTRACT_UNKNOWN" {
		t.Fatalf("report=%#v", report)
	}
}
