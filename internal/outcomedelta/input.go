// Package outcomedelta compares saved outcomes across changing source contracts.
package outcomedelta

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

const MaxInputBytes = 32 << 20

//go:generate go run ../../scripts/receipt-schema --source delta.gooo --root OutcomeDelta --go delta.generated.go --json delta.schema.json
//go:embed delta.gooo
var declaration []byte

func digest(b []byte) string { return fmt.Sprintf("sha256:%x", sha256.Sum256(b)) }

// Only the latest evaluation is selected. Saved construction attempts and runtime
// history are context, not extra observations or new model calls.
func runtimeInput(raw []byte) (bodyexecution.CompositionRuntime, error) {
	var envelope map[string]json.RawMessage
	if err := bodyexecution.DecodeExecutionReceipt(raw, &envelope); err != nil {
		return bodyexecution.CompositionRuntime{}, err
	}
	if envelope["evaluation"] != nil && envelope["runtime"] != nil {
		return bodyexecution.CompositionRuntime{}, fmt.Errorf("ambiguous evaluation and runtime")
	}
	var runtime bodyexecution.CompositionRuntime
	var err error
	switch {
	case envelope["evaluation"] != nil:
		var wrapper struct {
			Generated    bool                            `json:"generated_now"`
			Construction bodyexecution.JointConstruction `json:"construction"`
			Evaluation   bodyexecution.JointEvaluation   `json:"evaluation"`
		}
		err = bodyexecution.DecodeExecutionReceipt(raw, &wrapper)
		runtime = wrapper.Evaluation.Runtime
	case envelope["runtime"] != nil && envelope["composition"] != nil:
		var wrapper struct {
			InputSchema string                               `json:"input_schema"`
			Generated   bool                                 `json:"generated_now"`
			Composition bodyexecution.Composition            `json:"composition"`
			Runtime     bodyexecution.CompositionRuntime     `json:"runtime"`
			History     []bodyexecution.CompositionRuntime   `json:"runtime_history,omitempty"`
			Series      *bodyexecution.CompositionCaseSeries `json:"case_series,omitempty"`
		}
		err = bodyexecution.DecodeExecutionReceipt(raw, &wrapper)
		runtime = wrapper.Runtime
	case envelope["runtime"] != nil:
		var evaluation bodyexecution.JointEvaluation
		err = bodyexecution.DecodeExecutionReceipt(raw, &evaluation)
		runtime = evaluation.Runtime
	default:
		err = bodyexecution.DecodeExecutionReceipt(raw, &runtime)
	}
	if err != nil {
		return runtime, err
	}
	switch runtime.Schema {
	case "gooo/body-composition-runtime/v1", "gooo/body-composition-runtime/v2", "gooo/body-composition-runtime/v3":
	default:
		return runtime, fmt.Errorf("expected a supported composition runtime or its command/evaluation envelope")
	}
	if runtime.Stage == "" {
		return runtime, fmt.Errorf("runtime stage is missing")
	}
	return runtime, nil
}

// Canonicalize JSON object order without converting integers through float64.
// These execution contracts carry signed integers, strings, booleans and records.
func canonical(raw []byte) (any, string, error) {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, "", err
	}
	value, err := integerValues(value)
	if err != nil {
		return nil, "", err
	}
	b, err := json.Marshal(value)
	return value, string(b), err
}

func integerValues(value any) (any, error) {
	switch v := value.(type) {
	case json.Number:
		n, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("runtime value requires an exact int64: %w", err)
		}
		return n, nil
	case []any:
		for i := range v {
			n, err := integerValues(v[i])
			if err != nil {
				return nil, err
			}
			v[i] = n
		}
	case map[string]any:
		for k := range v {
			n, err := integerValues(v[k])
			if err != nil {
				return nil, err
			}
			v[k] = n
		}
	}
	return value, nil
}

func runtimeSummary(r bodyexecution.CompositionRuntime) map[string]any {
	return map[string]any{"schema": r.Schema, "stage": r.Stage, "failure": r.Failure,
		"producer_source_sha": r.ProducerSourceSHA, "selected_source_sha256": r.SelectedSourceSHA256,
		"runtime_suite_sha256": r.RuntimeSuiteSHA256, "observed_trace_count": len(r.Traces),
		"reported_finite_passed": r.FinitePassed, "reported_finite_total": r.FiniteTotal}
}
