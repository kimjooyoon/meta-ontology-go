package query

import (
	"sort"
)

func sortFacts(facts []Fact) {
	sort.Slice(facts, func(i, j int) bool {
		left, right := facts[i], facts[j]
		if left.Subject != right.Subject {
			return left.Subject < right.Subject
		}
		if left.Predicate != right.Predicate {
			return left.Predicate < right.Predicate
		}
		if left.Object != right.Object {
			return left.Object < right.Object
		}
		if left.Status != right.Status {
			return left.Status < right.Status
		}
		return left.Reason < right.Reason
	})
}
