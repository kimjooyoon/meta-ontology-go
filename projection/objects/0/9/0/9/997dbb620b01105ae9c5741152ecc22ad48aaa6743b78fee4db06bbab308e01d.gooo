package query

func (graph Graph) traversePaths(
	start ID,
	options TraversalOptions,
	selection FactSelection,
	outputStatus FactStatus,
	resultLimit int,
	quota *queryWorkQuota,
) []Path {
	facts := graph.AllFacts()
	frontier := []Path{{IDs: []ID{start}, Status: FactDeterministic}}
	paths := make([]Path, 0)
	for depth := 1; depth <= options.MaxDepth && len(frontier) > 0; depth++ {
		next := make([]Path, 0)
		complete := true
		for _, path := range frontier {
			edges, edgesComplete := graph.edges(path.Last(), facts, options, selection, quota)
			if !edgesComplete {
				complete = false
			}
			for _, fact := range edges {
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
			if !complete {
				break
			}
		}
		sortPaths(next)
		for _, path := range next {
			if path.Status == outputStatus {
				paths = append(paths, path)
				if resultLimit > 0 && len(paths) == resultLimit {
					return paths
				}
			}
		}
		if !complete {
			return paths
		}
		frontier = next
	}
	return paths
}
func (graph Graph) edges(
	at ID, facts []Fact, options TraversalOptions, selection FactSelection,
	quota *queryWorkQuota,
) ([]Fact, bool) {
	matches := make([]Fact, 0)
	for _, fact := range facts {
		if !selection.includes(fact.Status) {
			continue
		}
		if !quota.take() {
			return matches, false
		}
		if options.Predicate != "" && fact.Predicate != options.Predicate {
			continue
		}
		if follows(at, fact, options.Direction) {
			matches = append(matches, fact)
		}
	}
	return matches, true
}
