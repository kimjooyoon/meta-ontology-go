package entityfieldsv1

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/lsp"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type NavigationSymbol struct {
	Name           string    `json:"name"`
	ID             string    `json:"id"`
	Kind           int       `json:"kind"`
	Range          lsp.Range `json:"range"`
	SelectionRange lsp.Range `json:"selection_range"`
	IdentityRange  lsp.Range `json:"identity_range"`
	HasIdentity    bool      `json:"has_identity"`
}

type NavigationReference struct {
	Name  string    `json:"name"`
	ID    string    `json:"id"`
	Range lsp.Range `json:"range"`
}

type CounterexampleEvidence struct {
	ID            string `json:"id"`
	Decision      string `json:"decision"`
	Resolution    string `json:"resolution"`
	Reason        string `json:"reason"`
	InputDigest   string `json:"input_digest"`
	OutputDigest  string `json:"output_digest"`
	EvidenceDigest string `json:"evidence_digest"`
	PartialOutput bool   `json:"partial_output"`
	Unknown       *UnknownEvidence `json:"unknown,omitempty"`
}

type UnknownEvidence struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type Observation struct {
	Schema           string                         `json:"schema"`
	Profile          syntax.EntityFieldsProfile     `json:"profile"`
	Source           string                         `json:"source"`
	Formatted        string                         `json:"formatted"`
	DeclarationOrder []string                       `json:"declaration_order"`
	StableIDs        []string                       `json:"stable_ids"`
	Semantic         generator.SemanticIR           `json:"semantic_ir"`
	Generated        []byte                         `json:"generated_go"`
	SourceMap        generator.SourceMap            `json:"source_map"`
	Symbols          []NavigationSymbol             `json:"symbols"`
	References       []NavigationReference          `json:"references"`
	SourceDigest     string                         `json:"source_digest"`
	FormattedDigest  string                         `json:"formatted_digest"`
	SemanticDigest   string                         `json:"semantic_digest"`
	GeneratedDigest  string                         `json:"generated_digest"`
	SourceMapDigest  string                         `json:"source_map_digest"`
	NavigationDigest string                         `json:"navigation_digest"`
	EvidenceDigests  map[string]string              `json:"evidence_digests"`
	GetPutRoundTrip       bool                         `json:"get_put_roundtrip"`
	PutGetRoundTrip       bool                         `json:"put_get_roundtrip"`
	GetPutOriginalDigest  string                       `json:"get_put_original_digest"`
	GetPutWrittenDigest   string                       `json:"get_put_written_digest"`
	PutGetInputDigest     string                       `json:"put_get_input_digest"`
	PutGetObservedDigest  string                       `json:"put_get_observed_digest"`
	Counterexamples       []CounterexampleEvidence     `json:"counterexamples"`
}

func Observe(filename, source string) (Observation, error) {
	support := support()
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport(filename, source, support)
	if len(diagnostics) != 0 || file == nil {
		return Observation{}, fmt.Errorf("REFUTED/EXACT: entity-fields parse: %s", diagnostics.Error().Error())
	}
	formatted, err := syntax.FormatWithEntityFieldsSupport(file, support)
	if err != nil {
		return Observation{}, fmt.Errorf("REFUTED/EXACT: entity-fields format: %w", err)
	}
	replayed, replayDiagnostics := syntax.ParseFileWithEntityFieldsSupport(filename, formatted, support)
	if len(replayDiagnostics) != 0 || replayed == nil {
		return Observation{}, fmt.Errorf("REFUTED/EXACT: formatted replay: %s", replayDiagnostics.Error().Error())
	}
	document, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(file, support)
	if err != nil {
		return Observation{}, fmt.Errorf("REFUTED/EXACT: document lowering: %w", err)
	}
	model, err := bidir.GetWithEntityFieldsSupport(document, support)
	if err != nil {
		return Observation{}, fmt.Errorf("REFUTED/EXACT: BX Get: %w", err)
	}
	if err := validateStableIdentityDomain(model); err != nil {
		return Observation{}, fmt.Errorf("REFUTED/EXACT: stable identity domain: %w", err)
	}
	getPutDocument, err := bidir.PutWithEntityFieldsSupport(document, model, support)
	if err != nil {
		return Observation{}, fmt.Errorf("REFUTED/EXACT: BX Get/Put: %w", err)
	}
	if !bidir.DocumentEquivalent(document, getPutDocument) {
		return Observation{}, fmt.Errorf("REFUTED/EXACT: BX Get/Put changed source document")
	}
	putGetInput := model.Clone()
	if len(putGetInput.Nodes) == 0 || len(putGetInput.Nodes[0].Fields) == 0 {
		return Observation{}, fmt.Errorf("REFUTED/EXACT: BX Put/Get mutation has no field")
	}
	putGetInput.Nodes[0].Fields[0].Name += "Changed"
	putGetDocument, err := bidir.PutWithEntityFieldsSupport(document, putGetInput, support)
	if err != nil {
		return Observation{}, fmt.Errorf("REFUTED/EXACT: BX Put/Get: %w", err)
	}
	putGetObserved, err := bidir.GetWithEntityFieldsSupport(putGetDocument, support)
	if err != nil || !bidir.SemanticEquivalent(putGetInput, putGetObserved) {
		return Observation{}, fmt.Errorf("REFUTED/EXACT: BX Put/Get semantic mutation was not observed")
	}
	ir, err := bidir.LowerContextWithEntityFieldsSupport(nil, replayed, support)
	if err != nil {
		return Observation{}, fmt.Errorf("REFUTED/EXACT: semantic lowering: %w", err)
	}
	semanticIR, err := projection(ir, model, replayed)
	if err != nil {
		return Observation{}, fmt.Errorf("REFUTED/EXACT: source-backed projection: %w", err)
	}
	generated, err := generator.GenerateEntityFieldsV1(semanticIR, nil)
	if err != nil {
		return Observation{}, fmt.Errorf("REFUTED/EXACT: Go projection: %w", err)
	}
	navigation := navigation(filename, formatted)
	return buildObservation(source, file, formatted, semanticIR, generated, navigation, model, document, getPutDocument, putGetInput, putGetObserved), nil
}

func projection(ir semantic.IR, model bidir.Model, source *syntax.File) (generator.SemanticIR, error) {
	result := generator.SemanticIR{Package: ir.Package}
	byID := make(map[string]bidir.Node, len(model.Nodes))
	for _, node := range model.Nodes {
		byID[string(node.ID)] = node
	}
	for _, declaration := range declarations(source) {
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			id := declarationID(model, value.Name, bidir.EntityKind, value.ID)
			node, ok := ir.Graph.Node(semantic.ID(id))
			if !ok || node.Kind != semantic.Entity {
				return generator.SemanticIR{}, fmt.Errorf("missing semantic entity %q", id)
			}
			sourceNode, ok := byID[id]
			if !ok {
				return generator.SemanticIR{}, fmt.Errorf("missing BX source node %q", id)
			}
			result.Entities = append(result.Entities, generatorEntity(sourceNode))
		case *syntax.ActivityDecl:
			id := declarationID(model, value.Name, bidir.ActivityKind, value.Name)
			node, ok := ir.Graph.Node(semantic.ID(id))
			if !ok || node.Kind != semantic.Activity {
				return generator.SemanticIR{}, fmt.Errorf("missing semantic activity %q", id)
			}
			sourceNode, ok := byID[id]
			if !ok {
				return generator.SemanticIR{}, fmt.Errorf("missing BX source activity %q", id)
			}
			result.Activities = append(result.Activities, generator.Activity{ID: id, Name: node.Name, GoName: node.Name, Source: span(sourceNode.Span)})
		}
	}
	if len(result.Entities)+len(result.Activities) != len(ir.Graph.Nodes()) {
		return generator.SemanticIR{}, fmt.Errorf("source declaration order does not cover semantic graph")
	}
	return result, nil
}

func generatorEntity(node bidir.Node) generator.Entity {
	entity := generator.Entity{ID: string(node.ID), Name: node.Name, GoName: node.Name, Source: span(node.Span)}
	for _, field := range node.Fields {
		entity.Fields = append(entity.Fields, generatorField(field))
	}
	return entity
}

func generatorField(field bidir.Field) generator.Field {
	typeID := string(field.TypeRef.ID)
	if typeID == "" {
		typeID = string(field.TypeRefUse.ResolvedID)
	}
	return generator.Field{ID: string(field.ID), Parent: string(field.Parent), Name: field.Name, TypeRefID: typeID, Presence: string(field.Presence), Cardinality: string(field.Cardinality), Source: span(field.Span), IDSpan: span(field.IDSpan), NameSpan: span(field.NameSpan), TypeRefSpan: span(field.TypeRefSpan), PresenceSpan: span(field.PresenceSpan), CardinalitySpan: span(field.CardinalitySpan), NameSource: span(field.NameSpan)}
}

func span(value bidir.SourceSpan) generator.SourceSpan {
	return generator.SourceSpan{URI: value.File, Start: generator.Position{Offset: value.Start, Line: value.StartLine, Column: value.StartColumn}, End: generator.Position{Offset: value.End, Line: value.EndLine, Column: value.EndColumn}}
}

func digest(value any) string {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
