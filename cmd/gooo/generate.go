package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func runGenerate(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	args, jsonMode := parseJSONFlag(args)
	options, err := parseGenerateArguments(args)
	if err != nil {
		if jsonMode {
			return reportUsage(true, stdout, stderr, "generate", err.Error())
		}
		fmt.Fprintln(stderr, err)
		return exitUsage
	}
	deadline := time.Now().Add(commandDeadline)
	input, code := readGenerateInput(options, reader, parser, jsonMode, stdout, stderr, deadline)
	if code != exitOK {
		return code
	}
	artifacts, code := buildGenerateArtifacts(options, input, jsonMode, stdout, stderr, deadline)
	if code != exitOK {
		return code
	}
	if code := writeGenerateArtifacts(artifacts, jsonMode, stdout, stderr); code != exitOK {
		return code
	}
	return reportGenerateSuccess(options, input, artifacts, jsonMode, stdout)
}

func projectionIR(ir semantic.IR) (generator.SemanticIR, error) {
	if err := ir.Validate(); err != nil {
		return generator.SemanticIR{}, err
	}
	model := generator.SemanticIR{Package: ir.Package}
	entities := make(map[string]int)
	activities := make(map[string]int)
	for _, node := range ir.Graph.Nodes() {
		item := generatorNode(node)
		switch node.Kind {
		case semantic.Entity:
			fields, err := projectionSemanticFields(node)
			if err != nil {
				return generator.SemanticIR{}, err
			}
			item.entity.Fields = fields
			entities[string(node.ID)] = len(model.Entities)
			model.Entities = append(model.Entities, item.entity)
		case semantic.Activity:
			activities[string(node.ID)] = len(model.Activities)
			model.Activities = append(model.Activities, item.activity)
		}
	}
	for _, fact := range ir.Graph.DeterministicFacts() {
		appendFactPort(&model, entities, activities, fact)
	}
	return model, nil
}

type generatorNodeResult struct {
	entity   generator.Entity
	activity generator.Activity
}

func generatorNode(node semantic.Node) generatorNodeResult {
	span := generatorSpan(node.Span)
	entity := generator.Entity{ID: string(node.ID), Name: node.Name, GoName: node.Name, Source: span}
	activity := generator.Activity{ID: string(node.ID), Name: node.Name, GoName: node.Name, Source: span}
	return generatorNodeResult{entity: entity, activity: activity}
}

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
