package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
)

func TestSampleRouteIsSeededAndBoundedToEligibleRoutes(t *testing.T) {
	options := []decisionroute.Option{
		{ID: preserveRoute, Description: "preserve"},
		{ID: guardReturnRoute, Description: "guard"},
		{ID: mergeResultRoute, Description: "join"},
	}
	seen := make(map[string]bool)
	for index := range 96 {
		seed := fmt.Sprintf("body-codegen-seed-%03d", index)
		first, err := sampleRoute(seed, "sha256:request", options, nil, "")
		if err != nil {
			t.Fatal(err)
		}
		replay, err := sampleRoute(seed, "sha256:request", options, nil, "")
		if err != nil {
			t.Fatal(err)
		}
		if first.ProposedRoute != replay.ProposedRoute || first.DrawHex != replay.DrawHex || first.Weights[preserveRoute] != 1.0/3.0 {
			t.Fatalf("seeded route selection did not replay: first=%#v replay=%#v", first, replay)
		}
		seen[first.ProposedRoute] = true
	}
	if len(seen) != len(options) {
		t.Fatalf("seeded distribution did not cover all eligible routes: %#v", seen)
	}
}

func TestGenerateWithPlannerAndSampleSeedUsesLayaWeights(t *testing.T) {
	const sourceText = `package bodycodegen
namespace bodycodegen
entity Integer id "bodycodegen://entity/integer"
activity Choose(Integer) -> Integer computes "if input > 5 { return input + 2 } else { return 0 }"
`
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/systemone" || request.Method != http.MethodPost {
			http.NotFound(writer, request)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"model": "route-model",
			"answers": map[string]any{"body_codegen_route": map[string]any{
				// Deliberately make the top-1 answer disagree with the distribution.
				"choice": mergeResultRoute,
				"probabilities": map[string]float64{
					preserveRoute: 1, guardReturnRoute: 0, mergeResultRoute: 0,
				},
			}},
		})
	}))
	defer server.Close()

	result, err := GenerateWithPlannerAndSampleSeed(context.Background(), "main.gooo", []byte(sourceText), "Choose", server.URL+"/v1/systemone", "", "replay-seed")
	if err != nil {
		t.Fatal(err)
	}
	if result.Report.Route != preserveRoute || result.Report.RouteDecision.Mode != "laya" || result.Report.RouteDecision.Provider != "laya" || result.Report.RouteDecision.Selected != mergeResultRoute {
		t.Fatalf("probability draw did not control the route: %#v", result.Report)
	}
	selection := result.Report.RouteSelection
	if selection.Method != seededSelectionV1 || selection.ProposedRoute != preserveRoute || selection.FinalRoute != preserveRoute || selection.DrawHex == "" || selection.SeedSHA256 != digest([]byte("replay-seed")) {
		t.Fatalf("seeded draw receipt is incomplete: %#v", selection)
	}
	if selection.Weights[preserveRoute] != 1 || selection.Weights[guardReturnRoute] != 0 || selection.Weights[mergeResultRoute] != 0 {
		t.Fatalf("receipt did not retain normalized Laya weights: %#v", selection.Weights)
	}
}

func TestGenerateWithPlannerSampleIsDeterministicWithoutLaya(t *testing.T) {
	source := []byte(`package bodycodegen
namespace bodycodegen
entity Integer id "bodycodegen://entity/integer"
activity Choose(Integer) -> Integer computes "if input > 5 { return input + 2 } else { return 0 }"
`)
	first, err := GenerateWithPlannerAndSampleSeed(context.Background(), "main.gooo", source, "Choose", "", "", "cohort-7")
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateWithPlannerAndSampleSeed(context.Background(), "main.gooo", source, "Choose", "", "", "cohort-7")
	if err != nil {
		t.Fatal(err)
	}
	if first.Report.Route != second.Report.Route || first.Report.GeneratedDigest != second.Report.GeneratedDigest || first.Report.RouteSelection.DrawHex != second.Report.RouteSelection.DrawHex {
		t.Fatalf("sampled route changed with identical source and seed: first=%#v second=%#v", first.Report, second.Report)
	}
	if first.Report.RouteDecision.Provider != "deterministic" || first.Report.RouteDecision.Mode != "deterministic_fallback" || first.Report.RouteDecision.FallbackReason != "NOT_CONFIGURED" || first.Report.RouteDecision.Selected != preserveRoute {
		t.Fatalf("offline sampling did not retain deterministic-provider evidence: %#v", first.Report.RouteDecision)
	}
	encoded, err := json.Marshal(first.Report.RouteSelection)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "cohort-7") {
		t.Fatalf("raw sample seed leaked into the receipt: %s", encoded)
	}
}
