package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
	"strings"
)

func graphRelations(facts []semantic.Fact) []graphRelation {
	result := make([]graphRelation, 0, len(facts))
	for _, fact := range facts {
		result = append(result, graphRelation{
			Status: fact.Status.String(), Subject: string(fact.Subject),
			Predicate: fact.Predicate.String(), Object: string(fact.Object),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		left, right := result[i], result[j]
		if left.Status != right.Status {
			return left.Status < right.Status
		}
		if left.Subject != right.Subject {
			return left.Subject < right.Subject
		}
		if left.Predicate != right.Predicate {
			return left.Predicate < right.Predicate
		}
		return left.Object < right.Object
	})
	return result
}
func authoritativeGraphHash(graph semantic.Graph) string {
	return semantic.StableHash([]byte(authoritativeGraphCanonical(graph)))
}
func authoritativeIRHash(ir semantic.IR) string {
	var canonical strings.Builder
	version := strings.TrimSpace(ir.Version)
	if version == "" {
		version = semantic.CurrentIRVersion
	}
	namespace := strings.TrimSpace(ir.Namespace.String())
	if parsed, err := semantic.ParseNamespace(namespace); err == nil {
		namespace = parsed.String()
	}
	canonical.WriteString("ir\t")
	canonical.WriteString(version)
	canonical.WriteByte('\t')
	canonical.WriteString(strings.TrimSpace(ir.Package))
	canonical.WriteByte('\t')
	canonical.WriteString(namespace)
	canonical.WriteByte('\n')
	canonical.WriteString(authoritativeGraphCanonical(ir.Graph))
	bindings := append([]semantic.RuntimeBinding(nil), ir.RuntimeBindings...)
	sort.Slice(bindings, func(i, j int) bool {
		left, right := bindings[i].Key(), bindings[j].Key()
		if left.ProducerActivity != right.ProducerActivity {
			return left.ProducerActivity < right.ProducerActivity
		}
		if left.ProducerPort != right.ProducerPort {
			return left.ProducerPort < right.ProducerPort
		}
		if left.ConsumerActivity != right.ConsumerActivity {
			return left.ConsumerActivity < right.ConsumerActivity
		}
		if left.ConsumerPort != right.ConsumerPort {
			return left.ConsumerPort < right.ConsumerPort
		}
		return bindings[i].Entity < bindings[j].Entity
	})
	for _, binding := range bindings {
		canonical.WriteString(binding.SemanticCanonical())
		canonical.WriteByte('\n')
	}
	return semantic.StableHash([]byte(canonical.String()))
}
func authoritativeGraphCanonical(graph semantic.Graph) string {
	var canonical strings.Builder
	for _, node := range canonicalNodes(graph.Nodes()) {
		canonical.WriteString(node.SemanticCanonical())
		canonical.WriteByte('\n')
	}
	for _, fact := range canonicalFacts(graph.DeterministicFacts()) {
		canonical.WriteString(fact.SemanticCanonical())
		canonical.WriteByte('\n')
	}
	return canonical.String()
}
