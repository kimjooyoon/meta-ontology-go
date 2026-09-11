package generation

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestBuildIsExactFailClosedAndNonAuthorizing(t *testing.T) {
	base, head := strings.Repeat("a", 40), strings.Repeat("b", 40)
	fixedReport := sourcepolicy.Report{Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(), Indicators: []sourcepolicy.Indicator{
		metric("floor", sourcepolicy.OperationSplitGo, true, true),
	}}
	fixed := Build(base, head, fixedReport)
	if fixed.Decision != DecisionFixedPoint || fixed.Reason != ReasonExactFixedPoint || fixed.ReplayDigest == "" {
		t.Fatalf("unexpected fixed point: %+v", fixed)
	}
	if replay := Build(base, head, fixedReport); !reflect.DeepEqual(fixed, replay) {
		t.Fatal("same exact input did not replay identically")
	}
	report := sourcepolicy.Report{Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(), Indicators: []sourcepolicy.Indicator{
		metric("expression", sourcepolicy.OperationCollapseAssign, false, false),
		metric("topology", sourcepolicy.OperationSplitGo, false, false),
	}}
	planned := Build(base, head, report)
	if planned.Decision != DecisionPlan || len(planned.Selected) != 2 {
		t.Fatalf("unexpected plan: %+v", planned)
	}
	if planned.PromotionAuthorized || planned.PromotionAuthorizedByPlan() || planned.ReplayProof != ProofCoherence {
		t.Fatal("a generation plan acquired authority or lost its replay proof")
	}
	if len(planned.Registry) != 5 {
		t.Fatalf("operation registry is not visible in the plan: %+v", planned.Registry)
	}
	for _, action := range planned.Selected {
		if !action.ReceiptRequired || action.Evaluator == "" || len(action.RequiredIndicatorIDs) == 0 ||
			action.Activity == "" || action.Output == "" || !validActionApplicability(action) {
			t.Fatalf("action lacks conformance obligations: %+v", action)
		}
	}
	if planned.Selected[0].ProofChoice != ProofRegress || planned.Selected[1].ProofChoice != ProofFoundation {
		t.Fatalf("unexpected trilemma choices: %+v", planned.Selected)
	}
}

func TestBuildRejectsShortfallAndUnboundMetrics(t *testing.T) {
	base, head := strings.Repeat("c", 40), strings.Repeat("d", 40)
	short := sourcepolicy.Report{Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(), Indicators: []sourcepolicy.Indicator{
		metric("only-one-group", sourcepolicy.OperationCollapseAssign, false, false),
	}}
	if plan := Build(base, head, short); plan.Decision != DecisionUnknown || plan.Reason != ReasonPressureShortfall || len(plan.Selected) != 0 {
		t.Fatalf("shortfall did not fail closed: %+v", plan)
	}
	unbound := sourcepolicy.Report{Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(), Indicators: []sourcepolicy.Indicator{
		metric("unbound", sourcepolicy.Operation("missing-operation"), false, false),
	}}
	if plan := Build(base, head, unbound); plan.Decision != DecisionUnknown || plan.Reason != ReasonMissingOperation {
		t.Fatalf("unbound metric did not fail closed: %+v", plan)
	}
}

func metric(subject string, operation sourcepolicy.Operation, satisfied, blocking bool) sourcepolicy.Indicator {
	proof := sourcepolicy.ProofFoundation
	dimension := sourcepolicy.DimensionRefactorAssign
	kind := sourcepolicy.SubjectKindFunction
	value, limit := 0, 0
	relation := sourcepolicy.RelationEqual
	role := sourcepolicy.IndicatorRole("")
	if operation == sourcepolicy.OperationCollapseAssign {
		proof = sourcepolicy.ProofRegression
	}
	switch operation {
	case sourcepolicy.OperationSplitGo:
		dimension, kind, value, limit = sourcepolicy.DimensionGoFileLines, sourcepolicy.SubjectKindFile, 76, 75
		relation, role, blocking = sourcepolicy.RelationLessOrEqual, sourcepolicy.IndicatorRoleDriver, false
	case sourcepolicy.OperationSplitGooo:
		dimension, kind, value, limit = sourcepolicy.DimensionGoooFileLines, sourcepolicy.SubjectKindFile, 76, 75
		relation, role, blocking = sourcepolicy.RelationLessOrEqual, sourcepolicy.IndicatorRoleDriver, false
	case sourcepolicy.OperationExtractFunction:
		dimension, kind, value, limit = sourcepolicy.DimensionFunctionLines, sourcepolicy.SubjectKindFunction, 76, 75
		relation, role, blocking = sourcepolicy.RelationLessOrEqual, sourcepolicy.IndicatorRoleDriver, false
	}
	if sourcepolicy.IsLineCapMetric(dimension) && satisfied {
		value = limit
	}
	return sourcepolicy.Indicator{MetricID: dimension, Subject: subject,
		SubjectKind: kind, Value: value, Limit: limit, Relation: relation, Role: role, Applicability: sourcepolicy.ApplicabilityApplicable,
		ApplicabilityRule:   sourcepolicy.ApplicabilityRuleDefault,
		ApplicabilityReason: sourcepolicy.ApplicabilityReasonCatalogApplicable,
		Satisfied:           satisfied, Blocking: blocking, Proof: proof,
		Producer: "test-metric", Consumer: "test-generation", Operation: operation}
}
