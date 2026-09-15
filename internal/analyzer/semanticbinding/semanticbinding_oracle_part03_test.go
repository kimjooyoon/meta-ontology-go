package semanticbinding

import (
	"testing"
)

func locationForSpan(span Span) *fixtureLocation {
	return &fixtureLocation{
		Filename: span.Filename,
		Start:    fixturePosition{Offset: span.Start.Offset, Line: span.Start.Line, Column: span.Start.Column},
		End:      fixturePosition{Offset: span.End.Offset, Line: span.End.Line, Column: span.End.Column},
	}
}
func assertRecord(t *testing.T, index int, got, want fixtureRecord) {
	t.Helper()
	if got.Directive != want.Directive || got.Name != want.Name || got.ID != want.ID ||
		got.Subject != want.Subject || got.Pressure != want.Pressure {
		t.Fatalf("record[%d] = %#v, want oracle %#v", index, got, want)
	}
	if want.Location != nil && (got.Location == nil || *got.Location != *want.Location) {
		t.Fatalf("record[%d] location = %#v, want oracle %#v", index, got.Location, want.Location)
	}
}
func oracleCode(diagnostic string) Code {
	switch diagnostic {
	case "detached-comment":
		return CodeDetachedComment
	case "unknown-field":
		return CodeUnknownField
	case "duplicate-field":
		return CodeDuplicateField
	case "duplicate-id":
		return CodeDuplicateID
	case "invalid-uri":
		return CodeInvalidIdentity
	case "multi-name-declaration":
		return CodeAmbiguousBinding
	default:
		return Code(diagnostic)
	}
}
