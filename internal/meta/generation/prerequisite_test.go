package generation

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestPrerequisitesSelectLaterCandidateWithoutErasingDebt(t *testing.T) {
	base, head := strings.Repeat("a", 40), strings.Repeat("b", 40)
	report := prerequisiteMetrics()
	var visited []string
	observe := func(action Action) CandidatePrerequisite {
		visited = append(visited, action.Subject)
		reason := "SPLIT_PLAN_READY"
		if action.Subject == "a.go" {
			reason = "RENDERED_DECLARATION_CAPACITY"
		}
		return NewSplitPrerequisite(action, head, strings.Repeat("c", 64), strings.Repeat("d", 64), reason)
	}
	plan := BuildWithPrerequisites(base, head, report, nil, observe)
	if plan.Decision != DecisionPlan || len(plan.Selected) != 2 || plan.Selected[1].Subject != "b.go" ||
		!reflect.DeepEqual(visited, []string{"a.go", "b.go"}) {
		t.Fatalf("candidate selection ignored prerequisites or eagerly inspected unused candidates: %+v %v", plan, visited)
	}
	if !receiptPlanKnown(plan) || validatePlanIndicatorDecisionLedger(plan) != nil ||
		plan.IndicatorDecisionLedger.IndicatorCount != len(report.Indicators) || plan.IndicatorDecisionLedger.DeferredCount != 2 {
		t.Fatal("selection lost the full source denominator or canonical plan", plan)
	}
	for _, entry := range plan.IndicatorDecisionLedger.Entries {
		if entry.SourceIndicator.Satisfied || entry.IndicatorOutcome != entry.SourceIndicator.Outcome() {
			t.Fatal("a prerequisite observation changed measured debt", entry)
		}
	}
	visited = nil
	if replay := BuildWithPrerequisites(base, head, report, nil, observe); !reflect.DeepEqual(plan, replay) {
		t.Fatal("same inputs did not replay")
	}
}

func TestPrerequisiteShortfallRetainsUnknownAndDeferredInputs(t *testing.T) {
	for _, reason := range []string{"LOGICAL_DIRECTORY_CAPACITY", "DIRECTORY_OBSERVATION_MISSING", "SUCCESSOR_PROJECTION_REQUIRED"} {
		t.Run(reason, func(t *testing.T) {
			head := strings.Repeat("b", 40)
			plan := BuildWithPrerequisites(strings.Repeat("a", 40), head, prerequisiteMetrics(), nil,
				func(action Action) CandidatePrerequisite {
					return NewSplitPrerequisite(action, head, strings.Repeat("c", 64), strings.Repeat("d", 64), reason)
				})
			if plan.Decision != DecisionUnknown || plan.Reason != ReasonPressureShortfall || len(plan.Selected) != 0 ||
				len(plan.Prerequisites) != 3 || plan.IndicatorDecisionLedger == nil || plan.IndicatorDecisionLedger.IndicatorCount != 4 ||
				!receiptPlanKnown(plan) || validatePlanIndicatorDecisionLedger(plan) != nil {
				t.Fatalf("unmet prerequisites disappeared or became a fixed point: %+v", plan)
			}
		})
	}
}

func TestPrerequisitePlanRejectsForgedOriginsAndReadiness(t *testing.T) {
	head := strings.Repeat("b", 40)
	plan := BuildWithPrerequisites(strings.Repeat("a", 40), head, prerequisiteMetrics(), nil,
		func(action Action) CandidatePrerequisite {
			return NewSplitPrerequisite(action, head, strings.Repeat("c", 64), strings.Repeat("d", 64), "SPLIT_PLAN_READY")
		})
	for name, mutate := range map[string]func(*CandidatePrerequisite){
		"head":      func(p *CandidatePrerequisite) { p.HeadSHA = strings.Repeat("e", 40) },
		"source":    func(p *CandidatePrerequisite) { p.SourceDigest = "" },
		"metrics":   func(p *CandidatePrerequisite) { p.MetricsDigest = "" },
		"contract":  func(p *CandidatePrerequisite) { p.ContractSourceDigest = strings.Repeat("e", 64) },
		"readiness": func(p *CandidatePrerequisite) { p.State = "BLOCKED" },
		"next":      func(p *CandidatePrerequisite) { p.NextOperation = "waive-capacity" },
	} {
		t.Run(name, func(t *testing.T) {
			forged := plan
			forged.Prerequisites = append([]CandidatePrerequisite{}, plan.Prerequisites...)
			mutate(&forged.Prerequisites[0])
			forged = finish(forged)
			if receiptPlanKnown(forged) {
				t.Fatal("resealed invalid prerequisite was accepted")
			}
		})
	}
}

func TestSplitPrerequisitesAreBoundToGoooRelations(t *testing.T) {
	contract, err := SplitGoInputContractEvidence()
	if err != nil || len(contract.Obligations) != 3 || !contract.UsedInputFact || !contract.GeneratedOutputFact {
		t.Fatalf("split prerequisites are not source-bound: %+v %v", contract, err)
	}
	for _, obligation := range contract.Obligations {
		if !obligation.UsedInputFact || !obligation.GeneratedOutputFact {
			t.Fatal("prerequisite is missing a PROV-O relation", obligation)
		}
	}
	tampered := strings.Replace(string(operationInputContractSource),
		"InspectSplitTopology(SplitRenderingObservation)", "InspectSplitTopology(FileInput)", 1)
	if _, err := parseOperationInputContract([]byte(tampered)); err == nil {
		t.Fatal("broken prerequisite chain accepted")
	}
}

func prerequisiteMetrics() sourcepolicy.Report {
	return sourcepolicy.Report{Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(),
		Indicators: []sourcepolicy.Indicator{
			metric("expression", sourcepolicy.OperationCollapseAssign, false, false),
			metric("a.go", sourcepolicy.OperationSplitGo, false, false),
			metric("b.go", sourcepolicy.OperationSplitGo, false, false),
			metric("c.go", sourcepolicy.OperationSplitGo, false, false),
		}}
}
