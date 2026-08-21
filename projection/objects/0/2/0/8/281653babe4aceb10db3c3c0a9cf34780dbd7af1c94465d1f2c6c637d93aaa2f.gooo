package cycles

// Position identifies a source location. Zero values are valid and mean that
// the source location is unavailable.
type Position struct {
	Offset int
	Line   int
	Column int
}

// Span carries optional source provenance into diagnostics.
type Span struct {
	File  string
	Start Position
	End   Position
}
