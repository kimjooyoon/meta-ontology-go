package partialknowledgecomposition

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
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
	rawEvidence, err := parseRawEvidence(input, model)
	if err != nil {
		return Receipt{}, err
	}
	intervention, err := interventionEvidence(input, model)
	if err != nil {
		return Receipt{}, err
	}
	results := make([]CaseResult, 0, len(model.Cases))
	for index, recipe := range model.Cases {
		rawCase := rawEvidence.Cases[index]
		value := Compose(rawCase.Left, rawCase.Right)
		decision, resolution, reason, topSuccess := classify(value)
		predicate := predicateFor(value.State)
		proposition := propositionFor(recipe, predicate)
		provenance := Provenance{
			SourcePath: input.SourcePath, SourceActivity: recipe.SourceActivity,
			Producer: recipe.Producer, Consumer: recipe.Consumer,
			MetaOperation: recipe.MetaOperation, ProofChoice: recipe.ProofChoice,
			RawSourceDigest: digestBytes(input.Source), SemanticIRDigest: model.SemanticIRDigest,
			RawEvidenceDigest: rawEvidence.Digest, ObservationDigest: digestValue(rawCase),
		}
		result := CaseResult{
			ID: recipe.ID, SourceActivity: recipe.SourceActivity, SourceActivityID: recipe.SourceActivityID,
			Producer: recipe.Producer, Consumer: recipe.Consumer, MetaOperation: recipe.MetaOperation,
			ProofChoice: recipe.ProofChoice, Left: rawCase.Left, Right: rawCase.Right, Result: value,
			Predicate: predicate, Proposition: proposition, PropositionDigest: digestValue(proposition),
			TargetAddress: targetAddress(recipe), TargetOperation: recipe.MetaOperation, TargetOutput: "MetaValue",
			Decision: decision, Resolution: resolution, Stage: "COMPOSITION", Step: "COMPOSE_OBSERVATIONS",
			Reason: reason, TopSuccess: topSuccess, Provenance: provenance,
		}
		result.EvidenceDigest = caseEvidenceDigest(result)
		results = append(results, result)
	}
	summary := summarize(results, rawEvidence.Workspace.RepositoryWrites)
	claims := buildClaims(results)
	receipt := Receipt{
		Schema: Schema, Repository: input.Repository, HeadSHA: input.HeadSHA, SourcePath: input.SourcePath,
		SourceDigest: digestBytes(input.Source), SemanticIRDigest: model.SemanticIRDigest, RawEvidenceDigest: rawEvidence.Digest,
		SourceCases: len(model.Cases), SourceCasesTotal: FixedDenominator,
		Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation, ProofChoice: ProofCoherence,
		Resolution: ResolutionCalculus, SubjectResolution: subjectResolution(summary), EvidenceCoverage: evidenceCoverage(rawEvidence),
		AuthorityState: rawEvidence.Authority.State, AuthorityResolution: rawEvidence.Authority.Resolution,
		AuthorityStage: rawEvidence.Authority.Stage, AuthorityStep: rawEvidence.Authority.Step, AuthorityReason: rawEvidence.Authority.Reason,
		Decision: DecisionCalculusProven, Reason: "COMPOSITION_RULES_REPLAYED",
		FixedDenominator: FixedDenominator, Cases: results, Claims: claims, Summary: summary,
		Intervention: intervention, RepositoryWrites: rawEvidence.Workspace.RepositoryWrites,
		PromotionAuthorized: promotionAuthorized(rawEvidence.Authority),
	}
	receipt.Indicators = buildIndicators(receipt, rawEvidence)
	receipt.SemanticProjectionDigest = semanticProjectionDigest(receipt)
	receipt.Digest = receiptDigest(receipt)
	return receipt, nil
}

func interventionEvidence(input Input, model sourceModel) (Intervention, error) {
	mode := input.Intervention
	if mode == "" {
		mode = InterventionNone
	}
	if !validIntervention(mode) {
		return Intervention{}, fmt.Errorf("intervention %q is invalid", mode)
	}
	intervention := Intervention{Mode: mode, SourceDigest: digestBytes(input.Source), SemanticIRDigest: model.SemanticIRDigest}
	switch mode {
	case InterventionNone:
		intervention.Comment = "no source variant"
	case InterventionCommentOnly:
		intervention.Comment = "source comment only; lowered recipe is unchanged"
	case InterventionSemantic:
		if model.Cases[1].Left.ObservationRecipe != "exact" {
			return Intervention{}, fmt.Errorf("semantic source variant did not change direct observation recipe")
		}
		intervention.Semantic = true
		intervention.Target = "direct-unknown.left.observation_recipe"
		intervention.From = "missing"
		intervention.To = "exact"
	}
	return intervention, nil
}

func summarize(results []CaseResult, repositoryWrites int) Summary {
	summary := Summary{FixedDenominator: FixedDenominator, CaseTotal: len(results), RepositoryWrites: repositoryWrites}
	predicates := make([]string, 0, len(results))
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
		predicates = append(predicates, result.Predicate)
	}
	summary.NonExactCases = summary.CaseTotal - summary.ExactCases
	summary.NonExactNotPromoted = summary.NonExactCases
	summary.ClaimTransitionTotal = len(results)
	summary.DischargedClaims = summary.ExactCases
	summary.OpenClaims = summary.NonExactCases
	summary.DistinctPredicateCount = len(slices.Compact(slices.Sorted(predicates)))
	summary.PredicateDenominator = FixedDenominator
	return summary
}

func buildClaims(results []CaseResult) []ClaimTransition {
	claims := make([]ClaimTransition, 0, len(results))
	previous := ""
	for index, result := range results {
		claim := ClaimTransition{
			Sequence: index + 1, ClaimID: "composition-claim/" + result.PropositionDigest,
			From: "OPEN", To: transitionState(result.Result.State), Predicate: result.Predicate,
			Proposition: result.Proposition, PropositionDigest: result.PropositionDigest,
			TargetAddress: result.TargetAddress, TargetOperation: result.TargetOperation, TargetOutput: result.TargetOutput,
			Stage: result.Stage, Step: result.Step, Reason: result.Reason,
			RawSourceDigest: result.Provenance.RawSourceDigest, SemanticDigest: result.Provenance.SemanticIRDigest,
			RawEvidenceDigest: result.Provenance.RawEvidenceDigest, EvidenceDigest: result.EvidenceDigest,
			Provenance: result.Provenance, PreviousDigest: previous,
		}
		claim.Digest = transitionDigest(claim)
		claims = append(claims, claim)
		previous = claim.Digest
	}
	return claims
}

func predicateFor(state State) string {
	switch state {
	case StateExact:
		return "all-observations-exact"
	case StateDirectUnknown:
		return "direct-observation-unavailable"
	case StateDependencyBlocked:
		return "upstream-dependency-unresolved"
	case StateInvariantOnly:
		return "known-invariant-without-target-proof"
	case StateMixedUnresolved:
		return "mixed-unresolved-causes"
	default:
		return "unknown-composition-state"
	}
}

func propositionFor(recipe Case, predicate string) string {
	return strings.Join([]string{
		"case=" + recipe.ID, "left=" + recipe.Left.Operation, "right=" + recipe.Right.Operation,
		"predicate=" + predicate, "target_operation=" + recipe.MetaOperation, "target_output=MetaValue",
	}, ";")
}

func targetAddress(recipe Case) string {
	return "partialknowledgecomposition/" + recipe.SourceActivity + "/output/MetaValue"
}

func caseEvidenceDigest(result CaseResult) string {
	result.EvidenceDigest = ""
	return digestValue(result)
}

func subjectResolution(summary Summary) string {
	if summary.NonExactCases == 0 {
		return "EXACT"
	}
	return "PARTIAL_KNOWLEDGE"
}

func evidenceCoverage(receipt RawEvidenceReceipt) string {
	if receipt.SourceCases == FixedDenominator && len(receipt.Cases) == FixedDenominator {
		return "COMPLETE"
	}
	return "INCOMPLETE"
}

func promotionAuthorized(authority CapabilityObservation) bool {
	return authority.Available && authority.State == "EXACT"
}

func buildIndicators(receipt Receipt, rawEvidence RawEvidenceReceipt) []Indicator {
	summary := receipt.Summary
	indicators := []Indicator{
		indicator("PKC-FOUNDATION-SOURCE-001", "foundation", ProofFoundation, "parse-source-and-lower", receipt.SourceCases, receipt.SourceCasesTotal),
		indicator("PKC-FOUNDATION-EVIDENCE-002", "foundation", ProofFoundation, "validate-raw-evidence-receipt", len(rawEvidence.Cases), rawEvidence.SourceCasesTotal),
		indicator("PKC-COHERENCE-DIRECT-UNKNOWN-003", "driver", ProofCoherence, "derive-direct-unknown", summary.DirectUnknownCases, 1),
		indicator("PKC-COHERENCE-DEPENDENCY-BLOCK-004", "driver", ProofCoherence, "derive-dependency-block", summary.DependencyBlockedCases, 1),
		indicator("PKC-COHERENCE-INVARIANT-005", "driver", ProofFoundation, "preserve-known-invariant", summary.InvariantOnlyCases, 1),
		indicator("PKC-COHERENCE-MIXED-006", "driver", ProofCoherence, "retain-mixed-causes", summary.MixedUnresolvedCases, 1),
		indicator("PKC-REGRESSION-PREDICATES-007", "guardrail", ProofRegression, "count-distinct-propositions", summary.DistinctPredicateCount, summary.PredicateDenominator),
		indicator("PKC-REGRESSION-CLAIMS-008", "driver", ProofRegression, "replay-claim-transitions", summary.ClaimTransitionTotal, receipt.FixedDenominator),
		indicator("PKC-GUARDRAIL-NON-PROMOTION-009", "guardrail", ProofCoherence, "reject-non-exact-promotion", summary.NonExactNotPromoted, summary.NonExactCases),
		indicator("PKC-GUARDRAIL-SNAPSHOT-010", "guardrail", ProofFoundation, "observe-pre-post-snapshot", boolCount(rawEvidence.Workspace.RepositoryWrites == 0), 1),
		indicator("PKC-GUARDRAIL-AUTHORITY-011", "guardrail", ProofFoundation, "retain-unknown-promotion-capability", boolCount(receipt.AuthorityState == "UNKNOWN" && receipt.AuthorityResolution == "LOWER_RESOLUTION"), 1),
		indicator("PKC-OUTCOME-CALCULUS-012", "outcome", ProofRegression, "replay-composition-calculus", summary.ExactCases+summary.DirectUnknownCases+summary.DependencyBlockedCases+summary.InvariantOnlyCases+summary.MixedUnresolvedCases, receipt.FixedDenominator),
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
