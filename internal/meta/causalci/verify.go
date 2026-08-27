package causalci

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Verify is deliberately separate from Evaluate. It reconstructs scenario
// decisions with a small independent reachability walk, then checks the
// digest-bound receipt and append-only transition chain. It never trusts the
// evaluator's in-memory result as evidence.
func Verify(inputRaw []byte, sourcePath string, source []byte, receipt Receipt) error {
	input, err := decodeInput(inputRaw)
	if err != nil {
		return err
	}
	if receipt.Schema != ReceiptSchema || receipt.Scope != ReceiptScope || receipt.InputDigest != digestBytes(inputRaw) || receipt.Source.Path != sourcePath || receipt.Source.Digest != digestBytes(source) || receipt.Source.Authority != "GOOO_SOURCE" {
		return fmt.Errorf("receipt binding mismatch")
	}
	if receipt.Operation != input.Operation || receipt.PolicySchema != input.Policy.Schema || receipt.FullSuiteID != input.Policy.FullSuiteID {
		return fmt.Errorf("receipt authority mismatch")
	}
	if receipt.IndependentVerifier.ID == "" || receipt.IndependentVerifier.Mode != "INDEPENDENT_REPLAY" || !receipt.IndependentVerifier.Required || !receipt.IndependentVerifier.ReadOnly {
		return fmt.Errorf("independent verifier declaration missing")
	}
	computedDigest, err := receiptDigest(receipt)
	if err != nil || computedDigest != receipt.Digest {
		return fmt.Errorf("receipt digest mismatch")
	}
	transitions, head, err := buildTransitions(input.ClaimTransitions)
	if err != nil || !equalJSON(transitions, receipt.ClaimTransitions) || head != receipt.ClaimTransitionHead || !validTransitionChain(receipt.ClaimTransitions) {
		return fmt.Errorf("claim transition ledger mismatch")
	}
	if len(receipt.Cases) != len(input.Cases) {
		return fmt.Errorf("case receipt denominator mismatch")
	}
	for index, value := range input.Cases {
		expected, err := independentCase(input.Policy, value)
		if err != nil {
			return err
		}
		actual := receipt.Cases[index]
		if actual.ID != value.ID || actual.Decision != expected.decision || actual.Resolution != resolutionFor(expected.decision) || actual.Reason != expected.reason || actual.Coordinate != expected.coordinate || !sameCheckIDs(actual.SelectedChecks, expected.selectedChecks) {
			return fmt.Errorf("independent case decision mismatch: %s", value.ID)
		}
		if err := validateChoices(actual, expected.decision); err != nil {
			return fmt.Errorf("%s: %w", value.ID, err)
		}
		if err := validatePaths(value, actual); err != nil {
			return fmt.Errorf("%s: %w", value.ID, err)
		}
		if expected.decision == DecisionFullFallback && (len(actual.UnknownCauses) == 0 || actual.UnknownCauses[0].Reason != expected.reason || actual.UnknownCauses[0].Reason != actual.Coordinate.Reason) {
			return fmt.Errorf("fallback cause is not carried: %s", value.ID)
		}
		if expected.decision == DecisionFullFallback && !allChecksSelected(input.Policy, actual.SelectedChecks) {
			return fmt.Errorf("full fallback is not total: %s", value.ID)
		}
		if expected.decision == DecisionRejected && len(actual.SelectedChecks) != 0 {
			return fmt.Errorf("rejected case selected work: %s", value.ID)
		}
	}
	if receipt.Decision != DecisionPass || receipt.Reason == "" || receipt.Metrics.FixedCheckDenominator != FixedCheckDenominator || receipt.Metrics.FixedIndicatorDenominator != FixedIndicatorDenominator || receipt.Metrics.FixedIndicatorSatisfied != FixedIndicatorDenominator || len(receipt.Indicators) != FixedIndicatorDenominator {
		return fmt.Errorf("receipt fixed contract mismatch")
	}
	for index, indicator := range receipt.Indicators {
		if indicator.ID != indicatorIDs[index] {
			return fmt.Errorf("indicator order mismatch: %s", indicator.ID)
		}
		if indicator.Denominator != 1 || !indicator.Satisfied || indicator.Observed != 1 {
			return fmt.Errorf("indicator is not satisfied: %s", indicator.ID)
		}
	}
	expectedMetrics := independentMetrics(input, receipt)
	if receipt.Metrics != expectedMetrics {
		return fmt.Errorf("independent metrics mismatch")
	}
	return nil
}

type independentResult struct {
	decision       string
	reason         string
	coordinate     Coordinate
	selectedChecks []string
}

func independentCase(policy Policy, value Case) (independentResult, error) {
	checks := map[string]struct{}{}
	for _, check := range policy.Checks {
		checks[check.ID] = struct{}{}
	}
	claims := map[string]struct{}{}
	for _, claim := range value.Claims {
		claims[claim.ID] = struct{}{}
	}
	byFrom := map[string][]ImpactEdge{}
	for _, edge := range value.ImpactEdges {
		if strings.HasPrefix(edge.To, "check:") {
			if _, ok := checks[strings.TrimPrefix(edge.To, "check:")]; !ok {
				return independentResult{decision: DecisionRejected, reason: ReasonUnregistered, coordinate: Coordinate{Stage: "CAUSAL_SELECTION", Step: "validate-graph", Reason: edge.To}}, nil
			}
		}
		if strings.HasPrefix(edge.To, "claim:") {
			if _, ok := claims[edge.To]; !ok {
				return independentResult{decision: DecisionRejected, reason: ReasonUnregistered, coordinate: Coordinate{Stage: "CAUSAL_SELECTION", Step: "validate-graph", Reason: edge.To}}, nil
			}
		}
		byFrom[edge.From] = append(byFrom[edge.From], edge)
	}
	unknown := ""
	knownChecks := map[string]bool{}
	for _, file := range sortedStrings(value.ChangedFiles) {
		queue := []struct {
			node  string
			claim bool
		}{{node: file}}
		visited := map[string]bool{file + "|false": true}
		for len(queue) > 0 {
			state := queue[0]
			queue = queue[1:]
			for _, edge := range byFrom[state.node] {
				if !edge.Known {
					if unknown == "" {
						unknown = edge.Reason
					}
					continue
				}
				if strings.HasPrefix(edge.To, "check:") {
					if !state.claim {
						if unknown == "" {
							unknown = ReasonClaimBypass
						}
						continue
					}
					knownChecks[strings.TrimPrefix(edge.To, "check:")] = true
					continue
				}
				if _, ok := claims[edge.To]; !ok {
					if unknown == "" {
						unknown = "UNKNOWN_IMPACT_NODE:" + edge.To
					}
					continue
				}
				key := edge.To + "|true"
				if !visited[key] {
					visited[key] = true
					queue = append(queue, struct {
						node  string
						claim bool
					}{node: edge.To, claim: true})
				}
			}
		}
	}
	if unknown != "" {
		return independentResult{decision: DecisionFullFallback, reason: ReasonUnknownPath, coordinate: Coordinate{Stage: "CAUSAL_SELECTION", Step: "descend-full-suite", Reason: unknown}, selectedChecks: append([]string(nil), requiredCheckIDs[:]...)}, nil
	}
	if len(knownChecks) == 0 {
		return independentResult{decision: DecisionRejected, reason: ReasonNoRoute, coordinate: Coordinate{Stage: "CAUSAL_SELECTION", Step: "validate-graph", Reason: "NO_CLAIM_TO_CHECK_PATH"}}, nil
	}
	selected := []string{}
	for _, check := range policy.Checks {
		if knownChecks[check.ID] {
			selected = append(selected, check.ID)
		}
	}
	return independentResult{decision: DecisionSelected, reason: ReasonCompletePaths, coordinate: Coordinate{Stage: "CAUSAL_SELECTION", Step: "select-checks", Reason: "ALL_CHANGED_FILES_HAVE_KNOWN_CLAIM_PATHS"}, selectedChecks: selected}, nil
}

func sameCheckIDs(values []CheckChoice, expected []string) bool {
	if len(values) != len(expected) {
		return false
	}
	for index, value := range values {
		if value.CheckID != expected[index] {
			return false
		}
	}
	return true
}

func validateChoices(value CaseReceipt, decision string) error {
	switch decision {
	case DecisionSelected:
		for _, choice := range value.SelectedChecks {
			if choice.ProofChoice != proofCausalPath || choice.Reason != ReasonCompletePaths || len(choice.ClaimIDs) == 0 || len(choice.PathIDs) == 0 {
				return fmt.Errorf("causal check choice is not path-bound")
			}
		}
	case DecisionFullFallback:
		for _, choice := range value.SelectedChecks {
			if choice.ProofChoice != proofFullDescent || choice.Reason != ReasonUnknownPath {
				return fmt.Errorf("fallback check choice is not full-suite-bound")
			}
		}
	case DecisionRejected:
		if len(value.SelectedChecks) != 0 {
			return fmt.Errorf("rejected case contains choices")
		}
	}
	return nil
}

func validatePaths(input Case, receipt CaseReceipt) error {
	edges := map[string]ImpactEdge{}
	claims := map[string]bool{}
	for _, claim := range input.Claims {
		claims[claim.ID] = true
	}
	for _, edge := range input.ImpactEdges {
		edges[edge.ID] = edge
	}
	for _, path := range receipt.Paths {
		if path.ProofChoice != proofCausalPath || path.Explanation == "" || len(path.EdgeIDs) == 0 || len(path.ClaimIDs) == 0 {
			return fmt.Errorf("path evidence is incomplete")
		}
		for _, claimID := range path.ClaimIDs {
			if !claims[claimID] {
				return fmt.Errorf("path names an undeclared claim")
			}
		}
		current := path.ChangedFile
		claimSeen := false
		for _, edgeID := range path.EdgeIDs {
			edge, exists := edges[edgeID]
			if !exists || !edge.Known || edge.From != current {
				return fmt.Errorf("path edge chain is not exact")
			}
			if strings.HasPrefix(edge.To, "claim:") {
				claimSeen = true
			}
			current = edge.To
		}
		if !claimSeen || current != "check:"+path.CheckID {
			return fmt.Errorf("path is not claim-mediated to its check")
		}
	}
	return nil
}

func independentMetrics(input Input, receipt Receipt) Metrics {
	metrics := Metrics{CaseTotal: len(input.Cases), ClaimTransitionTotal: len(input.ClaimTransitions), FixedCheckDenominator: FixedCheckDenominator, FixedIndicatorDenominator: FixedIndicatorDenominator, FixedIndicatorSatisfied: FixedIndicatorDenominator}
	for index, value := range input.Cases {
		metrics.ChangedFileTotal += len(value.ChangedFiles)
		metrics.ImpactEdgeTotal += len(value.ImpactEdges)
		for _, edge := range value.ImpactEdges {
			if edge.Known {
				metrics.KnownImpactEdgeTotal++
			}
		}
		if index >= len(receipt.Cases) {
			continue
		}
		metrics.SelectedCheckTotal += len(receipt.Cases[index].SelectedChecks)
		switch receipt.Cases[index].Decision {
		case DecisionFullFallback:
			metrics.FullFallbackCaseTotal++
		case DecisionRejected:
			metrics.RejectedCaseTotal++
		}
	}
	return metrics
}

func allChecksSelected(policy Policy, values []CheckChoice) bool {
	if len(values) != len(policy.Checks) {
		return false
	}
	seen := map[string]bool{}
	for _, value := range values {
		seen[value.CheckID] = true
	}
	for _, check := range policy.Checks {
		if !seen[check.ID] {
			return false
		}
	}
	return true
}

func equalJSON(left, right any) bool {
	leftRaw, leftErr := json.Marshal(left)
	rightRaw, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && string(leftRaw) == string(rightRaw)
}
