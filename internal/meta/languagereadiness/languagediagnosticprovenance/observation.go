package languagediagnosticprovenance

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/formatter"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
)

type Observation struct {
	Origin          string
	Stage           string
	Code            string
	Message         string
	Hardness        string
	Severity        formatter.Severity
	Physical        Span
	Logical         Span
	GeneratedOffset int
	SourceMap       generator.SourceMap
	RequireSemantic bool
}

type ProvenanceError struct {
	Code string
}

func (failure *ProvenanceError) Error() string {
	return failure.Code
}

func provenanceError(code string) *ProvenanceError {
	return &ProvenanceError{Code: code}
}
