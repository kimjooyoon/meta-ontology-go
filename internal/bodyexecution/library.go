package bodyexecution

import (
	"context"
	"fmt"
	"go/token"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

// GenerateLibrary projects an already authored entry and its producers. It does
// not run the program, select assembly candidates, load models, or create cases.
func GenerateLibrary(ctx context.Context, filename string, source []byte, entry, packageName string) (string, error) {
	if !token.IsIdentifier(packageName) || packageName == "_" || packageName == "main" || entry == "" {
		return "", fmt.Errorf("library generation requires an entry and an importable Go package name")
	}
	graph, err := prepareCompositionGraphForEntry(ctx, filename, source, entry)
	if err != nil {
		return "", err
	}
	if graph.hasAssembly() {
		return "", fmt.Errorf("library generation requires authored computes bodies; realize assembling activities first")
	}
	steps := make([]CompositionStep, 0, graph.count)
	for _, node := range graph.nodes[:graph.count] {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		generation, err := bodycodegen.GenerateWithPlanner(ctx, filename, source, node.Name, "", "")
		if err != nil {
			return "", fmt.Errorf("library activity %q: %w", node.ID, err)
		}
		steps = append(steps, CompositionStep{Generation: generation})
	}
	projection, err := emitCompositionProjection(steps, graph)
	if err != nil {
		return "", err
	}
	wrapper, err := graph.libraryEntry(entry)
	if err != nil {
		return "", err
	}
	return nameLibraryProjection(projection+wrapper, packageName, digest(source), graph.plan)
}

func (graph compositionGraph) libraryEntry(entry string) (string, error) {
	var parameters, body, origins strings.Builder
	root, selected := 0, -1
	for i, node := range graph.nodes[:graph.count] {
		if node.Name == entry {
			selected = i
		}
		arguments := make([]string, 0, len(node.inputSlots()))
		for _, slot := range node.inputSlots() {
			argument := fmt.Sprintf("value%d", slot.From)
			if slot.From < 0 {
				argument = fmt.Sprintf("input%d", root)
				fmt.Fprintf(&parameters, "%s %s,", argument, graph.valueGoType(slot.Type))
				fmt.Fprintf(&origins, "// %s receives activity=%q port=%q entity=%q\n", argument, node.ID, slot.Port, slot.EntityID)
				root++
			}
			arguments = append(arguments, argument)
		}
		fmt.Fprintf(&body, "value%d := %s(%s)\n", i, node.GoFunction, strings.Join(arguments, ","))
	}
	if selected < 0 {
		return "", fmt.Errorf("library entry is absent from the typed plan")
	}
	node := graph.nodes[selected]
	return fmt.Sprintf("\n//gooo:generated:start id=%q kind=\"library-entry\"\n%s"+
		"func Evaluate(%s) %s {\n%sreturn value%d\n}\n"+
		"//gooo:generated:end id=%q kind=\"library-entry\"\n",
		node.ID, origins.String(), parameters.String(), graph.valueGoType(node.OutputType), body.String(), selected, node.ID), nil
}
