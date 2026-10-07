package bodyexecution

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type CompositionStep struct {
	InputSourceSHA256 string             `json:"input_source_sha256"`
	Generation        bodycodegen.Result `json:"generation"`
}

type Composition struct {
	Schema               string                         `json:"schema"`
	Stage                string                         `json:"stage"`
	Failure              string                         `json:"failure,omitempty"`
	ActiveActivity       string                         `json:"active_activity,omitempty"`
	OriginalSourceSHA256 string                         `json:"original_source_sha256"`
	SelectedSourceSHA256 string                         `json:"selected_source_sha256"`
	GeneratedSHA256      string                         `json:"generated_sha256"`
	DriverSHA256         string                         `json:"driver_sha256"`
	Plan                 CompositionPlan                `json:"plan"`
	Steps                []CompositionStep              `json:"steps"`
	GoooSource           string                         `json:"gooo_source"`
	Source               string                         `json:"source"`
	Driver               string                         `json:"driver_source"`
	Model                *bodycodegen.RetainedModelInfo `json:"model,omitempty"`
	FillModel            *CompositionFillModelInfo      `json:"fill_model,omitempty"`
	ElapsedNS            int64                          `json:"elapsed_ns"`
	Scope                string                         `json:"scope"`
	Continuation         *CompositionContinuation       `json:"continuation,omitempty"`
}

// GenerateComposition constructs each activity in typed plan order, retaining
// one optional model across assembly activities. Selected source checkpoints
// become the next step's exact input. No native execution happens in this phase.
func GenerateComposition(ctx context.Context, filename string, source []byte,
	suite CompositionCases, modelPath string) (Composition, error) {
	return GenerateCompositionWithOptions(ctx, filename, source, suite, CompositionOptions{ModelPath: modelPath})
}

// GenerateCompositionWithOptions retains separate optional models for structural
// choices and source_fill assignments. Empty paths use deterministic selection.
func GenerateCompositionWithOptions(ctx context.Context, filename string, source []byte,
	suite CompositionCases, options CompositionOptions) (Composition, error) {
	start := time.Now()
	graph, err := prepareCompositionGraph(ctx, filename, source)
	result := Composition{Schema: "gooo/body-composition/v1", Stage: "PLAN", OriginalSourceSHA256: digest(source),
		Plan: graph.plan, Steps: make([]CompositionStep, 0, graph.count),
		Scope: "source-declared typed value graph; finite activity selection and independently compiled graph execution are separate observations"}
	finish := func(err error) (Composition, error) {
		result.ElapsedNS = time.Since(start).Nanoseconds()
		if err != nil {
			result.Failure = err.Error()
		}
		return result, err
	}
	if err != nil {
		return finish(err)
	}
	if (options.ModelPath != "" || options.FillModelPath != "") && !graph.hasAssembly() {
		return finish(fmt.Errorf("a composition model requires an assembling activity"))
	}
	if _, err := graph.inputRows(suite); err != nil {
		return finish(err)
	}
	result.Stage = "BODY_PREFLIGHT"
	if err := graph.preflight(ctx, filename, source); err != nil {
		return finish(err)
	}
	if err := graph.validateModelRoute(ctx, filename, source, options); err != nil {
		return finish(err)
	}
	result.Stage = "ACTIVITY_GENERATION"
	current, err := generateCompositionSteps(ctx, filename, source, graph, options, &result)
	if err != nil {
		return finish(err)
	}
	result.GoooSource, result.SelectedSourceSHA256 = string(current), digest(current)
	result.Stage, result.ActiveActivity = "GRAPH_EMISSION", ""
	result.Source, err = emitCompositionProjection(result.Steps, graph)
	if err != nil {
		return finish(err)
	}
	result.GeneratedSHA256 = digest([]byte(result.Source))
	result.Driver, err = compositionDriver(graph)
	if err != nil {
		return finish(err)
	}
	result.DriverSHA256 = digest([]byte(result.Driver))
	result.Stage = "COMPLETE"
	return finish(nil)
}

func generateCompositionSteps(ctx context.Context, filename string, source []byte,
	graph compositionGraph, options CompositionOptions, result *Composition) ([]byte, error) {
	generator := compositionAssemblyGenerator{options: options}
	current := append([]byte(nil), source...)
	for _, node := range graph.nodes[:graph.count] {
		result.ActiveActivity = node.Name
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var generation bodycodegen.Result
		var err error
		if node.Assembling {
			generation, err = generator.generate(ctx, filename, current, node.Name)
			result.FillModel = generator.fillInfo
			if generator.retained != nil {
				info := generator.retained.Info()
				result.Model = &info
			}
		} else {
			generation, err = bodycodegen.GenerateWithPlanner(ctx, filename, current, node.Name, "", "")
		}
		if err != nil {
			return nil, fmt.Errorf("activity %q generation: %w", node.Name, err)
		}
		result.Steps = append(result.Steps, CompositionStep{InputSourceSHA256: digest(current), Generation: generation})
		if node.Assembling {
			realized, err := bodycodegen.RealizeSourceAssembly(ctx, filename, current, generation)
			if err != nil {
				return nil, fmt.Errorf("activity %q checkpoint: %w", node.Name, err)
			}
			current = []byte(realized.Source)
		}
	}
	return current, nil
}

func (graph compositionGraph) hasAssembly() bool {
	for _, node := range graph.nodes[:graph.count] {
		if node.Assembling {
			return true
		}
	}
	return false
}

func DecodeComposition(raw []byte) (Composition, error) {
	var result Composition
	if err := decode(raw, &result, 32<<20); err != nil {
		return result, err
	}
	return result, nil
}

// VerifyComposition reconstructs all selected bodies and the exact graph Go
// projection without loading a model or repeating candidate selection.
func VerifyComposition(ctx context.Context, filename string, source []byte, prior Composition) error {
	_, err := replayComposition(ctx, filename, source, prior)
	return err
}

func replayComposition(ctx context.Context, filename string, source []byte, prior Composition) (compositionGraph, error) {
	graph, err := prepareCompositionGraph(ctx, filename, source)
	if err != nil {
		return graph, err
	}
	if prior.Schema != "gooo/body-composition/v1" || prior.Stage != "COMPLETE" || prior.Failure != "" ||
		prior.OriginalSourceSHA256 != digest(source) ||
		!sameCompositionPlan(prior.Plan, graph.plan) || len(prior.Steps) != graph.count {
		return graph, fmt.Errorf("composition original source, typed plan or step order differs")
	}
	current := append([]byte(nil), source...)
	for i, node := range graph.nodes[:graph.count] {
		step := prior.Steps[i]
		if step.InputSourceSHA256 != digest(current) || step.Generation.Report.Activity != node.Name {
			return graph, fmt.Errorf("composition step %d input or activity order differs", i)
		}
		if node.Assembling {
			realized, err := bodycodegen.RealizeSourceAssembly(ctx, filename, current, step.Generation)
			if err != nil {
				return graph, fmt.Errorf("composition step %d replay: %w", i, err)
			}
			current = []byte(realized.Source)
		} else if err := replayCompositionPlain(ctx, filename, current, node, step.Generation); err != nil {
			return graph, fmt.Errorf("composition step %d replay: %w", i, err)
		}
	}
	generated, err := emitCompositionProjection(prior.Steps, graph)
	if err != nil || generated != prior.Source || prior.GeneratedSHA256 != digest([]byte(generated)) ||
		prior.GoooSource != string(current) || prior.SelectedSourceSHA256 != digest(current) {
		return graph, fmt.Errorf("composition selected source or generated graph does not replay")
	}
	driver, err := compositionDriver(graph)
	if err != nil || prior.Driver != driver || prior.DriverSHA256 != digest([]byte(driver)) {
		return graph, fmt.Errorf("composition generated input delivery driver does not replay")
	}
	if err := verifyCompositionContinuation(prior); err != nil {
		return graph, err
	}
	return graph, ctx.Err()
}

func replayCompositionPlain(ctx context.Context, filename string, source []byte,
	node CompositionActivity, prior bodycodegen.Result) error {
	replayed, err := bodycodegen.GenerateWithPlanner(ctx, filename, source, node.Name, "", "")
	if err != nil {
		return err
	}
	r, expected := prior.Report, replayed.Report
	if prior.Source != replayed.Source || prior.GoooSource != "" || r.BodyPaths != nil || r.BodyFill != nil ||
		r.BodySearch != nil || r.RecordAssembly != nil || r.Route != "preserve" || r.Decision != "PASS" || !r.TypecheckPassed ||
		!r.DeterministicReplay || r.ActivityID != node.ID || r.SourceDigest != expected.SourceDigest ||
		r.GeneratedDigest != expected.GeneratedDigest || r.ProgramDigest != expected.ProgramDigest ||
		r.InputType != expected.InputType || r.OutputType != expected.OutputType ||
		!reflect.DeepEqual(r.InputParameters, expected.InputParameters) || !reflect.DeepEqual(r.RecordTypes, expected.RecordTypes) {
		return fmt.Errorf("ordinary body projection differs from its source")
	}
	return nil
}

func compositionDigest(value any) string { raw, _ := json.Marshal(value); return digest(raw) }
