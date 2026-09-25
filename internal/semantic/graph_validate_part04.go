package semantic

func (g Graph) Clone() Graph {
	clone := NewGraph()
	for _, node := range g.Nodes() {
		if err := clone.AddNode(node); err != nil {
			return NewGraph()
		}
	}
	for _, fact := range g.AllFacts() {
		if err := clone.AddFact(fact); err != nil {
			return NewGraph()
		}
	}
	return clone
}
