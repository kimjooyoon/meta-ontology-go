package bodycodegen

import (
	"context"
	"fmt"
)

// ReplayIRBodyFill reconstructs an externally planned fill using its recorded
// decision. It scores the bound candidates without loading or calling a model.
func ReplayIRBodyFill(ctx context.Context, filename string, source []byte, plan IRBodyFillPlan, prior Result) (string, error) {
	if ctx == nil || len(prior.Source) > 256<<10 || len(prior.GoooSource) > 128<<10 {
		return "", fmt.Errorf("body-fill replay requires bounded source and context")
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	r := prior.Report
	if r.BodyFill == nil || r.BodyFill.OriginalSourceDigest != digest(source) || r.BodySearch != nil ||
		r.RecordAssembly != nil || r.BodyPaths != nil || r.Schema != schema || r.Decision != "PASS" ||
		!r.TypecheckPassed || !r.DeterministicReplay {
		return "", fmt.Errorf("body-fill observation is missing or source binding differs")
	}
	options := IRBodyFillOptions{recordedDecision: &r.BodyFill.Decision, TinyModelLoadMS: r.BodyFill.Timing.TinyModelLoadMS}
	expected, err := generateWithIRBodyFillOptions(ctx, filename, source, r.Activity, plan, "", "", options, nil)
	if err != nil {
		return "", fmt.Errorf("replay external body fill: %w", err)
	}
	if err := compareSourceFillReplay(prior, expected); err != nil {
		return "", err
	}
	return expected.GoooSource, nil
}
