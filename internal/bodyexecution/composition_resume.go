package bodyexecution

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type CompositionContinuation struct {
	ParentCompositionSHA256 string                   `json:"parent_composition_sha256"`
	Activities              []RecordContinuationStep `json:"activities"`
	NewModelCalls           int                      `json:"new_model_calls"`
}

type RecordContinuationStep struct {
	Activity          string `json:"activity"`
	RetainedAttempts  int    `json:"retained_attempts"`
	AddedAttempts     int    `json:"added_attempts"`
	RecheckedAttempts int    `json:"rechecked_attempts,omitempty"`
}

// ResumeComposition verifies the saved graph, resumes record-choice activities
// under an explicit Gooo policy, and projects their updated typed graph. Plain
// activities are regenerated from their source. No model is loaded or called.
func ResumeComposition(ctx context.Context, filename string, source []byte, prior Composition,
	suite CompositionCases, policy bodycodegen.RecordAssemblyPolicy) (Composition, error) {
	started := time.Now()
	graph, err := replayComposition(ctx, filename, source, prior)
	if err != nil {
		return Composition{}, fmt.Errorf("resume composition: %w", err)
	}
	if err = bodycodegen.ValidateRecordAssemblyPolicy(ctx, policy); err != nil {
		return Composition{}, err
	}
	if _, err = graph.inputRows(suite); err != nil {
		return Composition{}, err
	}
	found := false
	for _, step := range prior.Preparations {
		if step.Generation.Report.RecordAssembly == nil {
			return Composition{}, fmt.Errorf("resume composition currently requires record-choice assembly at %s", step.Generation.Report.Activity)
		}
		found = true
	}
	for i, node := range graph.nodes[:graph.count] {
		if node.Assembling && !node.Prepared {
			if prior.Steps[i].Generation.Report.RecordAssembly == nil {
				return Composition{}, fmt.Errorf("resume composition currently requires record-choice assembly at %s", node.Name)
			}
			found = true
		}
	}
	if !found {
		return Composition{}, fmt.Errorf("resume composition requires a record-choice activity")
	}
	result := Composition{Schema: "gooo/body-composition/v1", Stage: "ACTIVITY_GENERATION",
		OriginalSourceSHA256: digest(source), Plan: graph.plan,
		Continuation: &CompositionContinuation{ParentCompositionSHA256: compositionDigest(prior)},
		Scope:        "continued source-owned record choices; original ranking and cumulative attempt budget; prior observations reconstructed; zero new model calls"}
	current, err := resumeCompositionSteps(ctx, filename, source, prior, graph, policy, &result)
	if err == nil {
		err = finishResumedComposition(current, graph, &result)
	}
	result.ElapsedNS = time.Since(started).Nanoseconds()
	if err != nil {
		result.Failure = err.Error()
	}
	return result, err
}

func resumeCompositionSteps(ctx context.Context, filename string, source []byte, prior Composition,
	graph compositionGraph, policy bodycodegen.RecordAssemblyPolicy, result *Composition) ([]byte, error) {
	oldCurrent, current := source, source
	for i, helper := range graph.plan.Preparations {
		result.ActiveActivity = helper.Name
		generation, err := bodycodegen.ResumeRecordAssembly(ctx, filename, oldCurrent, current, prior.Preparations[i].Generation, policy)
		if err != nil {
			return current, fmt.Errorf("resume called activity %s: %w", helper.Name, err)
		}
		result.Preparations = append(result.Preparations, CompositionStep{InputSourceSHA256: digest(current), Generation: generation})
		realized, err := bodycodegen.RealizeCalledAssembly(ctx, filename, current, generation)
		if err != nil {
			return current, err
		}
		oldRealized, err := bodycodegen.RealizeCalledAssembly(ctx, filename, oldCurrent, prior.Preparations[i].Generation)
		if err != nil {
			return current, err
		}
		current, oldCurrent = []byte(realized.Source), []byte(oldRealized.Source)
		appendRecordContinuation(result, generation)
	}
	for i, node := range graph.nodes[:graph.count] {
		result.ActiveActivity = node.Name
		var generation bodycodegen.Result
		var err error
		if node.Assembling && !node.Prepared {
			generation, err = bodycodegen.ResumeRecordAssembly(ctx, filename, oldCurrent, current, prior.Steps[i].Generation, policy)
		} else {
			generation, err = bodycodegen.GenerateWithPlanner(ctx, filename, current, node.Name, "", "")
		}
		if err != nil {
			return current, fmt.Errorf("resume activity %s: %w", node.Name, err)
		}
		result.Steps = append(result.Steps, CompositionStep{InputSourceSHA256: digest(current), Generation: generation})
		if node.Assembling && !node.Prepared {
			realized, err := bodycodegen.RealizeSourceAssembly(ctx, filename, current, generation)
			if err != nil {
				return current, err
			}
			current, oldCurrent = []byte(realized.Source), []byte(prior.Steps[i].Generation.GoooSource)
			appendRecordContinuation(result, generation)
		}
	}
	return current, nil
}

func appendRecordContinuation(result *Composition, generation bodycodegen.Result) {
	counts := generation.Report.RecordAssembly.Continuation
	result.Continuation.Activities = append(result.Continuation.Activities, RecordContinuationStep{
		Activity: generation.Report.Activity, RetainedAttempts: counts.RetainedAttempts,
		RecheckedAttempts: counts.RecheckedAttempts, AddedAttempts: counts.AddedAttempts})
}

func finishResumedComposition(source []byte, graph compositionGraph, result *Composition) error {
	result.GoooSource, result.SelectedSourceSHA256 = string(source), digest(source)
	result.Stage, result.ActiveActivity = "GRAPH_EMISSION", ""
	var err error
	result.Source, err = emitCompositionProjection(result.Steps, graph)
	if err != nil {
		return err
	}
	result.GeneratedSHA256 = digest([]byte(result.Source))
	result.Driver, err = compositionDriver(graph)
	if err != nil {
		return err
	}
	result.DriverSHA256 = digest([]byte(result.Driver))
	result.Stage = "COMPLETE"
	return nil
}

func verifyCompositionContinuation(prior Composition) error {
	var expected []RecordContinuationStep
	for _, step := range prior.ConstructionSteps() {
		if record := step.Generation.Report.RecordAssembly; record != nil && record.Continuation != nil {
			expected = append(expected, RecordContinuationStep{Activity: step.Generation.Report.Activity,
				RetainedAttempts: record.Continuation.RetainedAttempts, RecheckedAttempts: record.Continuation.RecheckedAttempts,
				AddedAttempts: record.Continuation.AddedAttempts})
		}
	}
	if prior.Continuation == nil {
		if len(expected) != 0 {
			return fmt.Errorf("composition continuation observation is missing")
		}
		return nil
	}
	if len(expected) == 0 || prior.Continuation.NewModelCalls != 0 ||
		!reflect.DeepEqual(prior.Continuation.Activities, expected) || !validCompositionParentDigest(prior.Continuation.ParentCompositionSHA256) {
		return fmt.Errorf("composition continuation counts or parent reference differs")
	}
	return nil
}

func validCompositionParentDigest(value string) bool {
	if len(value) != 71 || value[:7] != "sha256:" {
		return false
	}
	for _, c := range value[7:] {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}
