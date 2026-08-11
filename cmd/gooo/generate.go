package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func runGenerate(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	if len(args) != 3 || args[1] != "--out" || args[2] == "" {
		fmt.Fprintln(stderr, "usage: gooo generate <file.gooo> --out <directory>")
		return exitUsage
	}
	filename, outputDir := args[0], args[2]
	source, err := reader.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: read error: %v\n", filename, err)
		return exitFailure
	}
	file, diagnostics := parser.ParseFile(filename, string(source))
	for _, diagnostic := range diagnostics.SortBySpan() {
		fmt.Fprintln(stderr, diagnostic.String())
	}
	if diagnostics.HasErrors() {
		return exitFailure
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: semantic lowering failed: %v\n", filename, err)
		return exitFailure
	}
	model, err := projectionIR(ir)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: generator adapter failed: %v\n", filename, err)
		return exitFailure
	}
	result, err := generator.Generate(model, nil)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: generation failed: %v\n", filename, err)
		return exitFailure
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		fmt.Fprintf(stderr, "gooo: %s: create output directory: %v\n", outputDir, err)
		return exitFailure
	}
	output := filepath.Join(outputDir, "semantic.gooo.go")
	if err := os.WriteFile(output, result.Source, 0o644); err != nil {
		fmt.Fprintf(stderr, "gooo: %s: write generated source: %v\n", output, err)
		return exitFailure
	}
	fmt.Fprintf(stdout, "generated: %s\n", output)
	return exitOK
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
