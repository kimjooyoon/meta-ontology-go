package semanticdelta

import (
	"sort"
)

// Normalize sorts a report and returns a detached copy.
func (r *Report) Normalize() {
	if r == nil {
		return
	}
	sort.SliceStable(r.Violations, func(i, j int) bool {
		left, right := r.Violations[i], r.Violations[j]
		if left.Operation != right.Operation {
			return left.Operation < right.Operation
		}
		if left.Change != right.Change {
			return changeRank(left.Change) < changeRank(right.Change)
		}
		if left.ID != right.ID {
			return left.ID < right.ID
		}
		if left.Subject != right.Subject {
			return left.Subject < right.Subject
		}
		if left.Predicate != right.Predicate {
			return left.Predicate < right.Predicate
		}
		if left.Object != right.Object {
			return left.Object < right.Object
		}
		if left.Endpoint != right.Endpoint {
			return left.Endpoint < right.Endpoint
		}
		return left.Reason < right.Reason
	})
}
func changeRank(kind ChangeKind) int {
	if kind == ChangeNode {
		return 0
	}
	return 1
}

// Allowed reports whether the detector found no scope violation.
func (r Report) Passes() bool { return r.Allowed && len(r.Violations) == 0 }
