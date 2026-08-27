package proofchoicealgebra

import (
	"fmt"
	"strings"
)

func Evaluate(path string, source []byte) Receipt {
	bundle, issues := parseBundle(path, source)
	receipt := Receipt{
		Schema: Schema, Decision: Pass, Reason: "PROOF_CHOICES_COMPOSED",
		Resolution: Exact, SourcePath: path, SourceDigest: digestSource(source),
		FixedDenom: FixedDenominator, Items: bundle.Items, Transitions: bundle.Transitions,
		Effects: Effects{RepositoryWrites: 0, MutationAuthority: false},
	}
	receipt.Summary = summarize(bundle)
	for _, item := range bundle.Items {
		receipt.Indicators = append(receipt.Indicators, indicatorForItem(item))
	}
	for _, transition := range bundle.Transitions {
		receipt.Indicators = append(receipt.Indicators, indicatorForTransition(transition))
	}
	failure := firstIssue(issues)
	if failure == "" {
		failure = validateBundle(bundle)
	}
	if failure != "" {
		receipt.Decision, receipt.Reason, receipt.Resolution = FailClosed, failure, FailClosed
	}
	for index := range receipt.Indicators {
		if receipt.Decision == FailClosed {
			receipt.Indicators[index].Decision = FailClosed
		}
	}
	receipt.Summary.Unknowns = countUnknowns(bundle)
	receipt.Summary.Contradictions = countContradictions(bundle)
	receipt.Summary.Compositions = len(bundle.Items) + len(bundle.Transitions)
	digest, err := digestReceipt(receipt)
	if err != nil {
		receipt.Decision, receipt.Reason, receipt.Resolution = FailClosed, "RECEIPT_DIGEST_UNKNOWN", FailClosed
	} else {
		receipt.Digest = digest
	}
	return receipt
}

func validateBundle(bundle Bundle) string {
	if len(bundle.Items) == 0 {
		return "NO_PROOF_CHOICES"
	}
	byID := make(map[string]Item, len(bundle.Items))
	claims := make(map[string]Item)
	transitions := make(map[string]bool)
	for _, item := range bundle.Items {
		if item.ID == "" || item.Statement == "" || !item.Choice.Valid() {
			if !item.Choice.Valid() {
				return "PROOF_CHOICE_MISSING"
			}
			return "ITEM_METADATA_MISSING"
		}
		if item.Kind != Claim && item.Kind != Metric {
			return "ITEM_KIND_UNKNOWN"
		}
		if metadataUnknown(item.Producer, item.Consumer, item.MetaOperation, item.Stage, item.Step, item.Reason) {
			return "UNKNOWN_CONTEXT"
		}
		if item.Kind == Metric && (item.Denominator != FixedDenominator || item.Numerator < 0 || item.Numerator > item.Denominator) {
			return "FIXED_DENOMINATOR_MISMATCH"
		}
		if previous, exists := byID[item.ID]; exists && !sameItem(previous, item) {
			return "PROOF_CHOICE_CONTRADICTION"
		}
		byID[item.ID] = item
		if item.Kind == Claim {
			claims[item.ID] = item
		}
	}
	for _, transition := range bundle.Transitions {
		claim, exists := claims[transition.ClaimID]
		if !exists || !transition.Persistent || transition.From == "" || transition.To == "" {
			return "PERSISTENT_TRANSITION_MISMATCH"
		}
		if transition.Choice != claim.Choice || !transition.Choice.Valid() {
			return "PROOF_CHOICE_CONTRADICTION"
		}
		if metadataUnknown(transition.Producer, transition.Consumer, transition.MetaOperation, transition.Stage, transition.Step, transition.Reason) || strings.EqualFold(transition.From, "UNKNOWN") || strings.EqualFold(transition.To, "UNKNOWN") {
			return "UNKNOWN_CONTEXT"
		}
		transitions[transition.ClaimID] = true
	}
	for id := range claims {
		if !transitions[id] {
			return "PERSISTENT_TRANSITION_MISSING"
		}
	}
	return ""
}

func sameItem(left, right Item) bool {
	left.Line, right.Line = 0, 0
	return left == right
}

func metadataUnknown(values ...string) bool {
	for _, value := range values {
		if value == "" || strings.EqualFold(value, "UNKNOWN") {
			return true
		}
	}
	return false
}

func firstIssue(issues []issue) string {
	if len(issues) == 0 {
		return ""
	}
	return issues[0].Reason
}

func summarize(bundle Bundle) Summary {
	var summary Summary
	summary.Items, summary.Transitions = len(bundle.Items), len(bundle.Transitions)
	summary.FixedDenominator = FixedDenominator
	for _, item := range bundle.Items {
		if item.Kind == Claim {
			summary.Claims++
		} else if item.Kind == Metric {
			summary.Metrics++
		}
		if item.Choice.Valid() {
			summary.ChoicesExplicit++
		}
	}
	for _, transition := range bundle.Transitions {
		if transition.Persistent {
			summary.PersistentTransitions++
		}
	}
	if summary.Items > 0 {
		summary.ChoiceCoverageBPS = summary.ChoicesExplicit * 10000 / summary.Items
	}
	return summary
}

func countUnknowns(bundle Bundle) int {
	count := 0
	for _, item := range bundle.Items {
		if !item.Choice.Valid() || metadataUnknown(item.Producer, item.Consumer, item.MetaOperation, item.Stage, item.Step, item.Reason) {
			count++
		}
	}
	for _, transition := range bundle.Transitions {
		if !transition.Choice.Valid() || metadataUnknown(transition.Producer, transition.Consumer, transition.MetaOperation, transition.Stage, transition.Step, transition.Reason) {
			count++
		}
	}
	return count
}

func countContradictions(bundle Bundle) int {
	seen := map[string]Choice{}
	count := 0
	for _, item := range bundle.Items {
		if choice, exists := seen[item.ID]; exists && choice != item.Choice {
			count++
		}
		seen[item.ID] = item.Choice
	}
	return count
}

func indicatorForItem(item Item) Indicator {
	decision := Pass
	if !item.Choice.Valid() {
		decision = FailClosed
	}
	return Indicator{ID: item.ID, Kind: string(item.Kind), Choice: item.Choice,
		Decision: decision, Relation: "choice", Value: item.Choice.String(), Limit: "exactly-one"}
}

func indicatorForTransition(transition Transition) Indicator {
	decision := Pass
	if !transition.Choice.Valid() {
		decision = FailClosed
	}
	return Indicator{ID: "transition:" + transition.ClaimID, Kind: "PERSISTENT_TRANSITION",
		Choice: transition.Choice, Decision: decision, Relation: "preserves", Value: transition.To, Limit: transition.From}
}

// Combine is the proof-choice algebra's disjoint-union operator (⊕). Exact
// duplicate witnesses are idempotent; conflicting values for one ID are not.
func Combine(left, right Bundle) (Bundle, error) {
	if err := validateBundleShape(left); err != nil {
		return Bundle{}, err
	}
	if err := validateBundleShape(right); err != nil {
		return Bundle{}, err
	}
	result := Bundle{Items: append([]Item(nil), left.Items...), Transitions: append([]Transition(nil), left.Transitions...)}
	for _, item := range right.Items {
		found := false
		for _, existing := range result.Items {
			if existing.ID != item.ID {
				continue
			}
			found = true
			if !sameItem(existing, item) {
				return Bundle{}, fmt.Errorf("PROOF_CHOICE_CONTRADICTION: %s", item.ID)
			}
		}
		if !found {
			result.Items = append(result.Items, item)
		}
	}
	for _, transition := range right.Transitions {
		found := false
		for _, existing := range result.Transitions {
			if existing.ClaimID == transition.ClaimID && existing.From == transition.From && existing.To == transition.To {
				found = true
				if existing.Choice != transition.Choice || !sameTransition(existing, transition) {
					return Bundle{}, fmt.Errorf("PROOF_CHOICE_CONTRADICTION: transition:%s", transition.ClaimID)
				}
			}
		}
		if !found {
			result.Transitions = append(result.Transitions, transition)
		}
	}
	if failure := validateBundle(result); failure != "" {
		return Bundle{}, fmt.Errorf("%s", failure)
	}
	return result, nil
}

func validateBundleShape(bundle Bundle) error {
	for _, item := range bundle.Items {
		if !item.Choice.Valid() {
			return fmt.Errorf("PROOF_CHOICE_MISSING")
		}
	}
	return nil
}

func sameTransition(left, right Transition) bool {
	left.Line, right.Line = 0, 0
	return left == right
}
