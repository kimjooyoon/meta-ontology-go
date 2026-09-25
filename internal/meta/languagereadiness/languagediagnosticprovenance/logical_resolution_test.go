package languagediagnosticprovenance

import "testing"

func TestLogicalSpanPreservesUnknownColumnResolution(t *testing.T) {
	span := oneByteSpan(Position{
		Filename: "model.gooo",
		Offset:   12,
		Line:     40,
		Column:   0,
	})
	if span.End.Column != 0 {
		t.Fatalf("logical column resolution was invented: got %d, want 0", span.End.Column)
	}
	if !validLogicalSpan(span) {
		t.Fatal("column-zero logical span must remain valid at lower resolution")
	}
	if validSpan(span) {
		t.Fatal("physical span validator must continue to require an exact column")
	}
}

func TestLogicalSpanRejectsAsymmetricUnknownColumn(t *testing.T) {
	span := Span{
		Start: Position{Filename: "model.gooo", Offset: 12, Line: 40, Column: 0},
		End:   Position{Filename: "model.gooo", Offset: 13, Line: 40, Column: 1},
	}
	if validLogicalSpan(span) {
		t.Fatal("one-sided logical resolution loss must fail closed")
	}
}
