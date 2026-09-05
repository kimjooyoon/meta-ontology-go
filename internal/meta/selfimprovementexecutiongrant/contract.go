package selfimprovementexecutiongrant

import (
	"errors"
	"reflect"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	CanonicalExecutorInputEntityID        = "gooo://self-improvement/execution-grant/entity/canonical-executor-grant-input"
	CanonicalExecutorFixtureEntityID      = "gooo://self-improvement/execution-grant/entity/canonical-executor-grant-fixture"
	CanonicalExecutorVerificationEntityID = "gooo://self-improvement/execution-grant/entity/canonical-executor-grant-verification"
	CanonicalExecutorMaterializeProgram   = "fixture_schema=gooo/self-improvement-execution-grant-canonical-request/v1;decision_type=CANONICAL_EXECUTOR_CONFORMANCE_ONLY;scope=CALLER_OWNED_TEMP_ONLY;live_authority=false;user_decision=false;product_utility_evidence=false;repository_writes=0;local_test_executions=0;next=VerifyCanonicalExecutorGrantFixture"
	CanonicalExecutorVerifyProgram         = "verification=independent;candidate_execution=0;grant_consumption=0;repository_writes=0;local_test_executions=0;refuted_dominates_unknown=true;output=caller-owned-artifact"
)

// CanonicalExecutorActivityContract is the typed semantic contract compiled
// from the Gooo activity declarations. It is deliberately carried alongside
// the policy evidence so the fixture builder consumes the activity authority
// rather than reimplementing a YAML- or Go-only convention.
type CanonicalExecutorActivityContract struct {
	Name             string   `json:"name"`
	ID               string   `json:"id"`
	Kind             string   `json:"kind"`
	Inputs           []string `json:"inputs"`
	InputIDs         []string `json:"input_ids"`
	Outputs          []string `json:"outputs"`
	OutputIDs        []string `json:"output_ids"`
	Attributes       map[string]string `json:"attributes"`
	Program          string   `json:"program"`
	ProgramDigest    string   `json:"program_digest"`
	SemanticNodeDigest string `json:"semantic_node_digest"`
}

type CanonicalExecutorSemanticContract struct {
	Schema           string                             `json:"schema"`
	SourceDigest     string                             `json:"source_digest"`
	CanonicalDigest  string                             `json:"canonical_digest"`
	SemanticIRDigest string                             `json:"semantic_ir_digest"`
	FullIRDigest     string                             `json:"full_ir_digest"`
	Entities         []string                           `json:"entities"`
	Activities       []CanonicalExecutorActivityContract `json:"activities"`
	Digest           string                             `json:"digest"`
}

const canonicalExecutorSemanticContractSchema = "gooo/self-improvement-execution-grant-semantic-contract/v1"

func compileCanonicalExecutorSemanticContract(file *syntax.File, raw, canonical string, ir semantic.IR) (CanonicalExecutorSemanticContract, error) {
	document, err := bidir.DocumentFromSyntax(file)
	if err != nil {
		return CanonicalExecutorSemanticContract{}, err
	}
	model, err := bidir.Get(document)
	if err != nil {
		return CanonicalExecutorSemanticContract{}, err
	}
	if len(model.Nodes) != len(document.Declarations) {
		return CanonicalExecutorSemanticContract{}, errors.New("canonical executor model declaration denominator drifted")
	}
	contract := CanonicalExecutorSemanticContract{
		Schema: canonicalExecutorSemanticContractSchema,
		SourceDigest: digestBytes([]byte(raw)),
		CanonicalDigest: digestBytes([]byte(canonical)),
		SemanticIRDigest: digestBytes([]byte(ir.SemanticCanonical())),
		FullIRDigest: digestBytes([]byte(ir.Canonical())),
	}
	for _, declaration := range document.Declarations {
		switch declaration.Kind {
		case bidir.EntityKind:
			contract.Entities = append(contract.Entities, modelNodeID(model, declaration.Kind, declaration.Name))
		case bidir.ActivityKind:
			activity := CanonicalExecutorActivityContract{
				Name: declaration.Name, ID: modelNodeID(model, declaration.Kind, declaration.Name), Kind: string(declaration.Kind),
				Attributes: cloneCanonicalExecutorAttributes(declaration.Attributes),
			}
			activity.Program = declaration.Attributes[bidir.ActivityValueProgramAttribute]
			activity.ProgramDigest = digestBytes([]byte(activity.Program))
			for _, input := range declaration.Inputs {
				activity.Inputs = append(activity.Inputs, input.Name)
				activity.InputIDs = append(activity.InputIDs, resolveDeclarationReference(document, model, input))
			}
			for _, output := range declaration.Outputs {
				activity.Outputs = append(activity.Outputs, output.Name)
				activity.OutputIDs = append(activity.OutputIDs, resolveDeclarationReference(document, model, output))
			}
			if id, parseErr := semantic.ParseIdentity(activity.ID); parseErr == nil {
				if node, ok := ir.Graph.Node(id); ok {
					activity.SemanticNodeDigest = digestBytes([]byte(node.SemanticCanonical()))
				}
			}
			contract.Activities = append(contract.Activities, activity)
		}
	}
	sort.Strings(contract.Entities)
	sort.Slice(contract.Activities, func(i, j int) bool { return contract.Activities[i].ID < contract.Activities[j].ID })
	contract.Digest = digestCanonicalExecutorSemanticContract(contract)
	return contract, nil
}

func cloneCanonicalExecutorAttributes(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	result := make(map[string]string, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}

func modelNodeID(model bidir.Model, kind bidir.Kind, name string) string {
	for _, node := range model.Nodes {
		if node.Kind == kind && node.Name == name {
			return string(node.ID)
		}
	}
	return ""
}

func resolveDeclarationReference(document bidir.Document, model bidir.Model, reference bidir.Reference) string {
	if reference.ID != "" {
		return string(reference.ID)
	}
	for _, declaration := range document.Declarations {
		if declaration.Name == reference.Name && declaration.Kind == bidir.EntityKind {
			return modelNodeID(model, declaration.Kind, declaration.Name)
		}
	}
	return ""
}

func digestCanonicalExecutorSemanticContract(value CanonicalExecutorSemanticContract) string {
	value.Digest = ""
	return digestJSON(value)
}

func validateCanonicalExecutorSemanticContract(file *syntax.File, raw, canonical string, ir semantic.IR, actual CanonicalExecutorSemanticContract) error {
	expected, err := compileCanonicalExecutorSemanticContract(file, raw, canonical, ir)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(expected, actual) || actual.Digest != digestCanonicalExecutorSemanticContract(actual) {
		return errors.New("canonical executor semantic contract is not an exact compiled binding")
	}
	if len(actual.Entities) != 3 || len(actual.Activities) != 2 {
		return errors.New("canonical executor semantic contract denominator drifted")
	}
	return validateCanonicalExecutorActivityPrograms(actual)
}

func validateCanonicalExecutorActivityPrograms(contract CanonicalExecutorSemanticContract) error {
	programs := map[string]string{}
	for _, activity := range contract.Activities {
		if activity.Kind != string(bidir.ActivityKind) || activity.Program == "" || activity.ProgramDigest != digestBytes([]byte(activity.Program)) || activity.SemanticNodeDigest == "" {
			return errors.New("canonical executor activity semantic binding is incomplete")
		}
		programs[activity.Name] = activity.Program
	}
	if programs["MaterializeCanonicalExecutorGrantFixture"] != CanonicalExecutorMaterializeProgram || programs["VerifyCanonicalExecutorGrantFixture"] != CanonicalExecutorVerifyProgram {
		return errors.New("canonical executor activity program binding is not exact")
	}
	return nil
}

func canonicalExecutorActivityProgram(contract CanonicalExecutorSemanticContract, name string) (string, bool) {
	for _, activity := range contract.Activities {
		if activity.Name == name {
			return activity.Program, true
		}
	}
	return "", false
}

func canonicalExecutorActivityDigest(contract CanonicalExecutorSemanticContract, name string) string {
	for _, activity := range contract.Activities {
		if activity.Name == name {
			return digestJSON(activity)
		}
	}
	return ""
}

func canonicalExecutorProgramBinding(program PolicyProgram) (map[string]string, map[string]string, error) {
	if err := validateCanonicalExecutorActivityPrograms(program.ExecutorContract); err != nil {
		return nil, nil, err
	}
	materialize, ok := canonicalExecutorActivityProgram(program.ExecutorContract, "MaterializeCanonicalExecutorGrantFixture")
	if !ok {
		return nil, nil, errors.New("canonical executor materialization activity is missing")
	}
	verify, ok := canonicalExecutorActivityProgram(program.ExecutorContract, "VerifyCanonicalExecutorGrantFixture")
	if !ok {
		return nil, nil, errors.New("canonical executor verification activity is missing")
	}
	materializeValues := parseCanonicalExecutorProgram(materialize)
	verifyValues := parseCanonicalExecutorProgram(verify)
	if materializeValues["fixture_schema"] != CanonicalExecutorRequestSchema || materializeValues["decision_type"] != CanonicalExecutorDecisionType || materializeValues["scope"] != CanonicalExecutorScope || materializeValues["live_authority"] != "false" || materializeValues["user_decision"] != "false" || materializeValues["product_utility_evidence"] != "false" || materializeValues["repository_writes"] != "0" || materializeValues["local_test_executions"] != "0" || materializeValues["next"] != "VerifyCanonicalExecutorGrantFixture" {
		return nil, nil, errors.New("canonical executor materialization activity semantics are not exact")
	}
	if verifyValues["verification"] != "independent" || verifyValues["candidate_execution"] != "0" || verifyValues["grant_consumption"] != "0" || verifyValues["repository_writes"] != "0" || verifyValues["local_test_executions"] != "0" || verifyValues["refuted_dominates_unknown"] != "true" || verifyValues["output"] != "caller-owned-artifact" {
		return nil, nil, errors.New("canonical executor verification activity semantics are not exact")
	}
	return materializeValues, verifyValues, nil
}

func parseCanonicalExecutorProgram(program string) map[string]string {
	values := make(map[string]string)
	for _, part := range strings.Split(program, ";") {
		key, value, ok := strings.Cut(part, "=")
		if ok {
			values[key] = value
		}
	}
	return values
}

func validateCanonicalExecutorDeclaration(file *syntax.File) error {
	document, err := bidir.DocumentFromSyntax(file)
	if err != nil {
		return err
	}
	model, err := bidir.Get(document)
	if err != nil {
		return err
	}
	if len(document.Declarations) != 5 {
		return errors.New("canonical executor declaration denominator drifted")
	}
	entities := map[string]string{}
	activities := map[string]bidir.Declaration{}
	for _, declaration := range document.Declarations {
		switch declaration.Kind {
		case bidir.EntityKind:
			entities[declaration.Name] = modelNodeID(model, declaration.Kind, declaration.Name)
		case bidir.ActivityKind:
			activities[declaration.Name] = declaration
		}
	}
	if entities["CanonicalExecutorGrantInput"] != CanonicalExecutorInputEntityID || entities["CanonicalExecutorGrantFixture"] != CanonicalExecutorFixtureEntityID || entities["CanonicalExecutorGrantVerification"] != CanonicalExecutorVerificationEntityID {
		return errors.New("canonical executor entity declaration is not exact")
	}
	materialize, ok := activities["MaterializeCanonicalExecutorGrantFixture"]
	if !ok || len(materialize.Inputs) != 1 || materialize.Inputs[0].Name != "CanonicalExecutorGrantInput" || len(materialize.Outputs) != 1 || materialize.Outputs[0].Name != "CanonicalExecutorGrantFixture" || materialize.Attributes[bidir.ActivityValueProgramAttribute] != CanonicalExecutorMaterializeProgram {
		return errors.New("canonical executor materialization activity is not exact")
	}
	verify, ok := activities["VerifyCanonicalExecutorGrantFixture"]
	if !ok || len(verify.Inputs) != 1 || verify.Inputs[0].Name != "CanonicalExecutorGrantFixture" || len(verify.Outputs) != 1 || verify.Outputs[0].Name != "CanonicalExecutorGrantVerification" || verify.Attributes[bidir.ActivityValueProgramAttribute] != CanonicalExecutorVerifyProgram {
		return errors.New("canonical executor verification activity is not exact")
	}
	return nil
}
