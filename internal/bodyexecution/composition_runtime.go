package bodyexecution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type CompositionDelivery struct {
	ActivityID   string                    `json:"activity_id"`
	ProducerID   string                    `json:"producer_id,omitempty"`
	Input        json.RawMessage           `json:"input,omitempty"`
	InputFields  []CompositionRecordField  `json:"input_fields,omitempty"`
	Inputs       []CompositionPortDelivery `json:"inputs,omitempty"`
	Actual       json.RawMessage           `json:"actual"`
	ActualFields []CompositionRecordField  `json:"actual_fields,omitempty"`
	Expected     json.RawMessage           `json:"expected,omitempty"`
	Passed       *bool                     `json:"passed,omitempty"`
}

type CompositionPortDelivery struct {
	Port       string                   `json:"port"`
	EntityID   string                   `json:"entity_id"`
	ProducerID string                   `json:"producer_id,omitempty"`
	Value      json.RawMessage          `json:"value"`
	Fields     []CompositionRecordField `json:"fields,omitempty"`
}

type CompositionTrace struct {
	CaseIndex  int                   `json:"case_index"`
	Deliveries []CompositionDelivery `json:"deliveries"`
}

type CompositionRuntime struct {
	Schema               string               `json:"schema"`
	Stage                string               `json:"stage"`
	Failure              string               `json:"failure,omitempty"`
	CompositionSHA256    string               `json:"composition_sha256"`
	OriginalSourceSHA256 string               `json:"original_source_sha256"`
	SelectedSourceSHA256 string               `json:"selected_source_sha256"`
	TypedPlanSHA256      string               `json:"typed_plan_sha256"`
	GeneratedSHA256      string               `json:"generated_sha256"`
	DriverSHA256         string               `json:"driver_sha256"`
	RuntimeSuiteSHA256   string               `json:"runtime_suite_sha256"`
	ExecutableSHA256     string               `json:"executable_sha256"`
	GoToolSHA256         string               `json:"go_tool_sha256"`
	GoToolSelection      string               `json:"go_tool_selection"`
	GoVersion            string               `json:"go_version"`
	ProducerSourceSHA    string               `json:"producer_source_sha"`
	Toolchain            ProcessObservation   `json:"toolchain"`
	Build                ProcessObservation   `json:"build"`
	Runs                 []ProcessObservation `json:"runs"`
	Traces               []CompositionTrace   `json:"traces"`
	ProjectionReplayed   bool                 `json:"projection_replayed"`
	RuntimeReplayed      bool                 `json:"runtime_replayed"`
	FinitePassed         int                  `json:"finite_passed"`
	FiniteTotal          int                  `json:"finite_total"`
	ModelCalls           int                  `json:"model_calls"`
	ElapsedNS            int64                `json:"elapsed_ns"`
	Scope                string               `json:"scope"`
}

// ExecuteComposition rebuilds a source-replayed graph and immediately runs it
// twice. Stage input/output traces are reconstructed from actual native outputs.
// Generation-time model observations are retained separately, never re-attested.
func ExecuteComposition(ctx context.Context, filename string, source []byte, prior Composition,
	suite CompositionCases, goBinary string) (CompositionRuntime, error) {
	start := time.Now()
	r := CompositionRuntime{Schema: "gooo/body-composition-runtime/v1", Stage: "GRAPH_REPLAY",
		CompositionSHA256: compositionDigest(prior), OriginalSourceSHA256: digest(source),
		SelectedSourceSHA256: prior.SelectedSourceSHA256, TypedPlanSHA256: prior.Plan.TypedPlanSHA256,
		GeneratedSHA256: prior.GeneratedSHA256, DriverSHA256: prior.DriverSHA256, RuntimeSuiteSHA256: compositionDigest(suite),
		ProducerSourceSHA: producerSourceSHA(), Runs: make([]ProcessObservation, 0, 2), Traces: []CompositionTrace{},
		Scope: "two fresh compiled value-graph executions; finite named expectations; explicit edges and ordered actual values; zero inference during replay/execution"}
	finish := func(err error) (CompositionRuntime, error) {
		r.ElapsedNS = time.Since(start).Nanoseconds()
		if err != nil {
			r.Failure = err.Error()
		}
		return r, err
	}
	if ctx == nil {
		return finish(fmt.Errorf("composition requires a context"))
	}
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	graph, err := replayComposition(ctx, filename, source, prior)
	if err != nil {
		return finish(err)
	}
	err = executeCompositionNative(ctx, graph, prior, suite, goBinary, &r)
	return finish(err)
}

func executeCompositionNative(ctx context.Context, graph compositionGraph, prior Composition,
	suite CompositionCases, goBinary string, r *CompositionRuntime) error {
	r.ProjectionReplayed, r.Stage = true, "RUNTIME_CASES"
	rows, err := graph.inputRows(suite)
	if err != nil {
		return err
	}
	for _, test := range suite.Cases {
		r.FiniteTotal += len(test.Expected)
	}
	goBinary, err = prepareCompositionTool(ctx, goBinary, r)
	if err != nil {
		return err
	}
	root, err := os.MkdirTemp("", "gooo-composition-runtime-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(root)
	executable, err := buildCompositionExecutable(ctx, root, prior.Source, prior.Driver, goBinary, r)
	if err != nil {
		return err
	}
	return observeCompositionRuns(ctx, root, executable, rows, graph, suite, r)
}

func prepareCompositionTool(ctx context.Context, requested string, r *CompositionRuntime) (string, error) {
	native := &Observation{ProducerSourceSHA: r.ProducerSourceSHA}
	goBinary, selection, err := selectGoTool(requested)
	r.GoToolSelection = selection
	if err != nil {
		return "", err
	}
	native.GoToolSHA256, err = fileDigest(goBinary)
	if err != nil {
		return "", err
	}
	r.GoToolSHA256, r.Stage = native.GoToolSHA256, "TOOLCHAIN"
	err = observeToolchain(ctx, goBinary, native, nil)
	r.Toolchain, r.GoVersion = native.Toolchain, native.GoVersion
	return goBinary, err
}

func buildCompositionExecutable(ctx context.Context, root, projection, driver string,
	goBinary string, r *CompositionRuntime) (string, error) {
	for name, text := range map[string]string{"go.mod": "module gooo.observed.composition\n\ngo 1.27.1\n",
		"generated.go": projection, "main.go": driver} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(text), 0600); err != nil {
			return "", err
		}
	}
	executable := filepath.Join(root, "observed-composition")
	if runtime.GOOS == "windows" {
		executable += ".exe"
	}
	r.Stage = "BUILD"
	var err error
	_, r.Build, err = process(ctx, root, goBinary, nil, "build", "-trimpath", "-buildvcs=false", "-o", executable, ".")
	if err != nil {
		return "", err
	}
	r.ExecutableSHA256, err = fileDigest(executable)
	return executable, err
}

func observeCompositionRuns(ctx context.Context, root, executable string, rows [][]json.RawMessage,
	graph compositionGraph, suite CompositionCases, r *CompositionRuntime) error {
	input, _ := json.Marshal(rows)
	var first []byte
	for run := range 2 {
		r.Stage = fmt.Sprintf("EXECUTE_%d", run+1)
		output, observation, runErr := runCompositionProcess(ctx, root, executable, input)
		r.Runs = append(r.Runs, observation)
		if runErr != nil {
			return runErr
		}
		if run == 0 {
			first = output
			traces, passed, err := graph.nativeTraces(output, suite)
			if err != nil {
				return err
			}
			r.Traces, r.FinitePassed = traces, passed
		} else if !bytes.Equal(first, output) {
			return fmt.Errorf("compiled composition replay outputs differ")
		}
	}
	r.RuntimeReplayed, r.Stage = true, "COMPLETE"
	return nil
}

func runCompositionProcess(ctx context.Context, root, executable string, input []byte) ([]byte, ProcessObservation, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return process(ctx, root, executable, input)
}

func (graph compositionGraph) nativeTraces(output []byte, suite CompositionCases) ([]CompositionTrace, int, error) {
	var rows [][]json.RawMessage
	if err := json.Unmarshal(output, &rows); err != nil || len(rows) != len(suite.Cases) {
		return nil, 0, fmt.Errorf("compiled composition output case count differs")
	}
	traces := make([]CompositionTrace, len(rows))
	passed := 0
	for c, row := range rows {
		if len(row) != graph.count {
			return nil, 0, fmt.Errorf("compiled composition output activity count differs")
		}
		trace := CompositionTrace{CaseIndex: c, Deliveries: make([]CompositionDelivery, graph.count)}
		for i, node := range graph.nodes[:graph.count] {
			actual, err := graph.canonicalValue(row[i], node.OutputType)
			if err != nil {
				return nil, 0, fmt.Errorf("activity %q runtime output: %w", node.Name, err)
			}
			entry := CompositionDelivery{ActivityID: node.ID, Actual: actual,
				ActualFields: graph.recordFieldValues(node.OutputType, actual)}
			if len(node.Inputs) > 0 {
				entry.Inputs = make([]CompositionPortDelivery, len(node.Inputs))
			}
			for p, slot := range node.inputSlots() {
				input := suite.Cases[c].Inputs[node.inputKey(slot.Port)]
				producer := ""
				if slot.From >= 0 {
					input, producer = trace.Deliveries[slot.From].Actual, graph.nodes[slot.From].ID
				}
				input, err = graph.canonicalValue(input, slot.Type)
				if err != nil {
					return nil, 0, fmt.Errorf("runtime input %q: %w", node.inputKey(slot.Port), err)
				}
				if len(node.Inputs) == 0 {
					entry.Input, entry.ProducerID = input, producer
					entry.InputFields = graph.recordFieldValues(slot.Type, input)
				} else {
					entry.Inputs[p] = CompositionPortDelivery{Port: slot.Port, EntityID: slot.EntityID,
						ProducerID: producer, Value: input, Fields: graph.recordFieldValues(slot.Type, input)}
				}
			}
			if expected, present := suite.Cases[c].Expected[node.Name]; present {
				entry.Expected, _ = graph.canonicalValue(expected, node.OutputType)
				match := bytes.Equal(entry.Expected, actual)
				entry.Passed = &match
				if match {
					passed++
				}
			}
			trace.Deliveries[i] = entry
		}
		traces[c] = trace
	}
	return traces, passed, nil
}
