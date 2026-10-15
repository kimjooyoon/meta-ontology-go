package sourceimpact

import "github.com/kimjooyoon/meta-ontology-go/internal/semantic"

func (s *snapshot) addEdge(dependency, dependent, kind, surface string) {
	if _, ok := s.nodes[dependency]; !ok {
		return
	}
	if _, ok := s.nodes[dependent]; !ok {
		return
	}
	s.edges = append(s.edges, Edge{Dependency: dependency, Dependent: dependent, Kind: kind, Surface: surface})
}

func addStructuralEdges(s *snapshot, ir semantic.IR) {
	for _, f := range ir.Graph.DeterministicFacts() {
		subject, object := string(f.Subject), string(f.Object)
		switch f.Predicate {
		case semantic.Used:
			s.addEdge(object, subject, "input_type", "signature")
		case semantic.WasGeneratedBy:
			// Shared output types describe a signature. Value flow belongs to an
			// explicit runtime binding; activity edits do not change the type.
			s.addEdge(subject, object, "output_type", "signature")
		default:
			s.addEdge(object, subject, "declared_relation", string(f.Predicate))
		}
	}
	for _, binding := range ir.RuntimeBindings {
		kind := "runtime_binding"
		if binding.Schema == semantic.RuntimeFeedbackSchema {
			kind = "runtime_feedback"
		}
		s.addEdge(string(binding.ProducerActivity), string(binding.ConsumerActivity), kind, binding.ProducerPort+"->"+binding.ConsumerPort+":"+string(binding.Entity))
	}
	for _, node := range ir.Graph.SortedNodes() {
		for _, field := range node.Fields {
			s.addEdge(string(field.TypeRef.ID), string(node.ID), "field_type", string(field.ID))
		}
	}
}
