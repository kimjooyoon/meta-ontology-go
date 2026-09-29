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
	if result.Report.CandidateRoutes[0] != preserveRoute || result.Report.CandidateRoutes[1] != guardReturnRoute {
		t.Fatalf("unexpected route candidates: %#v", result.Report.CandidateRoutes)
	}
	if strings.Contains(result.Source, "} else {") || !strings.Contains(result.Source, "\n\treturn 0\n}") {
		t.Fatalf("guard-return route was not emitted:\n%s", result.Source)
	}
	if result.Report.EquivalenceRule != "if-return-else-return-to-guard-return-v1" || result.Report.SourceSemanticUnits != result.Report.LoweredSemanticUnits || result.Report.CompletenessPercent != 100 {
		t.Fatalf("lowering completeness or equivalence witness is missing: %#v", result.Report)
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
