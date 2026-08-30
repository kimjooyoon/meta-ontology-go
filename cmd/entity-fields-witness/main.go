package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/entityfieldsv1"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/entityfields"
)

func main() {
	input := flag.String("input", "", "EntityFields source file")
	output := flag.String("output", "", "caller-owned output directory")
	flag.Parse()
	if *input == "" || *output == "" { fail(entityfields.FailureReport("ENTITY_FIELDS_INPUT_ARGUMENT_MISSING"), *output); os.Exit(1) }
	source, err := os.ReadFile(*input)
	if err != nil { fail(entityfields.FailureReport("ENTITY_FIELDS_SOURCE_MISSING"), *output); os.Exit(1) }
	observation, err := entityfieldsv1.Observe(filepath.ToSlash(*input), string(source))
	if err != nil { fail(entityfields.FailureReport(err.Error()), *output); os.Exit(1) }
	report := entityfields.Evaluate(observation)
	if err := entityfields.Verify(observation, report); err != nil { fail(entityfields.FailureReport(err.Error()), *output); os.Exit(1) }
	if err := writeSuccess(*output, observation, report); err != nil { fail(entityfields.FailureReport(err.Error()), *output); os.Exit(1) }
}

func writeSuccess(root string, observation entityfieldsv1.Observation, report entityfields.Report) error {
	if err := os.MkdirAll(root, 0o755); err != nil { return err }
	files := map[string][]byte{
		"source.gooo":       []byte(observation.Source),
		"formatted.gooo":    []byte(observation.Formatted),
		"generated.go":      observation.Generated,
		"source-map.json":   mustJSON(observation.SourceMap),
		"semantic-ir.json":  mustJSON(observation.Semantic),
		"lsp.json":          mustJSON(struct{ Symbols []entityfieldsv1.NavigationSymbol `json:"symbols"`; References []entityfieldsv1.NavigationReference `json:"references"` }{observation.Symbols, observation.References}),
		"report.json":       mustJSON(report),
		"report.md":         []byte(report.HumanReport + "\n"),
	}
	for name, data := range files { if err := os.WriteFile(filepath.Join(root, name), data, 0o644); err != nil { return err } }
	return nil
}

func fail(report entityfields.Report, root string) {
	if root == "" { return }
	if err := os.MkdirAll(root, 0o755); err != nil { return }
	_ = os.WriteFile(filepath.Join(root, "report.json"), mustJSON(report), 0o644)
	_ = os.WriteFile(filepath.Join(root, "report.md"), []byte(report.HumanReport+"\n"), 0o644)
	fmt.Fprintln(os.Stderr, report.Reason)
}

func mustJSON(value any) []byte {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil { return []byte(`{"decision":"REFUTED","reason":"ENTITY_FIELDS_JSON_ENCODING_FAILED"}`) }
	return append(data, '\n')
}
