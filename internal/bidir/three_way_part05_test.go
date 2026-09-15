package bidir

func countRelation(model Model, want Relation) int {
	count := 0
	for _, relation := range model.Relations {
		if relationKey(relation.Kind, relation.Source, relation.Target) == relationKey(want.Kind, want.Source, want.Target) {
			count++
		}
	}
	return count
}
func reverseModelCollections(model *Model) {
	for left, right := 0, len(model.Nodes)-1; left < right; left, right = left+1, right-1 {
		model.Nodes[left], model.Nodes[right] = model.Nodes[right], model.Nodes[left]
	}
	for left, right := 0, len(model.Relations)-1; left < right; left, right = left+1, right-1 {
		model.Relations[left], model.Relations[right] = model.Relations[right], model.Relations[left]
	}
}
