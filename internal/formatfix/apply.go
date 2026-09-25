package formatfix

import "fmt"

func Apply(source string, plan Plan) (string, error) {
	if err := Validate(plan); err != nil {
		return "", err
	}
	if plan.Resolution != ResolutionExact || digestBytes([]byte(source)) != plan.SourceDigest ||
		len(source) != plan.SourceBytes {
		return "", fmt.Errorf("format/fix source binding invalid")
	}
	if plan.Decision == DecisionFixedPoint {
		return source, nil
	}
	edit := plan.Edits[0]
	result := source[:edit.Start] + edit.Replacement + source[edit.End:]
	if digestBytes([]byte(result)) != plan.ResultDigest || len(result) != plan.ResultBytes {
		return "", fmt.Errorf("format/fix result binding invalid")
	}
	return result, nil
}
