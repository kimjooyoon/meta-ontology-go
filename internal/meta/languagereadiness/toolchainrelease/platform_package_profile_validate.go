package toolchainrelease

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

func validateComputationPackage(raw, manifest, cases []byte, program workspaceexecution.Program, profile packageProfile, replay bool,
	selected boundedTextSmokeBinding) (boundedTextSmokeBinding, error) {
	rows, err := computationPackageReference(cases, profile)
	if err != nil {
		return boundedTextSmokeBinding{}, err
	}
	var r packageProfileReceipt
	if err := json.Unmarshal(raw, &r); err != nil {
		return boundedTextSmokeBinding{}, err
	}
	sha := r.Result.Composition.SHA
	if r.Schema != "gooo/workspace-body-execution-receipt/v1" || r.Decision != "PASS" || r.Error != "" ||
		r.Manifest != fmt.Sprintf("sha256:%x", sha256.Sum256(manifest)) ||
		r.Cases != fmt.Sprintf("sha256:%x", sha256.Sum256(cases)) ||
		!computationPackageReceiptProgramEqual(raw, program) ||
		!boundedTextDigest(r.Result.SourceSHA) ||
		r.Result.Program.Entry.PackagePath != profile.Package || r.Result.Program.Entry.Activity != profile.Activity ||
		!boundedTextDigest(sha) || selected.Program != "" && selected.Program != sha {
		return boundedTextSmokeBinding{}, fmt.Errorf("computation package source or program identity differs")
	}
	if replay && (selected.Program == "" || !boundedTextDigest(selected.Receipt) || r.From != selected.Receipt ||
		r.Result.Replay == nil || !jointSmokeInt(r.Result.Replay.Calls, 0)) {
		return boundedTextSmokeBinding{}, fmt.Errorf("computation package saved receipt is unbound")
	}
	runtime := r.Result.Runtime
	if runtime.Stage != "COMPLETE" || runtime.Failure != "" ||
		!jointSmokeInt(runtime.Passed, profile.Total) || !jointSmokeInt(runtime.Total, profile.Total) ||
		!jointSmokeInt(runtime.Calls, 0) || !jointSmokeBool(runtime.Projection, true) || !jointSmokeBool(runtime.Replay, true) ||
		len(runtime.Traces) != profile.Total || runtime.SHA != sha {
		return boundedTextSmokeBinding{}, fmt.Errorf("computation package native results are incomplete")
	}
	seen := make([]bool, profile.Total)
	key := profile.Package + ":" + profile.Activity
	for _, trace := range runtime.Traces {
		if trace.Index == nil {
			return boundedTextSmokeBinding{}, fmt.Errorf("computation package case index is missing")
		}
		i := *trace.Index
		if i < 0 || i >= len(rows) || seen[i] || len(trace.Deliveries) != 1 {
			return boundedTextSmokeBinding{}, fmt.Errorf("computation package case identity differs")
		}
		seen[i] = true
		d := trace.Deliveries[0]
		if d.ID != profile.ID || !jointSmokeBool(d.Passed, true) ||
			!samePackageSourceValue(d.Input, rows[i].Inputs[key]) ||
			!samePackageSourceValue(d.Actual, rows[i].Expected[key]) || !samePackageSourceValue(d.Expected, rows[i].Expected[key]) {
			return boundedTextSmokeBinding{}, fmt.Errorf("computation package value differs at row %d", i)
		}
	}
	if err := validateSingleConditionFill(r, program, profile); err != nil {
		return boundedTextSmokeBinding{}, err
	}
	return boundedTextSmokeBinding{Program: sha, Receipt: fmt.Sprintf("sha256:%x", sha256.Sum256(raw))}, nil
}
