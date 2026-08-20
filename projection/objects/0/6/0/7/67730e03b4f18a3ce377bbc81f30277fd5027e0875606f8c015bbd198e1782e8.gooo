package semanticdelta

func factLess(left, right Fact) bool {
	if left.Subject != right.Subject {
		return left.Subject < right.Subject
	}
	if left.Predicate != right.Predicate {
		return left.Predicate < right.Predicate
	}
	return left.Object < right.Object
}
func overlapNodes(left, right []Node) bool {
	seen := make(map[Node]struct{}, len(left))
	for _, node := range left {
		seen[node] = struct{}{}
	}
	for _, node := range right {
		if _, exists := seen[node]; exists {
			return true
		}
	}
	return false
}
func overlapFacts(left, right []Fact) bool {
	seen := make(map[factIdentity]struct{}, len(left))
	for _, fact := range left {
		seen[factIdentityOf(fact)] = struct{}{}
	}
	for _, fact := range right {
		if _, exists := seen[factIdentityOf(fact)]; exists {
			return true
		}
	}
	return false
}
