package query

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
)

func matchDatalogPatterns(patterns []DatalogAtom, facts []DatalogFact) []DatalogRow {
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
			next, ok := bindDatalogAtom(pattern, fact, binding)
			if ok {
				join(index+1, next, append(matched, fact))
			}
		}
	}
	join(0, make(datalogBinding), nil)
	sort.Slice(rows, func(i, j int) bool { return datalogRowCanonical(rows[i]) < datalogRowCanonical(rows[j]) })
	return rows
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

func datalogRowCanonical(row DatalogRow) string {
	var builder strings.Builder
	builder.WriteString(datalogBindingCanonical(row.Bindings))
	for _, fact := range row.Facts {
		builder.WriteString(strconv.Quote(fact.Key().String()))
	}
	return builder.String()
}

func sortDatalogFacts(facts []DatalogFact) {
	sort.Slice(facts, func(i, j int) bool {
		left, right := facts[i], facts[j]
		if left.Subject != right.Subject {
			return left.Subject < right.Subject
		}
		if left.Predicate != right.Predicate {
			return left.Predicate < right.Predicate
		}
		if left.Object != right.Object {
			return left.Object < right.Object
		}
		if left.Origin != right.Origin {
			return left.Origin < right.Origin
		}
		return left.RuleID < right.RuleID
	})
}

func (key DatalogFactKey) String() string {
	return key.Subject.String() + "\x00" + key.Predicate + "\x00" + key.Object.String()
}

// Canonical provides a stable receipt for replay and permutation tests.
func (result DatalogResult) Canonical() string {
	var builder strings.Builder
	for _, row := range result.Rows {
		builder.WriteString(datalogRowCanonical(row))
		builder.WriteByte('\n')
	}
	for _, fact := range result.Derived {
		builder.WriteString(fact.Key().String())
		builder.WriteByte('\t')
		builder.WriteString(fact.RuleID)
		builder.WriteByte('\n')
	}
	if result.Complete {
		builder.WriteString("complete\n")
	} else {
		builder.WriteString("incomplete\n")
	}
	return builder.String()
}

func (result DatalogResult) StableHash() string {
	digest := sha256.Sum256([]byte(result.Canonical()))
	return hex.EncodeToString(digest[:])
}
