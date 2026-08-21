package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"io"
	"time"
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
