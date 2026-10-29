package bodycodegen

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

// FixedProgram owns an immutable typed projection. Each Evaluate call gets its
// own bindings, call counter and context; programs may be shared by callers.
type FixedProgram struct {
	activityID, sourceSHA256, projectionSHA256 string
	template                                   *criteriaTupleEvaluator
}

// NewFixedProgram prepares a fixed computes activity once. The source owns its
// behavior and field bounds; preparation does not invoke a model or process.
func NewFixedProgram(ctx context.Context, filename string, source []byte, activity string) (*FixedProgram, error) {
	if ctx == nil || len(source) == 0 || len(source) > 128<<10 || !utf8.Valid(source) {
		return nil, fmt.Errorf("fixed preparation requires a context and bounded UTF-8 source")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	body, err := prepareActivityBody(filename, source, activity)
	if err != nil {
		return nil, err
	}
	if body.activity.Assembly != nil {
		return nil, fmt.Errorf("fixed evaluation requires a fixed computes activity")
	}
	template, err := newCriteriaTupleEvaluator(ctx, body.base.source,
		body.activity.Name, body.base.report.RecordTypes, nil)
	if err != nil {
		return nil, err
	}
	template.evaluator.context = nil
	return &FixedProgram{activityID: body.activityID, sourceSHA256: digest(source),
		projectionSHA256: digest(body.base.source), template: template}, nil
}

// Evaluate interprets one input tuple without retaining its values or context.
func (p *FixedProgram) Evaluate(ctx context.Context, input []byte) (FixedValue, error) {
	if p == nil || p.template == nil || ctx == nil || len(input) == 0 || len(input) > 1<<20 || !utf8.Valid(input) {
		return FixedValue{}, fmt.Errorf("fixed evaluation requires a prepared program, context and bounded UTF-8 inputs")
	}
	if err := ctx.Err(); err != nil {
		return FixedValue{}, err
	}
	if err := completeness.CheckJSON(input, 1<<20); err != nil {
		return FixedValue{}, err
	}
	evaluator := *p.template
	evaluator.evaluator.context = ctx
	value, err := evaluator.evaluate(string(input), 0)
	if err != nil {
		return FixedValue{}, err
	}
	if len(value) > 1<<20 {
		return FixedValue{}, fmt.Errorf("fixed evaluation output exceeds 1048576 bytes")
	}
	return FixedValue{ActivityID: p.activityID, SourceSHA256: p.sourceSHA256,
		ProjectionSHA256: p.projectionSHA256, Value: value}, nil
}
