package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"sort"
)

func validateCLIProjectedFields(model generator.SemanticIR, support syntax.EntityFieldsSupport) error {
	types, err := generator.EntityFieldTypeNames(model, support.Profile)
	if err != nil {
		return err
	}
	for _, entity := range model.Entities {
		if len(entity.Fields) == 0 {
			continue
		}
		capabilities, err := support.Profile.Capabilities()
		if err != nil {
			return err
		}
		seenIDs := make(map[string]struct{}, len(entity.Fields))
		seenNames := make(map[string]struct{}, len(entity.Fields))
		var sourceURI string
		var previousStart int
		for index, field := range entity.Fields {
			if field.ID == "" || field.Parent == "" || field.Name == "" || field.TypeRefID == "" || field.Presence == "" || field.Cardinality == "" {
				return fmt.Errorf("GOOO-EF-V1-INCOMPLETE-FIELD: entity %q field %d", entity.ID, index)
			}
			if field.Parent != entity.ID {
				return fmt.Errorf("GOOO-EF-V1-WRONG-PARENT: field %q parent %q does not match %q", field.ID, field.Parent, entity.ID)
			}
			if _, exists := seenIDs[field.ID]; exists {
				return fmt.Errorf("GOOO-EF-V1-ID-COLLISION: field %q is duplicated", field.ID)
			}
			seenIDs[field.ID] = struct{}{}
			if _, exists := seenNames[field.Name]; exists {
				return fmt.Errorf("GOOO-EF-V1-GO-NAME-COLLISION: field name %q is duplicated", field.Name)
			}
			seenNames[field.Name] = struct{}{}
			stringType := field.TypeRefID == "urn:gooo:type:string"
			if _, allowed := types[field.TypeRefID]; !allowed {
				return fmt.Errorf("GOOO-EF-V1-UNSUPPORTED-TYPE: field %q type %q", field.ID, field.TypeRefID)
			}
			if field.TextBound.Validate() != nil || field.TextBound.IsSet() && (!capabilities.TextBound || !stringType) {
				return fmt.Errorf("GOOO-EF-V1-UNSUPPORTED-TYPE: field %q max_utf8_bytes requires Text under V6", field.ID)
			}
			optionalSupported := capabilities.Optional && field.Presence == "optional"
			if field.Presence != "required" && !optionalSupported || field.Cardinality != "one" && !(capabilities.Many && field.Cardinality == "many") {
				return fmt.Errorf("GOOO-EF-V1-UNSUPPORTED-SHAPE: field %q", field.ID)
			}
			if err := validateCLIFieldSpans(entity, field, sourceURI, previousStart, index > 0); err != nil {
				return err
			}
			if sourceURI == "" {
				sourceURI = field.Source.URI
			}
			previousStart = field.Source.Start.Offset
		}
	}
	return nil
}
func validateCLIFieldSpans(entity generator.Entity, field generator.Field, sourceURI string, previousStart int, hasPrevious bool) error {
	if field.Source.URI == "" || field.Source.End.Offset <= field.Source.Start.Offset || field.Source.Start.Line <= 0 || field.Source.End.Line <= 0 {
		return fmt.Errorf("GOOO-EF-V1-INCOMPLETE-FIELD: field %q source span is missing", field.ID)
	}
	if sourceURI != "" && field.Source.URI != sourceURI {
		return fmt.Errorf("GOOO-EF-V1-UNREPRESENTABLE: field %q crosses source snapshots", field.ID)
	}
	if entity.Source.URI != "" && field.Source.URI != entity.Source.URI {
		return fmt.Errorf("GOOO-EF-V1-UNREPRESENTABLE: field %q does not match entity source snapshot", field.ID)
	}
	if hasPrevious && field.Source.Start.Offset <= previousStart {
		return fmt.Errorf("GOOO-EF-V1-ILLEGAL-REORDER: field %q is not source ordered", field.ID)
	}
	spans := []generator.SourceSpan{field.IDSpan, field.NameSpan, field.TypeRefSpan, field.PresenceSpan, field.CardinalitySpan}
	for _, span := range spans {
		if span.URI != field.Source.URI || span.End.Offset <= span.Start.Offset || span.Start.Offset < field.Source.Start.Offset || span.End.Offset > field.Source.End.Offset || span.Start.Line <= 0 || span.End.Line <= 0 {
			return fmt.Errorf("GOOO-EF-V1-UNREPRESENTABLE: field %q has an invalid source subspan", field.ID)
		}
	}
	sort.SliceStable(spans, func(left, right int) bool {
		return spans[left].Start.Offset < spans[right].Start.Offset
	})
	for index := 1; index < len(spans); index++ {
		if spans[index-1].End.Offset > spans[index].Start.Offset {
			return fmt.Errorf("GOOO-EF-V1-ILLEGAL-REORDER: field %q source subspans overlap", field.ID)
		}
	}
	return nil
}
