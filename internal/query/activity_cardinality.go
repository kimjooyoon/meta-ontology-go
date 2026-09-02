package query

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

var ErrInvalidActivitySelector = errors.New("invalid activity selector")

func (graph Graph) ResolveActivityCardinality(selector ActivitySelector) (ActivityCardinalityResolution, error) {
	normalized, err := selector.normalized()
	if err != nil {
		return ActivityCardinalityResolution{}, err
	}
	metadata := graph.Metadata()
	result := ActivityCardinalityResolution{
		Schema: ActivityCardinalityResolutionSchema, Selector: normalized,
		Matches: make([]ActivityResolutionMatch, 0),
		Subject: ActivityResolutionSubject{
			GraphHash: metadata.GraphHash, SemanticDigest: metadata.SemanticDigest,
			Namespace: metadata.Namespace, SourceStatus: metadata.SourceStatus,
		},
	}
	for _, node := range graph.Nodes() {
		if node.Kind != ActivityNodeKind || !normalized.matches(node) {
			continue
		}
		result.Matches = append(result.Matches, ActivityResolutionMatch{
			ID: node.ID.String(), Namespace: node.Namespace, Name: node.Name,
		})
	}
	result.Occurrences = len(result.Matches)
	result.Decision, result.Claim = activityCardinalityClaim(result.Occurrences)
	return result, nil
}

func (selector ActivitySelector) normalized() (ActivitySelector, error) {
	selector.Namespace = strings.TrimSpace(selector.Namespace)
	selector.Name = normalizeDisplayName(selector.Name)
	selector.IDPrefix = strings.TrimSpace(selector.IDPrefix)
	if strings.IndexFunc(selector.Namespace, unicode.IsSpace) >= 0 ||
		strings.IndexFunc(selector.IDPrefix, unicode.IsSpace) >= 0 {
		return ActivitySelector{}, fmt.Errorf("%w: namespace and ID prefix cannot contain whitespace", ErrInvalidActivitySelector)
	}
	if selector.Namespace == "" && selector.Name == "" && selector.IDPrefix == "" {
		return ActivitySelector{}, fmt.Errorf("%w: at least one field is required", ErrInvalidActivitySelector)
	}
	return selector, nil
}

func (selector ActivitySelector) matches(node Node) bool {
	return (selector.Namespace == "" || node.Namespace == selector.Namespace) &&
		(selector.Name == "" || node.hasName(selector.Name)) &&
		(selector.IDPrefix == "" || strings.HasPrefix(node.ID.String(), selector.IDPrefix))
}

func activityCardinalityClaim(count int) (ActivityResolutionDecision, ActivityResolutionClaim) {
	claim := ActivityResolutionClaim{Stage: "RESOLUTION", Step: "RESOLVE_ACTIVITY_CARDINALITY", BlockedBy: []string{}}
	switch count {
	case 0:
		claim.State, claim.Reason = ActivityResolutionUnknown, "ACTIVITY_NOT_FOUND"
		claim.UnknownClass, claim.NextOperation, claim.ProofChoice = "DIRECT_MISSING", "DECLARE_OR_WIDEN_ACTIVITY_SELECTOR", "FOUNDATION"
	case 1:
		claim.State, claim.Reason = ActivityResolutionClosed, "ACTIVITY_UNIQUELY_RESOLVED"
		claim.NextOperation, claim.ProofChoice = "USE_RESOLVED_ACTIVITY", "COHERENCE"
	default:
		claim.State, claim.Reason = ActivityResolutionRefuted, "AMBIGUOUS_ACTIVITY_BINDING"
		claim.NextOperation, claim.ProofChoice = "NARROW_ACTIVITY_SELECTOR", "REGRESSION"
	}
	return claim.State, claim
}
