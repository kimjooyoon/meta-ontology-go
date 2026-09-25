package languagediagnosticprovenance

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/formatter"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
)

func sourceMapObservation(fixture string) (Observation, bool) {
	ordinal, found := sourceMapOrdinal(fixture)
	if !found {
		return Observation{}, false
	}
	diagnostic := formatter.Diagnostic{
		Severity: formatter.SeverityError,
		Code:     formatter.CodeUnsupportedSyntax,
		Message:  "generated boundary cannot represent semantic declaration",
		Span: formatter.Span{
			Filename: "generated.go",
			Start:    formatter.Position{Offset: 15, Line: 2, Column: 6},
			End:      formatter.Position{Offset: 16, Line: 2, Column: 7},
		},
	}
	physical := formatterSpan(diagnostic.Span)
	sourceStart := generator.Position{
		Offset: 100 + ordinal*20, Line: 10 + ordinal, Column: 3,
	}
	sourceEnd := generator.Position{
		Offset: sourceStart.Offset + 8, Line: sourceStart.Line, Column: 11,
	}
	mapping := generator.SourceMapping{
		SemanticID: "billing://" + fixture + "/" + fixture,
		Kind:       fixture, Ordinal: ordinal,
		Source: generator.SourceSpan{
			URI: "model.gooo", Start: sourceStart, End: sourceEnd,
		},
		Generated: generator.SourceRange{
			Start: generator.Position{Offset: 10, Line: 2, Column: 1},
			End:   generator.Position{Offset: 30, Line: 2, Column: 21},
		},
	}
	return Observation{
		Origin: "FORMATTER", Stage: "FORMAT", Code: string(diagnostic.Code),
		Message: diagnostic.Message, Hardness: "NOT_APPLICABLE",
		Severity: diagnostic.Severity, Physical: physical, Logical: physical,
		GeneratedOffset: diagnostic.Span.Start.Offset,
		SourceMap:       generator.SourceMap{Mappings: []generator.SourceMapping{mapping}},
		RequireSemantic: true,
	}, true
}

func sourceMapOrdinal(fixture string) (int, bool) {
	switch fixture {
	case "entity":
		return 0, true
	case "field":
		return 1, true
	case "activity":
		return 2, true
	case "slot":
		return 3, true
	default:
		return 0, false
	}
}

func formatterSpan(span formatter.Span) Span {
	return Span{
		Start: Position{span.Filename, span.Start.Offset, span.Start.Line, span.Start.Column},
		End:   Position{span.Filename, span.End.Offset, span.End.Line, span.End.Column},
	}
}
