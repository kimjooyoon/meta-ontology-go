package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestConstructionPathObservationReportKeepsScopesSeparate(t *testing.T) {
	p := &bodycodegen.BodyPathReceipt{Observation: &bodycodegen.PathObservationReceipt{
		Status: "ONE_SURVIVING_CANDIDATE", InitialCases: []pathplan.TestCase{{Input: 31}}, EffectiveCases: 2}}
	c := &bodycodegen.PathCandidate{Observation: &bodycodegen.PathCandidateObservation{InitialCases: 1, AddedCases: 1}}
	raw, _ := json.Marshal(constructionPathObservationRows(p, c))
	for _, want := range []string{"Initial question outcome", "ONE_SURVIVING_CANDIDATE", "1 original; 2 effective",
		"1 original + 1 acquired; included in selected local checks"} {
		if !strings.Contains(string(raw), want) {
			t.Fatal("missing scope", want, string(raw))
		}
	}
	if rows := constructionPathObservationRows(&bodycodegen.BodyPathReceipt{}, nil); len(rows) != 0 {
		t.Fatal("invented an observation for a legacy path", rows)
	}
	for _, status := range []string{"ORACLE_UNAVAILABLE", "CANDIDATE_BUDGET", "ROUND_BUDGET"} {
		p.Observation.Status = status
		raw, _ := json.Marshal(constructionPathObservationRows(p, nil))
		if !strings.Contains(string(raw), status) || !strings.Contains(string(raw), "not recorded; inspect JSON") {
			t.Fatal("unresolved or missing evidence hidden", string(raw))
		}
	}
}
