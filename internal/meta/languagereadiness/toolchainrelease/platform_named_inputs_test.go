package toolchainrelease

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Synthetic envelopes check the reader, separately from native platform runs.
func newNamedInputFixture(t *testing.T) computationFixture {
	t.Helper()
	root := filepath.Join("../../../..", namedInputsRoot)
	read := func(name string) []byte {
		raw, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	f := computationFixture{Manifest: read("workspace.json"), Cases: read("cases.json")}
	var err error
	f.Program, err = computationPackageProgram(root, f.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(f.Program)
	if err != nil {
		t.Fatal(err)
	}
	var program map[string]any
	if err := json.Unmarshal(encoded, &program); err != nil {
		t.Fatal(err)
	}
	traces := []any{}
	for i, row := range namedInputExamples {
		ports := []any{
			map[string]any{"port": "input0", "entity_id": "urn:gooo:type:integer", "producer_id": namedCreditID, "value": row.credit},
			map[string]any{"port": "input1", "entity_id": "urn:gooo:type:integer", "value": row.debit},
		}
		traces = append(traces, map[string]any{"case_index": i, "deliveries": []any{
			map[string]any{"activity_id": namedCreditID, "input": row.input, "actual": row.credit, "expected": row.credit, "passed": true},
			map[string]any{"activity_id": namedNetID, "inputs": ports, "actual": row.net, "expected": row.net, "passed": true},
		}})
	}
	generated := "sha256:" + strings.Repeat("a", 64)
	f.Receipt = map[string]any{"schema": "gooo/workspace-body-execution-receipt/v1", "decision": "PASS",
		"manifest_digest": computationDigest(f.Manifest), "cases_digest": computationDigest(f.Cases),
		"result": map[string]any{"program": program, "execution_source_sha256": computationDigest([]byte(f.Program.Source)),
			"composition": map[string]any{"generated_sha256": generated}, "runtime": map[string]any{
				"stage": "COMPLETE", "generated_sha256": generated, "finite_passed": 4, "finite_total": 4,
				"model_calls": 0, "projection_replayed": true, "runtime_replayed": true, "traces": traces}}}
	return f
}

func validateNamedFixture(f computationFixture, replay bool, saved boundedTextSmokeBinding) (boundedTextSmokeBinding, error) {
	raw, err := json.Marshal(f.Receipt)
	if err != nil {
		return boundedTextSmokeBinding{}, err
	}
	return validateNamedInputs(raw, f.Manifest, f.Cases, f.Program, replay, saved)
}

func TestNamedInputsReleaseFourOutputsAndBoundSavedReplay(t *testing.T) {
	f := newNamedInputFixture(t)
	saved, err := validateNamedFixture(f, false, boundedTextSmokeBinding{})
	if err != nil {
		t.Fatal(err)
	}
	f.Receipt["replayed_from_sha256"] = saved.Receipt
	f.Receipt["result"].(map[string]any)["replay"] = map[string]any{"model_calls": 0}
	if _, err := validateNamedFixture(f, true, saved); err != nil {
		t.Fatal(err)
	}
	for _, wrong := range []boundedTextSmokeBinding{{}, {Program: saved.Program, Receipt: "other"},
		{Program: "other", Receipt: saved.Receipt}} {
		if _, err := validateNamedFixture(f, true, wrong); err == nil {
			t.Fatal("unbound named-input replay accepted")
		}
	}
	if _, err := validateNamedFixture(f, false, boundedTextSmokeBinding{}); err == nil {
		t.Fatal("a replay was accepted as the first source execution")
	}
}

func TestNamedInputsReleaseRefusesWrongOriginsCountsValuesAndPorts(t *testing.T) {
	mutations := []computationMutation{
		{"schema", "other", false}, {"decision", "OBSERVED", false}, {"error", "failed", false},
		{"manifest_digest", "other", false}, {"cases_digest", "other", false},
		{"result.program.entry.activity", "Credit", false},
		{"result.program.lowered_gooo_source", "other", false},
		{"result.execution_source_sha256", "other", false},
		{"result.runtime.stage", "FAILED", false}, {"result.runtime.finite_total", 2, false},
		{"result.runtime.finite_passed", nil, true}, {"result.runtime.model_calls", 1, false},
		{"result.runtime.projection_replayed", nil, true}, {"result.runtime.runtime_replayed", false, false},
		{"result.runtime.traces.0.case_index", nil, true}, {"result.runtime.traces.0.case_index", 1, false},
		{"result.runtime.traces.0.deliveries.0.activity_id", namedNetID, false},
		{"result.runtime.traces.0.deliveries.0.input", nil, true},
		{"result.runtime.traces.0.deliveries.0.actual", "124", false},
		{"result.runtime.traces.0.deliveries.0.expected", nil, true},
		{"result.runtime.traces.0.deliveries.0.passed", false, false},
		{"result.runtime.traces.0.deliveries.0.fault", map[string]any{"kind": "fault"}, false},
		{"result.runtime.traces.0.deliveries.1.blocked_by", []string{namedCreditID}, false},
		{"result.runtime.traces.0.deliveries.1.inputs.0.port", "input1", false},
		{"result.runtime.traces.0.deliveries.1.inputs.0.entity_id", "other", false},
		{"result.runtime.traces.0.deliveries.1.inputs.0.producer_id", "other", false},
		{"result.runtime.traces.0.deliveries.1.inputs.0.value", 101, false},
		{"result.runtime.traces.0.deliveries.1.inputs.1.value", nil, true},
	}
	for _, mutation := range mutations {
		t.Run(mutation.path, func(t *testing.T) {
			f := newNamedInputFixture(t)
			mutateComputationReceipt(t, f, mutation)
			if _, err := validateNamedFixture(f, false, boundedTextSmokeBinding{}); err == nil {
				t.Fatal("inconsistent named-input synthetic receipt accepted")
			}
		})
	}
}

func TestNativeNamedInputsReleaseProfile(t *testing.T) {
	binary := os.Getenv("GOOO_PACKAGE_SMOKE_BINARY")
	if binary == "" {
		t.Skip("set GOOO_PACKAGE_SMOKE_BINARY for an actual packaged compiler")
	}
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	output := os.Getenv("GOOO_NAMED_INPUTS_SMOKE_OUTPUT")
	if output == "" {
		output = t.TempDir()
	}
	if err := os.MkdirAll(output, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := smokeNamedInputs(binary, BuildInput{Root: root, OutputDir: output, Target: Target{ID: "host"}}); err != nil {
		t.Fatal(err)
	}
}
