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
		body, _, err := rewriteBodySurface(a.ValueProgram)
		start, end := node.Span.Start, node.Span.End
		if err != nil || start < 0 || end <= start || end > len(body) {
			continue
		}
		result[i] = *bindBodyExpressionSource(source, sha, a, digest([]byte(body)),
			BodyElementOrigin{Start: start, End: end, Expression: body[start:end]})
	}
	return result, nil
}
