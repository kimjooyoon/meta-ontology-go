package main

import (
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type graphDump struct {
	SchemaVersion string              `json:"schema_version"`
	GraphHash     string              `json:"graph_hash"`
	SourceDigest  string              `json:"source_digest"`
	IR            graphIRStatus       `json:"ir"`
	Evidence      graphReferenceState `json:"evidence"`
	Provenance    graphReferenceState `json:"provenance"`
	Projection    graphStatus         `json:"projection"`
	Lowering      graphStatus         `json:"lowering"`
	Output        graphStatus         `json:"output"`
	Authorities   graphAuthorities    `json:"authorities"`
	Nodes         []graphNode         `json:"nodes"`
	Relations     []graphRelation     `json:"relations"`
}

type graphIRStatus struct {
	Status         string `json:"status"`
	SemanticDigest string `json:"semantic_digest,omitempty"`
	Reason         string `json:"reason,omitempty"`
}

type graphReferenceState struct {
	Status string   `json:"status"`
	Refs   []string `json:"refs,omitempty"`
	Reason string   `json:"reason,omitempty"`
}

type graphStatus struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

type graphAuthorities struct {
	GoooSource  string `json:".gooo"`
	SemanticIR  string `json:"ir"`
	Handwritten string `json:"handwritten_go"`
	Provenance  string `json:"provenance"`
	Graph       string `json:"graph"`
}

type graphNode struct {
	ID        string       `json:"id"`
	Kind      string       `json:"kind"`
	Namespace string       `json:"namespace"`
	Name      string       `json:"name"`
	Aliases   []string     `json:"aliases,omitempty"`
	Fields    []graphField `json:"fields,omitempty"`
}

type graphField struct {
	ID          string    `json:"id"`
	Parent      string    `json:"parent"`
	Name        string    `json:"name"`
	Aliases     []string  `json:"aliases,omitempty"`
	TypeRefID   string    `json:"type_ref_id"`
	Presence    string    `json:"presence"`
	Cardinality string    `json:"cardinality"`
	Source      graphSpan `json:"source"`
}

type graphSpan struct {
	File  string        `json:"file"`
	Start graphPosition `json:"start"`
	End   graphPosition `json:"end"`
}

type graphPosition struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

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

func graphNodes(nodes []semantic.Node) []graphNode {
	nodes = canonicalNodes(nodes)
	result := make([]graphNode, 0, len(nodes))
	for _, node := range nodes {
		aliases := append([]string(nil), node.Aliases...)
		sort.Strings(aliases)
		fields := make([]graphField, 0, len(node.Fields))
		for _, field := range node.Fields {
			fieldAliases := append([]string(nil), field.Aliases...)
			sort.Strings(fieldAliases)
			fields = append(fields, graphField{
				ID: string(field.ID), Parent: string(field.Parent), Name: field.Name, Aliases: fieldAliases,
				TypeRefID: string(field.TypeRef.ID), Presence: string(field.Presence), Cardinality: string(field.Cardinality),
				Source: graphSpan{File: field.Span.File, Start: graphPosition{Offset: field.Span.Start.Offset, Line: field.Span.Start.Line, Column: field.Span.Start.Column}, End: graphPosition{Offset: field.Span.End.Offset, Line: field.Span.End.Line, Column: field.Span.End.Column}},
			})
		}
		result = append(result, graphNode{
			ID: string(node.ID), Kind: node.Kind.String(), Namespace: node.Namespace.String(),
			Name: node.Name, Aliases: aliases, Fields: fields,
		})
	}
	return result
}

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
