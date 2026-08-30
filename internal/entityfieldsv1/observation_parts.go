package entityfieldsv1

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/lsp"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func navigation(filename, source string) lsp.ParseResult {
	return (lsp.EntityFieldsSyntaxParser{}).Parse(filename, source)
}

func buildObservation(source string, file *syntax.File, formatted string, semanticIR generator.SemanticIR, generated generator.Result, parsed lsp.ParseResult, model bidir.Model) Observation {
	order := []string{}
	stableIDs := []string{}
	for _, declaration := range declarations(file) {
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			id := declarationID(model, value.Name, bidir.EntityKind, value.ID)
			order = append(order, id)
			stableIDs = append(stableIDs, id)
			stableIDs = append(stableIDs, fieldIDs(model, id)...)
		case *syntax.ActivityDecl:
			id := declarationID(model, value.Name, bidir.ActivityKind, value.Name)
			order = append(order, id)
			stableIDs = append(stableIDs, id)
		}
	}
	symbols := []NavigationSymbol{}
	for _, value := range append(append([]lsp.Symbol{}, parsed.Headers...), parsed.Symbols...) {
		symbols = append(symbols, NavigationSymbol{Name: value.Name, ID: value.ID, Kind: int(value.Kind)})
	}
	references := []NavigationReference{}
	for _, value := range parsed.References {
		references = append(references, NavigationReference{Name: value.Name, ID: value.ID})
	}
	modelDigest := digest(model)
	navigationDigest := digest(struct{ Symbols []NavigationSymbol; References []NavigationReference }{symbols, references})
	evidence := map[string]string{
		"OBSERVATION_SOURCE_PIN": boundDigest("OBSERVATION_SOURCE_PIN", digest(source)),
		"CONTRACT_PROFILE_PIN":   boundDigest("CONTRACT_PROFILE_PIN", digest(syntax.EntityFieldsV1Support().Profile)),
		"SYNTAX_FORMAT_REPLAY":   boundDigest("SYNTAX_FORMAT_REPLAY", digest(formatted)),
		"SEMANTIC_IR_LOWERING":   boundDigest("SEMANTIC_IR_LOWERING", digest(semanticIR)),
		"GLOBAL_ID_VALIDATION":   boundDigest("GLOBAL_ID_VALIDATION", modelDigest),
		"DECLARATION_ORDER":      boundDigest("DECLARATION_ORDER", digest(order)),
		"BX_GET_ROUNDTRIP":       boundDigest("BX_GET_ROUNDTRIP", modelDigest),
		"BX_PUT_ROUNDTRIP":       boundDigest("BX_PUT_ROUNDTRIP", digest(struct{ Model bidir.Model; Source string }{model, formatted})),
		"GO_STRUCT_PROJECTION":   boundDigest("GO_STRUCT_PROJECTION", digest(generated.Source)),
		"SOURCE_MAP_PROJECTION":  boundDigest("SOURCE_MAP_PROJECTION", digest(generated.SourceMap)),
		"LSP_NAVIGATION":         boundDigest("LSP_NAVIGATION", navigationDigest),
		"ENTITY_FIELDS_RECEIPT":  boundDigest("ENTITY_FIELDS_RECEIPT", digest(struct{ Source, Formatted, Generated string }{digestBytes([]byte(source)), digestBytes([]byte(formatted)), digestBytes(generated.Source)})),
	}
	return Observation{Schema: Schema, Profile: syntax.EntityFieldsV1Support().Profile, Source: source, Formatted: formatted, DeclarationOrder: order, StableIDs: stableIDs, Semantic: semanticIR, Generated: generated.Source, SourceMap: generated.SourceMap, Symbols: symbols, References: references, SourceDigest: digestBytes([]byte(source)), FormattedDigest: digestBytes([]byte(formatted)), SemanticDigest: digest(semanticIR), GeneratedDigest: digestBytes(generated.Source), SourceMapDigest: digest(generated.SourceMap), NavigationDigest: navigationDigest, EvidenceDigests: evidence, GetPutRoundTrip: true}
}

func fieldIDs(model bidir.Model, entityID string) []string {
	for _, node := range model.Nodes {
		if string(node.ID) != entityID {
			continue
		}
		ids := make([]string, 0, len(node.Fields))
		for _, field := range node.Fields {
			ids = append(ids, string(field.ID))
		}
		return ids
	}
	return nil
}

func declarations(file *syntax.File) []syntax.Declaration {
	if len(file.Decls) > 0 {
		return file.Decls
	}
	return file.Declarations
}

func declarationID(model bidir.Model, name string, kind bidir.Kind, fallback string) string {
	for _, node := range model.Nodes {
		if node.Name == name && node.Kind == kind {
			return string(node.ID)
		}
	}
	return fallback
}

func validateStableIdentityDomain(model bidir.Model) error {
	seen := make(map[string]string, len(model.Nodes))
	for _, node := range model.Nodes {
		if owner, exists := seen[string(node.ID)]; exists {
			return fmt.Errorf("identity %q is shared by %s and %s", node.ID, owner, node.Kind)
		}
		seen[string(node.ID)] = string(node.Kind)
		for _, field := range node.Fields {
			if owner, exists := seen[string(field.ID)]; exists {
				return fmt.Errorf("identity %q is shared by %s and field", field.ID, owner)
			}
			seen[string(field.ID)] = "field"
		}
	}
	return nil
}

func digestBytes(value []byte) string { return digest(string(value)) }

func boundDigest(key, value string) string { return digest(struct{ Key, Value string }{key, value}) }
