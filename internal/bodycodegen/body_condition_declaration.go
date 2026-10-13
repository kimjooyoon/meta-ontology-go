package bodycodegen

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// ConditionDeclarationSource identifies an authored declaration by its stable
// semantic ID and byte interval. Display names do not own semantic identity.
type ConditionDeclarationSource struct {
	Status       string `json:"status"`
	Reason       string `json:"reason,omitempty"`
	View         string `json:"view"`
	Kind         string `json:"kind"`
	StableID     string `json:"stable_id"`
	Name         string `json:"name"`
	SourceSHA256 string `json:"source_sha256"`
	Start        int    `json:"start_byte"`
	End          int    `json:"end_byte"`
}

func conditionDeclaration(sourceSHA, kind, id, name string, span syntax.Span, size int) *ConditionDeclarationSource {
	decl := &ConditionDeclarationSource{Status: "UNMAPPED", Reason: "DECLARATION_RANGE_UNAVAILABLE",
		View: "selected_gooo_source", Kind: kind, StableID: id, Name: name, SourceSHA256: sourceSHA}
	if span.Start.Offset >= 0 && span.End.Offset > span.Start.Offset && span.End.Offset <= size {
		decl.Status, decl.Reason = "MAPPED", ""
		decl.Start, decl.End = span.Start.Offset, span.End.Offset
	}
	return decl
}

func conditionFieldDeclarations(source []byte, file *syntax.File, model bidir.Model) map[string]*ConditionDeclarationSource {
	entities, fields := map[string]string{}, map[string]string{}
	for _, node := range model.Nodes {
		if node.Kind != bidir.EntityKind {
			continue
		}
		entities[node.Name] = string(node.ID)
		for _, field := range node.Fields {
			fields[node.Name+"\x00"+field.Name] = string(field.ID)
		}
	}
	result := map[string]*ConditionDeclarationSource{}
	sourceSHA := digest(source)
	for _, declaration := range file.Declarations {
		entity, ok := declaration.(*syntax.EntityDecl)
		if !ok {
			continue
		}
		for _, field := range entity.Fields {
			id := fields[entity.Name+"\x00"+field.Name]
			result[entities[entity.Name]+"\x00"+id] = conditionDeclaration(sourceSHA, "field", id, field.Name, field.Span, len(source))
		}
	}
	return result
}
