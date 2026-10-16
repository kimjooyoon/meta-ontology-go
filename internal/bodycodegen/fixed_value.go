package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

// FixedValue retains the source and projection which produced one pure value.
// Evaluation is interpretation of the existing typed profile, not a native run.
type FixedValue struct {
	ActivityID       string          `json:"activity_id"`
	SourceSHA256     string          `json:"source_sha256"`
	ProjectionSHA256 string          `json:"projection_sha256"`
	Value            json.RawMessage `json:"value"`
}

// EvaluateFixedValue evaluates positional JSON inputs using the existing pure
// body evaluator. Fixed source owns the decision; no model or process is used.
func EvaluateFixedValue(ctx context.Context, filename string, source []byte,
	activity string, input []byte) (FixedValue, error) {
	if ctx == nil || len(source) == 0 || len(source) > 128<<10 ||
		len(input) == 0 || len(input) > 1<<20 || !utf8.Valid(source) || !utf8.Valid(input) {
		return FixedValue{}, fmt.Errorf("fixed evaluation requires a context and bounded UTF-8 source and inputs")
	}
	if err := ctx.Err(); err != nil {
		return FixedValue{}, err
	}
	if err := completeness.CheckJSON(input, 1<<20); err != nil {
		return FixedValue{}, err
	}
	body, err := prepareActivityBody(filename, source, activity)
	if err != nil {
		return FixedValue{}, err
	}
	if body.activity.Assembly != nil {
		return FixedValue{}, fmt.Errorf("fixed evaluation requires a fixed computes activity")
	}
	evaluator, err := newCriteriaTupleEvaluator(ctx, body.base.source,
		body.activity.Name, body.base.report.RecordTypes, nil)
	if err != nil {
		return FixedValue{}, err
	}
	value, err := evaluator.evaluate(string(input), 0)
	if err != nil {
		return FixedValue{}, err
	}
	if len(value) > 1<<20 {
		return FixedValue{}, fmt.Errorf("fixed evaluation output exceeds 1048576 bytes")
	}
	return FixedValue{ActivityID: body.activityID, SourceSHA256: digest(source),
		ProjectionSHA256: digest(body.base.source), Value: value}, nil
}
