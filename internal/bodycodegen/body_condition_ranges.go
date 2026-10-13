package bodycodegen

import "go/ast"

// BodyConditionRanges reads source positions without evaluating a body or model.
func BodyConditionRanges(body string, arity int) ([]BodyConditionOrigin, error) {
	file, set, header, err := parseRenameBody(body, arity)
	if err != nil {
		return nil, err
	}
	normalized, err := rewriteLetDeclarations(body)
	if err != nil {
		return nil, err
	}
	var result []BodyConditionOrigin
	ast.Inspect(file, func(node ast.Node) bool {
		stmt, ok := node.(*ast.IfStmt)
		if !ok {
			return true
		}
		a, b := set.Position(stmt.Cond.Pos()).Offset-header, set.Position(stmt.Cond.End()).Offset-header
		result = append(result, BodyConditionOrigin{SourceOrdinal: len(result), BodySHA256: digest([]byte(normalized)), Start: a, End: b, Expression: normalized[a:b], Relation: "retained_condition_node"})
		return true
	})
	return result, nil
}

func NormalizedConditionBodySHA256(body string) (string, error) {
	normalized, err := rewriteLetDeclarations(body)
	return digest([]byte(normalized)), err
}

// BodyConditionSourceRanges returns positions in the original decoded Gooo body.
// Projection reports use BodyConditionRanges and normalized offsets instead.
func BodyConditionSourceRanges(body string, arity int) ([]BodyConditionOrigin, error) {
	ranges, err := BodyConditionRanges(body, arity)
	if err != nil {
		return nil, err
	}
	_, mapping, err := rewriteBodySurface(body)
	if err != nil {
		return nil, err
	}
	for i := range ranges {
		span := mapping.originalRange(FillOriginRange{ranges[i].Start, ranges[i].End})
		ranges[i].Start, ranges[i].End = span.Start, span.End
		ranges[i].Expression = body[span.Start:span.End]
	}
	return ranges, nil
}
