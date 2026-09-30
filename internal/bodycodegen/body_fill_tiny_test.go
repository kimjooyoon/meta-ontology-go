package bodycodegen

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
)

func TestTinyGoBodyFillMapsConditionRootOperationsToDeclaredIDs(t *testing.T) {
	fixture := tinyBodyFillFixture(`if input < 0 { return __GOOO_BODY_HOLE_floor__ } else { return input }`)
	plan := tinyBodyFillPlan("Choose the declared operation for the negative branch.", "floor",
		IRBodyFillCandidate{ID: "zetaadd", Expression: "input + 0"},
		IRBodyFillCandidate{ID: "alpha", Expression: "(input - 0)"},
	)
	plan.TestCases = []IRBodyFillTestCase{{Input: -1, Expected: -1}, {Input: 0, Expected: 0}, {Input: 4, Expected: 4}}
	provider := &recordingTinyGoBodyFillProvider{operation: decisionroute.TinyGoOperationSubtract}

	result, err := generateWithIRBodyFillOptions(context.Background(), "tiny.gooo", fixture, "Choose",
		plan, "", "", IRBodyFillOptions{}, provider)
	if err != nil {
		t.Fatalf("GenerateWithIRBodyFillWithOptions: %v", err)
	}
	if provider.calls != 1 || provider.request.Intent != plan.Intent {
		t.Fatalf("provider calls or explicit intent mismatch: calls=%d intent=%q", provider.calls, provider.request.Intent)
	}
	if got := []string{provider.request.Question.Options[0].Operation, provider.request.Question.Options[1].Operation}; got[0] != decisionroute.TinyGoOperationAdd || got[1] != decisionroute.TinyGoOperationSubtract {
		t.Fatalf("root operations were not mapped by expression, in plan order: %v", got)
	}
	fill := result.Report.BodyFill
	if fill == nil || fill.ProposedCandidateID != "alpha" || fill.SelectedCandidateID != "alpha" ||
		fill.FunctionalAccuracyPct != 100 || fill.TestCasesTotal != 3 {
		t.Fatalf("unexpected selected finite-suite result: %+v", fill)
	}
	if fill.Decision.TinyGoPredictedOperation != decisionroute.TinyGoOperationSubtract ||
		fill.Decision.TinyGoPredictionApplied == nil || !*fill.Decision.TinyGoPredictionApplied {
		t.Fatalf("mapped model operation was not preserved in the receipt: %+v", fill.Decision)
	}
	if result.Report.RouteEquivalence.Decision != "PASS" || !result.Report.TypecheckPassed {
		t.Fatalf("tiny-filled conditional lost compiler validation: route=%+v typecheck=%t", result.Report.RouteEquivalence, result.Report.TypecheckPassed)
	}
	if fill.LocalModelPredictions != 1 || fill.ExternalProviderCalls != 0 || !fill.ExternalProviderCallsKnown ||
		fill.Timing.TinyDecisionMS != fill.Timing.ProviderDecisionMS || fill.Timing.LayaDecisionMS != 0 {
		t.Fatalf("tiny decision accounting is not separate from Laya: %+v", fill)
	}
	if dimension := bodyFillDimension(result, "external_network_boundary"); dimension.Status != "PASS" || dimension.Numerator != 1 {
		t.Fatalf("tiny path counted as external network use: %+v", dimension)
	}
	if dimension := bodyFillDimension(result, "laya_decision_observation"); dimension.Denominator != 0 {
		t.Fatalf("local choice was counted as an eligible Laya decision: %+v", dimension)
	}
	if provider.request.Fallback != "zetaadd" || provider.request.Question.Options[1].ID != "alpha" {
		t.Fatalf("candidate names or positions replaced explicit ID mapping: %+v", provider.request.Question)
	}
}

func TestTinyGoRootOperationUsesOnlyTheClosedEightBinaryTokens(t *testing.T) {
	tests := []struct {
		expression string
		operation  string
	}{
		{"a + b", decisionroute.TinyGoOperationAdd},
		{"((a - b))", decisionroute.TinyGoOperationSubtract},
		{"a * b", decisionroute.TinyGoOperationMultiply},
		{"a < b", decisionroute.TinyGoOperationLessThan},
		{"a <= b", decisionroute.TinyGoOperationLessEqual},
		{"a == b", decisionroute.TinyGoOperationEqual},
		{"a && b", decisionroute.TinyGoOperationAnd},
		{"a || b", decisionroute.TinyGoOperationOr},
	}
	for _, test := range tests {
		got, err := tinyGoRootOperation(test.expression)
		if err != nil || got != test.operation {
			t.Errorf("tinyGoRootOperation(%q) = %q, %v; want %q", test.expression, got, err, test.operation)
		}
	}
	if _, err := tinyGoRootOperation("a != b"); err == nil {
		t.Fatal("operator outside the closed eight-operation set was accepted")
	}
}

func TestTinyGoBodyFillMapsAssignmentRootOperationsAndPreservesSourceEquivalence(t *testing.T) {
	fixture := tinyBodyFillFixture(`let result = input + __GOOO_BODY_HOLE_value__
return result`)
	plan := tinyBodyFillPlan("Add a small offset to the input.", "value",
		IRBodyFillCandidate{ID: "multid", Expression: "input * 2"},
		IRBodyFillCandidate{ID: "plusid", Expression: "((input + 1))"},
	)
	plan.TestCases = []IRBodyFillTestCase{{Input: 0, Expected: 0}, {Input: 2, Expected: 6}}
	provider := &recordingTinyGoBodyFillProvider{operation: decisionroute.TinyGoOperationMultiply}

	result, err := generateWithIRBodyFillOptions(context.Background(), "tiny-assignment.gooo", fixture, "Choose",
		plan, "", "", IRBodyFillOptions{}, provider)
	if err != nil {
		t.Fatalf("GenerateWithIRBodyFillWithOptions: %v", err)
	}
	if provider.calls != 1 || provider.request.Question.Options[0].Operation != decisionroute.TinyGoOperationMultiply ||
		provider.request.Question.Options[1].Operation != decisionroute.TinyGoOperationAdd {
		t.Fatalf("assignment operation mapping is wrong: %+v", provider.request.Question.Options)
	}
	if result.Report.BodyFill == nil || result.Report.BodyFill.ProposedCandidateID != "multid" ||
		result.Report.BodyFill.FunctionalAccuracyPct != 100 || result.Report.RouteEquivalence.Decision != "PASS" {
		t.Fatalf("assignment result did not preserve the declared finite score and equivalence: %+v", result.Report.BodyFill)
	}
}

func TestTinyGoBodyFillRejectsUnsupportedDuplicateAndIllTypedCandidatesBeforeProvider(t *testing.T) {
	fixture := tinyBodyFillFixture(`if __GOOO_BODY_HOLE_floor__ { return input } else { return 0 }`)
	tests := []struct {
		name       string
		candidates []IRBodyFillCandidate
	}{
		{name: "unsupported root", candidates: []IRBodyFillCandidate{{ID: "a", Expression: "input != 0"}, {ID: "b", Expression: "input < 0"}}},
		{name: "duplicate operation", candidates: []IRBodyFillCandidate{{ID: "a", Expression: "input < 1"}, {ID: "b", Expression: "(input < 2)"}}},
		{name: "not a binary operation", candidates: []IRBodyFillCandidate{{ID: "a", Expression: "true"}, {ID: "b", Expression: "input < 0"}}},
		{name: "ill typed before inference", candidates: []IRBodyFillCandidate{{ID: "a", Expression: "input + true"}, {ID: "b", Expression: "input < 0"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan := tinyBodyFillPlan("Add or adjust the input.", "floor", test.candidates...)
			plan.TestCases = []IRBodyFillTestCase{{Input: -1, Expected: 0}}
			provider := &recordingTinyGoBodyFillProvider{operation: decisionroute.TinyGoOperationAdd}
			_, err := generateWithIRBodyFillOptions(context.Background(), "tiny-invalid.gooo", fixture, "Choose",
				plan, "", "", IRBodyFillOptions{}, provider)
			if err == nil || provider.calls != 0 {
				t.Fatalf("invalid candidate set reached tiny_go: err=%v calls=%d", err, provider.calls)
			}
		})
	}
}

func TestTinyGoBodyFillRejectsLayaConfigurationAndUsesDeterministicFallback(t *testing.T) {
	fixture := tinyBodyFillFixture(`if input < 0 { return __GOOO_BODY_HOLE_floor__ } else { return input }`)
	plan := tinyBodyFillPlan("Clamp negative input to zero.", "floor",
		IRBodyFillCandidate{ID: "first", Expression: "input + 0"},
		IRBodyFillCandidate{ID: "second", Expression: "input - 0"},
	)
	plan.TestCases = []IRBodyFillTestCase{{Input: -1, Expected: -1}}
	provider := &recordingTinyGoBodyFillProvider{operation: decisionroute.TinyGoOperationSubtract,
		abstain: true}
	if _, err := generateWithIRBodyFillOptions(context.Background(), "tiny.gooo", fixture, "Choose", plan,
		"http://laya.invalid", "secret", IRBodyFillOptions{}, provider); err == nil || provider.calls != 0 {
		t.Fatalf("tiny_go accepted simultaneous Laya configuration: err=%v calls=%d", err, provider.calls)
	}
	plan.ProviderModel = "english"
	if _, err := generateWithIRBodyFillOptions(context.Background(), "tiny.gooo", fixture, "Choose", plan,
		"", "", IRBodyFillOptions{}, provider); err == nil || provider.calls != 0 {
		t.Fatalf("tiny_go accepted Laya model selector: err=%v calls=%d", err, provider.calls)
	}
	plan.ProviderModel = ""
	result, err := generateWithIRBodyFillOptions(context.Background(), "tiny.gooo", fixture, "Choose", plan,
		"", "", IRBodyFillOptions{}, provider)
	if err != nil {
		t.Fatalf("abstaining tiny provider should fall back deterministically: %v", err)
	}
	fill := result.Report.BodyFill
	if fill == nil || fill.Decision.FallbackReason != decisionroute.TinyGoFallbackLowConfidence ||
		fill.ProposedCandidateID != plan.Candidates[0].ID || fill.SelectedCandidateID != plan.Candidates[0].ID ||
		fill.LocalModelPredictions != 1 || fill.ExternalProviderCalls != 0 {
		t.Fatalf("abstention did not preserve explicit local fallback accounting: %+v", fill)
	}
	if fill.Decision.TinyGoPredictedOperation != decisionroute.TinyGoOperationSubtract ||
		fill.Decision.TinyGoPredictionApplied == nil || *fill.Decision.TinyGoPredictionApplied {
		t.Fatalf("abstention did not preserve its unapplied raw operation: %+v", fill.Decision)
	}
}

func TestTinyGoUnOfferedOperationFallbackScoreIsNotModelAccuracy(t *testing.T) {
	fixture := tinyBodyFillFixture(`let result = __GOOO_BODY_HOLE_value__
return result`)
	plan := tinyBodyFillPlan("Add or subtract zero from the input.", "value",
		IRBodyFillCandidate{ID: "first", Expression: "input + 0"},
		IRBodyFillCandidate{ID: "second", Expression: "input - 0"},
	)
	plan.TestCases = []IRBodyFillTestCase{{Input: -4, Expected: -4}, {Input: 0, Expected: 0}, {Input: 9, Expected: 9}}
	provider := &recordingTinyGoBodyFillProvider{operation: decisionroute.TinyGoOperationLessEqual}
	result, err := generateWithIRBodyFillOptions(context.Background(), "tiny-unoffered.gooo", fixture, "Choose",
		plan, "", "", IRBodyFillOptions{}, provider)
	if err != nil {
		t.Fatalf("GenerateWithIRBodyFillWithOptions: %v", err)
	}
	fill := result.Report.BodyFill
	if fill == nil || fill.Decision.Mode != "deterministic_fallback" ||
		fill.Decision.FallbackReason != decisionroute.TinyGoFallbackOperationNotOffered ||
		fill.Decision.TinyGoPredictedOperation != decisionroute.TinyGoOperationLessEqual ||
		fill.Decision.TinyGoPredictionApplied == nil || *fill.Decision.TinyGoPredictionApplied {
		t.Fatalf("unoffered raw prediction was lost or treated as applied: %+v", fill)
	}
	if fill.ProposedCandidateID != plan.Candidates[0].ID || fill.ProposedAccuracyPct != 100 ||
		fill.FunctionalAccuracyPct != 100 {
		t.Fatalf("fallback's finite-suite score was not preserved separately: %+v", fill)
	}
}

func TestTinyGoBodyFillHonorsCancellationWithoutCallingProvider(t *testing.T) {
	fixture := tinyBodyFillFixture(`if input < 0 { return __GOOO_BODY_HOLE_floor__ } else { return input }`)
	plan := tinyBodyFillPlan("Clamp the input.", "floor",
		IRBodyFillCandidate{ID: "add", Expression: "input + 0"},
		IRBodyFillCandidate{ID: "subtract", Expression: "input - 0"},
	)
	plan.TestCases = []IRBodyFillTestCase{{Input: 1, Expected: 1}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	provider := &recordingTinyGoBodyFillProvider{operation: decisionroute.TinyGoOperationAdd}
	_, err := generateWithIRBodyFillOptions(ctx, "tiny.gooo", fixture, "Choose", plan,
		"", "", IRBodyFillOptions{}, provider)
	if !errors.Is(err, context.Canceled) || provider.calls != 0 {
		t.Fatalf("canceled local decision was not stopped before inference: err=%v calls=%d", err, provider.calls)
	}
}

func TestTinyGoBodyFillReturnsCancellationAfterSynchronousProviderCall(t *testing.T) {
	fixture := tinyBodyFillFixture(`if input < 0 { return __GOOO_BODY_HOLE_floor__ } else { return input }`)
	plan := tinyBodyFillPlan("Clamp the input.", "floor",
		IRBodyFillCandidate{ID: "add", Expression: "input + 0"},
		IRBodyFillCandidate{ID: "subtract", Expression: "input - 0"},
	)
	plan.TestCases = []IRBodyFillTestCase{{Input: 1, Expected: 1}}
	ctx, cancel := context.WithCancel(context.Background())
	provider := &recordingTinyGoBodyFillProvider{operation: decisionroute.TinyGoOperationAdd, afterResolve: cancel}
	result, err := generateWithIRBodyFillOptions(ctx, "tiny.gooo", fixture, "Choose", plan,
		"", "", IRBodyFillOptions{}, provider)
	if !errors.Is(err, context.Canceled) || provider.calls != 1 || result.Source != "" {
		t.Fatalf("post-inference cancellation was not returned before emission: err=%v calls=%d result=%+v", err, provider.calls, result)
	}
}

func TestTinyGoBodyFillRejectsMissingModelProvenance(t *testing.T) {
	fixture := tinyBodyFillFixture(`if input < 0 { return __GOOO_BODY_HOLE_floor__ } else { return input }`)
	plan := tinyBodyFillPlan("Choose an operation.", "floor",
		IRBodyFillCandidate{ID: "add", Expression: "input + 0"},
		IRBodyFillCandidate{ID: "subtract", Expression: "input - 0"},
	)
	provider := &recordingTinyGoBodyFillProvider{operation: decisionroute.TinyGoOperationAdd, omitWeightsSHA: true}
	result, err := generateWithIRBodyFillOptions(context.Background(), "tiny.gooo", fixture, "Choose", plan,
		"", "", IRBodyFillOptions{}, provider)
	if err == nil || provider.calls != 1 || result.Source != "" || !strings.Contains(err.Error(), "incomplete or mismatched model provenance") {
		t.Fatalf("missing model provenance was accepted: err=%v calls=%d result=%+v", err, provider.calls, result)
	}
}

func TestTinyGoBodyFillRejectsMissingMetadataProvenance(t *testing.T) {
	fixture := tinyBodyFillFixture(`if input < 0 { return __GOOO_BODY_HOLE_floor__ } else { return input }`)
	plan := tinyBodyFillPlan("Choose an operation.", "floor",
		IRBodyFillCandidate{ID: "add", Expression: "input + 0"},
		IRBodyFillCandidate{ID: "subtract", Expression: "input - 0"},
	)
	provider := &recordingTinyGoBodyFillProvider{operation: decisionroute.TinyGoOperationAdd, omitMetadataSHA: true}
	result, err := generateWithIRBodyFillOptions(context.Background(), "tiny.gooo", fixture, "Choose", plan,
		"", "", IRBodyFillOptions{}, provider)
	if err == nil || provider.calls != 1 || result.Source != "" || !strings.Contains(err.Error(), "incomplete or mismatched model provenance") {
		t.Fatalf("missing metadata provenance was accepted: err=%v calls=%d result=%+v", err, provider.calls, result)
	}
}

func TestTinyGoReceiptRejectsMalformedMetadataSHA256(t *testing.T) {
	request := decisionroute.Request{Question: decisionroute.Question{Options: []decisionroute.Option{
		{ID: "add", Operation: decisionroute.TinyGoOperationAdd},
		{ID: "subtract", Operation: decisionroute.TinyGoOperationSubtract},
	}}, Fallback: "add"}
	apply := true
	valid := decisionroute.Receipt{
		Schema: decisionroute.ReceiptSchema, Provider: decisionroute.ProviderTinyGo,
		Mode: decisionroute.ProviderTinyGo, Selected: "add",
		TinyGoPredictedOperation: decisionroute.TinyGoOperationAdd, TinyGoPredictionApplied: &apply,
		TinyGoVariant: "fp32", TinyGoWeightsSHA256: strings.Repeat("a", 64),
		TinyGoMetadataSHA256: strings.Repeat("b", 64), RequestSHA256: "request",
	}
	if !validTinyGoDecisionReceipt(valid, "request", request) {
		t.Fatal("valid metadata SHA-256 was rejected")
	}
	for name, digest := range map[string]string{
		"missing":   "",
		"short":     strings.Repeat("b", 63),
		"uppercase": strings.Repeat("B", 64),
		"nonhex":    strings.Repeat("g", 64),
	} {
		t.Run(name, func(t *testing.T) {
			candidate := valid
			candidate.TinyGoMetadataSHA256 = digest
			if validTinyGoDecisionReceipt(candidate, "request", request) {
				t.Fatal("malformed metadata SHA-256 was accepted")
			}
		})
	}
}

func TestTinyGoReceiptRequiresSupportedPredictionAndConsistentMode(t *testing.T) {
	request := decisionroute.Request{Question: decisionroute.Question{Options: []decisionroute.Option{
		{ID: "add", Operation: decisionroute.TinyGoOperationAdd},
		{ID: "subtract", Operation: decisionroute.TinyGoOperationSubtract},
	}}, Fallback: "add"}
	apply := true
	valid := decisionroute.Receipt{
		Schema: decisionroute.ReceiptSchema, Mode: decisionroute.ProviderTinyGo,
		Provider: decisionroute.ProviderTinyGo, Selected: "add",
		TinyGoPredictedOperation: decisionroute.TinyGoOperationAdd, TinyGoPredictionApplied: &apply,
		FallbackReason: "", TinyGoVariant: "fp32",
		TinyGoWeightsSHA256: strings.Repeat("a", 64), TinyGoMetadataSHA256: strings.Repeat("b", 64),
		RequestSHA256: "request",
	}
	if !validTinyGoDecisionReceipt(valid, "request", request) {
		t.Fatal("valid applied prediction was rejected")
	}
	cases := []struct {
		name   string
		change func(*decisionroute.Receipt)
	}{
		{name: "missing application flag", change: func(receipt *decisionroute.Receipt) { receipt.TinyGoPredictionApplied = nil }},
		{name: "unsupported raw operation", change: func(receipt *decisionroute.Receipt) { receipt.TinyGoPredictedOperation = "emit_source" }},
		{name: "unknown mode", change: func(receipt *decisionroute.Receipt) { receipt.Mode = "unknown" }},
		{name: "tiny mode marked unapplied", change: func(receipt *decisionroute.Receipt) { *receipt.TinyGoPredictionApplied = false }},
		{name: "tiny selected operation mismatch", change: func(receipt *decisionroute.Receipt) { receipt.Selected = "subtract" }},
		{name: "fallback marked applied", change: func(receipt *decisionroute.Receipt) {
			receipt.Mode = "deterministic_fallback"
			receipt.Selected = request.Fallback
			receipt.FallbackReason = decisionroute.TinyGoFallbackLowConfidence
		}},
		{name: "fallback without supported reason", change: func(receipt *decisionroute.Receipt) {
			*receipt.TinyGoPredictionApplied = false
			receipt.Mode = "deterministic_fallback"
			receipt.Selected = request.Fallback
			receipt.FallbackReason = "UNEXPECTED"
		}},
		{name: "unoffered reason for offered operation", change: func(receipt *decisionroute.Receipt) {
			*receipt.TinyGoPredictionApplied = false
			receipt.Mode = "deterministic_fallback"
			receipt.Selected = request.Fallback
			receipt.FallbackReason = decisionroute.TinyGoFallbackOperationNotOffered
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			candidate := valid
			value := true
			candidate.TinyGoPredictionApplied = &value
			test.change(&candidate)
			if validTinyGoDecisionReceipt(candidate, "request", request) {
				t.Fatal("inconsistent prediction receipt was accepted")
			}
		})
	}

	// An unoffered operation can be recorded only as an unapplied fallback.
	apply = false
	valid.Mode = "deterministic_fallback"
	valid.Selected = request.Fallback
	valid.FallbackReason = decisionroute.TinyGoFallbackOperationNotOffered
	valid.TinyGoPredictedOperation = decisionroute.TinyGoOperationLessEqual
	valid.TinyGoPredictionApplied = &apply
	if !validTinyGoDecisionReceipt(valid, "request", request) {
		t.Fatal("valid unoffered-operation fallback was rejected")
	}
}

type recordingTinyGoBodyFillProvider struct {
	operation       string
	calls           int
	request         decisionroute.Request
	abstain         bool
	afterResolve    func()
	omitWeightsSHA  bool
	omitMetadataSHA bool
}

func (provider *recordingTinyGoBodyFillProvider) Resolve(ctx context.Context, request decisionroute.Request) (decisionroute.Receipt, error) {
	provider.calls++
	provider.request = request
	if err := ctx.Err(); err != nil {
		return decisionroute.Receipt{}, err
	}
	digest, err := decisionroute.Validate(request)
	if err != nil {
		return decisionroute.Receipt{}, err
	}
	selected := request.Fallback
	reason := decisionroute.TinyGoFallbackLowConfidence
	predictionApplied := false
	if !provider.abstain {
		reason = decisionroute.TinyGoFallbackOperationNotOffered
		for _, option := range request.Question.Options {
			if option.Operation == provider.operation {
				selected = option.ID
				reason = ""
				predictionApplied = true
				break
			}
		}
	}
	mode := decisionroute.ProviderTinyGo
	if reason != "" {
		mode = "deterministic_fallback"
	}
	weightsSHA := strings.Repeat("a", 64)
	if provider.omitWeightsSHA {
		weightsSHA = ""
	}
	metadataSHA := strings.Repeat("b", 64)
	if provider.omitMetadataSHA {
		metadataSHA = ""
	}
	receipt := decisionroute.Receipt{
		Schema: decisionroute.ReceiptSchema, Mode: mode, Provider: decisionroute.ProviderTinyGo,
		Selected: selected, FallbackReason: reason, TinyGoVariant: "fp32",
		TinyGoPredictedOperation: provider.operation, TinyGoPredictionApplied: &predictionApplied,
		TinyGoWeightsSHA256: weightsSHA, TinyGoMetadataSHA256: metadataSHA, RequestSHA256: digest,
	}
	if provider.afterResolve != nil {
		provider.afterResolve()
	}
	return receipt, nil
}

func tinyBodyFillFixture(body string) []byte {
	body = strings.ReplaceAll(body, "\n", `\n`)
	return []byte("package sample\nnamespace sample\n" +
		"entity Integer id \"sample://entity/integer\"\n" +
		"activity Choose(Integer) -> Integer computes \"" + body + "\"\n")
}

func tinyBodyFillPlan(intent, holeID string, candidates ...IRBodyFillCandidate) IRBodyFillPlan {
	return IRBodyFillPlan{Schema: bodyFillPlanSchema, Intent: intent, HoleID: holeID,
		Candidates: candidates, TestCases: []IRBodyFillTestCase{{Input: -1, Expected: 0}}}
}

func bodyFillDimension(result Result, id string) CompletenessDimension {
	for _, dimension := range result.Report.CompletenessReceipt.Dimensions {
		if dimension.ID == id {
			return dimension
		}
	}
	return CompletenessDimension{}
}
