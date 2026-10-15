package bodycodegen

import (
	"context"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func declaredPathObservation(spec *assemblyspec.Spec) *PathObservationOptions {
	if spec == nil || spec.Observation == nil {
		return nil
	}
	o := spec.Observation
	return &PathObservationOptions{Inputs: append([]int64(nil), o.Inputs...), MaxCandidates: o.MaxCandidates,
		MaxRounds: o.MaxRounds, OracleActivity: o.OracleActivity, ReuseProbeOutputs: o.ReuseProbeOutputs,
		ResolveUniqueCandidate: o.ResolveUniqueCandidate}
}

// Source-owned observation is resolved once, after source binding and before
// oracle execution or model loading. External options cannot replace that meaning.
func sourcePathObservation(activity *syntax.ActivityDecl, external *PathObservationOptions) (*PathObservationOptions, error) {
	options := external
	if activity.Assembly != nil && activity.Assembly.Spec.Observation != nil {
		if external != nil {
			return nil, fmt.Errorf("observe is declared by the Gooo activity; omit external path observation options")
		}
		options = declaredPathObservation(&activity.Assembly.Spec)
	}
	if activity.Assembly != nil {
		if err := observationAvoidsHoldout(&activity.Assembly.Spec, options); err != nil {
			return nil, err
		}
	}
	return copyPathObservation(options)
}

func verifySourcePathObservation(ctx context.Context, filename string, source []byte, activity string,
	prior *PathObservationReceipt) error {
	spec, err := SourceAssembly(ctx, filename, source, activity)
	if err != nil {
		return err
	}
	declared := declaredPathObservation(spec)
	if declared != nil && (prior == nil || !reflect.DeepEqual(*declared, prior.Options)) {
		return fmt.Errorf("saved observation differs from the source observe declaration")
	}
	if prior != nil {
		return observationAvoidsHoldout(spec, &prior.Options)
	}
	return nil
}

func observationAvoidsHoldout(spec *assemblyspec.Spec, options *PathObservationOptions) error {
	if spec == nil || options == nil {
		return nil
	}
	for _, input := range options.Inputs {
		for _, held := range spec.HoldoutCases {
			if input == held.Input {
				return fmt.Errorf("path observation input %d is reserved by holdout_case", input)
			}
		}
	}
	return nil
}
