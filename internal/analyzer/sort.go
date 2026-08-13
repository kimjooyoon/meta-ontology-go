package analyzer

import (
	"go/ast"
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

func identityEqual(left, right Identity) bool {
	return left.Namespace == right.Namespace && left.ID == right.ID
}

func identitySliceLess(left, right []Identity) bool {
	for index := 0; index < len(left) && index < len(right); index++ {
		if identityEqual(left[index], right[index]) {
			continue
		}
		return identityLess(left[index], right[index])
	}
	return len(left) < len(right)
}

func sortRegistrations(entries []Registration) {
	sort.SliceStable(entries, func(i, j int) bool {
		left, right := entries[i], entries[j]
		if left.Ref.canonical() != right.Ref.canonical() {
			return left.Ref.canonical() < right.Ref.canonical()
		}
		if left.Kind != right.Kind {
			return left.Kind < right.Kind
		}
		if !identityEqual(left.Identity, right.Identity) {
			return identityLess(left.Identity, right.Identity)
		}
		return spanLess(left.Span, right.Span)
	})
}

func uniqueRegistrations(entries []Registration) []Registration {
	unique := make([]Registration, 0, len(entries))
	seen := make(map[string]bool, len(entries))
	for _, entry := range entries {
		key := entry.Ref.canonical() + "\x00" + string(entry.Kind) + "\x00" + entry.Identity.Namespace + "\x00" + entry.Identity.ID
		if seen[key] {
			continue
		}
		seen[key] = true
		unique = append(unique, entry)
	}
	sortRegistrations(unique)
	return unique
}

func isSelectorChild(parent ast.Node) bool {
	_, ok := parent.(*ast.SelectorExpr)
	return ok
}

func isCallTarget(parent ast.Node, ident *ast.Ident) bool {
	call, ok := parent.(*ast.CallExpr)
	return ok && call.Fun == ident
}

func isDeclarationName(parent ast.Node, ident *ast.Ident) bool {
	switch current := parent.(type) {
	case *ast.AssignStmt:
		for _, left := range current.Lhs {
			if left == ident {
				return true
			}
		}
	case *ast.ValueSpec:
		for _, name := range current.Names {
			if name == ident {
				return true
			}
		}
	case *ast.Field:
		for _, name := range current.Names {
			if name == ident {
				return true
			}
		}
	}
	return false
}
