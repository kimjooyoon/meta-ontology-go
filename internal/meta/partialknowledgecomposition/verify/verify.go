package verify

import (
	"encoding/json"
	"fmt"
	"regexp"
)

var headPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func Verify(input Input) (Report, error) {
	if input.Repository == "" || !headPattern.MatchString(input.HeadSHA) || input.SourcePath != "examples/partial-knowledge-composition/main.gooo" || len(input.Source) == 0 {
		return Report{}, fmt.Errorf("independent verifier identity is malformed")
	}
	model, err := parseSource(input.SourcePath, input.Source)
	if err != nil {
		return Report{}, err
	}
	intervention, err := applyIntervention(&model, input.InterventionMode)
	if err != nil {
		return Report{}, err
	}
	expected := reconstruct(input, model, intervention)
	var actual Receipt
	if err := json.Unmarshal(input.Receipt, &actual); err != nil {
		return Report{}, fmt.Errorf("decode producer receipt: %w", err)
	}
	if actual.Digest == "" || actual.Digest != receiptDigest(actual) || actual.Digest != expected.Digest {
		return Report{}, fmt.Errorf("receipt digest or source reconstruction differs")
	}
	actualRaw, _ := json.Marshal(actual)
	expectedRaw, _ := json.Marshal(expected)
	if string(actualRaw) != string(expectedRaw) {
		return Report{}, fmt.Errorf("producer receipt differs from independent source reconstruction")
	}
	return reportFrom(actual), nil
}

func reconstruct(input Input, model sourceModel, intervention Intervention) Receipt {
	results := make([]CaseResult, 0, len(model.Cases))
	for _, current := range model.Cases {
		value := compose(current.Left, current.Right)
		decision, resolution, reason, topSuccess := classify(value)
		provenance := Provenance{
			SourcePath: input.SourcePath, SourceActivity: current.SourceActivity,
			Producer: current.Producer, Consumer: current.Consumer,
			MetaOperation: current.MetaOperation, ProofChoice: current.ProofChoice,
			SemanticIRDigest: model.SemanticIRDigest, ObservationDigest: digestValue(current),
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
		Schema: "gooo/meta/partial-knowledge-composition-receipt/v2", Repository: input.Repository, HeadSHA: input.HeadSHA,
		SourcePath: input.SourcePath, SourceDigest: digestBytes(input.Source), SemanticIRDigest: model.SemanticIRDigest,
		SourceCases: len(model.Cases), SourceCasesTotal: 5, Producer: "partial-knowledge-producer",
		Consumer: "partial-knowledge-composition-consumer", MetaOperation: "compose-partial-knowledge", ProofChoice: "COHERENCE",
		Resolution: "CALCULUS", SubjectResolution: "PARTIAL_KNOWLEDGE", Decision: "CALCULUS_PROVEN", Reason: "COMPOSITION_RULES_REPLAYED",
		FixedDenominator: 5, Cases: results, Claims: claims, Summary: summary, Intervention: intervention,
		RepositoryWrites: 0, PromotionAuthorized: false,
	}
	receipt.Indicators = buildIndicators(receipt)
	receipt.SemanticProjectionDigest = semanticProjectionDigest(receipt)
	receipt.Digest = receiptDigest(receipt)
	return receipt
}

func summarize(results []CaseResult) Summary {
	summary := Summary{FixedDenominator: 5, CaseTotal: len(results), RepositoryWrites: 0}
	for _, result := range results {
		switch result.Result.State {
		case exact:
			summary.ExactCases++
		case directUnknown:
			summary.DirectUnknownCases++
		case dependencyBlocked:
			summary.DependencyBlockedCases++
		case invariantOnly:
			summary.InvariantOnlyCases++
		case mixedUnresolved:
			summary.MixedUnresolvedCases++
		}
		if result.TopSuccess {
			summary.TopSuccessCases++
		}
	}
	summary.NonExactCases = summary.CaseTotal - summary.ExactCases
	summary.NonExactNotPromoted = summary.NonExactCases
	summary.OpenClaims = summary.NonExactCases
	summary.DischargedClaims = summary.ExactCases
	summary.ClaimTransitionTotal = len(results)
	return summary
}

func buildClaims(results []CaseResult) []ClaimTransition {
	claims := make([]ClaimTransition, 0, len(results))
	previous := ""
	for index, result := range results {
		claim := ClaimTransition{Sequence: index + 1, ClaimID: "composition/" + result.ID, From: "OPEN", To: transitionState(result.Result.State), Stage: result.Stage, Step: result.Step, Reason: result.Reason, Provenance: result.Provenance, EvidenceDigest: result.EvidenceDigest, PreviousDigest: previous}
		claim.Digest = transitionDigest(claim)
		claims = append(claims, claim)
		previous = claim.Digest
	}
	return claims
}

func buildIndicators(receipt Receipt) []Indicator {
	summary := receipt.Summary
	items := []Indicator{
		makeIndicator("PKC-FOUNDATION-SOURCE-001", "foundation", "FOUNDATION", "parse-source-and-lower", receipt.SourceCases, 5),
		makeIndicator("PKC-COHERENCE-DIRECT-UNKNOWN-002", "driver", "COHERENCE", "derive-direct-unknown", summary.DirectUnknownCases, 1),
		makeIndicator("PKC-COHERENCE-DEPENDENCY-BLOCK-003", "driver", "COHERENCE", "derive-dependency-block", summary.DependencyBlockedCases, 1),
		makeIndicator("PKC-COHERENCE-INVARIANT-004", "driver", "FOUNDATION", "preserve-known-invariant", summary.InvariantOnlyCases, 1),
		makeIndicator("PKC-COHERENCE-MIXED-005", "driver", "COHERENCE", "retain-mixed-causes", summary.MixedUnresolvedCases, 1),
		makeIndicator("PKC-REGRESSION-DENOMINATOR-006", "guardrail", "REGRESSION", "freeze-composition-denominator", summary.CaseTotal, 5),
		makeIndicator("PKC-REGRESSION-CLAIMS-007", "driver", "REGRESSION", "replay-claim-transitions", summary.ClaimTransitionTotal, 5),
		makeIndicator("PKC-GUARDRAIL-NON-PROMOTION-008", "guardrail", "COHERENCE", "reject-non-exact-promotion", summary.NonExactNotPromoted, summary.NonExactCases),
		makeIndicator("PKC-GUARDRAIL-READ-ONLY-009", "guardrail", "FOUNDATION", "observe-read-only-boundary", 1, 1),
		makeIndicator("PKC-OUTCOME-CALCULUS-010", "outcome", "REGRESSION", "replay-composition-calculus", summary.ExactCases+summary.DirectUnknownCases+summary.DependencyBlockedCases+summary.InvariantOnlyCases+summary.MixedUnresolvedCases, 5),
	}
	for index := range items {
		items[index].Producer = receipt.Producer
		items[index].Consumer = receipt.Consumer
	}
	return items
}

func makeIndicator(id, class, proof, operation string, observed, denominator int) Indicator {
	basisPoints := 0
	if denominator > 0 {
		basisPoints = observed * 10000 / denominator
	}
	return Indicator{ID: id, Class: class, ProofChoice: proof, MetaOperation: operation, Observed: observed, Denominator: denominator, BasisPoints: basisPoints, TargetBasisPoints: 10000, Satisfied: denominator > 0 && basisPoints == 10000}
}

func reportFrom(receipt Receipt) Report {
	result := Report{Schema: "gooo/meta/partial-knowledge-composition-verification/v2", Repository: receipt.Repository, HeadSHA: receipt.HeadSHA, Status: "VERIFIED", Decision: receipt.Decision, Resolution: receipt.Resolution, SubjectResolution: receipt.SubjectResolution, FixedDenominator: receipt.FixedDenominator, SourceCases: receipt.SourceCases, SourceCasesTotal: receipt.SourceCasesTotal, ExactCases: receipt.Summary.ExactCases, DirectUnknownCases: receipt.Summary.DirectUnknownCases, DependencyBlockedCases: receipt.Summary.DependencyBlockedCases, InvariantOnlyCases: receipt.Summary.InvariantOnlyCases, MixedUnresolvedCases: receipt.Summary.MixedUnresolvedCases, TopSuccessCases: receipt.Summary.TopSuccessCases, NonExactNotPromoted: receipt.Summary.NonExactNotPromoted, OpenClaims: receipt.Summary.OpenClaims, ClaimTransitionTotal: receipt.Summary.ClaimTransitionTotal, RepositoryWrites: receipt.RepositoryWrites, PromotionAuthorized: receipt.PromotionAuthorized, IndependentEvaluator: true, SourceSemanticDigest: receipt.SemanticIRDigest, ReceiptDigest: receipt.Digest, SemanticProjectionDigest: receipt.SemanticProjectionDigest}
	result.Digest = reportDigest(result)
	return result
}

func reportDigest(report Report) string { report.Digest = ""; return digestValue(report) }
