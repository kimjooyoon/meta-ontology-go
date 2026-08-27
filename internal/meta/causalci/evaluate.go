package causalci

import (
	"fmt"
	"sort"
	"strings"
)

func Evaluate(inputRaw []byte, sourcePath string, source []byte) (Receipt, error) {
	input, err := decodeInput(inputRaw)
	if err != nil {
		return Receipt{}, err
	}
	if input.SourcePath != sourcePath || len(source) == 0 {
		return Receipt{}, fmt.Errorf("%s: source binding", ReasonMalformedInput)
	}
	transitions, head, err := buildTransitions(input.ClaimTransitions)
	if err != nil {
		return Receipt{}, err
	}
	cases := make([]CaseReceipt, 0, len(input.Cases))
	for _, value := range input.Cases {
		result, err := evaluateCase(input.Policy, value)
		if err != nil {
			return Receipt{}, err
		}
		cases = append(cases, result)
	}
	receipt := Receipt{
		Schema: ReceiptSchema, Scope: ReceiptScope,
		Source:      SourceEvidence{Path: sourcePath, Digest: digestBytes(source), Authority: "GOOO_SOURCE"},
		InputDigest: digestBytes(inputRaw), Operation: input.Operation,
		PolicySchema: input.Policy.Schema, FullSuiteID: input.Policy.FullSuiteID,
		ClaimTransitionHead: head, ClaimTransitions: transitions, Cases: cases,
		IndependentVerifier: IndependentVerifier{ID: "gooo://verifier/causal-ci-selection", Mode: "INDEPENDENT_REPLAY", Required: true, ReadOnly: true},
		Decision:            DecisionPass, Reason: "ALL_SCENARIO_BOUNDARIES_CLASSIFIED",
	}
	receipt.Indicators = deriveIndicators(receipt.Cases, receipt.ClaimTransitions)
	receipt.Metrics = deriveMetrics(input.Policy, input.Cases, receipt.Cases, len(input.ClaimTransitions))
	for _, indicator := range receipt.Indicators {
		if indicator.Satisfied {
			receipt.Metrics.FixedIndicatorSatisfied++
		}
	}
	receipt.Digest, err = receiptDigest(receipt)
	if err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}

func buildTransitions(values []ClaimTransition) ([]TransitionEvidence, string, error) {
	result := make([]TransitionEvidence, 0, len(values))
	previous := ""
	for _, value := range values {
		evidence := TransitionEvidence{Sequence: value.Sequence, ClaimID: value.ClaimID, Before: value.Before, After: value.After, Event: value.Event, Coordinate: value.Coordinate, EvidenceDigest: value.EvidenceDigest, PreviousDigest: previous}
		digest, err := transitionDigest(evidence)
		if err != nil {
			return nil, "", err
		}
		evidence.Digest = digest
		result = append(result, evidence)
		previous = digest
	}
	return result, previous, nil
}

type graphResult struct {
	paths    []PathEvidence
	unknown  []UnknownCause
	decision string
	reason   string
	coord    Coordinate
}

func evaluateCase(policy Policy, value Case) (CaseReceipt, error) {
	claims := make(map[string]Claim, len(value.Claims))
	for _, claim := range value.Claims {
		claims[claim.ID] = claim
	}
	checks := make(map[string]Check, len(policy.Checks))
	for _, check := range policy.Checks {
		checks[check.ID] = check
	}
	edges := append([]ImpactEdge(nil), value.ImpactEdges...)
	sort.Slice(edges, func(i, j int) bool { return edges[i].ID < edges[j].ID })
	if rejection := validateGraphEdges(edges, claims, checks); rejection != nil {
		return CaseReceipt{ID: value.ID, ChangedFiles: sortedStrings(value.ChangedFiles), Decision: DecisionRejected, Resolution: ResolutionRejected, Reason: rejection.reason, Coordinate: rejection.coordinate, SelectedChecks: []CheckChoice{}}, nil
	}
	result := graphResult{}
	for _, changedFile := range sortedStrings(value.ChangedFiles) {
		paths, unknowns := traceFile(changedFile, edges, claims, checks)
		result.paths = append(result.paths, paths...)
		result.unknown = append(result.unknown, unknowns...)
	}
	if len(result.unknown) > 0 {
		cause := result.unknown[0]
		result.decision, result.reason, result.coord = DecisionFullFallback, ReasonUnknownPath, Coordinate{Stage: "CAUSAL_SELECTION", Step: "descend-full-suite", Reason: cause.Reason}
		return makeCaseReceipt(value, result, policy, true), nil
	}
	if len(result.paths) == 0 {
		result.decision, result.reason, result.coord = DecisionRejected, ReasonNoRoute, Coordinate{Stage: "CAUSAL_SELECTION", Step: "validate-graph", Reason: "NO_CLAIM_TO_CHECK_PATH"}
		return makeCaseReceipt(value, result, policy, false), nil
	}
	result.decision, result.reason, result.coord = DecisionSelected, ReasonCompletePaths, Coordinate{Stage: "CAUSAL_SELECTION", Step: "select-checks", Reason: "ALL_CHANGED_FILES_HAVE_KNOWN_CLAIM_PATHS"}
	return makeCaseReceipt(value, result, policy, false), nil
}

type graphRejection struct {
	reason     string
	coordinate Coordinate
}

func validateGraphEdges(edges []ImpactEdge, claims map[string]Claim, checks map[string]Check) *graphRejection {
	pairs := map[string]ImpactEdge{}
	for _, edge := range edges {
		pair := edge.From + "\x00" + edge.To
		if previous, exists := pairs[pair]; exists && (previous.Kind != edge.Kind || previous.Known != edge.Known || previous.Reason != edge.Reason) {
			return &graphRejection{ReasonContradictory, Coordinate{Stage: "CAUSAL_SELECTION", Step: "validate-graph", Reason: edge.ID + ":CONFLICTS_WITH:" + previous.ID}}
		}
		pairs[pair] = edge
		if strings.HasPrefix(edge.To, "check:") {
			if _, exists := checks[strings.TrimPrefix(edge.To, "check:")]; !exists {
				return &graphRejection{ReasonUnregistered, Coordinate{Stage: "CAUSAL_SELECTION", Step: "validate-graph", Reason: edge.To}}
			}
		}
		if strings.HasPrefix(edge.To, "claim:") {
			if _, exists := claims[edge.To]; !exists {
				return &graphRejection{ReasonUnregistered, Coordinate{Stage: "CAUSAL_SELECTION", Step: "validate-graph", Reason: edge.To}}
			}
		}
	}
	return nil
}

type traceState struct {
	node     string
	claimIDs []string
	edgeIDs  []string
	hasClaim bool
}

func traceFile(changedFile string, edges []ImpactEdge, claims map[string]Claim, checks map[string]Check) ([]PathEvidence, []UnknownCause) {
	paths := []PathEvidence{}
	unknowns := []UnknownCause{}
	queue := []traceState{{node: changedFile}}
	seen := map[string]struct{}{changedFile + "|": {}}
	for len(queue) > 0 {
		state := queue[0]
		queue = queue[1:]
		outgoing := outgoingEdges(state.node, edges)
		if len(outgoing) == 0 {
			unknowns = append(unknowns, UnknownCause{ChangedFile: changedFile, Reason: "NO_OUTGOING_CAUSAL_EDGE", Coordinate: Coordinate{Stage: "CAUSAL_SELECTION", Step: "trace-impact", Reason: "NO_OUTGOING_CAUSAL_EDGE"}})
			continue
		}
		for _, edge := range outgoing {
			if !edge.Known {
				unknowns = append(unknowns, UnknownCause{ChangedFile: changedFile, EdgeID: edge.ID, Coordinate: edge.Coordinate, Reason: edge.Reason})
				continue
			}
			next := traceState{node: edge.To, claimIDs: append([]string(nil), state.claimIDs...), edgeIDs: append(append([]string(nil), state.edgeIDs...), edge.ID), hasClaim: state.hasClaim}
			if claim, isClaim := claims[edge.To]; isClaim {
				next.claimIDs = append(next.claimIDs, claim.ID)
				next.hasClaim = true
			}
			if strings.HasPrefix(edge.To, "check:") {
				checkID := strings.TrimPrefix(edge.To, "check:")
				if !next.hasClaim {
					unknowns = append(unknowns, UnknownCause{ChangedFile: changedFile, EdgeID: edge.ID, Coordinate: edge.Coordinate, Reason: ReasonClaimBypass})
					continue
				}
				paths = append(paths, PathEvidence{ChangedFile: changedFile, ClaimIDs: sortedStrings(next.claimIDs), CheckID: checkID, EdgeIDs: next.edgeIDs, Explanation: edge.Reason, ProofChoice: proofCausalPath})
				continue
			}
			if _, isClaim := claims[edge.To]; !isClaim {
				unknowns = append(unknowns, UnknownCause{ChangedFile: changedFile, EdgeID: edge.ID, Coordinate: edge.Coordinate, Reason: "UNKNOWN_IMPACT_NODE:" + edge.To})
				continue
			}
			key := next.node + "|" + strings.Join(next.claimIDs, ",")
			if _, exists := seen[key]; !exists {
				seen[key] = struct{}{}
				queue = append(queue, next)
			}
		}
	}
	return deduplicatePaths(paths), deduplicateUnknowns(unknowns)
}

func outgoingEdges(node string, edges []ImpactEdge) []ImpactEdge {
	result := []ImpactEdge{}
	for _, edge := range edges {
		if edge.From == node {
			result = append(result, edge)
		}
	}
	return result
}

func makeCaseReceipt(value Case, result graphResult, policy Policy, fallback bool) CaseReceipt {
	choices := []CheckChoice{}
	if fallback {
		for _, check := range policy.Checks {
			choices = append(choices, CheckChoice{CheckID: check.ID, ProofChoice: proofFullDescent, Reason: ReasonUnknownPath})
		}
	} else if result.decision == DecisionSelected {
		byCheck := map[string]CheckChoice{}
		for _, path := range result.paths {
			choice := byCheck[path.CheckID]
			choice.CheckID = path.CheckID
			choice.ProofChoice = proofCausalPath
			choice.Reason = ReasonCompletePaths
			choice.ClaimIDs = append(choice.ClaimIDs, path.ClaimIDs...)
			choice.PathIDs = append(choice.PathIDs, path.EdgeIDs...)
			byCheck[path.CheckID] = choice
		}
		for _, check := range policy.Checks {
			if choice, exists := byCheck[check.ID]; exists {
				choice.ClaimIDs = sortedStrings(choice.ClaimIDs)
				choice.PathIDs = sortedStrings(choice.PathIDs)
				choices = append(choices, choice)
			}
		}
	}
	return CaseReceipt{ID: value.ID, ChangedFiles: sortedStrings(value.ChangedFiles), Decision: result.decision, Resolution: resolutionFor(result.decision), Reason: result.reason, Coordinate: result.coord, Paths: result.paths, UnknownCauses: result.unknown, SelectedChecks: choices}
}

func resolutionFor(decision string) string {
	switch decision {
	case DecisionSelected:
		return ResolutionSelective
	case DecisionFullFallback:
		return ResolutionFull
	default:
		return ResolutionRejected
	}
}

func deriveMetrics(policy Policy, inputs []Case, cases []CaseReceipt, transitionTotal int) Metrics {
	metrics := Metrics{CaseTotal: len(cases), FixedCheckDenominator: len(policy.Checks), ClaimTransitionTotal: transitionTotal, FixedIndicatorDenominator: FixedIndicatorDenominator}
	for index, value := range cases {
		metrics.ChangedFileTotal += len(value.ChangedFiles)
		metrics.SelectedCheckTotal += len(value.SelectedChecks)
		if value.Decision == DecisionFullFallback {
			metrics.FullFallbackCaseTotal++
		}
		if value.Decision == DecisionRejected {
			metrics.RejectedCaseTotal++
		}
		if index < len(inputs) {
			metrics.ImpactEdgeTotal += len(inputs[index].ImpactEdges)
			for _, edge := range inputs[index].ImpactEdges {
				if edge.Known {
					metrics.KnownImpactEdgeTotal++
				}
			}
		}
	}
	return metrics
}

func deriveIndicators(cases []CaseReceipt, transitions []TransitionEvidence) []Indicator {
	values := make([]Indicator, 0, FixedIndicatorDenominator)
	checks := []bool{
		allCasesHaveClaimPathOrCause(cases), allCasesExplainReason(cases), hasCausalSelection(cases), hasSoundFallback(cases), hasClosedRejection(cases), validTransitionChain(transitions),
	}
	for index, id := range indicatorIDs {
		observed := 0
		if checks[index] {
			observed = 1
		}
		values = append(values, Indicator{ID: id, Observed: observed, Denominator: 1, Satisfied: checks[index]})
	}
	return values
}

func allCasesHaveClaimPathOrCause(cases []CaseReceipt) bool {
	for _, value := range cases {
		if value.Decision == DecisionRejected {
			continue
		}
		if len(value.Paths) == 0 && len(value.UnknownCauses) == 0 {
			return false
		}
	}
	return true
}
func allCasesExplainReason(cases []CaseReceipt) bool {
	for _, value := range cases {
		if value.Reason == "" || !coordinateKnown(value.Coordinate) {
			return false
		}
	}
	return true
}
func hasCausalSelection(cases []CaseReceipt) bool {
	for _, value := range cases {
		if value.ID == "selection" {
			return value.Decision == DecisionSelected && len(value.Paths) > 0 && len(value.SelectedChecks) > 0
		}
	}
	return false
}
func hasSoundFallback(cases []CaseReceipt) bool {
	for _, value := range cases {
		if value.ID == "full-fallback" {
			return value.Decision == DecisionFullFallback && len(value.UnknownCauses) > 0 && len(value.SelectedChecks) == FixedCheckDenominator
		}
	}
	return false
}
func hasClosedRejection(cases []CaseReceipt) bool {
	for _, value := range cases {
		if value.ID == "rejection" {
			return value.Decision == DecisionRejected && len(value.SelectedChecks) == 0
		}
	}
	return false
}
func validTransitionChain(values []TransitionEvidence) bool {
	previous := ""
	for index, value := range values {
		if value.Sequence != index+1 || value.PreviousDigest != previous {
			return false
		}
		computed, err := transitionDigest(value)
		if err != nil || computed != value.Digest {
			return false
		}
		previous = value.Digest
	}
	return len(values) > 0
}

func sortedStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}

func deduplicatePaths(values []PathEvidence) []PathEvidence {
	seen := map[string]struct{}{}
	result := []PathEvidence{}
	for _, value := range values {
		key := value.ChangedFile + "|" + value.CheckID + "|" + strings.Join(value.EdgeIDs, ",")
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ChangedFile+result[i].CheckID < result[j].ChangedFile+result[j].CheckID
	})
	return result
}

func deduplicateUnknowns(values []UnknownCause) []UnknownCause {
	seen := map[string]struct{}{}
	result := []UnknownCause{}
	for _, value := range values {
		key := value.ChangedFile + "|" + value.EdgeID + "|" + value.Reason
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ChangedFile+result[i].EdgeID < result[j].ChangedFile+result[j].EdgeID
	})
	return result
}
