package bodycodegen

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestCriteriaRejectionLeavesMissingEvidenceUnexplained(t *testing.T) {
	original := generatePriceCriteria(t, priceCriteriaSource).Report.RecordAssembly
	raw, _ := json.Marshal(original)
	for _, damage := range []string{"output index", "binding", "input", "round mask"} {
		t.Run(damage, func(t *testing.T) {
			var a RecordAssemblyReceipt
			if err := json.Unmarshal(raw, &a); err != nil {
				t.Fatal(err)
			}
			q := a.CriteriaQuestions
			switch damage {
			case "output index":
				q.Evaluations[0].OutputIndex = -1
			case "binding":
				q.Bindings = nil
			case "input":
				q.InitialInputs = nil
			case "round mask":
				q.Rounds[0].RejectedMasks = []uint16{0, 0}
			}
			r := ExplainCriteriaRejections(&a)
			if r.Status != "INCOMPLETE_OBSERVATION" && r.Status != "INCOMPLETE_REJECTION_EVIDENCE" {
				t.Fatal("missing links produced a complete explanation", r)
			}
		})
	}
}

func TestCriteriaRejectionHistoryKeepsItsOriginalPredicateMeaning(t *testing.T) {
	prior := generatePriceCriteria(t, priceCriteriaSource)
	changed := strings.Replace(priceCriteriaSource, "return input1.price <= 25", "return input1.price <= 35", 1)
	changed = strings.Replace(changed, "return input1.price >= 18", "return input1.price >= 8", 1)
	resumed, err := ResumeRecordAssembly(context.Background(), "quote.gooo", []byte(priceCriteriaSource), []byte(changed), prior, assemblyPolicyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	a := resumed.Report.RecordAssembly
	r := ExplainCriteriaRejections(a)
	if a.Criteria.Passed != a.Criteria.Total || r.Status != "ALL_DECLARED_CANDIDATES_REJECTED" ||
		r.SourceSHA256 != prior.Report.RecordAssembly.CriteriaQuestions.SourceSHA256 {
		t.Fatal("historical failures were relabelled as current conditions", r, a.Criteria)
	}
}
