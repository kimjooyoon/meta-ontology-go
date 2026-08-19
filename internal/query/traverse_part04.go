package query

import (
	"sort"
)

func follows(at ID, fact Fact, direction Direction) bool {
	return (direction == Outgoing || direction == Both) && fact.Subject == at ||
		(direction == Incoming || direction == Both) && fact.Object == at
}
func nextNode(at ID, fact Fact, direction Direction) ID {
	if (direction == Outgoing || direction == Both) && fact.Subject == at {
		return fact.Object
	}
	return fact.Subject
}
func containsID(ids []ID, target ID) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}
func extendPath(path Path, fact Fact, next ID, status FactStatus) Path {
	ids := append(append([]ID(nil), path.IDs...), next)
	facts := append(append([]Fact(nil), path.Facts...), fact)
	return Path{IDs: ids, Facts: facts, Status: status}
}
func sortPaths(paths []Path) {
	sort.Slice(paths, func(i, j int) bool {
		left, right := paths[i], paths[j]
		if len(left.Facts) != len(right.Facts) {
			return len(left.Facts) < len(right.Facts)
		}
		for index := range left.IDs {
			if left.IDs[index] != right.IDs[index] {
				return left.IDs[index] < right.IDs[index]
			}
		}
		if left.Status != right.Status {
			return left.Status < right.Status
		}
		for index := range left.Facts {
			if left.Facts[index].Predicate != right.Facts[index].Predicate {
				return left.Facts[index].Predicate < right.Facts[index].Predicate
			}
		}
		return false
	})
}
