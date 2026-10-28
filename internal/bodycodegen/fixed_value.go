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
	program, err := NewFixedProgram(ctx, filename, source, activity)
	if err != nil {
		return FixedValue{}, err
	}
	return program.Evaluate(ctx, input)
}
