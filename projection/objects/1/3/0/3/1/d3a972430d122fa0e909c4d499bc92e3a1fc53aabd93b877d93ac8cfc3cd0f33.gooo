package formatter

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func syntaxSpan(span syntax.Span) Span {
	return Span{
		Filename: span.Filename,
		Start: Position{
			Offset: span.Start.Offset,
			Line:   span.Start.Line,
			Column: span.Start.Column,
		},
		End: Position{
			Offset: span.End.Offset,
			Line:   span.End.Line,
			Column: span.End.Column,
		},
	}
}
