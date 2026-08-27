package causality

import (
	"encoding/json"
	"fmt"
)

func Evaluate(input []byte, expectedMode string) (Receipt, error) {
	report, mode, err := parseInputReport(input, expectedMode)
	if err != nil {
		return Receipt{}, err
	}

	registered := report.OperationClaimTransitions[:ClaimTotal]
	resolved := report.OperationClaimTransitions[ClaimTotal:]
	claimIDs := make([]string, ClaimTotal)
	for index := range registered {
		claimIDs[index] = registered[index].ClaimID
	}
	graph, err := buildGraph(claimIDs)
	if err != nil {
		return Receipt{}, err
	}

	unavailable := make([]bool, ClaimTotal)
	for index := range resolved {
		unavailable[index] = isUnavailableEvent(resolved[index].Event)
	}
	resolutions := make([]Resolution, 0, ClaimTotal)
	for index, transition := range resolved {
		if !unavailable[index] {
			resolutions = append(resolutions, Resolution{
				ClaimID:       claimIDs[index],
				Axis:          claimAxes[index],
				State:         "DISCHARGED",
				Kind:          "EVIDENCE_ACCEPTED",
				ObservedEvent: transition.Event,
				Coordinate:    transition.Coordinate,
			})
			continue
		}

		blockingEdges := incomingUnavailableEdges(graph, claimIDs[index], unavailable, claimIDs)
		if len(blockingEdges) == 0 {
			coordinate := nonEmptyCoordinate(transition.Coordinate)
			resolutions = append(resolutions, Resolution{
				ClaimID:               claimIDs[index],
				Axis:                  claimAxes[index],
				State:                 "OPEN",
				Kind:                  "DIRECT_MISSING",
				ObservedEvent:         transition.Event,
				Coordinate:            coordinate,
				MissingEvidenceIDs:    []string{missingEvidenceID(claimIDs[index])},
				CausePath:             []string{claimIDs[index]},
				CauseTransitionDigest: transition.TransitionDigest,
				CauseCoordinate:       &coordinate,
			})
			continue
		}

		pathIndexes := primaryCausePath(index, unavailable)
		rootIndex := pathIndexes[0]
		causeCoordinate := nonEmptyCoordinate(resolved[rootIndex].Coordinate)
		resolution := Resolution{
			ClaimID:               claimIDs[index],
			Axis:                  claimAxes[index],
			State:                 "OPEN",
			Kind:                  "DEPENDENCY_BLOCKED",
			ObservedEvent:         transition.Event,
			Coordinate:            Coordinate{Stage: "RESOLVE_DEPENDENCY", Step: claimAxes[index], Reason: "UPSTREAM_CLAIM_OPEN"},
			MissingEvidenceIDs:    []string{missingEvidenceID(claimIDs[rootIndex])},
			CauseTransitionDigest: resolved[rootIndex].TransitionDigest,
			CauseCoordinate:       &causeCoordinate,
		}
		for _, edge := range blockingEdges {
			resolution.BlockedByClaimIDs = append(resolution.BlockedByClaimIDs, edge.FromClaimID)
			resolution.BlockedByEdgeIDs = append(resolution.BlockedByEdgeIDs, edge.EdgeID)
		}
		for _, pathIndex := range pathIndexes {
			resolution.CausePath = append(resolution.CausePath, claimIDs[pathIndex])
		}
		resolutions = append(resolutions, resolution)
	}

	metrics := deriveMetrics(graph, resolutions)
	receipt := Receipt{
		Schema: ReceiptSchema,
		Scope:  ReceiptScope,
		Subject: Subject{
			InputReportSchema:      report.Schema,
			InputReportDigest:      digestBytes(input),
			TransitionHeadDigest:   report.transitionHead(),
			GraphDigest:            graph.Digest,
			BindingStatus:          "UNKNOWN",
			MissingBindingEvidence: []string{"source_digest", "semantic_ir_digest", "source_ir_binding_digest"},
		},
		Graph:       graph,
		Metrics:     metrics,
		Indicators:  buildIndicators(metrics),
		Resolutions: resolutions,
		Decision:    decisionFor(mode, metrics),
	}
	receipt.ReceiptDigest, err = receiptDigest(receipt)
	if err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}

func parseInputReport(input []byte, expectedMode string) (inputReport, string, error) {
	var report inputReport
	if err := json.Unmarshal(input, &report); err != nil {
		return inputReport{}, "", fmt.Errorf("decode operation catalog report: %w", err)
	}
	mode, err := validateInputReport(report)
	if err != nil {
		return inputReport{}, "", err
	}
	if expectedMode != "" && expectedMode != mode {
		return inputReport{}, "", fmt.Errorf("report mode: got %q want %q", mode, expectedMode)
	}
	return report, mode, nil
}

func validateInputReport(report inputReport) (string, error) {
	if report.Schema != InputReportSchema {
		return "", fmt.Errorf("report schema: got %q want %q", report.Schema, InputReportSchema)
	}
	if len(report.OperationClaimTransitions) != TransitionTotal {
		return "", fmt.Errorf("transition total: got %d want %d", len(report.OperationClaimTransitions), TransitionTotal)
	}
	seen := make(map[string]struct{}, ClaimTotal)
	for index := 0; index < ClaimTotal; index++ {
		registered := report.OperationClaimTransitions[index]
		resolved := report.OperationClaimTransitions[index+ClaimTotal]
		if registered.Sequence != index+1 || resolved.Sequence != index+ClaimTotal+1 {
			return "", fmt.Errorf("transition sequence mismatch at claim %d", index+1)
		}
		if !isRegisteredEvent(registered.Event) {
			return "", fmt.Errorf("claim %d registration event: %q", index+1, registered.Event)
		}
		if registered.ClaimID == "" || registered.ClaimID != resolved.ClaimID {
			return "", fmt.Errorf("claim %d transition identity mismatch", index+1)
		}
		if _, exists := seen[registered.ClaimID]; exists {
			return "", fmt.Errorf("duplicate claim id %q", registered.ClaimID)
		}
		seen[registered.ClaimID] = struct{}{}
		if registered.TransitionDigest == "" || resolved.TransitionDigest == "" {
			return "", fmt.Errorf("claim %d transition digest missing", index+1)
		}
	}
	accepted := 0
	unavailable := 0
	for _, transition := range report.OperationClaimTransitions[ClaimTotal:] {
		switch {
		case isAcceptedEvent(transition.Event):
			accepted++
		case isUnavailableEvent(transition.Event):
			unavailable++
		default:
			return "", fmt.Errorf("unsupported resolution event %q", transition.Event)
		}
	}
	switch {
	case accepted == ClaimTotal && unavailable == 0:
		return ModeSuccess, nil
	case accepted == 0 && unavailable == ClaimTotal:
		return ModeUnknown, nil
	default:
		return "", fmt.Errorf("mixed claim resolution is outside v1 contract: accepted=%d unavailable=%d", accepted, unavailable)
	}
}

func (report inputReport) transitionHead() string {
	if report.OperationClaimTransitionHead != "" {
		return report.OperationClaimTransitionHead
	}
	return report.OperationClaimTransitionHeadDigest
}

func isRegisteredEvent(event string) bool {
	return event == "CLAIM_REGISTERED" || event == "REGISTERED"
}

func isAcceptedEvent(event string) bool {
	return event == "EVIDENCE_ACCEPTED"
}

func isUnavailableEvent(event string) bool {
	return event == "EVIDENCE_UNAVAILABLE"
}

func incomingUnavailableEdges(graph GraphContract, claimID string, unavailable []bool, claimIDs []string) []GraphEdge {
	indexes := make(map[string]int, len(claimIDs))
	for index, id := range claimIDs {
		indexes[id] = index
	}
	var edges []GraphEdge
	for _, edge := range graph.Edges {
		if edge.ToClaimID == claimID && unavailable[indexes[edge.FromClaimID]] {
			edges = append(edges, edge)
		}
	}
	return edges
}

func primaryCausePath(index int, unavailable []bool) []int {
	for _, edge := range edgeContract {
		if edge.to == index && unavailable[edge.from] {
			path := primaryCausePath(edge.from, unavailable)
			return append(path, index)
		}
	}
	return []int{index}
}

func missingEvidenceID(claimID string) string {
	return "evidence:" + claimID
}

func nonEmptyCoordinate(coordinate Coordinate) Coordinate {
	if coordinate.Stage == "" {
		coordinate.Stage = "RESOLVE"
	}
	if coordinate.Step == "" {
		coordinate.Step = "resolve-operation-spec"
	}
	if coordinate.Reason == "" {
		coordinate.Reason = "VALUE_PROGRAM_UNKNOWN"
	}
	return coordinate
}

func deriveMetrics(graph GraphContract, resolutions []Resolution) Metrics {
	metrics := Metrics{
		ContractClaimTotal:   graph.NodeTotal,
		ContractEdgeTotal:    graph.EdgeTotal,
		ClassifiedClaimTotal: len(resolutions),
	}
	for _, resolution := range resolutions {
		switch resolution.Kind {
		case "EVIDENCE_ACCEPTED":
			metrics.DischargedClaimTotal++
		case "DIRECT_MISSING":
			metrics.UnknownClaimTotal++
			metrics.DirectMissingClaimTotal++
		case "DEPENDENCY_BLOCKED":
			metrics.UnknownClaimTotal++
			metrics.DependencyBlockedClaimTotal++
			metrics.ObservedBlockingEdgeTotal += len(resolution.BlockedByEdgeIDs)
		}
		depth := len(resolution.CausePath) - 1
		if depth > metrics.MaximumCausePathDepth {
			metrics.MaximumCausePathDepth = depth
		}
	}
	if graph.NodeTotal > 0 {
		metrics.ClassificationBasisPoints = metrics.ClassifiedClaimTotal * 10000 / graph.NodeTotal
		metrics.DischargeBasisPoints = metrics.DischargedClaimTotal * 10000 / graph.NodeTotal
	}
	return metrics
}

func buildIndicators(metrics Metrics) []Indicator {
	return []Indicator{
		{IndicatorID: "claim-contract-node-total", Class: "DRIVER", Trilemma: "FOUNDATION", Value: metrics.ContractClaimTotal, Target: ClaimTotal, Comparator: "EQ", Satisfied: metrics.ContractClaimTotal == ClaimTotal},
		{IndicatorID: "claim-contract-edge-total", Class: "DRIVER", Trilemma: "FOUNDATION", Value: metrics.ContractEdgeTotal, Target: EdgeTotal, Comparator: "EQ", Satisfied: metrics.ContractEdgeTotal == EdgeTotal},
		{IndicatorID: "claim-causally-classified-total", Class: "OUTCOME", Trilemma: "COHERENCE", Value: metrics.ClassifiedClaimTotal, Target: ClaimTotal, Comparator: "EQ", Satisfied: metrics.ClassifiedClaimTotal == ClaimTotal},
		{IndicatorID: "claim-direct-missing-total", Class: "GUARDRAIL", Trilemma: "REGRESSION", Value: metrics.DirectMissingClaimTotal, Target: 0, Comparator: "EQ", Satisfied: metrics.DirectMissingClaimTotal == 0},
		{IndicatorID: "claim-dependency-blocked-total", Class: "GUARDRAIL", Trilemma: "REGRESSION", Value: metrics.DependencyBlockedClaimTotal, Target: 0, Comparator: "EQ", Satisfied: metrics.DependencyBlockedClaimTotal == 0},
		{IndicatorID: "semantic-promotion-authority-total", Class: "GUARDRAIL", Trilemma: "FOUNDATION", Value: 0, Target: 0, Comparator: "EQ", Satisfied: true},
	}
}

func decisionFor(mode string, metrics Metrics) Decision {
	if mode == ModeSuccess && metrics.DischargedClaimTotal == ClaimTotal && metrics.UnknownClaimTotal == 0 {
		return Decision{Value: "PASS", Resolution: "CAUSAL_CLASSIFICATION_EXACT", Reason: "ALL_CLAIMS_EVIDENCE_ACCEPTED", SemanticPromotionAuthorized: false}
	}
	return Decision{Value: "FAIL_CLOSED", Resolution: "DEPENDENCY_LOCAL", Reason: "DIRECT_EVIDENCE_MISSING", SemanticPromotionAuthorized: false}
}

func receiptDigest(receipt Receipt) (string, error) {
	receipt.ReceiptDigest = ""
	return digestJSON(receipt)
}
