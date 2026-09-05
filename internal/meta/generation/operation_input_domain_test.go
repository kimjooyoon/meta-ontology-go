package generation

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestOperationInputContractUsesExactEmbeddedGoooIR(t *testing.T) {
	contract, err := loadOperationInputContract()
	if err != nil {
		t.Fatal(err)
	}
	if len(contract.Bindings) != 4 || contract.SourceDigest == "" || contract.SemanticDigest == "" {
		t.Fatalf("unexpected operation input contract: %+v", contract)
	}
	want := map[sourcepolicy.Operation]sourcepolicy.SubjectKind{
		sourcepolicy.OperationCollapseAssign:  sourcepolicy.SubjectKindFunction,
		sourcepolicy.OperationSplitGo:          sourcepolicy.SubjectKindFile,
		sourcepolicy.OperationSplitGooo:        sourcepolicy.SubjectKindFile,
		sourcepolicy.OperationExtractFunction:  sourcepolicy.SubjectKindFunction,
	}
	for operation, kind := range want {
		binding, ok := contract.Bindings[operation]
		if !ok || binding.InputSubjectKind != kind {
			t.Fatalf("operation %q input binding = %+v", operation, binding)
		}
	}
}

func TestOperationInputContractRejectsMissingDuplicateAndUnknownDeclarations(t *testing.T) {
	base := string(operationInputContractSource)
	cases := map[string]string{
		"missing activity": strings.Replace(base, "activity ExtractFunction(FunctionInput) -> OperationResult\n", "", 1),
		"duplicate activity": base + "\nactivity ExtractFunction(FunctionInput) -> OperationResult\n",
		"unknown input": strings.Replace(base, "activity ExtractFunction(FunctionInput) -> OperationResult", "activity ExtractFunction(UnknownInput) -> OperationResult", 1),
	}
	for name, source := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parseOperationInputContract([]byte(source)); err == nil {
				t.Fatal("malformed operation input contract was accepted")
			}
		})
	}
}

func TestBuildRefutesMismatchedSourceFragmentBeforeSelection(t *testing.T) {
	base, head := strings.Repeat("a", 40), strings.Repeat("b", 40)
	duplicate := duplicateDomainMetric("fixture.go#func:Duplicate")
	report := sourcepolicy.Report{Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(), Indicators: []sourcepolicy.Indicator{
		metric("fixture.go", sourcepolicy.OperationSplitGo, false, false),
		metric("fixture.go:1:Selected", sourcepolicy.OperationExtractFunction, false, false),
		duplicate,
	}}
	first := Build(base, head, report)
	second := Build(base, head, report)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("domain-aware plan did not replay deterministically")
	}
	if first.Decision != DecisionPlan || len(first.Selected) != 2 || len(first.RefutedIndicatorIDs) != 1 || len(first.Counterexamples) != 1 {
		t.Fatalf("unexpected mixed domain plan: %+v", first)
	}
	for _, action := range first.Selected {
		if action.SubjectKind != action.InputSubjectKind {
			t.Fatalf("mismatched action entered selection: %+v", action)
		}
	}
	duplicateID := indicatorID(duplicate)
	if first.RefutedIndicatorIDs[0] != duplicateID || first.Counterexamples[0].IndicatorID != duplicateID ||
		!reflect.DeepEqual(first.Counterexamples[0].SourceIndicator, duplicate) || first.Counterexamples[0].Reason != "INPUT_SUBJECT_KIND_MISMATCH" {
		t.Fatalf("domain counterexample is not exact: %+v", first.Counterexamples)
	}
	if first.IndicatorDecisionLedger == nil || first.IndicatorDecisionLedger.RefutedCount != 1 {
		t.Fatalf("source refutation was not retained in ledger: %+v", first.IndicatorDecisionLedger)
	}
	for _, entry := range first.IndicatorDecisionLedger.Entries {
		if entry.IndicatorID == duplicateID {
			if entry.Disposition != IndicatorDispositionRepairRefuted || entry.Action != nil || !reflect.DeepEqual(entry.SourceIndicator, duplicate) {
				t.Fatalf("source fragment observation was not retained: %+v", entry)
			}
		}
	}
}

func TestBuildDoesNotReportFixedPointForMismatchOnlyOrPressureShortfall(t *testing.T) {
	base, head := strings.Repeat("c", 40), strings.Repeat("d", 40)
	duplicate := duplicateDomainMetric("fixture.go#func:Duplicate")
	mismatchOnly := Build(base, head, sourcepolicy.Report{Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(), Indicators: []sourcepolicy.Indicator{duplicate}})
	if mismatchOnly.Decision != DecisionUnknown || mismatchOnly.Reason != ReasonPressureShortfall || len(mismatchOnly.Selected) != 0 || len(mismatchOnly.RefutedIndicatorIDs) != 1 {
		t.Fatalf("mismatch-only case was not a typed shortfall: %+v", mismatchOnly)
	}
	validOne := Build(base, head, sourcepolicy.Report{Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(), Indicators: []sourcepolicy.Indicator{
		metric("fixture.go", sourcepolicy.OperationSplitGo, false, false), duplicate,
	}})
	if validOne.Decision != DecisionUnknown || validOne.Reason != ReasonPressureShortfall || len(validOne.Selected) != 0 || len(validOne.UnselectedIndicatorIDs) != 1 || len(validOne.RefutedIndicatorIDs) != 1 {
		t.Fatalf("valid-plus-mismatch case lost pressure shortfall: %+v", validOne)
	}
}

func TestBuildAndExecutionRejectForgedInputDomainContract(t *testing.T) {
	plan := Build(strings.Repeat("e", 40), strings.Repeat("f", 40), sourcepolicy.Report{Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(), Indicators: []sourcepolicy.Indicator{
		metric("fixture.go", sourcepolicy.OperationSplitGo, false, false),
		metric("fixture.go:1:Selected", sourcepolicy.OperationExtractFunction, false, false),
	}})
	if plan.Decision != DecisionPlan {
		t.Fatalf("fixture did not produce a plan: %+v", plan)
	}
	forgedExpected := plan
	forgedExpected.Selected = append([]Action{}, plan.Selected...)
	forgedExpected.Selected[0].InputSubjectKind = sourcepolicy.SubjectKindSourceFragment
	forgedExpected = finish(forgedExpected)
	if manifest := BuildExecutionManifest(forgedExpected); manifest.Decision != ExecutionDecisionUnknown {
		t.Fatalf("forged expected input kind was executable: %+v", manifest)
	}

	forgedRegistry := plan
	forgedRegistry.Registry = append([]Binding{}, plan.Registry...)
	forgedRegistry.Registry[0].InputSubjectKind = sourcepolicy.SubjectKindSourceFragment
	forgedRegistry.RegistryDigest = digestJSON(forgedRegistry.Registry)
	forgedRegistry = finish(forgedRegistry)
	if manifest := BuildExecutionManifest(forgedRegistry); manifest.Decision != ExecutionDecisionUnknown {
		t.Fatalf("forged registry input kind was executable: %+v", manifest)
	}

	action := plan.Selected[0]
	receipt := SealReceipt(plan, action, nil)
	receipt.SubjectKind = sourcepolicy.SubjectKindSourceFragment
	receipt.ReceiptDigest = operationReceiptDigest(receipt)
	if report := VerifyReceipts(plan, []OperationReceipt{receipt}); report.Decision != ReceiptDecisionUnknown {
		t.Fatalf("forged receipt input kind was accepted: %+v", report)
	}
}

func duplicateDomainMetric(subject string) sourcepolicy.Indicator {
	return sourcepolicy.Indicator{
		MetricID: sourcepolicy.DimensionRefactorDuplicate, Family: sourcepolicy.FamilyDuplication,
		Subject: subject, SubjectKind: sourcepolicy.SubjectKindSourceFragment,
		Value: 1, Limit: 0, Relation: sourcepolicy.RelationLessOrEqual,
		Applicability: sourcepolicy.ApplicabilityApplicable,
		ApplicabilityRule: sourcepolicy.ApplicabilityRuleDefault,
		ApplicabilityReason: sourcepolicy.ApplicabilityReasonCatalogApplicable,
		Blocking: true, Satisfied: false, Proof: sourcepolicy.ProofFoundation,
		Producer: "duplicate-detector", Consumer: "deduplicator",
		Operation: sourcepolicy.OperationExtractFunction,
	}
}
