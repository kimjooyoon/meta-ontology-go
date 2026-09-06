package syntaxregistration

import (
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
}

func bindContract() (contractBinding, error) {
	file, diagnostics := syntax.ParseFile("contract.gooo", string(contractSource))
	if file == nil || diagnostics.HasErrors() {
		return contractBinding{}, fmt.Errorf("registration contract cannot be parsed")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return contractBinding{}, err
	}
	outputs := []string{"RegistrationCandidate", "SyntaxCorpus", "NativeRegistry",
		"SyntaxDenominator", "SyntaxConformance", "DenominatorAdmission",
		"DenominatorSelection", "DenominatorDigest", "DenominatorEvidence", "MigrationConformance"}
	activities := []string{Operation, "GenerateSyntaxCorpus", "GenerateNativeRegistry",
		"GenerateSyntaxDenominator", "GenerateSyntaxConformance", "GenerateDenominatorAdmission",
		"GenerateDenominatorSelection", "GenerateDenominatorDigest",
		"GenerateDenominatorEvidence", "GenerateMigrationConformance"}
	if len(ir.Graph.Nodes()) != 21 {
		return contractBinding{}, fmt.Errorf("registration contract node inventory is not exact")
	}
	input, ok := ir.Graph.NodeByName(ir.Namespace, "RegistrationRequest")
	if !ok || input.Kind != semantic.Entity || input.ID.String() != "gooo://syntax-registration/request" {
		return contractBinding{}, fmt.Errorf("registration request identity mismatch")
	}
	binding := contractBinding{source: digest(contractSource), semantic: ir.StableHash()}
	for index, name := range activities {
		activity, found := ir.Graph.NodeByName(ir.Namespace, name)
		output, outputFound := ir.Graph.NodeByName(ir.Namespace, outputs[index])
		if !found || !outputFound || activity.Kind != semantic.Activity || output.Kind != semantic.Entity ||
			!ir.Graph.HasFact(semantic.FactKey{Subject: activity.ID, Predicate: semantic.Used, Object: input.ID}) ||
			!ir.Graph.HasFact(semantic.FactKey{Subject: output.ID, Predicate: semantic.WasGeneratedBy, Object: activity.ID}) {
			return contractBinding{}, fmt.Errorf("registration activity ABI mismatch: %s", name)
		}
		if index == 0 && activity.ValueProgram != "syntax.register:v1" {
			return contractBinding{}, fmt.Errorf("registration native program mismatch")
		}
		binding.activities = append(binding.activities, activity.ID.String())
		binding.outputs = append(binding.outputs, output.ID.String())
	}
	return binding, nil
}
