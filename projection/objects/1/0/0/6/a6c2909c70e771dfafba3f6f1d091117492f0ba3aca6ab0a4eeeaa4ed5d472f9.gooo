package query

import (
	"sort"
	"strconv"
	"strings"
)

func datalogRowCanonical(row DatalogRow) string {
	var builder strings.Builder
	builder.WriteString(datalogBindingCanonical(row.Bindings))
	for _, fact := range row.Facts {
		builder.WriteString(datalogFactCanonical(fact))
	}
	return builder.String()
}
func datalogFactCanonical(fact DatalogFact) string {
	var builder strings.Builder
	builder.WriteString(strconv.Quote(fact.Key().String()))
	builder.WriteByte('\t')
	builder.WriteString(fact.Origin.String())
	builder.WriteByte('\t')
	builder.WriteString(strconv.Itoa(fact.Depth))
	builder.WriteByte('\t')
	builder.WriteString(strconv.Quote(fact.RuleID))
	for _, support := range fact.Support {
		builder.WriteByte('\t')
		builder.WriteString(strconv.Quote(support.String()))
	}
	return builder.String()
}
func sortDatalogFacts(facts []DatalogFact) {
	sort.Slice(facts, func(i, j int) bool {
		left, right := facts[i], facts[j]
		if left.Namespace != right.Namespace {
			return left.Namespace < right.Namespace
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
		if left.Origin != right.Origin {
			return left.Origin < right.Origin
		}
		if left.RuleID != right.RuleID {
			return left.RuleID < right.RuleID
		}
		if left.Depth != right.Depth {
			return left.Depth < right.Depth
		}
		return datalogSupportCanonical(left.Support) < datalogSupportCanonical(right.Support)
	})
}
func (key DatalogFactKey) String() string {
	return key.Namespace + "\x00" + key.Subject.String() + "\x00" + key.Predicate + "\x00" + key.Object.String()
}
func datalogSupportCanonical(support []DatalogFactKey) string {
	var builder strings.Builder
	for _, key := range support {
		builder.WriteString(strconv.Quote(key.String()))
		builder.WriteByte(';')
	}
	return builder.String()
}
