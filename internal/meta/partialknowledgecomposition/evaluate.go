package partialknowledgecomposition

import (
	"fmt"
	"slices"
)

func Evaluate(input Input) (Receipt, error) {
	if err := ValidateInput(input); err != nil {
		return Receipt{}, err
	}
	results := make([]CaseResult, 0, len(input.Fixture.Cases))
	for _, current := range input.Fixture.Cases {
		value := Compose(current.Left, current.Right)
		decision, reason, topSuccess := classify(value)
		result := CaseResult{
			ID: current.ID, Producer: current.Producer, Consumer: current.Consumer,
			MetaOperation: current.MetaOperation, ProofChoice: current.ProofChoice,
			Left: current.Left, Right: current.Right, Result: value,
			Decision: decision, Reason: reason, TopSuccess: topSuccess,
		}
		result.EvidenceDigest = digestValue(struct {
			ID     string  `json:"id"`
			Left   Operand `json:"left"`
			Right  Operand `json:"right"`
			Result Value   `json:"result"`
		}{current.ID, current.Left, current.Right, value})
		results = append(results, result)
	}
	summary := summarize(results)
	claims := buildClaims(results)
	receipt := Receipt{
		Schema: Schema, Repository: input.Repository, HeadSHA: input.HeadSHA,
		SourcePath: input.SourcePath, SourceDigest: digestBytes(input.Source),
		Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation,
		ProofChoice: ProofCoherence, Resolution: "COMPOSITION_CALCULUS",
		Decision: "PROVEN", Reason: "COMPOSITION_RULES_REPLAYED",
		FixedDenominator: FixedDenominator,
		DenominatorDigest: digestValue(struct {
			Count int      `json:"count"`
			IDs   []string `json:"ids"`
		}{FixedDenominator, fixedCaseIDs}),
		Cases: results, Claims: claims, Summary: summary,
		RepositoryWrites: 0, PromotionAuthorized: false,
	}
	receipt.Indicators = buildIndicators(receipt)
	receipt.Digest = receiptDigest(receipt)
	return receipt, nil
}

func summarize(results []CaseResult) Summary {
	summary := Summary{FixedDenominator: FixedDenominator, CaseTotal: len(results), RepositoryWrites: 0}
	for _, result := range results {
		switch result.Result.State {
		case StateExact:
			summary.ExactCases++
		case StateDirectUnknown:
			summary.DirectUnknownCases++
		case StateDependencyBlocked:
			summary.DependencyBlockedCases++
		case StateInvariantOnly:
			summary.InvariantOnlyCases++
		case StateMixedUnresolved:
			summary.MixedUnresolvedCases++
		}
		if result.TopSuccess {
			summary.TopSuccessCases++
		}
	}
	summary.NonExactCases = summary.CaseTotal - summary.ExactCases
	summary.NonExactNotPromoted = summary.NonExactCases - countPromotedNonExact(results)
	summary.ClaimTransitionTotal = len(results)
	return summary
}

func countPromotedNonExact(results []CaseResult) int {
	count := 0
	for _, result := range results {
		if result.Result.State != StateExact && result.TopSuccess {
			count++
		}
	}
	return count
}

func buildClaims(results []CaseResult) []ClaimTransition {
	claims := make([]ClaimTransition, 0, len(results))
	previous := ""
	for index, result := range results {
		claim := ClaimTransition{
			Sequence: index + 1, ClaimID: "composition/" + result.ID,
			From: "OPEN", To: transitionState(result.Result.State),
			MetaOperation: result.MetaOperation, ProofChoice: result.ProofChoice,
			EvidenceDigest: result.EvidenceDigest, PreviousDigest: previous,
		}
		claim.Digest = transitionDigest(claim)
		claims = append(claims, claim)
		previous = claim.Digest
	}
	return claims
}

func transitionState(state State) string {
	switch state {
	case StateExact:
		return "DISCHARGED"
	case StateDirectUnknown:
		return "UNKNOWN"
	case StateDependencyBlocked:
		return "BLOCKED"
	case StateInvariantOnly:
		return "INVARIANT_PRESERVED"
	default:
		return "UNRESOLVED"
	}
}

func buildIndicators(receipt Receipt) []Indicator {
	summary := receipt.Summary
	indicators := []Indicator{
		indicator("PKC-FOUNDATION-SOURCE-001", "driver", ProofFoundation, "bind-partial-knowledge-source", 1, 1),
		indicator("PKC-COHERENCE-DIRECT-UNKNOWN-002", "driver", ProofCoherence, "compose-partial-knowledge", summary.DirectUnknownCases, 1),
		indicator("PKC-COHERENCE-DEPENDENCY-BLOCK-003", "driver", ProofCoherence, "compose-partial-knowledge", summary.DependencyBlockedCases, 1),
		indicator("PKC-COHERENCE-INVARIANT-004", "driver", ProofFoundation, "preserve-known-invariant", summary.InvariantOnlyCases, 1),
		indicator("PKC-COHERENCE-MIXED-005", "driver", ProofCoherence, "compose-partial-knowledge", summary.MixedUnresolvedCases, 1),
		indicator("PKC-REGRESSION-DENOMINATOR-006", "guardrail", ProofFoundation, "freeze-composition-denominator", summary.CaseTotal, FixedDenominator),
		indicator("PKC-REGRESSION-CLAIMS-007", "driver", ProofRegression, "replay-claim-transitions", summary.ClaimTransitionTotal, FixedDenominator),
		indicator("PKC-GUARDRAIL-NON-PROMOTION-008", "guardrail", ProofCoherence, "reject-non-exact-promotion", summary.NonExactNotPromoted, summary.NonExactCases),
		indicator("PKC-GUARDRAIL-READ-ONLY-009", "guardrail", ProofFoundation, "preserve-read-only-boundary", boolCount(receipt.RepositoryWrites == 0), 1),
		indicator("PKC-OUTCOME-CALCULUS-010", "outcome", ProofRegression, "replay-composition-corpus", summary.ExactCases+summary.DirectUnknownCases+summary.DependencyBlockedCases+summary.InvariantOnlyCases+summary.MixedUnresolvedCases, FixedDenominator),
	}
	for index := range indicators {
		indicators[index].Producer = receipt.Producer
		indicators[index].Consumer = receipt.Consumer
	}
	return indicators
}

func indicator(id, class string, proof ProofChoice, operation string, observed, denominator int) Indicator {
	basisPoints := 0
	if denominator > 0 {
		basisPoints = observed * 10000 / denominator
	}
	return Indicator{
		ID: id, Class: class, ProofChoice: proof, MetaOperation: operation,
		Observed: observed, Denominator: denominator, BasisPoints: basisPoints,
		TargetBasisPoints: 10000, Satisfied: denominator > 0 && basisPoints == 10000,
	}
}

func boolCount(value bool) int {
	if value {
		return 1
	}
	return 0
}

func ValidateReceipt(receipt Receipt) error {
	if receipt.Schema != Schema || receipt.Decision != "PROVEN" ||
		receipt.Reason != "COMPOSITION_RULES_REPLAYED" || receipt.Resolution != "COMPOSITION_CALCULUS" ||
		receipt.Producer != Producer || receipt.Consumer != Consumer || receipt.MetaOperation != MetaOperation ||
		receipt.FixedDenominator != FixedDenominator || len(receipt.Cases) != FixedDenominator ||
		len(receipt.Claims) != FixedDenominator || len(receipt.Indicators) != 10 ||
		receipt.RepositoryWrites != 0 || receipt.PromotionAuthorized || receipt.SourceDigest == "" ||
		receipt.DenominatorDigest != digestValue(struct {
			Count int      `json:"count"`
			IDs   []string `json:"ids"`
		}{FixedDenominator, fixedCaseIDs}) || receipt.Digest != receiptDigest(receipt) {
		return fmt.Errorf("partial-knowledge receipt is not closed")
	}
	if receipt.Summary.CaseTotal != FixedDenominator || receipt.Summary.TopSuccessCases != 1 ||
		receipt.Summary.NonExactCases != 4 || receipt.Summary.NonExactNotPromoted != 4 ||
		receipt.Summary.ClaimTransitionTotal != FixedDenominator {
		return fmt.Errorf("partial-knowledge receipt summary is not exact")
	}
	previous := ""
	for index, result := range receipt.Cases {
		if result.ID != fixedCaseIDs[index] || result.Producer != Producer || result.Consumer != Consumer ||
			result.MetaOperation == "" || !validProofChoice(result.ProofChoice) {
			return fmt.Errorf("case %d identity is not closed", index+1)
		}
		decision, reason, topSuccess := classify(result.Result)
		if result.Decision != decision || result.Reason != reason || result.TopSuccess != topSuccess {
			return fmt.Errorf("case %q classification is not closed", result.ID)
		}
		if result.EvidenceDigest != digestValue(struct {
			ID     string  `json:"id"`
			Left   Operand `json:"left"`
			Right  Operand `json:"right"`
			Result Value   `json:"result"`
		}{result.ID, result.Left, result.Right, result.Result}) {
			return fmt.Errorf("case %q evidence digest differs", result.ID)
		}
		claim := receipt.Claims[index]
		if claim.Sequence != index+1 || claim.ClaimID != "composition/"+result.ID || claim.From != "OPEN" ||
			claim.To != transitionState(result.Result.State) || claim.MetaOperation != result.MetaOperation ||
			claim.ProofChoice != result.ProofChoice || claim.EvidenceDigest != result.EvidenceDigest ||
			claim.PreviousDigest != previous || claim.Digest != transitionDigest(claim) {
			return fmt.Errorf("claim transition %d is not append-only", index+1)
		}
		previous = claim.Digest
	}
	for _, indicator := range receipt.Indicators {
		if !indicator.Satisfied || indicator.TargetBasisPoints != 10000 || indicator.Producer != Producer || indicator.Consumer != Consumer {
			return fmt.Errorf("indicator %q is not satisfied", indicator.ID)
		}
	}
	if !slices.Equal(receipt.Indicators, buildIndicators(receipt)) {
		return fmt.Errorf("indicator reconstruction differs")
	}
	return nil
}
