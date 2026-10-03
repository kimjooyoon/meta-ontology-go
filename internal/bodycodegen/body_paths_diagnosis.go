package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

type PathDiagnosisOptions struct {
	Inputs        []int64 `json:"inputs"`
	MaxCandidates int     `json:"max_candidates"`
}

func copyPathDiagnosis(options *PathDiagnosisOptions) (*PathDiagnosisOptions, error) {
	if options == nil {
		return nil, nil
	}
	if len(options.Inputs) == 0 || len(options.Inputs) > 32 || options.MaxCandidates < 1 || options.MaxCandidates > 64 {
		return nil, fmt.Errorf("path diagnosis requires 1..32 inputs and 1..64 candidates")
	}
	owned := *options
	owned.Inputs = append([]int64(nil), options.Inputs...)
	return &owned, nil
}

func resolveTypedPathOptions(options TypedPathOptions, hasModel bool) (*typedPathFeedback, *PathDiagnosisOptions, error) {
	if options.StepAttempts < 0 || options.StepAttempts > 64 || options.FeedbackRounds < 0 || options.FeedbackRounds > 16 {
		return nil, nil, fmt.Errorf("step must be 0..64 and feedback rounds 0..16")
	}
	if options.FeedbackRounds == 0 && (options.FeedbackUnfixed || options.CI != nil) {
		return nil, nil, fmt.Errorf("feedback options require explicit feedback rounds")
	}
	diagnosis, err := copyPathDiagnosis(options.Diagnosis)
	if err != nil {
		return nil, nil, err
	}
	var feedback *typedPathFeedback
	if options.FeedbackRounds != 0 {
		if !hasModel || options.StepAttempts == 0 {
			return nil, nil, fmt.Errorf("feedback requires a local model and 1..64 step attempts")
		}
		if err := options.CI.Validate(); err != nil {
			return nil, nil, err
		}
		feedback = &typedPathFeedback{rounds: options.FeedbackRounds, ci: options.CI, unfixed: options.FeedbackUnfixed}
	}
	return feedback, diagnosis, nil
}

// GenerateWithTypedPathOptions keeps model loading after source binding and
// adds an optional deterministic diagnosis after selection. The probe outputs
// do not authorize source edits or become expectations. Callers synchronize input.
func GenerateWithTypedPathOptions(ctx context.Context, filename string, source []byte, activityName string,
	document pathplan.Document, modelPath string, options TypedPathOptions) (Result, error) {
	feedback, diagnosis, err := resolveTypedPathOptions(options, modelPath != "")
	if err != nil {
		return Result{}, err
	}
	observation, err := copyPathObservation(options.Observation)
	if err != nil {
		return Result{}, err
	}
	return generateTypedPathRequest(ctx, filename, source, activityName, document,
		typedPathModel{path: modelPath, diagnosis: diagnosis, observation: observation}, options.StepAttempts, feedback)
}

func diagnoseSelectedPath(ctx context.Context, prepared *pathplan.PreparedPlan, document pathplan.Document,
	options *PathDiagnosisOptions, receipt *BodyPathReceipt) error {
	if options == nil {
		return nil
	}
	started := time.Now()
	raw, _ := json.Marshal(options)
	receipt.DiagnosisOptionsSHA256 = digest(raw)
	receipt.DiagnosisBudget = options.MaxCandidates
	value, err := prepared.Diagnose(ctx, bodyPathSelection(receipt).Choices, document.TestCases,
		options.Inputs, options.MaxCandidates)
	receipt.Timing.DiagnosisMS = elapsedMS(started)
	if value.Schema != "" {
		receipt.Diagnosis = &value
	}
	receipt.DiagnosisScope = "typed arena candidate outputs; probe expectations unset; alternatives not native-executed"
	return err
}
