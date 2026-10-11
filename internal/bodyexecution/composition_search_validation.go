package bodyexecution

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type CompositionSearchValidation struct {
	Stage       string                             `json:"stage"`
	Observation bodycodegen.SourceSearchValidation `json:"observation"`
}

func validateCompositionAssembly(ctx context.Context, filename string, source []byte, activity, stage string,
	checks *[]CompositionSearchValidation) error {
	if checks == nil {
		return bodycodegen.ValidateSourceAssembly(ctx, filename, source, activity)
	}
	spec, err := bodycodegen.SourceAssembly(ctx, filename, source, activity)
	if err != nil {
		return err
	}
	if !bodycodegen.IsSourceIRSearch(spec) {
		return bodycodegen.ValidateSourceAssembly(ctx, filename, source, activity)
	}
	observation, err := bodycodegen.ObserveSourceIRSearchValidation(ctx, filename, source, activity)
	*checks = append(*checks, CompositionSearchValidation{Stage: stage, Observation: observation})
	return err
}

type compositionValidationReplay struct {
	checks []CompositionSearchValidation
	next   int
}

// Historical absence does not acquire new observations. When a producer did
// record validation, replay checks that boundary against the original source
// checkpoint. This can repeat pure preflight probes, but never model selection.
func (r *compositionValidationReplay) check(ctx context.Context, filename string, source []byte,
	activity, stage string) error {
	if r.checks == nil {
		return nil
	}
	spec, err := bodycodegen.SourceAssembly(ctx, filename, source, activity)
	if err != nil || !bodycodegen.IsSourceIRSearch(spec) {
		return err
	}
	if r.next >= len(r.checks) {
		return fmt.Errorf("source search validation observation is missing")
	}
	prior := r.checks[r.next]
	if prior.Stage != stage || prior.Observation.SourceSHA256 != digest(source) || prior.Observation.Activity != activity {
		return fmt.Errorf("source search validation checkpoint or order differs")
	}
	actual, err := bodycodegen.ObserveSourceIRSearchValidation(ctx, filename, source, activity)
	if err != nil {
		return err
	}
	if compositionDigest(prior.Observation) != compositionDigest(actual) {
		return fmt.Errorf("source search validation inputs differ from replay")
	}
	r.next++
	return nil
}

func (graph compositionGraph) replaySearchPreflights(ctx context.Context, filename string, source []byte,
	r *compositionValidationReplay) error {
	if r.checks == nil {
		return nil
	}
	for _, node := range graph.nodes[:graph.count] {
		if node.Assembling && !graph.deferred[node.Name] {
			if err := r.check(ctx, filename, source, node.Name, "BODY_PREFLIGHT"); err != nil {
				return err
			}
		}
	}
	for _, helper := range graph.plan.Preparations {
		if !graph.deferred[helper.Name] {
			if err := r.check(ctx, filename, source, helper.Name, "BODY_PREFLIGHT"); err != nil {
				return err
			}
		}
	}
	return nil
}
