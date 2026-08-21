package analyzer

import (
	"sort"
)

func (d *SemanticDelta) sort() {
	sort.SliceStable(d.Added, func(i, j int) bool { return factLess(d.Added[i], d.Added[j]) })
	sort.SliceStable(d.Candidates, func(i, j int) bool { return candidateLess(d.Candidates[i], d.Candidates[j]) })
	sort.SliceStable(d.ImplementationDetails, func(i, j int) bool {
		left, right := d.ImplementationDetails[i], d.ImplementationDetails[j]
		if spanLess(left.Span, right.Span) {
			return true
		}
		if spanLess(right.Span, left.Span) {
			return false
		}
		if left.Reference != right.Reference {
			return left.Reference < right.Reference
		}
		if left.IdentityState != right.IdentityState {
			return left.IdentityState < right.IdentityState
		}
		return left.Reason < right.Reason
	})
}
func factLess(left, right Fact) bool {
	if spanLess(left.Span, right.Span) {
		return true
	}
	if spanLess(right.Span, left.Span) {
		return false
	}
	if !identityEqual(left.Subject, right.Subject) {
		return identityLess(left.Subject, right.Subject)
	}
	if left.Relation != right.Relation {
		return left.Relation < right.Relation
	}
	return identityLess(left.Object, right.Object)
}
func candidateLess(left, right Candidate) bool {
	if spanLess(left.Span, right.Span) {
		return true
	}
	if spanLess(right.Span, left.Span) {
		return false
	}
	if !identityEqual(left.Subject, right.Subject) {
		return identityLess(left.Subject, right.Subject)
	}
	if left.Relation != right.Relation {
		return left.Relation < right.Relation
	}
	if left.Reference != right.Reference {
		return left.Reference < right.Reference
	}
	return identitySliceLess(left.Options, right.Options)
}
func spanLess(left, right Span) bool {
	if left.Filename != right.Filename {
		return left.Filename < right.Filename
	}
	if left.Start.Offset != right.Start.Offset {
		return left.Start.Offset < right.Start.Offset
	}
	return left.End.Offset < right.End.Offset
}
func identityLess(left, right Identity) bool {
	if left.Namespace != right.Namespace {
		return left.Namespace < right.Namespace
	}
	return left.ID < right.ID
}
