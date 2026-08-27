package ambiguitybudget

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func observeSource(path string, raw []byte) (SourceObservation, error) {
	file, diagnostics := syntax.ParseFile(path, string(raw))
	if file == nil || diagnostics.HasErrors() {
		return SourceObservation{Path: path, Digest: digestBytes(raw)}, fmt.Errorf("source parse is unknown")
	}
	observation := SourceObservation{Path: path, Digest: digestBytes(raw)}
	if file.Package != nil {
		observation.Package = file.Package.Name
	}
	if file.Namespace != nil {
		observation.Namespace = file.Namespace.Name
	}
	for _, declaration := range file.Declarations {
		switch declaration.(type) {
		case *syntax.EntityDecl:
			observation.Entities++
		case *syntax.ActivityDecl:
			observation.Activities++
		}
	}
	return observation, nil
}
