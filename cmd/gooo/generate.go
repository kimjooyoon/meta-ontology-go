package main

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
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

type generationResult struct {
	ir     semantic.IR
	result generator.Result
	err    error
}

type generateInput struct {
	source      []byte
	file        *syntax.File
	diagnostics syntax.Diagnostics
	previousGo  []byte
}

type generateArtifacts struct {
	ir           semantic.IR
	result       generator.Result
	output       string
	manifestPath string
	manifest     projectionManifest
}

func reportGenerateUsage(jsonMode bool, stdout, stderr io.Writer, err error) int {
	if jsonMode {
		return reportUsage(true, stdout, stderr, "generate", err.Error())
	}
	fmt.Fprintln(stderr, err)
	return exitUsage
}

func readGenerateInput(options generateOptions, reader SourceReader, parser SourceParser, jsonMode bool, stdout, stderr io.Writer, deadline time.Time) (generateInput, int) {
	source, err := readSourceWithDeadline(reader, options.filename, remainingDeadline(deadline))
	if err != nil {
		if jsonMode {
			return generateInput{}, reportFailure(true, stdout, stderr, "generate", options.filename, "io.read", err.Error(), syntax.Span{})
		}
		fmt.Fprintf(stderr, "gooo: %s: read error: %v\n", options.filename, err)
		return generateInput{}, exitFailure
	}
	file, diagnostics, err := parseWithDeadline(parser, options.filename, string(source), remainingDeadline(deadline))
	if err != nil {
		if jsonMode {
			return generateInput{}, reportFailure(true, stdout, stderr, "generate", options.filename, "parse", err.Error(), syntaxFileSpan(file))
		}
		fmt.Fprintf(stderr, "gooo: %s: parse error: %v\n", options.filename, err)
		return generateInput{}, exitFailure
	}
	if diagnostics.HasErrors() {
		if jsonMode {
			if err := writeJSONReport(stdout, newJSONReport("generate", "error", options.filename, syntaxCLIDiagnostics(diagnostics))); err != nil {
				return generateInput{}, exitFailure
			}
		} else if !reportDiagnostics(diagnostics, stderr) {
			return generateInput{}, exitFailure
		}
		return generateInput{}, exitFailure
	}
	if !jsonMode && !reportDiagnostics(diagnostics, stderr) {
		return generateInput{}, exitFailure
	}
	var previousGo []byte
	if options.previousGo != "" {
		previousGo, err = readPreviousWithDeadline(reader, options.previousGo, remainingDeadline(deadline))
		if err != nil {
			if jsonMode {
				return generateInput{}, reportFailure(true, stdout, stderr, "generate", options.previousGo, "io.read-previous-go", err.Error(), syntax.Span{})
			}
			fmt.Fprintf(stderr, "gooo: %s: read previous Go error: %v\n", options.previousGo, err)
			return generateInput{}, exitFailure
		}
	}
	return generateInput{source: source, file: file, diagnostics: diagnostics, previousGo: previousGo}, exitOK
}

func buildGenerateArtifacts(options generateOptions, input generateInput, jsonMode bool, stdout, stderr io.Writer, deadline time.Time) (generateArtifacts, int) {
	generation, err := generateWithDeadline(input.file, input.previousGo, remainingDeadline(deadline))
	if err != nil {
		return generateArtifacts{}, reportGenerateError(jsonMode, stdout, stderr, options.filename, "generator.generate", "generation failed", err, input.file)
	}
	output := filepath.Join(options.outputDir, generatedFileName)
	manifestPath := options.manifestPath
	if manifestPath == "" {
		manifestPath = filepath.Join(options.outputDir, generatedManifestFileName)
	}
	manifest, err := buildProjectionManifest(options.filename, output, input.source, input.previousGo, generation.ir, generation.result)
	if err != nil {
		return generateArtifacts{}, reportGenerateError(jsonMode, stdout, stderr, options.filename, "manifest.build", "manifest failed", err, input.file)
	}
	root, err := canonicalOutputRoot(options.outputDir)
	if err != nil {
		return generateArtifacts{}, reportGenerateError(jsonMode, stdout, stderr, options.outputDir, "io.mkdir", "output root", err, input.file)
	}
	output, err = resolveOutputPath(root, generatedFileName)
	if err != nil {
		return generateArtifacts{}, reportGenerateError(jsonMode, stdout, stderr, options.outputDir, "io.output-path", "output path", err, input.file)
	}
	manifestPath, err = resolveManifestPath(root, manifestPath)
	if err != nil {
		return generateArtifacts{}, reportGenerateError(jsonMode, stdout, stderr, options.outputDir, "io.manifest-path", "manifest path", err, input.file)
	}
	manifest.GeneratedFile = output
	return generateArtifacts{ir: generation.ir, result: generation.result, output: output, manifestPath: manifestPath, manifest: manifest}, exitOK
}

func reportGenerateError(jsonMode bool, stdout, stderr io.Writer, filename, code, prefix string, err error, file *syntax.File) int {
	if jsonMode {
		return reportFailure(true, stdout, stderr, "generate", filename, code, err.Error(), syntaxFileSpan(file))
	}
	fmt.Fprintf(stderr, "gooo: %s: %s: %v\n", filename, prefix, err)
	return exitFailure
}

func writeGenerateArtifacts(artifacts generateArtifacts, jsonMode bool, stdout, stderr io.Writer) int {
	if err := writeGeneratedOutput(artifacts.output, artifacts.result.Source); err != nil {
		return reportGenerateIOError(jsonMode, stdout, stderr, artifacts.output, "io.write", "write generated source", err)
	}
	manifestBytes, err := jsonManifestBytes(artifacts.manifest)
	if err != nil {
		return reportGenerateIOError(jsonMode, stdout, stderr, artifacts.manifestPath, "io.write-manifest", "write manifest", err)
	}
	if err := writeGeneratedOutput(artifacts.manifestPath, manifestBytes); err != nil {
		return reportGenerateIOError(jsonMode, stdout, stderr, artifacts.manifestPath, "io.write-manifest", "write manifest", err)
	}
	return exitOK
}

func reportGenerateIOError(jsonMode bool, stdout, stderr io.Writer, filename, code, prefix string, err error) int {
	if jsonMode {
		return reportFailure(true, stdout, stderr, "generate", filename, code, err.Error(), syntax.Span{})
	}
	fmt.Fprintf(stderr, "gooo: %s: %s: %v\n", filename, prefix, err)
	return exitFailure
}

func reportGenerateSuccess(options generateOptions, input generateInput, artifacts generateArtifacts, jsonMode bool, stdout io.Writer) int {
	if !jsonMode {
		fmt.Fprintf(stdout, "generated: %s\n", filepath.Join(options.outputDir, generatedFileName))
		return exitOK
	}
	report := newJSONReport("generate", "ok", options.filename, syntaxCLIDiagnostics(input.diagnostics))
	report.Output = artifacts.output
	report.Manifest = artifacts.manifestPath
	report.PreviousGo = options.previousGo
	report.ProtectedBytesEqual = &artifacts.manifest.ProtectedBytesEqual
	report.SemanticHash = artifacts.ir.StableHash()
	if err := writeJSONReport(stdout, report); err != nil {
		return exitFailure
	}
	return exitOK
}

const generatedManifestFileName = "semantic.gooo.manifest.jsonl"

type generateOptions struct {
	filename     string
	outputDir    string
	previousGo   string
	manifestPath string
}

func parseGenerateArguments(args []string) (generateOptions, error) {
	usage := "usage: gooo generate <file.gooo> --out <directory>"
	if len(args) == 0 {
		return generateOptions{}, fmt.Errorf("%s", usage)
	}
	options := generateOptions{filename: args[0]}
	for index := 1; index < len(args); index++ {
		if index+1 >= len(args) {
			return generateOptions{}, fmt.Errorf("%s", usage)
		}
		value := args[index+1]
		if value == "" {
			return generateOptions{}, fmt.Errorf("%s", usage)
		}
		switch args[index] {
		case "--out":
			if options.outputDir != "" {
				return generateOptions{}, fmt.Errorf("%s", usage)
			}
			options.outputDir = value
		case "--previous-go":
			if options.previousGo != "" {
				return generateOptions{}, fmt.Errorf("%s", usage)
			}
			options.previousGo = value
		case "--manifest":
			if options.manifestPath != "" {
				return generateOptions{}, fmt.Errorf("%s", usage)
			}
			options.manifestPath = value
		default:
			return generateOptions{}, fmt.Errorf("%s", usage)
		}
		index++
	}
	if options.outputDir == "" {
		return generateOptions{}, fmt.Errorf("%s", usage)
	}
	return options, nil
}

func generateWithDeadline(file *syntax.File, previous []byte, timeout time.Duration) (generationResult, error) {
	if timeout <= 0 {
		return generationResult{}, errCommandDeadline
	}
	result := make(chan generationResult, 1)
	go func() {
		ir, err := bidir.Lower(file)
		if err != nil {
			result <- generationResult{err: fmt.Errorf("semantic lowering failed: %w", err)}
			return
		}
		model, err := projectionIR(ir)
		if err != nil {
			result <- generationResult{err: fmt.Errorf("generator adapter failed: %w", err)}
			return
		}
		generated, err := generator.Generate(model, previous)
		result <- generationResult{ir: ir, result: generated, err: err}
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case generated := <-result:
		return generated, generated.err
	case <-timer.C:
		return generationResult{}, errCommandDeadline
	}
}

func generateSource(file *syntax.File) ([]byte, error) {
	result, err := generateWithDeadline(file, nil, commandDeadline)
	if err != nil {
		return nil, err
	}
	return result.result.Source, nil
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
