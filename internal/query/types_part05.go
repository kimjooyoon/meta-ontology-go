package query

// TraversalResult keeps paths containing candidate facts separate from paths
// made entirely of deterministic facts.
type TraversalResult struct {
	Deterministic []Path
	Candidates    []Path
	Metadata      ProjectionMetadata
}

func (result TraversalResult) All() []Path {
	paths := append([]Path(nil), result.Deterministic...)
	paths = append(paths, result.Candidates...)
	sortPaths(paths)
	return paths
}
