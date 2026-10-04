package bodycodegen

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
)

type bodyRouteChoice struct {
	selected  string
	receipt   decisionroute.Receipt
	selection RouteSelectionReceipt
	latencyMS float64
	ids       []string
}

func chooseBodyRoute(ctx context.Context, p preparedBody, source []byte, endpoint, apiKey, seed string) (bodyRouteChoice, error) {
	choice := bodyRouteChoice{selected: preserveRoute, receipt: decisionroute.Receipt{
		Schema: decisionroute.ReceiptSchema, Mode: "deterministic_fallback", Selected: preserveRoute,
		FallbackReason: "NO_ALTERNATIVE_ROUTE", Provider: "deterministic"},
		selection: RouteSelectionReceipt{Method: "single_eligible_route", ProposedRoute: preserveRoute, FinalRoute: preserveRoute}}
	if seed != "" {
		choice.selection.SeedSHA256 = digest([]byte(seed))
		choice.selection.Weights = map[string]float64{preserveRoute: 1}
	}
	routes, shape, err := candidateRoutes(p.packageName, p.activity.Name, p.body)
	if err != nil {
		return choice, err
	}
	choice.ids = make([]string, 0, len(routes))
	for _, option := range routes {
		choice.ids = append(choice.ids, option.ID)
	}
	if len(routes) == 1 {
		return choice, nil
	}
	request := bodyRouteRequest(p, source, routes, shape)
	started := time.Now()
	decisionContext, cancel := context.WithTimeout(ctx, routeDecisionBudget)
	choice.receipt, err = decisionroute.Resolve(decisionContext, request, endpoint, apiKey)
	cancel()
	choice.latencyMS = float64(time.Since(started)) / float64(time.Millisecond)
	if err != nil {
		return choice, fmt.Errorf("select code generation route: %w", err)
	}
	err = choice.selectRankedRoute(seed, routes)
	return choice, err
}

func bodyRouteRequest(p preparedBody, source []byte, routes []decisionroute.Option, shape string) decisionroute.Request {
	inputs := p.activity.Inputs
	inputDescription := inputs[0].Name
	if len(inputs) > 1 {
		names := make([]string, len(inputs))
		for i, input := range inputs {
			names[i] = input.Name
		}
		inputDescription = "(" + strings.Join(names, ",") + ")"
	}
	return decisionroute.Request{Schema: decisionroute.RequestSchema,
		State: fmt.Sprintf("activity=%s; source_sha256=%s; program_sha256=%s; input=%s; output=%s; body_shape=%s; source_semantic_units=%d",
			p.activity.Name, digest(source), digest([]byte(p.activity.ValueProgram)), inputDescription,
			p.activity.Output, shape, p.base.report.SourceSemanticUnits),
		Question: decisionroute.Question{ID: "body_codegen_route",
			Instructions: "Choose one listed, semantics-preserving lowering route for the described Gooo activity shape. Do not invent code or routes. Prefer the route whose generated control flow is clearest for this shape.",
			Options:      routes}, Fallback: preserveRoute}
}

func (choice *bodyRouteChoice) selectRankedRoute(seed string, routes []decisionroute.Option) error {
	choice.selected = choice.receipt.Selected
	if seed != "" {
		weights := choice.receipt.Probabilities
		if choice.receipt.Mode == "laya" && len(weights) == 0 && choice.receipt.Selected != "" {
			weights = map[string]float64{choice.receipt.Selected: 1}
		}
		selection, err := sampleRoute(seed, choice.receipt.RequestSHA256, routes, weights, choice.receipt.Selected)
		if err != nil {
			return fmt.Errorf("sample code generation route: %w", err)
		}
		choice.selection, choice.selected = selection, selection.ProposedRoute
	} else if choice.receipt.Mode == "laya" {
		choice.selection = RouteSelectionReceipt{Method: "laya_top1", ProposedRoute: choice.selected, FinalRoute: choice.selected}
	} else {
		choice.selection = RouteSelectionReceipt{Method: "deterministic_fallback", ProposedRoute: choice.selected, FinalRoute: choice.selected}
	}
	return nil
}
