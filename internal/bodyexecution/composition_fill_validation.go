package bodyexecution

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type CompositionFillValidation struct {
	Stage       string                           `json:"stage"`
	Observation bodycodegen.SourceFillValidation `json:"observation"`
}

func (r *compositionValidationReplay) checkFill(ctx context.Context, filename string, source []byte,
	activity, stage string) error {
	if r.fills == nil {
		return nil // Older producers did not retain this boundary.
	}
	if r.nextFill >= len(r.fills) {
		return fmt.Errorf("source fill validation observation is missing")
	}
	prior := r.fills[r.nextFill]
	if prior.Stage != stage || prior.Observation.SourceSHA256 != digest(source) || prior.Observation.Activity != activity {
		return fmt.Errorf("source fill validation checkpoint or order differs")
	}
	actual, err := bodycodegen.ObserveSourceIRBodyFillValidation(ctx, filename, source, activity)
	if err != nil {
		return err
	}
	if compositionDigest(prior.Observation) != compositionDigest(actual) {
		return fmt.Errorf("source fill validation inputs differ from replay")
	}
	r.nextFill++
	return nil
}
