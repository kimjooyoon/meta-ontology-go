package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"unicode"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func runCheck(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: gooo check <file.gooo>")
		return exitUsage
	}
	ir, err := loadIR(args[0])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	nodes := ir.Graph.Nodes()
	facts := ir.Graph.Facts()
	if _, err := fmt.Fprintf(stdout, "ok: %s (nodes=%d facts=%d hash=%s)\n", args[0], len(nodes), len(facts), ir.StableHash()); err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	return exitOK
}

type queryOutput struct {
	Version   string          `json:"version"`
	Package   string          `json:"package"`
	Namespace string          `json:"namespace"`
	Hash      string          `json:"hash"`
	Nodes     []semantic.Node `json:"nodes"`
	Facts     []semantic.Fact `json:"facts"`
}

func runQuery(command string, args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintf(stderr, "usage: gooo %s <file.gooo>\n", command)
		return exitUsage
	}
	ir, err := loadIR(args[0])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	payload := queryOutput{
		Version:   ir.Version,
		Package:   ir.Package,
		Namespace: ir.Namespace.String(),
		Hash:      ir.StableHash(),
		Nodes:     ir.Graph.Nodes(),
		Facts:     ir.Graph.Facts(),
	}
	if err := json.NewEncoder(stdout).Encode(payload); err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	return exitOK
}

func loadIR(path string) (semantic.IR, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return semantic.IR{}, fmt.Errorf("read %s: %w", path, err)
	}
	file, diagnostics := syntax.ParseFile(path, string(source))
	if err := diagnostics.Error(); err != nil {
		return semantic.IR{}, err
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return semantic.IR{}, fmt.Errorf("lower %s: %w", path, err)
	}
	return ir, nil
}

func runGenerate(args []string, stdout, stderr io.Writer) int {
	if len(args) != 3 || args[1] != "--out" {
		fmt.Fprintln(stderr, "usage: gooo generate <file.gooo> --out <directory>")
		return exitUsage
	}
	sourcePath, outputDir := args[0], args[2]
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		fmt.Fprintln(stderr, fmt.Errorf("read %s: %w", sourcePath, err))
		return exitFailure
	}
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if err := diagnostics.Error(); err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		fmt.Fprintln(stderr, fmt.Errorf("lower %s: %w", sourcePath, err))
		return exitFailure
	}
	previous, err := readPrevious(filepath.Join(outputDir, "semantic.gooo.go"))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	result, err := generator.Generate(toGeneratorIR(file, ir), previous)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		fmt.Fprintln(stderr, fmt.Errorf("create %s: %w", outputDir, err))
		return exitFailure
	}
	outputPath := filepath.Join(outputDir, "semantic.gooo.go")
	if err := os.WriteFile(outputPath, result.Source, 0o644); err != nil {
		fmt.Fprintln(stderr, fmt.Errorf("write %s: %w", outputPath, err))
		return exitFailure
	}
	fmt.Fprintf(stdout, "generated: %s\n", outputPath)
	return exitOK
}

func readPrevious(path string) ([]byte, error) {
	previous, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read previous output %s: %w", path, err)
	}
	return previous, nil
}

func toGeneratorIR(file *syntax.File, ir semantic.IR) generator.SemanticIR {
	result := generator.SemanticIR{Package: ir.Package}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Entity {
			continue
		}
		result.Entities = append(result.Entities, generator.Entity{
			ID:     node.ID.String(),
			Name:   node.Name,
			GoName: node.Name,
			Source: toGeneratorSpan(node.Span),
		})
	}
	for _, declaration := range file.Decls {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok {
			continue
		}
		node, exists := ir.Graph.NodeByName(ir.Namespace, activity.Name)
		if !exists {
			continue
		}
		projected := generator.Activity{
			ID:     node.ID.String(),
			Name:   node.Name,
			GoName: node.Name,
			Source: toGeneratorSpan(node.Span),
		}
		for _, input := range activity.Parameters {
			if entity, found := ir.Graph.NodeByName(ir.Namespace, input.Name); found {
				projected.Inputs = append(projected.Inputs, generator.Port{
					ID:       entity.ID.String(),
					Name:     lowerCamel(entity.Name),
					GoName:   lowerCamel(entity.Name),
					EntityID: entity.ID.String(),
				})
			}
		}
		if entity, found := ir.Graph.NodeByName(ir.Namespace, activity.Result.Name); found {
			projected.Outputs = append(projected.Outputs, generator.Port{
				ID:       entity.ID.String(),
				Name:     lowerCamel(entity.Name),
				GoName:   lowerCamel(entity.Name),
				EntityID: entity.ID.String(),
			})
		}
		result.Activities = append(result.Activities, projected)
	}
	return result
}

func toGeneratorSpan(span semantic.Span) generator.SourceSpan {
	return generator.SourceSpan{
		URI: span.File,
		Start: generator.Position{
			Offset: span.Start.Offset,
			Line:   span.Start.Line,
			Column: span.Start.Column,
		},
		End: generator.Position{
			Offset: span.End.Offset,
			Line:   span.End.Line,
			Column: span.End.Column,
		},
	}
}

func lowerCamel(name string) string {
	runes := []rune(name)
	if len(runes) == 0 {
		return name
	}
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
