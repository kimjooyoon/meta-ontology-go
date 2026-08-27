package operationprovenance

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func lowerSource(source []byte) (semantic.IR, error) {
	file, diagnostics := syntax.ParseFile("main.gooo", string(source))
	if diagnostics.HasErrors() || file == nil {
		return semantic.IR{}, fmt.Errorf("Gooo source has syntax errors")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return semantic.IR{}, fmt.Errorf("lower Gooo source: %w", err)
	}
	return ir, nil
}

func computedFields(value string) (map[string]string, error) {
	parts := strings.Split(value, "|")
	if len(parts) < 2 || (parts[0] != "metric" && parts[0] != "scenario") {
		return nil, fmt.Errorf("unsupported computes value %q", value)
	}
	fields := make(map[string]string, len(parts)-1)
	for _, part := range parts[1:] {
		key, raw, ok := strings.Cut(part, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("malformed field %q", part)
		}
		if _, exists := fields[key]; exists {
			return nil, fmt.Errorf("duplicate field %q", key)
		}
		fields[key] = raw
	}
	return fields, nil
}
