package analyzer

import (
	"go/ast"
)

func (c *factCollector) recordResolution(
	result resolution, relation Relation, expression ast.Expr, reason string, origin ObservationOrigin,
) {
	if result.state == unresolved {
		c.addImplementationDetail(expression, reason, IdentityUnresolved)
		return
	}
	if result.state == ambiguous {
		options := uniqueIdentities(result.entries)
		if len(options) == 0 {
			c.addImplementationDetail(expression, reason, IdentityAmbiguous)
			return
		}
		c.delta.Candidates = append(c.delta.Candidates, Candidate{
			Subject:   c.subject,
			Relation:  relation,
			Reference: expressionName(expression),
			Options:   options,
			Span:      c.span(expression),
			Reason:    "multiple registered semantic symbols match",
			Origin:    origin,
		})
		return
	}
	if len(result.entries) != 1 || !result.entries[0].Identity.Valid() {
		c.addImplementationDetail(expression, reason, IdentityInvalid)
		return
	}
	c.delta.Added = append(c.delta.Added, Fact{
		Subject:  c.subject,
		Relation: relation,
		Object:   result.entries[0].Identity,
		Span:     c.span(expression),
		Origin:   origin,
	})
}
func (c *factCollector) addImplementationDetail(
	expression ast.Expr, reason string, state IdentityState,
) {
	c.delta.ImplementationDetails = append(c.delta.ImplementationDetails, ImplementationDetail{
		Reference: expressionName(expression), Span: c.span(expression),
		Reason: reason + ": " + identityStateReason(state), IdentityState: state,
	})
}
func identityStateReason(state IdentityState) string {
	switch state {
	case IdentityAmbiguous:
		return "ambiguous semantic symbol"
	case IdentityInvalid:
		return "invalid registered identity"
	default:
		return "unregistered semantic symbol"
	}
}
func (c *factCollector) span(node ast.Node) Span {
	return spanFor(c.resolver.fileSet, node)
}
func relationForCall(entries []Registration) Relation {
	if allOfKind(entries, KindActivity) {
		return RelationInvokes
	}
	if allOfKind(entries, KindEntity) {
		return RelationUses
	}
	return RelationReferences
}
