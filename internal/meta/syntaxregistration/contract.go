package syntaxregistration

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

//go:embed contract.gooo
var contractSource []byte

type contractBinding struct {
	source, semantic string
	activities       []string
	outputs          []string

	inputActivity, inputOutput string
}

func bindContractSource(source []byte) (contractBinding, error) {
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport("contract.gooo", string(source), syntax.EntityFieldsV5Support())
	if file == nil || diagnostics.HasErrors() {
		return contractBinding{}, fmt.Errorf("registration contract cannot be parsed: %v", diagnostics)
	}
	ir, err := bidir.LowerContextWithEntityFieldsSupport(context.Background(), file, bidir.EntityFieldsV5Support())
	if err != nil {
		return contractBinding{}, err
	}
	if err := bindCaseComposition(ir); err != nil {
		return contractBinding{}, err
	}
	if err := bindAppendComposition(ir); err != nil {
		return contractBinding{}, err
	}
	outputs, activities := registrationRoles()
	request, ok := ir.Graph.NodeByName(ir.Namespace, "RegistrationRequest")
	if !ok || request.Kind != semantic.Entity || request.ID.String() != "gooo://syntax-registration/request" {
		return contractBinding{}, fmt.Errorf("registration request identity mismatch")
	}
	input, inputFound := ir.Graph.NodeByName(ir.Namespace, "PinnedRegistrationInput")
	pin, pinFound := ir.Graph.NodeByName(ir.Namespace, "PinRegistrationExecutionIdentity")
	if !inputFound || !pinFound || input.Kind != semantic.Entity || pin.Kind != semantic.Activity ||
		input.ID.String() != "gooo://syntax-registration/pinned-input" ||
		!ir.Graph.HasFact(semantic.FactKey{Subject: pin.ID, Predicate: semantic.Used, Object: request.ID}) ||
		!ir.Graph.HasFact(semantic.FactKey{Subject: input.ID, Predicate: semantic.WasGeneratedBy, Object: pin.ID}) {
		return contractBinding{}, fmt.Errorf("registration execution identity ABI mismatch")
	}
	binding := contractBinding{source: digest(source), semantic: ir.StableHash(),
		inputActivity: pin.ID.String(), inputOutput: input.ID.String()}
	for index, name := range activities {
		activity, found := ir.Graph.NodeByName(ir.Namespace, name)
		output, outputFound := ir.Graph.NodeByName(ir.Namespace, outputs[index])
		if !found || !outputFound || activity.Kind != semantic.Activity || output.Kind != semantic.Entity ||
			!ir.Graph.HasFact(semantic.FactKey{Subject: activity.ID, Predicate: semantic.Used, Object: input.ID}) ||
			!ir.Graph.HasFact(semantic.FactKey{Subject: output.ID, Predicate: semantic.WasGeneratedBy, Object: activity.ID}) {
			return contractBinding{}, fmt.Errorf("registration activity ABI mismatch: %s", name)
		}
		if index == 0 && activity.ValueProgram != "syntax.register:v2" {
			return contractBinding{}, fmt.Errorf("registration native program mismatch")
		}
		if name == "GenerateDenominatorAdmission" && activity.ValueProgram != "syntax.history.append:v1" {
			return contractBinding{}, fmt.Errorf("registration history append program mismatch")
		}
		if name == "GenerateDenominatorSelection" && activity.ValueProgram != "syntax.selection.append:v1" {
			return contractBinding{}, fmt.Errorf("registration selection append program mismatch")
		}
		binding.activities = append(binding.activities, activity.ID.String())
		binding.outputs = append(binding.outputs, output.ID.String())
	}
	return binding, nil
}
