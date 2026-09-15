package generation

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

//go:embed callback-extraction-contract.gooo
var callbackExtractionContractSource []byte

type CallbackExtractionStep struct {
	ID       string
	Activity string
	Input    string
	Output   string
	Program  string
}

type CallbackExtractionContract struct {
	SourceDigest   string                   `json:"source_digest"`
	SemanticDigest string                   `json:"semantic_digest"`
	Steps          []CallbackExtractionStep `json:"steps"`
}

type CallbackExtractionRecord struct {
	Activity string            `json:"activity"`
	Entity   string            `json:"entity"`
	Program  string            `json:"program"`
	Fields   map[string]string `json:"fields"`
}

var callbackExtractionSteps = []CallbackExtractionStep{
	{ID: "source", Activity: "BindCallbackExtractionSource", Input: "CallbackExtractionInput", Output: "CallbackSourceBinding"},
	{ID: "structure", Activity: "ProveCallbackExtractionStructure", Input: "CallbackSourceBinding", Output: "CallbackStructureProof"},
	{ID: "capacity", Activity: "RenderCallbackExtractionPackage", Input: "CallbackStructureProof", Output: "CallbackPackageCapacity"},
	{ID: "types", Activity: "CheckCallbackExtractionPackage", Input: "CallbackPackageCapacity", Output: "CallbackPackageTypes"},
	{ID: "observers", Activity: "BindCallbackExtractionObservers", Input: "CallbackPackageTypes", Output: "CallbackObserverEvidence"},
	{ID: "admission", Activity: "RecordCallbackExtractionAdmission", Input: "CallbackObserverEvidence", Output: "CallbackExtractionAdmission"},
}

func LoadCallbackExtractionContract() (CallbackExtractionContract, error) {
	return parseCallbackExtractionContract(callbackExtractionContractSource)
}

func parseCallbackExtractionContract(raw []byte) (CallbackExtractionContract, error) {
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport("callback-extraction-contract.gooo", string(raw), syntax.EntityFieldsV1Support())
	if file == nil || diagnostics.HasErrors() {
		return CallbackExtractionContract{}, fmt.Errorf("callback extraction contract syntax is invalid")
	}
	ir, err := bidir.LowerContextWithEntityFieldsSupport(context.Background(), file, bidir.EntityFieldsV1Support())
	if err != nil {
		return CallbackExtractionContract{}, err
	}
	if file.Package == nil || file.Namespace == nil || file.Package.Name != "callbackextraction" || file.Namespace.Name != "callbackextraction" ||
		len(file.Decls) != 13 || len(ir.Graph.Nodes()) != 13 || len(file.Bindings) != 5 || len(ir.RuntimeBindings) != 5 {
		return CallbackExtractionContract{}, fmt.Errorf("callback extraction contract identity or cardinality differs")
	}
	if err := validateCallbackExtractionEntity(file, ir, "CallbackExtractionInput", []string{"LogicalPath", "Subject", "SourceDigest"}); err != nil {
		return CallbackExtractionContract{}, err
	}
	steps := make([]CallbackExtractionStep, 0, len(callbackExtractionSteps))
	for index, expected := range callbackExtractionSteps {
		step, err := validateCallbackExtractionStep(file, ir, index, expected)
		if err != nil {
			return CallbackExtractionContract{}, err
		}
		steps = append(steps, step)
	}
	digest := sha256.Sum256(raw)
	return CallbackExtractionContract{SourceDigest: hex.EncodeToString(digest[:]), SemanticDigest: ir.StableHash(), Steps: steps}, nil
}

func validateCallbackExtractionEntity(file *syntax.File, ir semantic.IR, name string, fields []string) error {
	id := "gooo://callback-extraction/entity/" + name
	declaration, ok := findCallbackPreviewEntity(file, name)
	if !ok || declaration.ID != id || len(declaration.Fields) != len(fields) {
		return fmt.Errorf("callback extraction entity %s differs", name)
	}
	node, ok := ir.Graph.NodeByName(ir.Namespace, name)
	if !ok || node.Kind != semantic.Entity || node.ID.String() != id || len(node.Fields) != len(fields) {
		return fmt.Errorf("callback extraction semantic entity %s differs", name)
	}
	for index, field := range node.Fields {
		if field.Name != fields[index] || field.ID.String() != "gooo://callback-extraction/field/"+name+"/"+fields[index] ||
			field.Presence != semantic.Required || field.Cardinality != semantic.One || field.TypeRef.ID != semantic.BuiltinStringTypeID {
			return fmt.Errorf("callback extraction field %s.%s differs", name, fields[index])
		}
	}
	return nil
}

func validateCallbackExtractionStep(file *syntax.File, ir semantic.IR, index int, step CallbackExtractionStep) (CallbackExtractionStep, error) {
	fields := []string{"State", "EvidenceDigest", "ObservedCount", "RequiredCount"}
	if step.ID == "observers" {
		fields = append([]string{"Scope", "SourcePackageDigest", "FinalPackageDigest", "TestEventDigest", "ObservationDecision"}, fields...)
	}
	if err := validateCallbackExtractionEntity(file, ir, step.Output, fields); err != nil {
		return step, err
	}
	step.Program = "callback-extraction." + step.ID + ":v1;authority=PROPOSAL_ONLY"
	if step.ID == "observers" {
		step.Program = "callback-extraction.observers:v2;scope=PACKAGE_TEST_EVENTS_ONLY;authority=PROPOSAL_ONLY"
	}
	declaration, ok := findCallbackPreviewActivity(file, step.Activity)
	if !ok || len(declaration.Inputs) != 1 || declaration.Inputs[0].Name != step.Input || declaration.Output != step.Output ||
		!declaration.ValueProgramPresent || declaration.ValueProgram != step.Program {
		return step, fmt.Errorf("callback extraction activity %s differs", step.Activity)
	}
	node, found := ir.Graph.NodeByName(ir.Namespace, step.Activity)
	input, inputFound := ir.Graph.NodeByName(ir.Namespace, step.Input)
	output, outputFound := ir.Graph.NodeByName(ir.Namespace, step.Output)
	if !found || !inputFound || !outputFound || node.Kind != semantic.Activity || node.ValueProgram != step.Program ||
		!ir.Graph.HasFact(semantic.FactKey{Subject: node.ID, Predicate: semantic.Used, Object: input.ID}) ||
		!ir.Graph.HasFact(semantic.FactKey{Subject: output.ID, Predicate: semantic.WasGeneratedBy, Object: node.ID}) {
		return step, fmt.Errorf("callback extraction provenance for %s differs", step.Activity)
	}
	if index == 0 {
		return step, nil
	}
	for _, binding := range ir.RuntimeBindings {
		producer, producerFound := ir.Graph.Node(binding.ProducerActivity)
		if producerFound && producer.Name == callbackExtractionSteps[index-1].Activity && binding.ConsumerActivity == node.ID &&
			binding.Entity == input.ID && binding.ProducerPort == semantic.RuntimeOutputPort && binding.ConsumerPort == semantic.RuntimeInputPort {
			return step, nil
		}
	}
	return step, fmt.Errorf("callback extraction predecessor binding for %s is missing", step.Activity)
}

// BuildRecord binds native counters to fields in the lowered Gooo contract.
// Proposal records cannot constitute an OperationResult or write authorization.
func (contract CallbackExtractionContract) BuildRecord(index int, state, evidence string, observed, required int) (CallbackExtractionRecord, error) {
	if len(contract.Steps) != 6 || contract.SourceDigest == "" || contract.SemanticDigest == "" || index < 0 || index >= len(contract.Steps) ||
		observed < 0 || required <= 0 || observed > required {
		return CallbackExtractionRecord{}, fmt.Errorf("callback extraction record is not bound")
	}
	if index < 4 && (state != "CLOSED" || evidence == "" || observed != required) {
		return CallbackExtractionRecord{}, fmt.Errorf("callback extraction observation is incomplete")
	}
	if index >= 4 && state != "UNKNOWN" {
		return CallbackExtractionRecord{}, fmt.Errorf("callback extraction proposal has no observer admission authority")
	}
	step := contract.Steps[index]
	fields := map[string]string{
		"State": state, "EvidenceDigest": evidence, "ObservedCount": strconv.Itoa(observed), "RequiredCount": strconv.Itoa(required),
	}
	if index == 4 {
		fields["Scope"], fields["ObservationDecision"] = "PACKAGE_TEST_EVENTS_ONLY", "UNKNOWN"
		fields["SourcePackageDigest"], fields["FinalPackageDigest"], fields["TestEventDigest"] = "", "", ""
	}
	return CallbackExtractionRecord{Activity: step.Activity, Entity: step.Output, Program: step.Program, Fields: fields}, nil
}
