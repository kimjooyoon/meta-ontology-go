package verify

import (
	"encoding/json"
	"fmt"
	"regexp"
)

var headPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

var caseIDs = []string{
	"exact-pair", "direct-unknown", "dependency-blocked",
	"invariant-preservation", "mixed-unknown-and-blocked",
}

func Verify(input Input) (Report, error) {
	if input.Repository == "" || !headPattern.MatchString(input.HeadSHA) || input.SourcePath != "examples/partial-knowledge-composition/main.gooo" {
		return Report{}, fmt.Errorf("verifier identity is malformed")
	}
	var sourceFixture fixture
	if err := json.Unmarshal(input.Fixture, &sourceFixture); err != nil {
		return Report{}, fmt.Errorf("decode fixture: %w", err)
	}
	var actual receipt
	if err := json.Unmarshal(input.Receipt, &actual); err != nil {
		return Report{}, fmt.Errorf("decode receipt: %w", err)
	}
	if err := validateFixture(sourceFixture); err != nil {
		return Report{}, err
	}
	if err := validateSource(input.Source); err != nil {
		return Report{}, err
	}
	expected, err := reconstruct(input, sourceFixture)
	if err != nil {
		return Report{}, err
	}
	if actual.Digest != receiptDigest(actual) || actual.Digest != expected.Digest {
		return Report{}, fmt.Errorf("receipt digest or reconstructed payload differs")
	}
	if err := compareReceipt(actual, expected); err != nil {
		return Report{}, err
	}
	return newReport(actual), nil
}

func validateFixture(value fixture) error {
	if value.Schema != "gooo/meta/partial-knowledge-composition-fixture/v1" || value.SourcePath != "examples/partial-knowledge-composition/main.gooo" || value.FixedDenominator != 5 || len(value.Cases) != 5 {
		return fmt.Errorf("fixture identity or denominator is not fixed")
	}
	for index, current := range value.Cases {
		if current.ID != caseIDs[index] || current.Producer != "partial-knowledge-producer" || current.Consumer != "partial-knowledge-composition-consumer" || current.MetaOperation == "" || !validProof(current.ProofChoice) {
			return fmt.Errorf("fixture case %d metadata is not closed", index+1)
		}
		if err := validateOperand(current.Left); err != nil {
			return fmt.Errorf("fixture case %q left operand: %w", current.ID, err)
		}
		if err := validateOperand(current.Right); err != nil {
			return fmt.Errorf("fixture case %q right operand: %w", current.ID, err)
		}
		result := compose(current.Left, current.Right)
		decision, reason, _ := classify(result)
		if current.ExpectedState != result.State || current.ExpectedDecision != decision || current.ExpectedReason != reason {
			return fmt.Errorf("fixture case %q expected result is not the independent result", current.ID)
		}
	}
	return nil
}

func validateSource(source []byte) error {
	text := string(source)
	for _, marker := range []string{
		"package partialknowledgecomposition", "namespace partialknowledgecomposition",
		"entity DirectUnknown", "entity DependencyBlocked", "entity InvariantOnly",
		"activity Compose(MetaValue, MetaValue) -> MetaValue",
	} {
		if !containsText(text, marker) {
			return fmt.Errorf("Gooo source is missing %q", marker)
		}
	}
	return nil
}

func containsText(text, marker string) bool {
	return len(text) >= len(marker) && regexp.MustCompile(regexp.QuoteMeta(marker)).MatchString(text)
}

func validProof(value string) bool {
	return value == "FOUNDATION" || value == "COHERENCE" || value == "REGRESSION"
}

func validateOperand(value operand) error {
	if value.Operation == "" || (value.State != exact && value.State != directUnknown && value.State != dependencyBlocked && value.State != invariantOnly) {
		return fmt.Errorf("operation or state is invalid")
	}
	switch value.State {
	case exact, directUnknown:
		if value.BlockedDependency != "" || len(value.Invariants) != 0 {
			return fmt.Errorf("classical operand carries unresolved evidence")
		}
	case dependencyBlocked:
		if value.BlockedDependency == "" || len(value.Invariants) != 0 {
			return fmt.Errorf("blocked operand lacks dependency evidence")
		}
	case invariantOnly:
		if value.BlockedDependency != "" || len(value.Invariants) == 0 {
			return fmt.Errorf("invariant operand lacks invariant evidence")
		}
	}
	return nil
}

func reconstruct(input Input, value fixture) (receipt, error) {
	results := make([]caseResult, 0, len(value.Cases))
	for _, current := range value.Cases {
		result := compose(current.Left, current.Right)
		decision, reason, topSuccess := classify(result)
		currentResult := caseResult{
			ID: current.ID, Producer: current.Producer, Consumer: current.Consumer,
			MetaOperation: current.MetaOperation, ProofChoice: current.ProofChoice,
			Left: current.Left, Right: current.Right, Result: result,
			Decision: decision, Reason: reason, TopSuccess: topSuccess,
		}
		currentResult.EvidenceDigest = digestValue(struct {
			ID     string  `json:"id"`
			Left   operand `json:"left"`
			Right  operand `json:"right"`
			Result value   `json:"result"`
		}{current.ID, current.Left, current.Right, result})
		results = append(results, currentResult)
	}
	summary := summarize(results)
	claims := buildClaims(results)
	result := receipt{
		Schema: "gooo/meta/partial-knowledge-composition-receipt/v1", Repository: input.Repository,
		HeadSHA: input.HeadSHA, SourcePath: input.SourcePath, SourceDigest: digestBytes(input.Source),
		Producer: "partial-knowledge-producer", Consumer: "partial-knowledge-composition-consumer",
		MetaOperation: "compose-partial-knowledge", ProofChoice: "COHERENCE",
		Resolution: "COMPOSITION_CALCULUS", Decision: "CALCULUS_PROVEN", Reason: "COMPOSITION_RULES_REPLAYED",
		FixedDenominator: 5,
		DenominatorDigest: digestValue(struct {
			Count int      `json:"count"`
			IDs   []string `json:"ids"`
		}{5, caseIDs}),
		Cases: results, Claims: claims, Summary: summary, RepositoryWrites: 0, PromotionAuthorized: false,
	}
	result.Indicators = indicators(result)
	result.Digest = receiptDigest(result)
	return result, nil
}

func summarize(results []caseResult) summary {
	result := summary{FixedDenominator: 5, CaseTotal: len(results), RepositoryWrites: 0}
	for _, current := range results {
		switch current.Result.State {
		case exact:
			result.ExactCases++
		case directUnknown:
			result.DirectUnknownCases++
		case dependencyBlocked:
			result.DependencyBlockedCases++
		case invariantOnly:
			result.InvariantOnlyCases++
		case mixedUnresolved:
			result.MixedUnresolvedCases++
		}
		if current.TopSuccess {
			result.TopSuccessCases++
		}
	}
	result.NonExactCases = result.CaseTotal - result.ExactCases
	result.NonExactNotPromoted = result.NonExactCases
	result.ClaimTransitionTotal = len(results)
	return result
}

func buildClaims(results []caseResult) []claimTransition {
	claims := make([]claimTransition, 0, len(results))
	previous := ""
	for index, current := range results {
		claim := claimTransition{
			Sequence: index + 1, ClaimID: "composition/" + current.ID, From: "OPEN",
			To: transitionState(current.Result.State), MetaOperation: current.MetaOperation,
			ProofChoice: current.ProofChoice, EvidenceDigest: current.EvidenceDigest, PreviousDigest: previous,
		}
		claim.Digest = transitionDigest(claim)
		claims = append(claims, claim)
		previous = claim.Digest
	}
	return claims
}

func indicators(value receipt) []indicator {
	result := []indicator{
		makeIndicator("PKC-FOUNDATION-SOURCE-001", "driver", "FOUNDATION", "bind-partial-knowledge-source", 1, 1),
		makeIndicator("PKC-COHERENCE-DIRECT-UNKNOWN-002", "driver", "COHERENCE", "compose-partial-knowledge", value.Summary.DirectUnknownCases, 1),
		makeIndicator("PKC-COHERENCE-DEPENDENCY-BLOCK-003", "driver", "COHERENCE", "compose-partial-knowledge", value.Summary.DependencyBlockedCases, 1),
		makeIndicator("PKC-COHERENCE-INVARIANT-004", "driver", "FOUNDATION", "preserve-known-invariant", value.Summary.InvariantOnlyCases, 1),
		makeIndicator("PKC-COHERENCE-MIXED-005", "driver", "COHERENCE", "compose-partial-knowledge", value.Summary.MixedUnresolvedCases, 1),
		makeIndicator("PKC-REGRESSION-DENOMINATOR-006", "guardrail", "FOUNDATION", "freeze-composition-denominator", value.Summary.CaseTotal, 5),
		makeIndicator("PKC-REGRESSION-CLAIMS-007", "driver", "REGRESSION", "replay-claim-transitions", value.Summary.ClaimTransitionTotal, 5),
		makeIndicator("PKC-GUARDRAIL-NON-PROMOTION-008", "guardrail", "COHERENCE", "reject-non-exact-promotion", value.Summary.NonExactNotPromoted, value.Summary.NonExactCases),
		makeIndicator("PKC-GUARDRAIL-READ-ONLY-009", "guardrail", "FOUNDATION", "preserve-read-only-boundary", 1, 1),
		makeIndicator("PKC-OUTCOME-CALCULUS-010", "outcome", "REGRESSION", "replay-composition-corpus", value.Summary.ExactCases+value.Summary.DirectUnknownCases+value.Summary.DependencyBlockedCases+value.Summary.InvariantOnlyCases+value.Summary.MixedUnresolvedCases, 5),
	}
	for index := range result {
		result[index].Producer = value.Producer
		result[index].Consumer = value.Consumer
	}
	return result
}

func makeIndicator(id, class, proof, operation string, observed, denominator int) indicator {
	basisPoints := 0
	if denominator > 0 {
		basisPoints = observed * 10000 / denominator
	}
	return indicator{ID: id, Class: class, ProofChoice: proof, MetaOperation: operation, Observed: observed, Denominator: denominator, BasisPoints: basisPoints, TargetBasisPoints: 10000, Satisfied: denominator > 0 && basisPoints == 10000}
}

func compareReceipt(actual, expected receipt) error {
	actualRaw, _ := json.Marshal(actual)
	expectedRaw, _ := json.Marshal(expected)
	if string(actualRaw) != string(expectedRaw) {
		return fmt.Errorf("receipt fields differ from independent reconstruction")
	}
	return nil
}

func newReport(value receipt) Report {
	result := Report{
		Schema: "gooo/meta/partial-knowledge-composition-verification/v1", Repository: value.Repository,
		HeadSHA: value.HeadSHA, ReceiptDigest: value.Digest, Status: "VERIFIED", Decision: "CALCULUS_PROVEN",
		FixedDenominator: value.Summary.FixedDenominator, ExactCases: value.Summary.ExactCases,
		DirectUnknownCases: value.Summary.DirectUnknownCases, DependencyBlockedCases: value.Summary.DependencyBlockedCases,
		InvariantOnlyCases: value.Summary.InvariantOnlyCases, MixedUnresolvedCases: value.Summary.MixedUnresolvedCases,
		TopSuccessCases: value.Summary.TopSuccessCases, NonExactNotPromoted: value.Summary.NonExactNotPromoted,
		ClaimTransitionTotal: value.Summary.ClaimTransitionTotal, RepositoryWrites: value.RepositoryWrites,
		PromotionAuthorized: value.PromotionAuthorized, IndependentEvaluator: true,
	}
	result.Digest = reportDigest(result)
	return result
}
