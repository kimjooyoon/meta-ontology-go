package toolchainrelease

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

const namedInputsRoot = "examples/named-inputs"
const namedCreditID = "urn:gooo:example:named-inputs:credit"
const namedNetID = "urn:gooo:example:named-inputs:net"

// The source's two cases have four independent output expectations. Input names
// remain source bindings; the external input and bind ports retain their order.
var namedInputExamples = [2]struct{ input, debit, credit, net int64 }{
	{101, 29, 124, 95}, {-53, 11, -30, -41},
}

func smokeNamedInputs(binary string, input BuildInput) error {
	root := filepath.Join(input.Root, namedInputsRoot)
	manifest, err := os.ReadFile(filepath.Join(root, "workspace.json"))
	if err != nil {
		return err
	}
	cases, err := os.ReadFile(filepath.Join(root, "cases.json"))
	if err != nil {
		return err
	}
	program, err := computationPackageProgram(root, manifest)
	if err != nil {
		return err
	}
	original := filepath.Join(input.OutputDir, input.Target.ID+"-named-inputs-execute.json")
	selected := boundedTextSmokeBinding{}
	for _, replay := range []bool{false, true} {
		mode := "execute"
		if replay {
			mode = "replay"
		}
		args := []string{"package", mode, "--json", "--cases", filepath.Join(namedInputsRoot, "cases.json")}
		if replay {
			args = append(args, "--receipt", original)
		}
		args = append(args, filepath.Join(namedInputsRoot, "workspace.json"))
		raw, runErr := commandOutput(input.Root, nil, binary, args...)
		raw, err = retainLanguageCommandOutput(input, "named-inputs-"+mode, raw, runErr)
		if err != nil {
			return err
		}
		selected, err = validateNamedInputs(raw, manifest, cases, program, replay, selected)
		if err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_NAMED_INPUTS %s: %w", mode, err)
		}
	}
	return nil
}

func namedInputCaseReference(raw []byte) error {
	var suite struct {
		Schema string
		Cases  []packageProfileCase
	}
	if err := json.Unmarshal(raw, &suite); err != nil {
		return err
	}
	if suite.Schema != "gooo/body-composition-cases/v1" || len(suite.Cases) != len(namedInputExamples) {
		return fmt.Errorf("named input case inventory differs")
	}
	for i, row := range suite.Cases {
		want := namedInputExamples[i]
		if len(row.Inputs) != 2 || len(row.Expected) != 2 ||
			!namedInputInteger(row.Inputs["named/inputs:Credit"], want.input) ||
			!namedInputInteger(row.Inputs["named/inputs:Net.input1"], want.debit) ||
			!namedInputInteger(row.Expected["named/inputs:Credit"], want.credit) ||
			!namedInputInteger(row.Expected["named/inputs:Net"], want.net) {
			return fmt.Errorf("named input independent row %d differs", i)
		}
	}
	return nil
}

func namedInputInteger(raw json.RawMessage, expected int64) bool {
	return samePackageSourceValue(raw, json.RawMessage(fmt.Sprintf("%d", expected)))
}

func validateNamedInputs(raw, manifest, cases []byte, program workspaceexecution.Program, replay bool,
	selected boundedTextSmokeBinding) (boundedTextSmokeBinding, error) {
	if err := namedInputCaseReference(cases); err != nil {
		return boundedTextSmokeBinding{}, err
	}
	var r packageProfileReceipt
	if err := json.Unmarshal(raw, &r); err != nil {
		return boundedTextSmokeBinding{}, err
	}
	generated := r.Result.Composition.SHA
	if r.Schema != "gooo/workspace-body-execution-receipt/v1" || r.Decision != "PASS" || r.Error != "" ||
		r.Manifest != computationDigest(manifest) || r.Cases != computationDigest(cases) ||
		!computationPackageReceiptProgramEqual(raw, program) ||
		r.Result.SourceSHA != computationDigest([]byte(program.Source)) ||
		r.Result.Program.Entry.PackagePath != "named/inputs" || r.Result.Program.Entry.Activity != "Net" ||
		!boundedTextDigest(generated) || selected.Program != "" && selected.Program != generated {
		return boundedTextSmokeBinding{}, fmt.Errorf("named input source or generated program differs")
	}
	if replay {
		if selected.Program == "" || !boundedTextDigest(selected.Receipt) || r.From != selected.Receipt ||
			r.Result.Replay == nil || !jointSmokeInt(r.Result.Replay.Calls, 0) {
			return boundedTextSmokeBinding{}, fmt.Errorf("named input saved receipt is unbound")
		}
	} else if r.From != "" || r.Result.Replay != nil {
		return boundedTextSmokeBinding{}, fmt.Errorf("named input first execution is a replay")
	}
	if err := validateNamedInputRuntime(r.Result.Runtime, generated); err != nil {
		return boundedTextSmokeBinding{}, err
	}
	return boundedTextSmokeBinding{Program: generated, Receipt: computationDigest(raw)}, nil
}

func validateNamedInputRuntime(runtime packageProfileRuntime, generated string) error {
	if runtime.Stage != "COMPLETE" || runtime.Failure != "" || runtime.SHA != generated ||
		!jointSmokeInt(runtime.Passed, 4) || !jointSmokeInt(runtime.Total, 4) ||
		!jointSmokeInt(runtime.Calls, 0) || !jointSmokeBool(runtime.Projection, true) ||
		!jointSmokeBool(runtime.Replay, true) || len(runtime.Traces) != 2 {
		return fmt.Errorf("named input native output population is incomplete")
	}
	seen := [2]bool{}
	for _, trace := range runtime.Traces {
		if trace.Index == nil || *trace.Index < 0 || *trace.Index >= len(seen) || seen[*trace.Index] || len(trace.Deliveries) != 2 {
			return fmt.Errorf("named input case identity differs")
		}
		i := *trace.Index
		seen[i] = true
		ids := map[string]bool{}
		for _, delivery := range trace.Deliveries {
			if ids[delivery.ID] || !validateNamedInputDelivery(delivery, i) {
				return fmt.Errorf("named input original value or port differs at row %d", i)
			}
			ids[delivery.ID] = true
		}
		if !ids[namedCreditID] || !ids[namedNetID] {
			return fmt.Errorf("named input activity identity differs")
		}
	}
	return nil
}

func validateNamedInputDelivery(delivery packageProfileDelivery, index int) bool {
	if !jointSmokeBool(delivery.Passed, true) || len(delivery.Fault) != 0 || len(delivery.BlockedBy) != 0 {
		return false
	}
	want := namedInputExamples[index]
	if delivery.ID == namedCreditID {
		return namedInputInteger(delivery.Input, want.input) && len(delivery.Inputs) == 0 &&
			namedInputInteger(delivery.Actual, want.credit) && namedInputInteger(delivery.Expected, want.credit)
	}
	if delivery.ID != namedNetID || !namedInputInteger(delivery.Actual, want.net) ||
		!namedInputInteger(delivery.Expected, want.net) || len(delivery.Inputs) != 2 {
		return false
	}
	for i, port := range delivery.Inputs {
		value, producer := want.credit, namedCreditID
		if i == 1 {
			value, producer = want.debit, ""
		}
		if port.Port != fmt.Sprintf("input%d", i) || port.EntityID != "urn:gooo:type:integer" ||
			port.ProducerID != producer || !namedInputInteger(port.Value, value) {
			return false
		}
	}
	return true
}
