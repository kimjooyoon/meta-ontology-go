package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
	"strings"
)

func canonicalNodes(nodes []semantic.Node) []semantic.Node {
	result := append([]semantic.Node(nil), nodes...)
	for index := range result {
		result[index].Aliases = append([]string(nil), result[index].Aliases...)
		sort.Strings(result[index].Aliases)
	}
	sort.Slice(result, func(i, j int) bool {
		left, right := result[i], result[j]
		if left.ID != right.ID {
			return left.ID < right.ID
		}
		if left.Kind != right.Kind {
			return left.Kind < right.Kind
		}
		if left.Namespace != right.Namespace {
			return left.Namespace < right.Namespace
		}
		if left.Name != right.Name {
			return left.Name < right.Name
		}
		return strings.Join(left.Aliases, "\x00") < strings.Join(right.Aliases, "\x00")
	})
	return result
}
func canonicalFacts(facts []semantic.Fact) []semantic.Fact {
	result := append([]semantic.Fact(nil), facts...)
	sort.Slice(result, func(i, j int) bool {
		left, right := result[i], result[j]
		if left.Subject != right.Subject {
			return left.Subject < right.Subject
		}
		if left.Predicate != right.Predicate {
			return left.Predicate < right.Predicate
		}
		if left.Object != right.Object {
			return left.Object < right.Object
		}
		return left.Status < right.Status
	})
	return result
}
