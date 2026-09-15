package bidir

func inputSpans(document Document) []SourceSpan {
	for _, declaration := range document.Declarations {
		if declaration.ID != "billing://activity/process" {
			continue
		}
		spans := make([]SourceSpan, len(declaration.Inputs))
		for index, input := range declaration.Inputs {
			spans[index] = input.Span
		}
		return spans
	}
	return nil
}
func outputIDs(document Document) []ID {
	for _, declaration := range document.Declarations {
		if declaration.ID != "billing://activity/process" {
			continue
		}
		ids := make([]ID, len(declaration.Outputs))
		for index, output := range declaration.Outputs {
			ids[index] = output.ID
		}
		return ids
	}
	return nil
}
func outputSpans(document Document) []SourceSpan {
	for _, declaration := range document.Declarations {
		if declaration.ID != "billing://activity/process" {
			continue
		}
		spans := make([]SourceSpan, len(declaration.Outputs))
		for index, output := range declaration.Outputs {
			spans[index] = output.Span
		}
		return spans
	}
	return nil
}
