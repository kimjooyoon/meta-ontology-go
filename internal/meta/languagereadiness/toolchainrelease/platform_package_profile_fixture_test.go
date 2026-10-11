package toolchainrelease

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

type computationFixture struct {
	Receipt         map[string]any
	Manifest, Cases []byte
	Program         workspaceexecution.Program
	Profile         packageProfile
}

// Synthetic receipts test the reader; they do not report native execution.
func newComputationFixture(t *testing.T, profile packageProfile) computationFixture {
	t.Helper()
	root := filepath.Join("../../../..", profile.Root)
	read := func(name string) []byte {
		raw, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	f := computationFixture{Manifest: read("gooo.workspace.json"), Cases: read("cases.json"), Profile: profile}
	var err error
	f.Program, err = computationPackageProgram(root, f.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := computationPackageReference(f.Cases, profile)
	if err != nil {
		t.Fatal(err)
	}
	key, traces := profile.Package+":"+profile.Activity, []any{}
	for i, row := range rows {
		traces = append(traces, map[string]any{"case_index": i, "deliveries": []any{map[string]any{
			"activity_id": profile.ID, "input": row.Inputs[key], "actual": row.Expected[key],
			"expected": row.Expected[key], "passed": true}}})
	}
	program, _ := json.Marshal(f.Program)
	var p map[string]any
	if err := json.Unmarshal(program, &p); err != nil {
		t.Fatal(err)
	}
	sha := "sha256:" + strings.Repeat("a", 64)
	f.Receipt = map[string]any{"schema": "gooo/workspace-body-execution-receipt/v1", "decision": "PASS",
		"manifest_digest": computationDigest(f.Manifest), "cases_digest": computationDigest(f.Cases),
		"result": map[string]any{"program": p, "execution_source_sha256": computationDigest([]byte(f.Program.Source)),
			"composition": map[string]any{"generated_sha256": sha}, "runtime": map[string]any{
				"stage": "COMPLETE", "generated_sha256": sha, "model_calls": 0, "finite_passed": profile.Total,
				"finite_total": profile.Total, "projection_replayed": true, "runtime_replayed": true, "traces": traces}}}
	if profile.Name == "single-condition-byte" {
		addComputationFillFixture(f)
	}
	return f
}

func (f computationFixture) validate(replay bool, saved boundedTextSmokeBinding) (boundedTextSmokeBinding, error) {
	raw, err := json.Marshal(f.Receipt)
	if err != nil {
		return boundedTextSmokeBinding{}, err
	}
	return validateComputationPackage(raw, f.Manifest, f.Cases, f.Program, f.Profile, replay, saved)
}
