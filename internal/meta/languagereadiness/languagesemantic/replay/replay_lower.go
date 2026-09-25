package replay

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func lower(path string, source []byte) (semantic.IR, error) {
	file, diagnostics := syntax.ParseFile(path, string(source))
	if diagnostics.HasErrors() {
		return semantic.IR{}, fmt.Errorf("parse %s: diagnostics contain errors", path)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return semantic.IR{}, fmt.Errorf("lower %s: %w", path, err)
	}
	normalized, err := ir.Normalized()
	if err != nil {
		return semantic.IR{}, fmt.Errorf("normalize %s: %w", path, err)
	}
	if err := normalized.Validate(); err != nil {
		return semantic.IR{}, fmt.Errorf("validate %s: %w", path, err)
	}
	return normalized, nil
}
