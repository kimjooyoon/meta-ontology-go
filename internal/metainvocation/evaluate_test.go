package metainvocation

import "testing"

const testSource = `package ciplan
namespace ciplan

entity ChangeSet id "gooo://meta/ci-plan/entity/change-set"
entity CheckPlan id "gooo://meta/ci-plan/entity/check-plan"
entity VerificationReceipt id "gooo://meta/ci-plan/entity/verification-receipt"

activity SelectGoCheck(ChangeSet) -> CheckPlan computes "ci.rule:go:v1"
activity SelectDocsCheck(ChangeSet) -> CheckPlan computes "ci.rule:docs:v1"
activity SelectYAMLCheck(ChangeSet) -> CheckPlan computes "ci.rule:yaml:v1"
activity PlanCI(ChangeSet) -> CheckPlan computes "ci.plan:v1"
activity VerifyCIPlan(CheckPlan) -> VerificationReceipt computes "ci.verify:v1"
`

func TestInvokeUsesGoooProgramsAndPreservesClaims(t *testing.T) {
	program, err := Compile("main.gooo", []byte(testSource), StandardRegistry())
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name       string
		input      string
		decision   string
		resolution string
	}{
		{name: "pass", input: `{"schema":"gooo/ci-plan-input/v1","case_id":"pass","files":["internal/a.go"]}`, decision: DecisionPass, resolution: ResolutionExact},
		{name: "closed", input: `{"schema":"gooo/ci-plan-input/v1","case_id":"closed","files":[]}`, decision: DecisionClosed, resolution: ResolutionExact},
		{name: "unknown", input: `{"schema":"gooo/ci-plan-input/v1","case_id":"unknown","files":["LICENSE"]}`, decision: DecisionUnknown, resolution: ResolutionLower},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			report, err := Invoke(program, "PlanCI", []byte(test.input))
			if err != nil {
				t.Fatal(err)
			}
			if report.Decision != test.decision || report.Resolution != test.resolution || len(report.Claims) != 3 {
				t.Fatalf("unexpected report: decision=%s resolution=%s claims=%d", report.Decision, report.Resolution, len(report.Claims))
			}
			if err := Validate(report); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestUnknownDoesNotAuthorizePartialPlan(t *testing.T) {
	program, err := Compile("main.gooo", []byte(testSource), StandardRegistry())
	if err != nil {
		t.Fatal(err)
	}
	report, err := Invoke(program, "PlanCI", []byte(`{"schema":"gooo/ci-plan-input/v1","case_id":"mixed","files":["internal/a.go","tool.py"]}`))
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionUnknown || len(report.Plan.Checks) != 0 || len(report.Unknowns) != 1 {
		t.Fatalf("unknown input authorized a partial plan: %+v", report)
	}
	if report.Unknowns[0].Stage != "RULE_SELECTION" || report.Unknowns[0].Step != "classify-change" || report.Unknowns[0].Reason != "NO_REGISTERED_RULE" {
		t.Fatalf("unknown causality was compressed: %+v", report.Unknowns[0])
	}
}
