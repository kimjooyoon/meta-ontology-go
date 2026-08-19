package query

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func matchDatalogPatterns(patterns []DatalogAtom, facts []DatalogFact, budget *datalogWorkBudget) ([]DatalogRow, error) {
	byPredicate := make(map[string][]DatalogFact)
	for _, fact := range facts {
		byPredicate[fact.Predicate] = append(byPredicate[fact.Predicate], fact)
	}
	for predicate := range byPredicate {
		sortDatalogFacts(byPredicate[predicate])
	}
	rows := make([]DatalogRow, 0)
	seen := make(map[string]struct{})
	var join func(int, datalogBinding, []DatalogFact)
	join = func(index int, binding datalogBinding, matched []DatalogFact) {
		if index == len(patterns) {
			row := DatalogRow{Bindings: copyDatalogBinding(binding), Facts: append([]DatalogFact(nil), matched...)}
			key := datalogBindingCanonical(binding)
			if _, exists := seen[key]; !exists {
				seen[key] = struct{}{}
				rows = append(rows, row)
			}
			return
		}
		pattern := patterns[index]
		for _, fact := range byPredicate[pattern.Predicate] {
			if !budget.take() {
				return
			}
			next, ok := bindDatalogAtom(pattern, fact, binding)
			if ok {
				join(index+1, next, append(matched, fact))
			}
		}
	}
	join(0, make(datalogBinding), nil)
	if budget.exhausted {
		return nil, fmt.Errorf("%w: maximum work %d", ErrDatalogBudget, budget.limit)
	}
	sort.Slice(rows, func(i, j int) bool { return datalogRowCanonical(rows[i]) < datalogRowCanonical(rows[j]) })
	return rows, nil
}
func copyDatalogBinding(binding datalogBinding) map[string]ID {
	copy := make(map[string]ID, len(binding))
	for name, value := range binding {
		copy[name] = value
	}
	return copy
}
func datalogBindingCanonical(binding datalogBinding) string {
	names := make([]string, 0, len(binding))
	for name := range binding {
		names = append(names, name)
	}
	sort.Strings(names)
	var builder strings.Builder
	for _, name := range names {
		builder.WriteString(strconv.Quote(name))
		builder.WriteByte('=')
		builder.WriteString(strconv.Quote(binding[name].String()))
		builder.WriteByte(';')
	}
	return builder.String()
}
