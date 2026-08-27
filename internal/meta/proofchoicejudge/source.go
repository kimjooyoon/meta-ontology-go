package proofchoicejudge

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func lowerSource(path string, source []byte) (lowered, error) {
	file, diagnostics := syntax.ParseFile(path, string(source))
	if file == nil || len(diagnostics) != 0 {
		return lowered{}, fmt.Errorf("SOURCE_PARSE_UNKNOWN")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return lowered{}, fmt.Errorf("SOURCE_LOWER_UNKNOWN: %w", err)
	}
	return collectValues(ir)
}
