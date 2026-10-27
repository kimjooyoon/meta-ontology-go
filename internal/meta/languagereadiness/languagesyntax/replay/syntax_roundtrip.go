package replay

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type syntaxRoundTrip struct {
	before, after             *syntax.File
	shape, canonical          string
	astReplayed, byteReplayed bool
}

func replaySourceSyntax(path, source string, support syntax.EntityFieldsSupport) (syntaxRoundTrip, error) {
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport(path, source, support)
	if diagnostics.HasErrors() {
		return syntaxRoundTrip{}, diagnostics.Error()
	}
	canonical, err := syntax.FormatWithEntityFieldsSupport(file, support)
	if err != nil {
		return syntaxRoundTrip{}, fmt.Errorf("format: %w", err)
	}
	replayed, diagnostics := syntax.ParseFileWithEntityFieldsSupport(path, canonical, support)
	if diagnostics.HasErrors() {
		return syntaxRoundTrip{}, diagnostics.Error()
	}
	replayCanonical, err := syntax.FormatWithEntityFieldsSupport(replayed, support)
	if err != nil {
		return syntaxRoundTrip{}, fmt.Errorf("replay format: %w", err)
	}
	left, err := astShape(file)
	if err != nil {
		return syntaxRoundTrip{}, err
	}
	right, err := astShape(replayed)
	if err != nil {
		return syntaxRoundTrip{}, err
	}
	return syntaxRoundTrip{file, replayed, left, canonical, left == right, canonical == replayCanonical}, nil
}
