package bodycodegen

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateLowersLetAssignmentAndConditionalReturn(t *testing.T) {
	source := []byte(`package bodycodegen
namespace bodycodegen
entity Integer id "bodycodegen://entity/integer"
activity ClampBelowZero(Integer) -> Integer computes "let value = input\nvalue = input\nlet accepted = value >= 0\nif accepted { return value } else { return 0 }"
`)
	result, err := Generate("main.gooo", source, "ClampBelowZero")
	if err != nil {
		t.Fatal(err)
	}
	if result.Report.Decision != "PASS" || !result.Report.TypecheckPassed || !result.Report.DeterministicReplay {
		t.Fatalf("report does not prove accepted lowering: %#v", result.Report)
	}
	if result.Report.CompletenessPercent != 100 || result.Report.SourceConstructs != result.Report.LoweredConstructs {
		t.Fatalf("accepted constructs are not fully accounted for: %#v", result.Report)
	}
	if result.Report.Route != preserveRoute || result.Report.RouteDecision.FallbackReason != "NO_ALTERNATIVE_ROUTE" || result.Report.SourceSemanticUnits != result.Report.LoweredSemanticUnits {
		t.Fatalf("single-route report is incomplete: %#v", result.Report)
	}
	for _, want := range []string{
		`//gooo:generated:start id="bodycodegen://activity/clamp-below-zero" kind="activity"`,
		"func ClampBelowZero(input int64) int64",
		"var value = input",
		"value = input",
		"if accepted {",
	} {
		if !strings.Contains(result.Source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, result.Source)
		}
	}
	if result.Report.GeneratedDigest != result.Report.ReplayDigest || result.Report.RepositoryWrites != 0 {
		t.Fatalf("output replay or write boundary is incorrect: %#v", result.Report)
	}
	if result.Report.RouteEquivalence.Decision != "PASS" || !result.Report.RouteEquivalence.Equivalent ||
		result.Report.RouteEquivalence.SourceSemanticDigest != result.Report.RouteEquivalence.GeneratedSemanticDigest {
		t.Fatalf("source-preserving route lacks matching semantic digests: %#v", result.Report.RouteEquivalence)
	}
	assertCompletenessReceipt(t, result.Report.CompletenessReceipt)
	planSHA, planOK := result.Report.CompletenessReceipt.Scope["plan_sha256"].(string)
	compilerSHA, compilerOK := result.Report.CompletenessReceipt.Scope["compiler_source_sha"].(string)
	if !planOK || !compilerOK || result.Report.PlanSHA256 != planSHA || result.Report.CompilerSourceSHA != compilerSHA {
		t.Fatal("top-level report identities are not bound to the completeness receipt scope")
	}
}

func TestCompletenessReceiptPreservesPartialSourceCoverageAndUnknowns(t *testing.T) {
	receipt := buildCompletenessReceipt(Report{
		Decision: "PASS", Activity: "Partial", ActivityID: "sample://activity/partial",
		SourceSemanticUnits: 4, LoweredSemanticUnits: 3,
	}, "")
	assertCompletenessReceipt(t, receipt)
	dimension := completenessDimensionByID(t, receipt, "source_ast_coverage")
	if dimension.Status != "PROGRESS" || dimension.Numerator != 3 || dimension.Denominator != 4 {
		t.Fatalf("partial source coverage was not preserved: %#v", dimension)
	}
	for _, id := range []string{"execution_boundary", "reverse_observation_coverage", "use_case_coverage", "semantic_profile_delta"} {
		if got := completenessDimensionByID(t, receipt, id).Status; got != "UNKNOWN" {
			t.Fatalf("%s status = %q, want UNKNOWN", id, got)
		}
	}
	if receipt.Decision == "PASS_WITHIN_DECLARED_FIXTURE_SCOPE" {
		t.Fatal("incomplete core source coverage received a scoped PASS")
	}
}

func TestFailedBodyCodegenRetainsFailureAndFirstUnresolvedStage(t *testing.T) {
	receipt := FailureCompletenessReceipt("Clamp", []byte("invalid source"), "parse .gooo source failed")
	assertCompletenessReceipt(t, receipt)
	if receipt.Decision != "FAIL_CLOSED" || receipt.FailClosedReason == nil || *receipt.FailClosedReason != "parse .gooo source failed" {
		t.Fatalf("failure was not retained: %#v", receipt)
	}
	if receipt.FirstUnresolved == nil || receipt.FirstUnresolved.ID != "declaration_coverage" ||
		receipt.FirstUnresolved.NextOperation == "" {
		t.Fatalf("first unresolved stage was not retained: %#v", receipt.FirstUnresolved)
	}
}

func assertCompletenessReceipt(t *testing.T, receipt *CompletenessReceipt) {
	t.Helper()
	if receipt == nil || receipt.Schema != completenessReceiptSchema || receipt.ProfileID == "" || receipt.DecisionBasis == "" {
		t.Fatalf("shared completeness receipt is missing or malformed: %#v", receipt)
	}
	if receipt.AggregateCompletenessScore != nil {
		t.Fatalf("aggregate completeness must remain unscored: %#v", receipt.AggregateCompletenessScore)
	}
	counts := map[string]int{"PASS": 0, "PROGRESS": 0, "UNKNOWN": 0, "FAIL_CLOSED": 0}
	for _, dimension := range receipt.Dimensions {
		counts[dimension.Status]++
		if dimension.ID == "" || dimension.Unit == "" || dimension.Reason == "" || len(dimension.Evidence) == 0 {
			t.Fatalf("dimension lacks reasoned evidence: %#v", dimension)
		}
	}
	for status, count := range counts {
		if receipt.StatusCounts[status] != count {
			t.Fatalf("status_counts[%s]=%d, dimension count=%d", status, receipt.StatusCounts[status], count)
		}
	}
	if len(receipt.UnresolvedClaims) == 0 {
		t.Fatal("the non-executing body-codegen profile must retain unresolved dimensions")
	}
	if receipt.FirstUnresolved == nil || *receipt.FirstUnresolved != receipt.UnresolvedClaims[0] {
		t.Fatalf("first_unresolved does not match the first unresolved claim: %#v", receipt.FirstUnresolved)
	}
	for _, claim := range receipt.UnresolvedClaims {
		if claim.ID == "" || claim.Reason == "" || claim.NextOperation == "" {
			t.Fatalf("unresolved claim lacks a reason or next operation: %#v", claim)
		}
	}
}

func completenessDimensionByID(t *testing.T, receipt *CompletenessReceipt, id string) CompletenessDimension {
	t.Helper()
	for _, dimension := range receipt.Dimensions {
		if dimension.ID == id {
			return dimension
		}
	}
	t.Fatalf("missing completeness dimension %q", id)
	return CompletenessDimension{}
}

func TestGenerateWithPlannerUsesLayaOnlyForBoundedEquivalentRoutes(t *testing.T) {
	const sourceText = `package bodycodegen
namespace bodycodegen
entity Integer id "bodycodegen://entity/integer"
activity Choose(Integer) -> Integer computes "if input > 5 { return input + 2 } else { return 0 }"
`
	const modelRevision = "0123456789abcdef0123456789abcdef01234567"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/health" {
			_ = json.NewEncoder(writer).Encode(map[string]any{"revisions": map[string]string{"route-model": modelRevision}})
			return
		}
		if request.URL.Path != "/v1/systemone" || request.Method != http.MethodPost {
			http.NotFound(writer, request)
			return
		}
		var payload map[string]any
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Errorf("decode Laya request: %v", err)
			return
		}
		encoded, _ := json.Marshal(payload)
		if strings.Contains(string(encoded), "input + 2") {
			t.Errorf("raw activity source was sent to Laya: %s", encoded)
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"model":   "route-model",
			"routing": map[string]any{"model": "route-model"},
			"answers": map[string]any{"body_codegen_route": map[string]any{
				"choice":        guardReturnRoute,
				"probabilities": map[string]float64{preserveRoute: 0.08, guardReturnRoute: 0.92},
				"confidence":    0.92,
			}},
		})
	}))
	defer server.Close()

	result, err := GenerateWithPlanner(context.Background(), "main.gooo", []byte(sourceText), "Choose", server.URL+"/v1/systemone", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Report.Route != guardReturnRoute || result.Report.RouteDecision.Mode != "laya" || result.Report.RouteDecision.ModelRevision != modelRevision {
		t.Fatalf("Laya decision was not bound to code generation: %#v", result.Report)
	}
	if len(result.Report.CandidateRoutes) != 3 || result.Report.CandidateRoutes[0] != preserveRoute || result.Report.CandidateRoutes[1] != guardReturnRoute || result.Report.CandidateRoutes[2] != mergeResultRoute {
		t.Fatalf("unexpected route candidates: %#v", result.Report.CandidateRoutes)
	}
	if strings.Contains(result.Source, "} else {") || !strings.Contains(result.Source, "\n\treturn 0\n}") {
		t.Fatalf("guard-return route was not emitted:\n%s", result.Source)
	}
	if result.Report.EquivalenceRule != "if-return-else-return-to-guard-return-v1" || result.Report.SourceSemanticUnits != result.Report.LoweredSemanticUnits || result.Report.CompletenessPercent != 100 ||
		result.Report.RouteEquivalence.Decision != "PASS" || !result.Report.RouteEquivalence.Equivalent ||
		result.Report.RouteEquivalence.SourceSemanticDigest != result.Report.RouteEquivalence.GeneratedSemanticDigest ||
		result.Report.RouteEquivalence.Rule != result.Report.EquivalenceRule {
		t.Fatalf("lowering completeness or equivalence witness is missing: %#v", result.Report)
	}
}

func TestGenerateWithPlannerCanSelectExplicitResultJoin(t *testing.T) {
	const sourceText = `package bodycodegen
namespace bodycodegen
entity Integer id "bodycodegen://entity/integer"
activity Choose(Integer) -> Integer computes "if input > 5 { return input + 2 } else { return 0 }"
`
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"model": "route-model",
			"answers": map[string]any{"body_codegen_route": map[string]any{
				"choice": mergeResultRoute,
			}},
		})
	}))
	defer server.Close()

	result, err := GenerateWithPlanner(context.Background(), "main.gooo", []byte(sourceText), "Choose", server.URL+"/v1/systemone", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Report.Route != mergeResultRoute || result.Report.RouteDecision.Mode != "laya" {
		t.Fatalf("explicit result join was not selected: %#v", result.Report)
	}
	for _, want := range []string{"var _goooResult int64", "_goooResult = input + 2", "_goooResult = 0", "return _goooResult"} {
		if !strings.Contains(result.Source, want) {
			t.Fatalf("merge-result output missing %q:\n%s", want, result.Source)
		}
	}
	if result.Report.EquivalenceRule != "if-return-else-return-to-explicit-result-join-v1" || result.Report.CompletenessPercent != 100 || result.Report.SourceSemanticUnits == 0 || result.Report.LoweredSemanticUnits < result.Report.SourceSemanticUnits ||
		result.Report.RouteEquivalence.Decision != "PASS" || !result.Report.RouteEquivalence.Equivalent ||
		result.Report.RouteEquivalence.SourceSemanticDigest != result.Report.RouteEquivalence.GeneratedSemanticDigest ||
		result.Report.RouteEquivalence.Rule != result.Report.EquivalenceRule {
		t.Fatalf("result-join report is incomplete: %#v", result.Report)
	}
}

func TestRouteEquivalenceFailsClosedWhenBranchSemanticsDiffer(t *testing.T) {
	const sourceBody = "if input > 5 { return input + 2 } else { return 0 }"
	generated := []byte(`package bodycodegen
func Choose(input int64) int64 {
	if input > 5 {
		return input + 2
	}
	return 1
}
`)
	receipt, err := routeEquivalence("bodycodegen", "Choose", "int64", "int64", sourceBody, generated, "if-return-else-return-to-guard-return-v1")
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != "FAIL_CLOSED" || receipt.Equivalent || receipt.SourceSemanticDigest == receipt.GeneratedSemanticDigest {
		t.Fatalf("different branch behavior received an equivalence pass: %#v", receipt)
	}
}

func TestGenerateWithPlannerFallsBackDeterministicallyWithoutLaya(t *testing.T) {
	source := []byte(`package bodycodegen
namespace bodycodegen
entity Integer id "bodycodegen://entity/integer"
activity Choose(Integer) -> Integer computes "if input > 5 { return input + 2 } else { return 0 }"
`)
	first, err := GenerateWithPlanner(context.Background(), "main.gooo", source, "Choose", "", "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateWithPlanner(context.Background(), "main.gooo", source, "Choose", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if first.Report.Route != preserveRoute || first.Report.RouteDecision.Mode != "deterministic_fallback" || first.Report.RouteDecision.FallbackReason != "NOT_CONFIGURED" {
		t.Fatalf("missing deterministic fallback receipt: %#v", first.Report)
	}
	if first.Source != second.Source || first.Report.GeneratedDigest != second.Report.GeneratedDigest || first.Report.RouteDecision.RequestSHA256 != second.Report.RouteDecision.RequestSHA256 {
		t.Fatalf("unconfigured route selection was not deterministic: first=%#v second=%#v", first.Report, second.Report)
	}
}

func TestGenerateFailsClosedForUnsupportedOrIncompletePrograms(t *testing.T) {
	tests := []struct {
		name, program string
	}{
		{name: "call", program: `return helper(input)`},
		{name: "missing else", program: `if input > 0 { return input } return 0`},
		{name: "missing return", program: `let value = input`},
		{name: "type mismatch", program: `return input >= 0`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := []byte("package bodycodegen\nnamespace bodycodegen\nentity Integer id \"bodycodegen://entity/integer\"\nactivity Clamp(Integer) -> Integer computes \"" + test.program + "\"\n")
			if result, err := Generate("main.gooo", source, "Clamp"); err == nil || result.Source != "" {
				t.Fatalf("unsupported program produced output: result=%#v err=%v", result, err)
			}
		})
	}
}
