package bodycodegen

import (
	"context"
	"fmt"
)

// SourceFillValidation retains calls made by the existing deterministic fill
// preflight, including entered prefixes when validation fails or is canceled.
type SourceFillValidation struct {
	Schema       string            `json:"schema"`
	SourceSHA256 string            `json:"source_sha256"`
	Activity     string            `json:"activity"`
	ActivityID   string            `json:"activity_id"`
	Completed    bool              `json:"completed"`
	Error        string            `json:"error,omitempty"`
	LocalCalls   *LocalCallHistory `json:"local_calls,omitempty"`
}

// ObserveSourceIRBodyFillValidation runs the same bounded deterministic fill
// used by ValidateSourceAssembly once. It loads no model and retains no timing
// or proposed body; Completed describes validation, not functional accuracy.
func ObserveSourceIRBodyFillValidation(ctx context.Context, filename string, source []byte,
	activity string) (SourceFillValidation, error) {
	r := SourceFillValidation{Schema: "gooo/source-fill-validation/v1",
		SourceSHA256: digest(source), Activity: activity}
	finish := func(err error) (SourceFillValidation, error) {
		r.Completed = err == nil
		if err != nil {
			r.Error = err.Error()
		}
		return r, err
	}
	if ctx == nil {
		return finish(fmt.Errorf("source fill validation requires a context"))
	}
	spec, err := SourceAssembly(ctx, filename, source, activity)
	if err != nil {
		return finish(err)
	}
	if !IsSourceIRBodyFill(spec) {
		return finish(fmt.Errorf("activity has no source-declared body fill"))
	}
	r.ActivityID, err = fillActivityID(filename, source, activity)
	if err != nil {
		return finish(err)
	}
	_, err = GenerateWithSourceIRBodyFill(ctx, filename, source, activity, spec, "", "",
		IRBodyFillOptions{localCalls: &r.LocalCalls})
	return finish(err)
}
