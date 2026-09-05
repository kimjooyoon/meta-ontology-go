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
		!reflect.DeepEqual(first.Counterexamples[0].SourceIndicator, duplicate) || first.Counterexamples[0].Reason != "INPUT_SUBJECT_KIND_MISMATCH" ||
		len(first.Counterexamples[0].BlockedBy) != 0 {
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

	receipts := passingReceipts(plan)
	if report := VerifyReceipts(plan, receipts); report.Decision != ReceiptDecisionConformant {
		t.Fatalf("valid receipt fixture was not accepted: %+v", report)
	}
	receipts[0].SubjectKind = sourcepolicy.SubjectKindSourceFragment
	receipts[0].ReceiptDigest = operationReceiptDigest(receipts[0])
	if report := VerifyReceipts(plan, receipts); report.Decision != ReceiptDecisionUnknown {
		t.Fatalf("forged receipt input kind was accepted: %+v", report)
	}
}

func TestRefutedInputDomainClaimsRequireCanonicalCause(t *testing.T) {
	base, head := strings.Repeat("a", 40), strings.Repeat("b", 40)
	duplicate := duplicateDomainMetric("fixture.go#func:Duplicate")
	plan := Build(base, head, sourcepolicy.Report{Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(), Indicators: []sourcepolicy.Indicator{
		metric("fixture.go", sourcepolicy.OperationSplitGo, false, false),
		metric("fixture.go:1:Selected", sourcepolicy.OperationExtractFunction, false, false),
		duplicate,
	}})
	if plan.Decision != DecisionPlan || len(plan.Counterexamples) != 1 || plan.IndicatorDecisionLedger == nil {
		t.Fatalf("fixture did not produce a canonical refuted plan: %+v", plan)
	}
	if manifest := BuildExecutionManifest(plan); manifest.Decision != ExecutionDecisionProposed {
		t.Fatalf("canonical refuted plan was not executable before mutation: %+v", manifest)
	}

	matching := Build(base, head, sourcepolicy.Report{Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(), Indicators: []sourcepolicy.Indicator{
		metric("expression", sourcepolicy.OperationCollapseAssign, false, false),
		metric("fixture.go", sourcepolicy.OperationSplitGo, false, false),
		metric("fixture.go:1:Selected", sourcepolicy.OperationExtractFunction, false, false),
	}})
	if manifest := BuildExecutionManifest(matching); manifest.Decision != ExecutionDecisionProposed {
		t.Fatalf("matching fixture was not executable before mutation: %+v", manifest)
	}
	refutedAction := matching.Selected[0]
	matching.Selected = append([]Action{}, matching.Selected[1:]...)
	matching.RefutedIndicatorIDs = []string{refutedAction.IndicatorID}
	binding, ok := BindingForOperation(matching.Registry, refutedAction.Operation)
	if !ok {
		t.Fatal("matching fixture binding is unavailable")
	}
	matching.Counterexamples = []Counterexample{inputDomainCounterexample(refutedAction.SourceIndicator, binding)}
	ledger, err := buildPlanIndicatorDecisionLedgerWithRefuted(
		ledgerSourceIndicators(*matching.IndicatorDecisionLedger),
		matching.Selected,
		matching.UnselectedIndicatorIDs,
		matching.RefutedIndicatorIDs,
	)
	if err != nil {
		t.Fatalf("rebuild matching fixture ledger: %v", err)
	}
	matching.IndicatorDecisionLedger = &ledger
	matching = finish(matching)
	if manifest := BuildExecutionManifest(matching); manifest.Decision != ExecutionDecisionUnknown {
		t.Fatalf("matching source was hidden by forged refutation: %+v", manifest)
	}

	unknownOperation := duplicate
	unknownOperation.Operation = sourcepolicy.Operation("unregistered-operation")
	unknownOperationID := indicatorID(unknownOperation)
	unknownIndicators := ledgerSourceIndicators(*plan.IndicatorDecisionLedger)
	for index, indicator := range unknownIndicators {
		if indicatorID(indicator) == indicatorID(duplicate) {
			unknownIndicators[index] = unknownOperation
		}
	}
	unknownLedger, err := buildPlanIndicatorDecisionLedgerWithRefuted(
		unknownIndicators,
		plan.Selected,
		plan.UnselectedIndicatorIDs,
		[]string{unknownOperationID},
	)
	if err != nil {
		t.Fatalf("rebuild unknown-operation fixture ledger: %v", err)
	}
	unknown := plan
	unknown.IndicatorsDigest = digestJSON(normalizeIndicators(unknownIndicators))
	unknown.RefutedIndicatorIDs = []string{unknownOperationID}
	unknown.Counterexamples = []Counterexample{{
		ID:            "input-domain:" + unknownOperationID,
		IndicatorID:   unknownOperationID,
		SourceIndicator: unknownOperation,
		BlockerID:     "binding-input-domain:unregistered-operation:FILE:SOURCE_FRAGMENT",
		Stage:         "binding",
		Step:          "validate-input-subject-kind",
		Reason:        "INPUT_SUBJECT_KIND_MISMATCH",
		UnknownClass:  "KNOWN_CONTRADICTION",
		NextOperation: "select-valid-domain-action",
		BlockedBy:     []string{},
	}}
	unknown.IndicatorDecisionLedger = &unknownLedger
	unknown = finish(unknown)
	if manifest := BuildExecutionManifest(unknown); manifest.Decision != ExecutionDecisionUnknown {
		t.Fatalf("unknown operation hid forged input-domain cause: %+v", manifest)
	}

	mutations := map[string]func(*Counterexample){
		"reason": func(counterexample *Counterexample) { counterexample.Reason = "FORGED_REASON" },
		"expected-binding": func(counterexample *Counterexample) { counterexample.BlockerID = "binding-input-domain:extract-function:FILE:SOURCE_FRAGMENT" },
		"cause": func(counterexample *Counterexample) { counterexample.Step = "forged-step" },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			forged := plan
			forged.Counterexamples = append([]Counterexample{}, plan.Counterexamples...)
			mutate(&forged.Counterexamples[0])
			forged = finish(forged)
			if manifest := BuildExecutionManifest(forged); manifest.Decision != ExecutionDecisionUnknown {
				t.Fatalf("forged refutation %s was accepted: %+v", name, manifest)
			}
		})
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
