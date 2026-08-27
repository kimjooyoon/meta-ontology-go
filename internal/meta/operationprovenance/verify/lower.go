package verify

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const receiptSchema = "gooo/meta-operation-provenance-receipt/v3"
const toolchain = "go1.27.0"

func lower(source []byte) (semantic.IR, error) {
	file, diagnostics := syntax.ParseFile("main.gooo", string(source))
	if diagnostics.HasErrors() || file == nil {
		return semantic.IR{}, fmt.Errorf("source has syntax errors")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return semantic.IR{}, fmt.Errorf("lower source: %w", err)
	}
	return ir, nil
}
