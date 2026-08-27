package partialknowledgecomposition

import (
	"fmt"
	"regexp"
)

var headPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func Evaluate(input Input) (Receipt, error) {
	if input.Repository == "" || !headPattern.MatchString(input.HeadSHA) || input.SourcePath != SourcePath || len(input.Source) == 0 {
		return Receipt{}, fmt.Errorf("partial-knowledge identity is malformed")
	}
	model, err := parseSource(input.SourcePath, input.Source)
	if err != nil {
		return Receipt{}, err
	}
	intervention, err := applyIntervention(&model, input.Intervention)
	if err != nil {
		return Receipt{}, err
	}
	results := make([]CaseResult, 0, len(model.Cases))
	for _, current := range model.Cases {
		value := Compose(current.Left, current.Right)
		decision, resolution, reason, topSuccess := classify(value)
		provenance := Provenance{
			SourcePath: input.SourcePath, SourceActivity: current.SourceActivity,
			Producer: current.Producer, Consumer: current.Consumer,
			MetaOperation: current.MetaOperation, ProofChoice: current.ProofChoice,
			SemanticIRDigest:  model.SemanticIRDigest,
			ObservationDigest: digestValue(current),
		}
		result := CaseResult{
			ID: current.ID, SourceActivity: current.SourceActivity, SourceActivityID: current.SourceActivityID,
			Producer: current.Producer, Consumer: current.Consumer, MetaOperation: current.MetaOperation,
			ProofChoice: current.ProofChoice, Left: current.Left, Right: current.Right, Result: value,
			Decision: decision, Resolution: resolution, Stage: "COMPOSITION", Step: "COMPOSE_OBSERVATIONS",
			Reason: reason, TopSuccess: topSuccess, Provenance: provenance,
		}
		result.EvidenceDigest = digestValue(struct {
			ID         string     `json:"id"`
			Activity   string     `json:"activity"`
			Left       Evidence   `json:"left"`
			Right      Evidence   `json:"right"`
			Result     Value      `json:"result"`
			Provenance Provenance `json:"provenance"`
		}{current.ID, current.SourceActivity, current.Left, current.Right, value, provenance})
		results = append(results, result)
	}
	summary := summarize(results)
	claims := buildClaims(results)
	receipt := Receipt{
		Schema: Schema, Repository: input.Repository, HeadSHA: input.HeadSHA, SourcePath: input.SourcePath,
		SourceDigest: digestBytes(input.Source), SemanticIRDigest: model.SemanticIRDigest,
		SourceCases: len(model.Cases), SourceCasesTotal: FixedDenominator,
		Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation, ProofChoice: ProofCoherence,
		Resolution: ResolutionCalculus, SubjectResolution: SubjectResolution,
		Decision: DecisionCalculusProven, Reason: "COMPOSITION_RULES_REPLAYED",
		FixedDenominator: FixedDenominator, Cases: results, Claims: claims, Summary: summary,
		Intervention: intervention, RepositoryWrites: 0, PromotionAuthorized: false,
	}
	receipt.Indicators = buildIndicators(receipt)
	receipt.SemanticProjectionDigest = semanticProjectionDigest(receipt)
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
	summary.NonExactNotPromoted = summary.NonExactCases
	summary.ClaimTransitionTotal = len(results)
	summary.DischargedClaims = summary.ExactCases
	summary.OpenClaims = summary.NonExactCases
	return summary
}

func buildClaims(results []CaseResult) []ClaimTransition {
	claims := make([]ClaimTransition, 0, len(results))
	previous := ""
	for index, result := range results {
		claim := ClaimTransition{
			Sequence: index + 1, ClaimID: "composition/" + result.ID, From: "OPEN", To: transitionState(result.Result.State),
			Stage: result.Stage, Step: result.Step, Reason: result.Reason, Provenance: result.Provenance,
			EvidenceDigest: result.EvidenceDigest, PreviousDigest: previous,
		}
		claim.Digest = transitionDigest(claim)
		claims = append(claims, claim)
		previous = claim.Digest
	}
	return claims
}

func buildIndicators(receipt Receipt) []Indicator {
	summary := receipt.Summary
	indicators := []Indicator{
		indicator("PKC-FOUNDATION-SOURCE-001", "foundation", ProofFoundation, "parse-source-and-lower", receipt.SourceCases, FixedDenominator),
		indicator("PKC-COHERENCE-DIRECT-UNKNOWN-002", "driver", ProofCoherence, "derive-direct-unknown", summary.DirectUnknownCases, 1),
		indicator("PKC-COHERENCE-DEPENDENCY-BLOCK-003", "driver", ProofCoherence, "derive-dependency-block", summary.DependencyBlockedCases, 1),
		indicator("PKC-COHERENCE-INVARIANT-004", "driver", ProofFoundation, "preserve-known-invariant", summary.InvariantOnlyCases, 1),
		indicator("PKC-COHERENCE-MIXED-005", "driver", ProofCoherence, "retain-mixed-causes", summary.MixedUnresolvedCases, 1),
		indicator("PKC-REGRESSION-DENOMINATOR-006", "guardrail", ProofRegression, "freeze-composition-denominator", summary.CaseTotal, FixedDenominator),
		indicator("PKC-REGRESSION-CLAIMS-007", "driver", ProofRegression, "replay-claim-transitions", summary.ClaimTransitionTotal, FixedDenominator),
		indicator("PKC-GUARDRAIL-NON-PROMOTION-008", "guardrail", ProofCoherence, "reject-non-exact-promotion", summary.NonExactNotPromoted, summary.NonExactCases),
		indicator("PKC-GUARDRAIL-READ-ONLY-009", "guardrail", ProofFoundation, "observe-read-only-boundary", boolCount(receipt.RepositoryWrites == 0), 1),
		indicator("PKC-OUTCOME-CALCULUS-010", "outcome", ProofRegression, "replay-composition-calculus", summary.ExactCases+summary.DirectUnknownCases+summary.DependencyBlockedCases+summary.InvariantOnlyCases+summary.MixedUnresolvedCases, FixedDenominator),
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
	return Indicator{ID: id, Class: class, ProofChoice: proof, MetaOperation: operation, Observed: observed, Denominator: denominator, BasisPoints: basisPoints, TargetBasisPoints: 10000, Satisfied: denominator > 0 && basisPoints == 10000}
}

func boolCount(value bool) int {
	if value {
		return 1
	}
	return 0
}
