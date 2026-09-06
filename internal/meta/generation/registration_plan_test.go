package generation

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

// These hashes are structural proposal fixtures, never runtime evidence.
func registrationPlanFixture(t *testing.T) (sourcepolicy.Report, syntaxregistration.Request) {
	t.Helper()
	hash := "sha256:" + strings.Repeat("a", 64)
	request := syntaxregistration.Request{BaseVersion: 30, SnapshotDigest: hash, SourceDigest: hash,
		Toolchain: "go1.27.0", ExecutionIdentity: syntaxregistration.ExecutionIdentity{
			GoVersion: "go1.27.0", GOOS: "linux", GOARCH: "amd64",
			ExecutableDigest: hash, GoCommandDigest: hash, CompilerDigest: hash},
		Case: languagesyntax.CaseDefinition{ID: "native-registration-plan-fixture",
			Path: "examples/native-registration-plan-fixture/main.gooo", Kind: languagesyntax.KindValid,
			ExpectedDecision: languagesyntax.DecisionPass, ProofChoice: "COHERENCE",
			MetaOperation: "replay-language-syntax", Scope: languagesyntax.ScopeLanguageCapability, EntityFields: true}}
	binding, err := syntaxregistration.NativeBinding()
	if err != nil {
		t.Fatal(err)
	}
	indicator := sourcepolicy.Indicator{MetricID: sourcepolicy.DimensionSyntaxRegistration,
		Subject: request.Case.Path, SubjectKind: sourcepolicy.SubjectKindRegistrationRequest,
		Value: 0, Limit: 1, Relation: sourcepolicy.RelationEqual,
		Applicability: sourcepolicy.ApplicabilityApplicable, ApplicabilityRule: sourcepolicy.ApplicabilityRuleDefault,
		ApplicabilityReason: sourcepolicy.ApplicabilityReasonCatalogApplicable,
		Role:                sourcepolicy.IndicatorRoleDriver, Proof: sourcepolicy.ProofCoherence,
		Producer: binding.InputActivityID, Consumer: binding.ActivityID, Operation: sourcepolicy.OperationRegisterSyntax,
		OperationInputDigest: syntaxregistration.RequestDigest(request)}
	return sourcepolicy.Report{Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(),
		Indicators: []sourcepolicy.Indicator{indicator, metric("fixture.go:3:CollapseFixture",
			sourcepolicy.OperationCollapseAssign, false, false)}}, request
}

func TestRegistrationCommonPlanCarriesExactTypedInput(t *testing.T) {
	report, request := registrationPlanFixture(t)
	inputs := map[string]syntaxregistration.Request{syntaxregistration.RequestDigest(request): request}
	first := BuildWithRegistrationInputs(strings.Repeat("a", 40), strings.Repeat("b", 40), report, inputs)
	second := BuildWithRegistrationInputs(strings.Repeat("a", 40), strings.Repeat("b", 40), report, inputs)
	if !reflect.DeepEqual(first, second) || first.Decision != DecisionPlan || len(first.Selected) != 2 ||
		first.MinimumIndependent != 2 || first.RequestedK != 2 || len(first.Registry) != 5 || first.PromotionAuthorized {
		t.Fatalf("common registration plan did not preserve exact selection: %+v", first)
	}
	registrationCount := 0
	for _, action := range first.Selected {
		if !ValidRegistrationActionInput(action) {
			t.Fatalf("action input is not exact: %+v", action)
		}
		if action.Operation == sourcepolicy.OperationRegisterSyntax {
			registrationCount++
			if !reflect.DeepEqual(action.RegistrationRequest, &request) || len(action.RequiredIndicatorIDs) != 4 {
				t.Fatalf("registration request or obligations lost: %+v", action)
			}
		}
	}
	if registrationCount != 1 || BuildExecutionManifest(first).Decision != ExecutionDecisionProposed {
		t.Fatal("exact typed request did not enter common manifest")
	}
	mutated := first
	mutated.Selected = append([]Action{}, first.Selected...)
	for index := range mutated.Selected {
		if mutated.Selected[index].RegistrationRequest != nil {
			requestCopy := *mutated.Selected[index].RegistrationRequest
			requestCopy.Case.ID = "substituted-request"
			mutated.Selected[index].RegistrationRequest = &requestCopy
		}
	}
	mutated = finish(mutated)
	if BuildExecutionManifest(mutated).Decision == ExecutionDecisionProposed {
		t.Fatal("resealed request substitution entered execution")
	}
}

func TestRegistrationCommonPlanPreservesMissingAndStaleCausality(t *testing.T) {
	report, request := registrationPlanFixture(t)
	key := syntaxregistration.RequestDigest(request)
	stale := request
	stale.SourceDigest = "sha256:" + strings.Repeat("b", 64)
	for _, item := range []struct {
		name   string
		inputs map[string]syntaxregistration.Request
		class  string
	}{
		{"missing", nil, "DIRECT_MISSING"},
		{"stale", map[string]syntaxregistration.Request{key: stale}, "STALE"},
	} {
		t.Run(item.name, func(t *testing.T) {
			plan := BuildWithRegistrationInputs(strings.Repeat("a", 40), strings.Repeat("b", 40), report, item.inputs)
			if plan.Decision != DecisionUnknown || len(plan.Selected) != 0 || len(plan.RegistrationInputFailures) != 1 {
				t.Fatalf("input uncertainty was hidden: %+v", plan)
			}
			failure := plan.RegistrationInputFailures[0]
			if failure.State != "UNKNOWN" || failure.UnknownClass != item.class || failure.Stage == "" ||
				failure.Step == "" || failure.Reason == "" || failure.NextOperation == "" || failure.BlockedBy == nil {
				t.Fatalf("input failure lost causal coordinates: %+v", failure)
			}
		})
	}
}

func TestRegistrationCommonPlanRetainsIndependenceAndNativeAuthority(t *testing.T) {
	report, request := registrationPlanFixture(t)
	inputs := map[string]syntaxregistration.Request{syntaxregistration.RequestDigest(request): request}
	single := report
	single.Indicators = append([]sourcepolicy.Indicator{}, report.Indicators[:1]...)
	plan := BuildWithRegistrationInputs(strings.Repeat("a", 40), strings.Repeat("b", 40), single, inputs)
	if plan.Decision != DecisionUnknown || plan.Reason != ReasonPressureShortfall || len(plan.Selected) != 0 {
		t.Fatalf("registration bypassed independence floor: %+v", plan)
	}
	tampered := report
	tampered.Indicators = append([]sourcepolicy.Indicator{}, report.Indicators...)
	tampered.Indicators[0].Producer = "prose-is-not-producer-authority"
	if plan := BuildWithRegistrationInputs(strings.Repeat("a", 40), strings.Repeat("b", 40), tampered, inputs); plan.Decision != DecisionRejected {
		t.Fatalf("forged metric producer was not refuted: %+v", plan)
	}
	legacy := sourcepolicy.Report{Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(),
		Indicators: []sourcepolicy.Indicator{report.Indicators[1]}}
	if plan := BuildWithRegistrationInputs(strings.Repeat("a", 40), strings.Repeat("b", 40), legacy, inputs); plan.Decision != DecisionRejected {
		t.Fatalf("unreferenced request was accepted: %+v", plan)
	}
	registry := DefaultRegistry()
	for index := range registry {
		if registry[index].Operation == sourcepolicy.OperationRegisterSyntax {
			registry[index].Executor = "caller-controlled-executor"
		}
	}
	if _, known := registryIndex(registry); known {
		t.Fatal("caller substituted native registration executor")
	}
}
