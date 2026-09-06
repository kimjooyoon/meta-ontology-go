package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

type graphRelation struct {
	Status    string `json:"status"`
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
}

func newGraphDump(source []byte, ir semantic.IR) graphDump {
	evidence := ir.Evidence()
	refs := make([]string, 0, len(evidence))
	for _, record := range evidence {
		refs = append(refs, record.ID.String())
	}
	sort.Strings(refs)
	return graphDump{
		SchemaVersion: graphDumpSchemaVersion,
		GraphHash:     authoritativeGraphHash(ir.Graph),
		SourceDigest:  semantic.StableHash(source),
		IR:            graphIRStatus{Status: "available", SemanticDigest: authoritativeIRHash(ir)},
		Evidence:      graphReferences(refs, "no semantic evidence records are attached"),
		Provenance: graphReferenceState{
			Status: "missing", Reason: "no provenance records are attached",
		},
		Projection: graphStatus{
			Status: "deferred", Reason: "read-only graph dump does not run projection",
		},
		Lowering: graphStatus{
			Status: "deferred", Reason: "bidir lowering has no cooperative cancellation contract",
		},
		Output: graphStatus{
			Status: "deferred", Reason: "generic writers have no cooperative cancellation contract",
		},
		Authorities: graphAuthorities{
			GoooSource: "authoritative", SemanticIR: "authoritative", Handwritten: "authoritative",
			Provenance: "authoritative", Graph: "derived",
		},
		Nodes:           graphNodes(ir.Graph.Nodes()),
		Relations:       graphRelations(ir.Graph.AllFacts()),
		RuntimeBindings: graphRuntimeBindings(ir.RuntimeBindings),
	}
}

func graphRuntimeBindings(bindings []semantic.RuntimeBinding) []graphRuntimeBinding {
	result := make([]graphRuntimeBinding, 0, len(bindings))
	for _, binding := range bindings {
		result = append(result, graphRuntimeBinding{
			Schema: binding.Schema, ProducerActivity: string(binding.ProducerActivity), ProducerPort: binding.ProducerPort,
			ConsumerActivity: string(binding.ConsumerActivity), ConsumerPort: binding.ConsumerPort, Entity: string(binding.Entity),
			Source: graphSpan{File: binding.Span.File,
				Start: graphPosition{Offset: binding.Span.Start.Offset, Line: binding.Span.Start.Line, Column: binding.Span.Start.Column},
				End:   graphPosition{Offset: binding.Span.End.Offset, Line: binding.Span.End.Line, Column: binding.Span.End.Column}},
		})
	}
	sort.Slice(result, func(i, j int) bool {
		left, right := result[i], result[j]
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
		return left.Entity < right.Entity
	})
	return result
}
func graphReferences(refs []string, missingReason string) graphReferenceState {
	if len(refs) == 0 {
		return graphReferenceState{Status: "missing", Reason: missingReason}
	}
	ordered := append([]string(nil), refs...)
	sort.Strings(ordered)
	return graphReferenceState{Status: "available", Refs: ordered}
}
