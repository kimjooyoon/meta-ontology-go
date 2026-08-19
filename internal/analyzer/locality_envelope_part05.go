package analyzer

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

func validateLocalityFacts(facts []semantic.FactKey) error {
	seen := make(map[semantic.FactKey]struct{}, len(facts))
	for _, key := range facts {
		if _, err := semantic.ParseIdentity(key.Subject.String()); err != nil {
			return err
		}
		if _, err := semantic.ParseIdentity(key.Object.String()); err != nil {
			return err
		}
		if !key.Predicate.Valid() {
			return fmt.Errorf("invalid preserved fact predicate: %s", key.Predicate)
		}
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate preserved fact: %v", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}
func sortedLocalityIDs(ids []semantic.ID) []semantic.ID {
	copyOf := append([]semantic.ID(nil), ids...)
	sort.Slice(copyOf, func(i, j int) bool { return copyOf[i] < copyOf[j] })
	return copyOf
}
func sortedLocalityIDsFromSet(ids map[semantic.ID]struct{}) []semantic.ID {
	output := make([]semantic.ID, 0, len(ids))
	for id := range ids {
		output = append(output, id)
	}
	return sortedLocalityIDs(output)
}
func sortedLocalityFacts(facts []semantic.FactKey) []semantic.FactKey {
	copyOf := append([]semantic.FactKey(nil), facts...)
	sort.Slice(copyOf, func(i, j int) bool {
		if copyOf[i].Subject != copyOf[j].Subject {
			return copyOf[i].Subject < copyOf[j].Subject
		}
		if copyOf[i].Predicate != copyOf[j].Predicate {
			return copyOf[i].Predicate < copyOf[j].Predicate
		}
		return copyOf[i].Object < copyOf[j].Object
	})
	return copyOf
}
func equalLocalityIDs(left, right []semantic.ID) bool {
	left, right = sortedLocalityIDs(left), sortedLocalityIDs(right)
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
