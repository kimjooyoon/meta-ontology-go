package analyzer

import (
	"go/ast"
	"sort"
)

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
