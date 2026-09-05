package generation

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// The native table is the deliberately small FOUNDATION ABI boundary. The
// embedded Gooo source describes the same contract, but cannot invent a new
// native executor or an input-kind alias.
type nativeOperationInput struct {
	Operation    sourcepolicy.Operation
	Activity     string
	InputEntity  string
	OutputEntity string
	InputKind    sourcepolicy.SubjectKind
}

var nativeOperationInputs = []nativeOperationInput{
	{Operation: sourcepolicy.OperationCollapseAssign, Activity: "CollapseAssignReturn", InputEntity: "FunctionInput", OutputEntity: "OperationResult", InputKind: sourcepolicy.SubjectKindFunction},
	{Operation: sourcepolicy.OperationSplitGo, Activity: "SplitGoDeclarations", InputEntity: "FileInput", OutputEntity: "OperationResult", InputKind: sourcepolicy.SubjectKindFile},
	{Operation: sourcepolicy.OperationSplitGooo, Activity: "SplitGoooSections", InputEntity: "FileInput", OutputEntity: "OperationResult", InputKind: sourcepolicy.SubjectKindFile},
	{Operation: sourcepolicy.OperationExtractFunction, Activity: "ExtractFunction", InputEntity: "FunctionInput", OutputEntity: "OperationResult", InputKind: sourcepolicy.SubjectKindFunction},
}

type operationInputContractBinding struct {
	InputSubjectKind sourcepolicy.SubjectKind
}

type operationInputContract struct {
	SourceDigest   string
	SemanticDigest string
	Bindings       map[sourcepolicy.Operation]operationInputContractBinding
}

//go:embed operation-input-contract.gooo
var operationInputContractSource []byte

func loadOperationInputContract() (operationInputContract, error) {
	return parseOperationInputContract(operationInputContractSource)
}

func parseOperationInputContract(raw []byte) (operationInputContract, error) {
	file, diagnostics := syntax.ParseFile("operation-input-contract.gooo", string(raw))
	if file == nil || diagnostics.HasErrors() {
		return operationInputContract{}, fmt.Errorf("operation input contract has syntax errors")
	}
	if file.Package == nil || file.Namespace == nil ||
		file.Package.Name != "operationinputcontract" || file.Namespace.Name != "operationinputcontract" {
		return operationInputContract{}, fmt.Errorf("operation input contract headers are not exact")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return operationInputContract{}, fmt.Errorf("lower operation input contract: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return operationInputContract{}, fmt.Errorf("validate operation input contract: %w", err)
	}

	entityIDs := map[string]string{
		"FunctionInput":   "gooo://meta-operation-input-contract/entity/function-input",
		"FileInput":       "gooo://meta-operation-input-contract/entity/file-input",
		"OperationResult": "gooo://meta-operation-input-contract/entity/result",
	}
	inputKinds := map[string]sourcepolicy.SubjectKind{
		"FunctionInput": sourcepolicy.SubjectKindFunction,
		"FileInput":     sourcepolicy.SubjectKindFile,
	}
	activities := make(map[string]*syntax.ActivityDecl, len(nativeOperationInputs))
	entities := make(map[string]*syntax.EntityDecl, len(entityIDs))
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *syntax.EntityDecl:
			if _, duplicate := entities[declaration.Name]; duplicate {
				return operationInputContract{}, fmt.Errorf("duplicate operation input contract entity %q", declaration.Name)
			}
			entities[declaration.Name] = declaration
		case *syntax.ActivityDecl:
			if _, duplicate := activities[declaration.Name]; duplicate {
				return operationInputContract{}, fmt.Errorf("duplicate operation input contract activity %q", declaration.Name)
			}
			activities[declaration.Name] = declaration
		default:
			return operationInputContract{}, fmt.Errorf("unknown operation input contract declaration")
		}
	}
	if len(entities) != len(entityIDs) || len(activities) != len(nativeOperationInputs) {
		return operationInputContract{}, fmt.Errorf("operation input contract declaration count is not exact")
	}
	for name, expectedID := range entityIDs {
		entity, ok := entities[name]
		if !ok || entity.ID != expectedID {
			return operationInputContract{}, fmt.Errorf("operation input contract entity %q is not exact", name)
		}
		node, ok := ir.Graph.NodeByName(ir.Namespace, name)
		if !ok || node.Kind != semantic.Entity || node.ID.String() != expectedID {
			return operationInputContract{}, fmt.Errorf("operation input contract semantic entity %q is not exact", name)
		}
	}

	expectedActivities := make(map[string]nativeOperationInput, len(nativeOperationInputs))
	bindings := make(map[sourcepolicy.Operation]operationInputContractBinding, len(nativeOperationInputs))
	for _, native := range nativeOperationInputs {
		expectedActivities[native.Activity] = native
		activity, ok := activities[native.Activity]
		if !ok || len(activity.Inputs) != 1 || activity.Inputs[0].Name != native.InputEntity || activity.Output != native.OutputEntity {
			return operationInputContract{}, fmt.Errorf("operation input contract activity %q signature is not exact", native.Activity)
		}
		inputKind, inputKnown := inputKinds[activity.Inputs[0].Name]
		if !inputKnown || inputKind != native.InputKind {
			return operationInputContract{}, fmt.Errorf("operation input contract activity %q input kind conflicts with native ABI", native.Activity)
		}
		bindings[native.Operation] = operationInputContractBinding{InputSubjectKind: inputKind}
		activityNode, ok := ir.Graph.NodeByName(ir.Namespace, native.Activity)
		if !ok || activityNode.Kind != semantic.Activity {
			return operationInputContract{}, fmt.Errorf("operation input contract activity %q is not exact", native.Activity)
		}
		inputNode, ok := ir.Graph.NodeByName(ir.Namespace, native.InputEntity)
		if !ok || inputNode.Kind != semantic.Entity {
			return operationInputContract{}, fmt.Errorf("operation input contract input %q is not exact", native.InputEntity)
		}
		outputNode, ok := ir.Graph.NodeByName(ir.Namespace, native.OutputEntity)
		if !ok || outputNode.Kind != semantic.Entity {
			return operationInputContract{}, fmt.Errorf("operation input contract output %q is not exact", native.OutputEntity)
		}
	}
	for activityName := range activities {
		if _, ok := expectedActivities[activityName]; !ok {
			return operationInputContract{}, fmt.Errorf("unknown operation input contract activity %q", activityName)
		}
	}
	if err := validateOperationInputContractFacts(ir, activities, entities); err != nil {
		return operationInputContract{}, err
	}

	sourceDigest := sha256.Sum256(raw)
	return operationInputContract{
		SourceDigest:   hex.EncodeToString(sourceDigest[:]),
		SemanticDigest: ir.StableHash(),
		Bindings:       bindings,
	}, nil
}

func validateOperationInputContractFacts(ir semantic.IR, activities map[string]*syntax.ActivityDecl, entities map[string]*syntax.EntityDecl) error {
	if len(ir.Graph.Nodes()) != len(activities)+len(entities) {
		return fmt.Errorf("operation input contract semantic node count is not exact")
	}
	seen := make(map[string]struct{}, len(activities)*2)
	for _, native := range nativeOperationInputs {
		activity, ok := ir.Graph.NodeByName(ir.Namespace, native.Activity)
		input, inputOK := ir.Graph.NodeByName(ir.Namespace, native.InputEntity)
		output, outputOK := ir.Graph.NodeByName(ir.Namespace, native.OutputEntity)
		if !ok || !inputOK || !outputOK {
			return fmt.Errorf("operation input contract fact endpoint is missing for %q", native.Activity)
		}
		used := semantic.FactKey{Subject: activity.ID, Predicate: semantic.Used, Object: input.ID}
		generated := semantic.FactKey{Subject: output.ID, Predicate: semantic.WasGeneratedBy, Object: activity.ID}
		if !ir.Graph.HasFact(used) || !ir.Graph.HasFact(generated) {
			return fmt.Errorf("operation input contract facts are incomplete for %q", native.Activity)
		}
		seen[fmt.Sprintf("%s:%s:%s", used.Subject, used.Predicate, used.Object)] = struct{}{}
		seen[fmt.Sprintf("%s:%s:%s", generated.Subject, generated.Predicate, generated.Object)] = struct{}{}
	}
	if len(ir.Graph.AllFacts()) != len(seen) || len(seen) != len(nativeOperationInputs)*2 {
		return fmt.Errorf("operation input contract has unexpected facts")
	}
	return nil
}
