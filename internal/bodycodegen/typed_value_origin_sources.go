package bodycodegen

func typedOriginSources(source []byte, nodes []RecordFlowNode) ([]BodyExpressionSource, error) {
	activities, err := bodySourceActivities(source)
	if err != nil {
		return nil, err
	}
	sha := digest(source)
	result := make([]BodyExpressionSource, len(nodes))
	for i, node := range nodes {
		result[i] = *UnmappedReadSource(sha, "BODY_RANGE_UNAVAILABLE")
		if node.Span.View == "record-field-projection" || node.Span.View == "derived-record-output" {
			result[i].Reason = "DERIVED_RECORD_FIELD_PROJECTION"
			continue
		}
		if node.Span.View == "input-port" {
			result[i].Reason = "INPUT_PORT_IDENTITY"
			continue
		}
		a := activities[node.Span.ActivityID]
		if a == nil || node.Span.View != "computes" {
			continue
		}
		start, end := node.Span.Start, node.Span.End
		if start < 0 || end <= start || end > len(a.ValueProgram) {
			continue
		}
		span := a.ValueProgramSpan
		if span.Start.Offset < 0 || span.End.Offset > len(source) || span.Start.Offset > span.End.Offset {
			continue
		}
		first, last, err := SourceLiteralByteRange(string(source[span.Start.Offset:span.End.Offset]), a.ValueProgram, start, end)
		if err != nil {
			continue
		}
		body, err := rewriteLetDeclarations(a.ValueProgram)
		if err != nil {
			continue
		}
		result[i] = BodyExpressionSource{Status: "MAPPED", View: "selected_gooo_source", SourceSHA256: sha,
			BodySHA256: digest([]byte(body)), Start: span.Start.Offset + first, End: span.Start.Offset + last,
			DecodedStart: start, DecodedEnd: end, Expression: a.ValueProgram[start:end]}
	}
	return result, nil
}
