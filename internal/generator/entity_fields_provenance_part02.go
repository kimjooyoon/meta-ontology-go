package generator

func validSourcePosition(position Position) bool {
	return position.Line > 0 && position.Column > 0 && position.Offset >= 0
}
func sourceLocationBefore(left, right Position) bool {
	if left.Line != right.Line {
		return left.Line < right.Line
	}
	return left.Column < right.Column
}
func sourcePositionWithin(position, start, end Position) bool {
	return position.Offset >= start.Offset && position.Offset <= end.Offset && !sourceLocationBefore(position, start) && !sourceLocationBefore(end, position)
}
func fieldNameSource(field Field) SourceSpan {
	return field.NameSpan
}
