package valuecatalog

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type coreObservation struct {
	activities  int
	fingerprint string
	programs    map[string]string
}

func observeCore(path string, source []byte) (coreObservation, error) {
	file, diagnostics := syntax.ParseFile(path, string(source))
	if diagnostics.HasErrors() {
		return coreObservation{}, diagnostics.Error()
	}
	names := make(map[string]struct{})
	for _, declaration := range file.Decls {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok {
			continue
		}
		if _, duplicate := names[activity.Name]; duplicate {
			return coreObservation{}, fmt.Errorf("duplicate activity %q", activity.Name)
		}
		names[activity.Name] = struct{}{}
	}
	if len(names) != 2 {
		return coreObservation{}, fmt.Errorf("activity denominator = %d, want 2", len(names))
	}
	if _, ok := names[BaselineActivity]; !ok {
		return coreObservation{}, fmt.Errorf("baseline activity is missing")
	}
	if _, ok := names[ExtensionActivity]; !ok {
		return coreObservation{}, fmt.Errorf("extension activity is missing")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return coreObservation{}, err
	}
	programs := make(map[string]string)
	for _, node := range ir.Graph.Nodes() {
		programs[node.Name] = node.ValueProgram
	}
	return coreObservation{activities: len(names), fingerprint: ir.StableHash(), programs: programs}, nil
}
