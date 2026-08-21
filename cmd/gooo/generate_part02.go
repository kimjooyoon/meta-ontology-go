package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

func appendFactPort(model *generator.SemanticIR, entities, activities map[string]int, fact semantic.Fact) {
	key := fact.Key()
	entityIndex, entityOK := entities[string(key.Object)]
	activityIndex, activityOK := activities[string(key.Subject)]
	if fact.Predicate == semantic.Used && entityOK && activityOK {
		entity := model.Entities[entityIndex]
		model.Activities[activityIndex].Inputs = append(model.Activities[activityIndex].Inputs, generatorPort(entity))
	}
	if fact.Predicate == semantic.WasGeneratedBy {
		entityIndex, entityOK = entities[string(key.Subject)]
		activityIndex, activityOK = activities[string(key.Object)]
		if entityOK && activityOK {
			entity := model.Entities[entityIndex]
			model.Activities[activityIndex].Outputs = append(model.Activities[activityIndex].Outputs, generatorPort(entity))
		}
	}
}
func generatorPort(entity generator.Entity) generator.Port {
	name := lowerCamel(entity.Name)
	return generator.Port{ID: entity.ID, Name: name, GoName: name, EntityID: entity.ID, GoType: entity.GoName, Source: entity.Source}
}
func generatorSpan(span semantic.Span) generator.SourceSpan {
	return generator.SourceSpan{
		URI:   span.File,
		Start: generator.Position{Offset: span.Start.Offset, Line: span.Start.Line, Column: span.Start.Column},
		End:   generator.Position{Offset: span.End.Offset, Line: span.End.Line, Column: span.End.Column},
	}
}
func lowerCamel(name string) string {
	if name == "" {
		return ""
	}
	return strings.ToLower(name[:1]) + name[1:]
}

var _ SourceParser = SyntaxSourceParser{}
