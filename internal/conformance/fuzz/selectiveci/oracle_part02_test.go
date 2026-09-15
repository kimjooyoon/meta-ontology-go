package selectiveci

import (
	"reflect"
	"sort"
	"testing"
)

func TestCanonicalDigestIgnoresInputOrder(t *testing.T) {
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	var fixture Case
	for _, candidate := range corpus.Cases {
		if candidate.Name == "order-permutations" {
			fixture = candidate
			break
		}
	}
	if fixture.Name == "" {
		t.Fatal("order-permutations fixture not found")
	}

	permuted := fixture
	permuted.Graph.Commands = append([]Command(nil), fixture.Graph.Commands...)
	permuted.Graph.Edges = append([]Edge(nil), fixture.Graph.Edges...)
	permuted.Graph.Roots = append([]string(nil), fixture.Graph.Roots...)
	permuted.Evidence.Paths = append([]PathEvidence(nil), fixture.Evidence.Paths...)
	permuted.Evidence.Changes = append([]PathChange(nil), fixture.Evidence.Changes...)
	reverseCommands(permuted.Graph.Commands)
	reverseEdges(permuted.Graph.Edges)
	sort.Sort(sort.Reverse(sort.StringSlice(permuted.Graph.Roots)))
	reversePaths(permuted.Evidence.Paths)
	reverseChanges(permuted.Evidence.Changes)

	if CanonicalDigest(permuted) != CanonicalDigest(fixture) {
		t.Fatal("canonical digest changed when input order changed")
	}
	if !reflect.DeepEqual(Evaluate(permuted), Evaluate(fixture)) {
		t.Fatal("oracle result changed when input order changed")
	}
}
func reverseCommands(items []Command) {
	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
}
func reverseEdges(items []Edge) {
	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
}
func reversePaths(items []PathEvidence) {
	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
}
func reverseChanges(items []PathChange) {
	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
}
