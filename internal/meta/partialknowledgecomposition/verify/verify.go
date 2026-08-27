package verify

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

var headPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func Verify(input Input) (Report, error) {
	if input.Repository == "" || !headPattern.MatchString(input.HeadSHA) || input.SourcePath != "examples/partial-knowledge-composition/main.gooo" || len(input.Source) == 0 || len(input.RawEvidence) == 0 {
		return Report{}, fmt.Errorf("independent verifier identity is malformed")
	}
	model, err := parseSource(input.SourcePath, input.Source)
	if err != nil {
		return Report{}, err
	}
	rawEvidence, err := parseRawEvidence(input, model)
	if err != nil {
		return Report{}, err
	}
	intervention, err := interventionEvidence(input, model)
	if err != nil {
		return Report{}, err
	}
	expected := reconstruct(input, model, rawEvidence, intervention)
	var actual Receipt
	if err := json.Unmarshal(input.Receipt, &actual); err != nil {
		return Report{}, fmt.Errorf("decode producer receipt: %w", err)
	}
	if actual.Digest == "" || actual.Digest != receiptDigest(actual) || actual.Digest != expected.Digest {
		return Report{}, fmt.Errorf("receipt digest or independent raw evidence reconstruction differs")
	}
	actualRaw, _ := json.Marshal(actual)
	expectedRaw, _ := json.Marshal(expected)
	if string(actualRaw) != string(expectedRaw) {
		return Report{}, fmt.Errorf("producer receipt differs from independent source/evidence reconstruction")
	}
	return reportFrom(actual, actual.Digest == expected.Digest), nil
}

func interventionEvidence(input Input, model sourceModel) (Intervention, error) {
	mode := input.InterventionMode
	if mode == "" {
		mode = "none"
	}
	if mode != "none" && mode != "semantic" && mode != "comment-only" {
		return Intervention{}, fmt.Errorf("intervention %q is invalid", mode)
	}
	intervention := Intervention{Mode: mode, SourceDigest: digestBytes(input.Source), SemanticIRDigest: model.SemanticIRDigest}
	switch mode {
	case "none":
		intervention.Comment = "no source variant"
	case "comment-only":
		intervention.Comment = "source comment only; lowered recipe is unchanged"
	case "semantic":
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

func reconstruct(input Input, model sourceModel, rawEvidence RawEvidenceReceipt, intervention Intervention) Receipt {
	results := make([]CaseResult, 0, len(model.Cases))
	for index, recipe := range model.Cases {
		rawCase := rawEvidence.Cases[index]
		value := compose(rawCase.Left, rawCase.Right)
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
		Schema: "gooo/meta/partial-knowledge-composition-receipt/v2", Repository: input.Repository, HeadSHA: input.HeadSHA, SourcePath: input.SourcePath,
		SourceDigest: digestBytes(input.Source), SemanticIRDigest: model.SemanticIRDigest, RawEvidenceDigest: rawEvidence.Digest,
		SourceCases: len(model.Cases), SourceCasesTotal: 5,
		Producer: "partial-knowledge-producer", Consumer: "partial-knowledge-composition-consumer", MetaOperation: "compose-partial-knowledge", ProofChoice: "COHERENCE",
		Resolution: "CALCULUS", SubjectResolution: subjectResolution(summary), EvidenceCoverage: evidenceCoverage(rawEvidence),
		AuthorityState: rawEvidence.Authority.State, AuthorityResolution: rawEvidence.Authority.Resolution,
		AuthorityStage: rawEvidence.Authority.Stage, AuthorityStep: rawEvidence.Authority.Step, AuthorityReason: rawEvidence.Authority.Reason,
		Decision: "CALCULUS_PROVEN", Reason: "COMPOSITION_RULES_REPLAYED", FixedDenominator: 5,
		Cases: results, Claims: claims, Summary: summary, Intervention: intervention,
		RepositoryWrites: rawEvidence.Workspace.RepositoryWrites, PromotionAuthorized: promotionAuthorized(rawEvidence.Authority),
	}
	receipt.Indicators = buildIndicators(receipt, rawEvidence)
	receipt.SemanticProjectionDigest = semanticProjectionDigest(receipt)
	receipt.Digest = receiptDigest(receipt)
	return receipt
}

func summarize(results []CaseResult, repositoryWrites int) Summary {
	summary := Summary{FixedDenominator: 5, CaseTotal: len(results), RepositoryWrites: repositoryWrites}
	predicates := make([]string, 0, len(results))
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
		predicates = append(predicates, result.Predicate)
	}
	summary.NonExactCases = summary.CaseTotal - summary.ExactCases
	summary.NonExactNotPromoted = summary.NonExactCases
	summary.OpenClaims = summary.NonExactCases
	summary.DischargedClaims = summary.ExactCases
	summary.ClaimTransitionTotal = len(results)
	summary.DistinctPredicateCount = len(slices.Compact(slices.Sorted(predicates)))
	summary.PredicateDenominator = 5
	return summary
}

func buildClaims(results []CaseResult) []ClaimTransition {
	claims := make([]ClaimTransition, 0, len(results))
	previous := ""
	for index, result := range results {
		claim := ClaimTransition{
			Sequence: index + 1, ClaimID: "composition-claim/" + result.PropositionDigest, From: "OPEN", To: transitionState(result.Result.State),
			Predicate: result.Predicate, Proposition: result.Proposition, PropositionDigest: result.PropositionDigest,
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

func predicateFor(state string) string {
	switch state {
	case exact:
		return "all-observations-exact"
	case directUnknown:
		return "direct-observation-unavailable"
	case dependencyBlocked:
		return "upstream-dependency-unresolved"
	case invariantOnly:
		return "known-invariant-without-target-proof"
	case mixedUnresolved:
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
	if receipt.SourceCases == 5 && len(receipt.Cases) == 5 {
		return "COMPLETE"
	}
	return "INCOMPLETE"
}

func promotionAuthorized(authority CapabilityObservation) bool {
	return authority.Available && authority.State == "EXACT"
}

func buildIndicators(receipt Receipt, rawEvidence RawEvidenceReceipt) []Indicator {
	summary := receipt.Summary
	items := []Indicator{
		makeIndicator("PKC-FOUNDATION-SOURCE-001", "foundation", "FOUNDATION", "parse-source-and-lower", receipt.SourceCases, receipt.SourceCasesTotal),
		makeIndicator("PKC-FOUNDATION-EVIDENCE-002", "foundation", "FOUNDATION", "validate-raw-evidence-receipt", len(rawEvidence.Cases), rawEvidence.SourceCasesTotal),
		makeIndicator("PKC-COHERENCE-DIRECT-UNKNOWN-003", "driver", "COHERENCE", "derive-direct-unknown", summary.DirectUnknownCases, 1),
		makeIndicator("PKC-COHERENCE-DEPENDENCY-BLOCK-004", "driver", "COHERENCE", "derive-dependency-block", summary.DependencyBlockedCases, 1),
		makeIndicator("PKC-COHERENCE-INVARIANT-005", "driver", "FOUNDATION", "preserve-known-invariant", summary.InvariantOnlyCases, 1),
		makeIndicator("PKC-COHERENCE-MIXED-006", "driver", "COHERENCE", "retain-mixed-causes", summary.MixedUnresolvedCases, 1),
		makeIndicator("PKC-REGRESSION-PREDICATES-007", "guardrail", "REGRESSION", "count-distinct-propositions", summary.DistinctPredicateCount, summary.PredicateDenominator),
		makeIndicator("PKC-REGRESSION-CLAIMS-008", "driver", "REGRESSION", "replay-claim-transitions", summary.ClaimTransitionTotal, receipt.FixedDenominator),
		makeIndicator("PKC-GUARDRAIL-NON-PROMOTION-009", "guardrail", "COHERENCE", "reject-non-exact-promotion", summary.NonExactNotPromoted, summary.NonExactCases),
		makeIndicator("PKC-GUARDRAIL-SNAPSHOT-010", "guardrail", "FOUNDATION", "observe-pre-post-snapshot", boolCount(rawEvidence.Workspace.RepositoryWrites == 0), 1),
		makeIndicator("PKC-GUARDRAIL-AUTHORITY-011", "guardrail", "FOUNDATION", "retain-unknown-promotion-capability", boolCount(receipt.AuthorityState == "UNKNOWN" && receipt.AuthorityResolution == "LOWER_RESOLUTION"), 1),
		makeIndicator("PKC-OUTCOME-CALCULUS-012", "outcome", "REGRESSION", "replay-composition-calculus", summary.ExactCases+summary.DirectUnknownCases+summary.DependencyBlockedCases+summary.InvariantOnlyCases+summary.MixedUnresolvedCases, receipt.FixedDenominator),
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

func boolCount(value bool) int {
	if value {
		return 1
	}
	return 0
}

func reportFrom(receipt Receipt, independent bool) Report {
	result := Report{
		Schema: "gooo/meta/partial-knowledge-composition-verification/v2", Repository: receipt.Repository, HeadSHA: receipt.HeadSHA, Status: "VERIFIED",
		Decision: receipt.Decision, Resolution: receipt.Resolution, SubjectResolution: receipt.SubjectResolution, EvidenceCoverage: receipt.EvidenceCoverage,
		AuthorityState: receipt.AuthorityState, AuthorityResolution: receipt.AuthorityResolution, FixedDenominator: receipt.FixedDenominator,
		SourceCases: receipt.SourceCases, SourceCasesTotal: receipt.SourceCasesTotal, ExactCases: receipt.Summary.ExactCases,
		DirectUnknownCases: receipt.Summary.DirectUnknownCases, DependencyBlockedCases: receipt.Summary.DependencyBlockedCases,
		InvariantOnlyCases: receipt.Summary.InvariantOnlyCases, MixedUnresolvedCases: receipt.Summary.MixedUnresolvedCases,
		TopSuccessCases: receipt.Summary.TopSuccessCases, NonExactNotPromoted: receipt.Summary.NonExactNotPromoted,
		OpenClaims: receipt.Summary.OpenClaims, ClaimTransitionTotal: receipt.Summary.ClaimTransitionTotal,
		DistinctPredicateCount: receipt.Summary.DistinctPredicateCount, PredicateDenominator: receipt.Summary.PredicateDenominator,
		RepositoryWrites: receipt.RepositoryWrites, PromotionAuthorized: receipt.PromotionAuthorized, IndependentEvaluator: independent,
		SourceSemanticDigest: receipt.SemanticIRDigest, RawEvidenceDigest: receipt.RawEvidenceDigest, ReceiptDigest: receipt.Digest,
		SemanticProjectionDigest: receipt.SemanticProjectionDigest,
	}
	result.Digest = reportDigest(result)
	return result
}

func reportDigest(report Report) string {
	report.Digest = ""
	return digestValue(report)
}
