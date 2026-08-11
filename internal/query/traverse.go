package query

import (
	"fmt"
	"sort"
)

// Traverse returns every simple path of depth 1 through MaxDepth from start.
// Deterministic paths use only deterministic facts. Candidate paths may mix
// deterministic and candidate edges, but contain at least one candidate edge.
func (graph Graph) Traverse(start ID, options TraversalOptions) (TraversalResult, error) {
	canonicalStart, err := ParseID(start.String())
	if err != nil {
		return TraversalResult{}, err
	}
	normalized, err := options.normalized()
	if err != nil {
		return TraversalResult{}, err
	}
	deterministic := graph.traversePaths(canonicalStart, normalized, false)
	all := graph.traversePaths(canonicalStart, normalized, true)
	candidates := make([]Path, 0, len(all))
	for _, path := range all {
		if path.Status == FactCandidate {
			candidates = append(candidates, path)
		}
	}
	return TraversalResult{Deterministic: deterministic, Candidates: candidates}, nil
}

func (options TraversalOptions) normalized() (TraversalOptions, error) {
	if options.MaxDepth <= 0 {
		return TraversalOptions{}, invalidTraversal("max depth must be positive")
	}
	if options.Direction == 0 {
		options.Direction = Outgoing
	}
	if options.Direction != Outgoing && options.Direction != Incoming && options.Direction != Both {
		return TraversalOptions{}, invalidTraversal("unknown traversal direction")
	}
	if options.Predicate != "" {
		predicate, err := ParseRelation(options.Predicate)
		if err != nil {
			return TraversalOptions{}, invalidTraversal(err.Error())
		}
		options.Predicate = predicate
	}
	return options, nil
}

func invalidTraversal(detail string) error {
	return fmt.Errorf("%w: %s", ErrInvalidTraversal, detail)
}

func (graph Graph) traversePaths(start ID, options TraversalOptions, includeCandidates bool) []Path {
	frontier := []Path{{IDs: []ID{start}, Status: FactDeterministic}}
	paths := make([]Path, 0)
	for depth := 1; depth <= options.MaxDepth && len(frontier) > 0; depth++ {
		next := make([]Path, 0)
		for _, path := range frontier {
			for _, fact := range graph.edges(path.Last(), options, includeCandidates) {
				nextID := nextNode(path.Last(), fact, options.Direction)
				if containsID(path.IDs, nextID) {
					continue
				}
				status := path.Status
				if fact.Status == FactCandidate {
					status = FactCandidate
				}
				next = append(next, extendPath(path, fact, nextID, status))
			}
		}
		sortPaths(next)
		paths = append(paths, next...)
		frontier = next
	}
	return paths
}

func (graph Graph) edges(at ID, options TraversalOptions, includeCandidates bool) []Fact {
	facts := graph.DeterministicFacts()
	if includeCandidates {
		facts = append(facts, graph.CandidateFacts()...)
	}
	matches := make([]Fact, 0)
	for _, fact := range facts {
		if options.Predicate != "" && fact.Predicate != options.Predicate {
			continue
		}
		if follows(at, fact, options.Direction) {
			matches = append(matches, fact)
		}
	}
	sortFacts(matches)
	return matches
}

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
