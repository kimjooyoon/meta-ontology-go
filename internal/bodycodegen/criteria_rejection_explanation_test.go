package bodycodegen

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

const priceCriteriaSource = `package quote
namespace quote
entity Integer id "quote://integer"
entity Boolean id "quote://boolean"
entity Quote id "quote://quote" fields { field price id "quote://price" type integer required one }
activity Price(Integer) -> Quote computes "return Quote{price:input}" assembling {
 choice "price" field_value at "0" alternative "input+20" intent "Cover cost 18 within budget 25"
 criteria {
  predicate "within-budget" activity "WithinBudget"
  predicate "cost-covered" activity "CostCovered"
  input "[10]"
 }
 observe { criteria_strategy "distinguish_then_check" criteria_input "[12]" max_candidates "2" max_rounds "1" }
 attempts "2"
}
activity WithinBudget(Integer,Quote) -> Boolean computes "return input1.price <= 25"
activity CostCovered(Integer,Quote) -> Boolean computes "return input1.price >= 18"
`

func generatePriceCriteria(t *testing.T, source string) Result {
	t.Helper()
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	r, err := g.GenerateSourceAssembly(context.Background(), "quote.gooo", []byte(source), "Price")
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestCriteriaRejectionExplainsBothSourceConditionsWithoutMutation(t *testing.T) {
	a := generatePriceCriteria(t, priceCriteriaSource).Report.RecordAssembly
	before, _ := json.Marshal(a)
	r := ExplainCriteriaRejections(a)
	if r.Status != "ALL_DECLARED_CANDIDATES_REJECTED" || r.CandidateSpace != 2 || r.Explained != 2 || len(r.Groups) != 2 ||
		r.ReceiptSHA256 != digest(before) || r.SourceSHA256 != a.CriteriaQuestions.SourceSHA256 {
		t.Fatal(r)
	}
	if r.Groups[0].Criterion.ID != "cost-covered" || r.Groups[0].Witnesses[0].Mask != 0 || r.Groups[0].Witnesses[0].ActualJSON != `{"price":10}` ||
		r.Groups[1].Criterion.ID != "within-budget" || r.Groups[1].Witnesses[0].Mask != 1 || r.Groups[1].Witnesses[0].ActualJSON != `{"price":30}` {
		t.Fatal(r.Groups)
	}
	again := ExplainCriteriaRejections(a)
	x, _ := json.Marshal(r)
	y, _ := json.Marshal(again)
	if string(x) != string(y) {
		t.Fatal("record-only explanation changed")
	}
	r.Groups[0].Criterion.Activity = "changed explanation"
	after, _ := json.Marshal(a)
	if string(before) != string(after) {
		t.Fatal("explanation changed recorded observations")
	}
}

func TestCriteriaRejectionDistinguishesBetterChoiceFromUnseenCandidates(t *testing.T) {
	repaired := strings.Replace(priceCriteriaSource, `alternative "input+20"`, `alternative "input+10"`, 1)
	r := ExplainCriteriaRejections(generatePriceCriteria(t, repaired).Report.RecordAssembly)
	if r.Status != "CANDIDATES_REMAIN" || r.Surviving != 1 || r.Rejected != 1 || r.Explained != 1 {
		t.Fatal("a permitted price of 20 satisfies both conditions", r)
	}
	bounded := strings.Replace(priceCriteriaSource, `max_candidates "2"`, `max_candidates "1"`, 1)
	r = ExplainCriteriaRejections(generatePriceCriteria(t, bounded).Report.RecordAssembly)
	if r.Status != "EXAMINED_CANDIDATES_REJECTED" || r.Unobserved != 1 || r.Explained != 1 {
		t.Fatal("unseen alternatives were ruled out", r)
	}
}

func TestCriteriaRejectionKeepsUnavailableCandidateOpen(t *testing.T) {
	source := strings.Replace(priceCriteriaSource, "return input1.price <= 25", "return input1.price/(input1.price-input1.price) > 0", 1)
	r := ExplainCriteriaRejections(generatePriceCriteria(t, source).Report.RecordAssembly)
	if r.Status != "CANDIDATES_REMAIN" || r.Unavailable != 1 || r.Rejected != 1 || r.Explained != 1 || r.Groups[0].Criterion.ID != "cost-covered" {
		t.Fatal(r)
	}
}

func TestCriteriaRejectionLinksExactCasesBeforePredicateChecks(t *testing.T) {
	source := strings.Replace(priceCriteriaSource, " criteria {", ` value_case "[10]" -> "{\"price\":20}"
 criteria {`, 1)
	r := ExplainCriteriaRejections(generatePriceCriteria(t, source).Report.RecordAssembly)
	if r.Status != "ALL_DECLARED_CANDIDATES_REJECTED" || r.Explained != 2 || len(r.Groups) != 1 ||
		r.Groups[0].Kind != "EXACT_CASE" || r.Groups[0].ExpectedJSON != `{"price":20}` || len(r.Groups[0].Witnesses) != 2 {
		t.Fatal(r)
	}
}
