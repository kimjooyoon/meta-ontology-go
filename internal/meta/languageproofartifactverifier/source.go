package languageproofartifactverifier

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type binding struct {
	Name string
	ID   string
}

type projection struct {
	Package        string
	Namespace      string
	Activity       string
	Inputs         []binding
	Output         binding
	SemanticDigest string
}

// projectSource is deliberately a consumer-side projection. It uses the core
// syntax and bidir boundaries, but it does not import the producer or the
// source-execution package. The lexer discards comments before lowering, so a
// comment-only intervention preserves this semantic projection.
func projectSource(raw []byte, selected string) (projection, error) {
	file, diagnostics := syntax.ParseFile("proof-carrying-consumer.gooo", string(raw))
	if file == nil || file.Package == nil || file.Namespace == nil || diagnostics.HasErrors() {
		return projection{}, fmt.Errorf("source parser rejected input")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return projection{}, fmt.Errorf("source lowerer rejected input: %w", err)
	}
	result := projection{Package: file.Package.Name, Namespace: file.Namespace.Name, SemanticDigest: "sha256:" + ir.StableHash()}
	entities := map[string]binding{}
	declarations := file.Decls
	if declarations == nil {
		declarations = file.Declarations
	}
	var activity *syntax.ActivityDecl
	for _, declaration := range declarations {
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			entities[value.Name] = binding{Name: value.Name, ID: value.ID}
		case *syntax.ActivityDecl:
			if value.Name == selected {
				activity = value
			}
		}
	}
	if activity == nil {
		return projection{}, fmt.Errorf("activity %q missing", selected)
	}
	result.Activity = activity.Name
	parameters := activity.Inputs
	if parameters == nil {
		parameters = activity.Parameters
	}
	result.Inputs = make([]binding, len(parameters))
	for index, parameter := range parameters {
		value, ok := entities[parameter.Name]
		if !ok {
			return projection{}, fmt.Errorf("input entity %q missing", parameter.Name)
		}
		result.Inputs[index] = value
	}
	outputName := activity.Output
	if outputName == "" {
		outputName = activity.Result.Name
	}
	value, ok := entities[outputName]
	if !ok {
		return projection{}, fmt.Errorf("output entity %q missing", outputName)
	}
	result.Output = value
	return result, nil
}
