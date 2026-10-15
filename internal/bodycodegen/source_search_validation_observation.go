package bodycodegen

import (
	"context"
	"fmt"
)

// SourceSearchValidation describes one original source-search preflight. Its
// probes can enter helpers before any candidate body has been selected.
type SourceSearchValidation struct {
	Schema       string            `json:"schema"`
	SourceSHA256 string            `json:"source_sha256"`
	Activity     string            `json:"activity"`
	ActivityID   string            `json:"activity_id"`
	Completed    bool              `json:"completed"`
	Error        string            `json:"error,omitempty"`
	LocalCalls   *LocalCallHistory `json:"local_calls,omitempty"`
}

// ObserveSourceIRSearchValidation records the existing validation invocation.
// It neither selects a final body nor calls a model. Ordinary validation keeps
// its existing entry point; consumers opt in to retaining this observation.
func ObserveSourceIRSearchValidation(ctx context.Context, filename string, source []byte,
	activity string) (SourceSearchValidation, error) {
	r := SourceSearchValidation{Schema: "gooo/source-search-validation/v1",
		SourceSHA256: digest(source), Activity: activity}
	finish := func(err error) (SourceSearchValidation, error) {
		r.Completed = err == nil
		if err != nil {
			r.Error = err.Error()
		}
		return r, err
	}
	if ctx == nil {
		return finish(fmt.Errorf("source IR search validation requires a context"))
	}
	spec, err := SourceAssembly(ctx, filename, source, activity)
	if err != nil {
		return finish(err)
	}
	if !IsSourceIRSearch(spec) {
		return finish(fmt.Errorf("activity has no source-declared IR search"))
	}
	ctx, r.LocalCalls = captureSearchCalls(ctx, filename, source, activity)
	capture := ctx.Value(localCaptureKey{}).(*searchCallCapture)
	r.ActivityID = capture.report.ActivityID
	return finish(ValidateSourceIRSearch(ctx, filename, source, activity, spec))
}
