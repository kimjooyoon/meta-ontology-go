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
		Nodes:     graphNodes(ir.Graph.Nodes()),
		Relations: graphRelations(ir.Graph.AllFacts()),
	}
}
func graphReferences(refs []string, missingReason string) graphReferenceState {
	if len(refs) == 0 {
		return graphReferenceState{Status: "missing", Reason: missingReason}
	}
	ordered := append([]string(nil), refs...)
	sort.Strings(ordered)
	return graphReferenceState{Status: "available", Refs: ordered}
}
