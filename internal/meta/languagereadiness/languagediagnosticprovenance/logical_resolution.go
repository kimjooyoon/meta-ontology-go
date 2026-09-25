package languagediagnosticprovenance

// validLogicalPosition accepts Go's explicit column-zero representation for a
// valid position whose logical column is unavailable after a //line remap.
func validLogicalPosition(position Position) bool {
	if position.Column < 0 {
		return false
	}
	if position.Column == 0 {
		position.Column = 1
	}
	return validPosition(position)
}

// validLogicalSpan never invents a missing logical column. Both ends must use
// the same lower resolution, while byte, line, and filename invariants remain
// identical to an exact span.
func validLogicalSpan(span Span) bool {
	unknownStart := span.Start.Column == 0
	unknownEnd := span.End.Column == 0
	if unknownStart != unknownEnd {
		return false
	}
	if !unknownStart {
		return validSpan(span)
	}
	return validLogicalPosition(span.Start) &&
		validLogicalPosition(span.End) &&
		span.Start.Filename == span.End.Filename &&
		span.End.Offset > span.Start.Offset &&
		span.End.Line >= span.Start.Line
}
