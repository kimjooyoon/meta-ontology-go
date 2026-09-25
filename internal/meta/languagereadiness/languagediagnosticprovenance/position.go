package languagediagnosticprovenance

type Position struct {
	Filename string `json:"filename"`
	Offset   int    `json:"offset"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
}

type Span struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

func oneByteSpan(position Position) Span {
	end := position
	end.Offset++
	if end.Column > 0 {
		end.Column++
	}
	return Span{Start: position, End: end}
}

func samePosition(left, right Position) bool {
	return left.Filename == right.Filename && left.Offset == right.Offset &&
		left.Line == right.Line && left.Column == right.Column
}

func sameSpan(left, right Span) bool {
	return samePosition(left.Start, right.Start) &&
		samePosition(left.End, right.End)
}
