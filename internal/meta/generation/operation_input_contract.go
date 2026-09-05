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

type nativeOperationObligation struct {
	Name         string
	Activity     string
	InputEntity  string
	OutputEntity string
}

type nativeContractPolicy struct {
	Name         string
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

var nativeOperationObligations = []nativeOperationObligation{
	{Name: "return-shape", Activity: "ProveReturnShape", InputEntity: "FunctionInput", OutputEntity: "ReturnShapeObligation"},
	{Name: "control-flow", Activity: "ProveControlFlow", InputEntity: "ReturnShapeObligation", OutputEntity: "ControlFlowObligation"},
	{Name: "free-bindings", Activity: "ProveFreeBindings", InputEntity: "ControlFlowObligation", OutputEntity: "FreeBindingsObligation"},
	{Name: "callee-effects", Activity: "ProveCalleeEffects", InputEntity: "FreeBindingsObligation", OutputEntity: "CalleeEffectsObligation"},
	{Name: "rendered-capacity", Activity: "ProveRenderedCapacity", InputEntity: "CalleeEffectsObligation", OutputEntity: "RenderedCapacityObligation"},
	{Name: "projected-conformance", Activity: "ProveProjectedConformance", InputEntity: "RenderedCapacityObligation", OutputEntity: "ProjectedConformanceObligation"},
}

var nativeContractPolicies = []nativeContractPolicy{
	{Name: "eligible-plain-import-group", Activity: "NormalizeEligibleImportGroup", InputEntity: "FileInput", OutputEntity: "ImportNormalizationPolicy", InputKind: sourcepolicy.SubjectKindFile},
}

type operationInputContractBinding struct {
	InputSubjectKind sourcepolicy.SubjectKind
}

type operationInputContractFacts struct {
	UsedInput       bool
	GeneratedOutput bool
}

type operationInputContract struct {
	SourceDigest    string
	SemanticDigest  string
	Bindings        map[sourcepolicy.Operation]operationInputContractBinding
	Facts           map[sourcepolicy.Operation]operationInputContractFacts
	ObligationFacts map[string]operationInputContractFacts
	PolicyFacts     map[string]operationInputContractFacts
}

// OperationInputContractEvidence is the native admission evidence for an
// operation executor. The booleans are derived from the lowered semantic
// graph, rather than from the native table alone.
type OperationInputContractEvidence struct {
	Operation           sourcepolicy.Operation
	Activity            string
	InputEntity         string
	OutputEntity        string
	InputSubjectKind    sourcepolicy.SubjectKind
	SourceDigest        string
	SemanticDigest      string
	UsedInputFact       bool
	GeneratedOutputFact bool
	Obligations         []OperationInputContractObligationEvidence
}

type OperationInputContractObligationEvidence struct {
	Name                string
	Activity            string
	InputEntity         string
	OutputEntity        string
	UsedInputFact       bool
	GeneratedOutputFact bool
}

type OperationInputContractPolicyEvidence struct {
	Name                string
	Activity            string
	InputEntity         string
	OutputEntity        string
	InputSubjectKind    sourcepolicy.SubjectKind
	SourceDigest        string
	SemanticDigest      string
	UsedInputFact       bool
	GeneratedOutputFact bool
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
		"FunctionInput":                  "gooo://meta-operation-input-contract/entity/function-input",
		"FileInput":                      "gooo://meta-operation-input-contract/entity/file-input",
		"OperationResult":                "gooo://meta-operation-input-contract/entity/result",
		"ImportNormalizationPolicy":     "gooo://meta-operation-input-contract/entity/import-normalization-policy",
		"ReturnShapeObligation":          "gooo://meta-operation-input-contract/entity/return-shape-obligation",
		"ControlFlowObligation":          "gooo://meta-operation-input-contract/entity/control-flow-obligation",
		"FreeBindingsObligation":         "gooo://meta-operation-input-contract/entity/free-bindings-obligation",
		"CalleeEffectsObligation":        "gooo://meta-operation-input-contract/entity/callee-effects-obligation",
		"RenderedCapacityObligation":     "gooo://meta-operation-input-contract/entity/rendered-capacity-obligation",
		"ProjectedConformanceObligation": "gooo://meta-operation-input-contract/entity/projected-conformance-obligation",
	}
	inputKinds := map[string]sourcepolicy.SubjectKind{
		"FunctionInput": sourcepolicy.SubjectKindFunction,
		"FileInput":     sourcepolicy.SubjectKindFile,
	}
	activities := make(map[string]*syntax.ActivityDecl, len(nativeOperationInputs)+len(nativeOperationObligations)+len(nativeContractPolicies))
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
	if len(entities) != len(entityIDs) || len(activities) != len(nativeOperationInputs)+len(nativeOperationObligations)+len(nativeContractPolicies) {
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
	for _, obligation := range nativeOperationObligations {
		expectedActivities[obligation.Activity] = nativeOperationInput{Activity: obligation.Activity, InputEntity: obligation.InputEntity, OutputEntity: obligation.OutputEntity, InputKind: sourcepolicy.SubjectKindFunction}
		activity, ok := activities[obligation.Activity]
		if !ok || len(activity.Inputs) != 1 || activity.Inputs[0].Name != obligation.InputEntity || activity.Output != obligation.OutputEntity {
			return operationInputContract{}, fmt.Errorf("operation input contract obligation %q signature is not exact", obligation.Name)
		}
		if _, inputKnown := entityIDs[activity.Inputs[0].Name]; !inputKnown {
			return operationInputContract{}, fmt.Errorf("operation input contract obligation %q input entity is not exact", obligation.Name)
		}
		activityNode, ok := ir.Graph.NodeByName(ir.Namespace, obligation.Activity)
		if !ok || activityNode.Kind != semantic.Activity {
			return operationInputContract{}, fmt.Errorf("operation input contract obligation activity %q is not exact", obligation.Activity)
		}
		inputNode, ok := ir.Graph.NodeByName(ir.Namespace, obligation.InputEntity)
		if !ok || inputNode.Kind != semantic.Entity {
			return operationInputContract{}, fmt.Errorf("operation input contract obligation input %q is not exact", obligation.InputEntity)
		}
		outputNode, ok := ir.Graph.NodeByName(ir.Namespace, obligation.OutputEntity)
		if !ok || outputNode.Kind != semantic.Entity {
			return operationInputContract{}, fmt.Errorf("operation input contract obligation output %q is not exact", obligation.OutputEntity)
		}
	}
	for _, policy := range nativeContractPolicies {
		expectedActivities[policy.Activity] = nativeOperationInput{Activity: policy.Activity, InputEntity: policy.InputEntity, OutputEntity: policy.OutputEntity, InputKind: policy.InputKind}
		activity, ok := activities[policy.Activity]
		if !ok || len(activity.Inputs) != 1 || activity.Inputs[0].Name != policy.InputEntity || activity.Output != policy.OutputEntity {
			return operationInputContract{}, fmt.Errorf("operation input contract policy %q signature is not exact", policy.Name)
		}
		inputKind, inputKnown := inputKinds[activity.Inputs[0].Name]
		if !inputKnown || inputKind != policy.InputKind {
			return operationInputContract{}, fmt.Errorf("operation input contract policy %q input kind conflicts with native ABI", policy.Name)
		}
		activityNode, ok := ir.Graph.NodeByName(ir.Namespace, policy.Activity)
		if !ok || activityNode.Kind != semantic.Activity {
			return operationInputContract{}, fmt.Errorf("operation input contract policy %q is not exact", policy.Name)
		}
		inputNode, ok := ir.Graph.NodeByName(ir.Namespace, policy.InputEntity)
		if !ok || inputNode.Kind != semantic.Entity {
			return operationInputContract{}, fmt.Errorf("operation input contract policy input %q is not exact", policy.InputEntity)
		}
		outputNode, ok := ir.Graph.NodeByName(ir.Namespace, policy.OutputEntity)
		if !ok || outputNode.Kind != semantic.Entity {
			return operationInputContract{}, fmt.Errorf("operation input contract policy output %q is not exact", policy.OutputEntity)
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
	facts, obligationFacts, policyFacts, err := operationInputContractFactsFor(ir, len(activities)+len(entities))
	if err != nil {
		return operationInputContract{}, err
	}

	sourceDigest := sha256.Sum256(raw)
	return operationInputContract{
		SourceDigest:    hex.EncodeToString(sourceDigest[:]),
		SemanticDigest:  ir.StableHash(),
		Bindings:        bindings,
		Facts:           facts,
		ObligationFacts: obligationFacts,
		PolicyFacts:     policyFacts,
	}, nil
}

// ExtractFunctionInputContractEvidence exposes only the exact native
// ExtractFunction relation needed by the repository projection executor.
// Parsing and lowering the embedded .gooo contract happens in
// loadOperationInputContract, so admission remains bound to the source and
// semantic digests as well as to the Used/WasGeneratedBy facts.
func ExtractFunctionInputContractEvidence() (OperationInputContractEvidence, error) {
	contract, err := loadOperationInputContract()
	if err != nil {
		return OperationInputContractEvidence{}, err
	}
	native := nativeOperationInputFor(sourcepolicy.OperationExtractFunction)
	binding, bindingOK := contract.Bindings[native.Operation]
	facts, factsOK := contract.Facts[native.Operation]
	if !bindingOK || !factsOK || !facts.UsedInput || !facts.GeneratedOutput {
		return OperationInputContractEvidence{}, fmt.Errorf("ExtractFunction operation input contract facts are incomplete")
	}
	obligations := make([]OperationInputContractObligationEvidence, 0, len(nativeOperationObligations))
	for _, obligation := range nativeOperationObligations {
		facts, ok := contract.ObligationFacts[obligation.Activity]
		if !ok || !facts.UsedInput || !facts.GeneratedOutput {
			return OperationInputContractEvidence{}, fmt.Errorf("ExtractFunction proof obligation %q facts are incomplete", obligation.Name)
		}
		obligations = append(obligations, OperationInputContractObligationEvidence{
			Name: obligation.Name, Activity: obligation.Activity, InputEntity: obligation.InputEntity, OutputEntity: obligation.OutputEntity,
			UsedInputFact: facts.UsedInput, GeneratedOutputFact: facts.GeneratedOutput,
		})
	}
	return OperationInputContractEvidence{
		Operation:           native.Operation,
		Activity:            native.Activity,
		InputEntity:         native.InputEntity,
		OutputEntity:        native.OutputEntity,
		InputSubjectKind:    binding.InputSubjectKind,
		SourceDigest:        contract.SourceDigest,
		SemanticDigest:      contract.SemanticDigest,
		UsedInputFact:       facts.UsedInput,
		GeneratedOutputFact: facts.GeneratedOutput,
		Obligations:         obligations,
	}, nil
}

// ImportNormalizationPolicyEvidence exposes the exact contract relation the
// native helper renderer consumes before normalizing eligible import groups.
// It shares the operation-input contract source and semantic identities with
// ExtractFunction rather than introducing a second receipt or operation.
func ImportNormalizationPolicyEvidence() (OperationInputContractPolicyEvidence, error) {
	contract, err := loadOperationInputContract()
	if err != nil {
		return OperationInputContractPolicyEvidence{}, err
	}
	if len(nativeContractPolicies) != 1 {
		return OperationInputContractPolicyEvidence{}, fmt.Errorf("import normalization policy cardinality is not exact")
	}
	policy := nativeContractPolicies[0]
	facts, ok := contract.PolicyFacts[policy.Activity]
	if !ok || !facts.UsedInput || !facts.GeneratedOutput {
		return OperationInputContractPolicyEvidence{}, fmt.Errorf("import normalization policy facts are incomplete")
	}
	return OperationInputContractPolicyEvidence{
		Name: policy.Name, Activity: policy.Activity, InputEntity: policy.InputEntity, OutputEntity: policy.OutputEntity,
		InputSubjectKind: policy.InputKind, SourceDigest: contract.SourceDigest, SemanticDigest: contract.SemanticDigest,
		UsedInputFact: facts.UsedInput, GeneratedOutputFact: facts.GeneratedOutput,
	}, nil
}

func nativeOperationInputFor(operation sourcepolicy.Operation) nativeOperationInput {
	for _, native := range nativeOperationInputs {
		if native.Operation == operation {
			return native
		}
	}
	return nativeOperationInput{}
}

func validateOperationInputContractFacts(ir semantic.IR, activities map[string]*syntax.ActivityDecl, entities map[string]*syntax.EntityDecl) error {
	_, _, _, err := operationInputContractFactsFor(ir, len(activities)+len(entities))
	return err
}

func operationInputContractFactsFor(ir semantic.IR, expectedNodes int) (map[sourcepolicy.Operation]operationInputContractFacts, map[string]operationInputContractFacts, map[string]operationInputContractFacts, error) {
	if len(ir.Graph.Nodes()) != expectedNodes {
		return nil, nil, nil, fmt.Errorf("operation input contract semantic node count is not exact")
	}
	seen := make(map[string]struct{}, len(nativeOperationInputs)*2)
	facts := make(map[sourcepolicy.Operation]operationInputContractFacts, len(nativeOperationInputs))
	obligationFacts := make(map[string]operationInputContractFacts, len(nativeOperationObligations))
	policyFacts := make(map[string]operationInputContractFacts, len(nativeContractPolicies))
	for _, native := range nativeOperationInputs {
		activity, ok := ir.Graph.NodeByName(ir.Namespace, native.Activity)
		input, inputOK := ir.Graph.NodeByName(ir.Namespace, native.InputEntity)
		output, outputOK := ir.Graph.NodeByName(ir.Namespace, native.OutputEntity)
		if !ok || !inputOK || !outputOK {
			return nil, nil, nil, fmt.Errorf("operation input contract fact endpoint is missing for %q", native.Activity)
		}
		used := semantic.FactKey{Subject: activity.ID, Predicate: semantic.Used, Object: input.ID}
		generated := semantic.FactKey{Subject: output.ID, Predicate: semantic.WasGeneratedBy, Object: activity.ID}
		if !ir.Graph.HasFact(used) || !ir.Graph.HasFact(generated) {
			return nil, nil, nil, fmt.Errorf("operation input contract facts are incomplete for %q", native.Activity)
		}
		facts[native.Operation] = operationInputContractFacts{UsedInput: ir.Graph.HasFact(used), GeneratedOutput: ir.Graph.HasFact(generated)}
		seen[fmt.Sprintf("%s:%s:%s", used.Subject, used.Predicate, used.Object)] = struct{}{}
		seen[fmt.Sprintf("%s:%s:%s", generated.Subject, generated.Predicate, generated.Object)] = struct{}{}
	}
	for _, obligation := range nativeOperationObligations {
		activity, ok := ir.Graph.NodeByName(ir.Namespace, obligation.Activity)
		input, inputOK := ir.Graph.NodeByName(ir.Namespace, obligation.InputEntity)
		output, outputOK := ir.Graph.NodeByName(ir.Namespace, obligation.OutputEntity)
		if !ok || !inputOK || !outputOK {
			return nil, nil, nil, fmt.Errorf("operation input contract obligation fact endpoint is missing for %q", obligation.Activity)
		}
		used := semantic.FactKey{Subject: activity.ID, Predicate: semantic.Used, Object: input.ID}
		generated := semantic.FactKey{Subject: output.ID, Predicate: semantic.WasGeneratedBy, Object: activity.ID}
		if !ir.Graph.HasFact(used) || !ir.Graph.HasFact(generated) {
			return nil, nil, nil, fmt.Errorf("operation input contract obligation facts are incomplete for %q", obligation.Activity)
		}
		obligationFacts[obligation.Activity] = operationInputContractFacts{UsedInput: ir.Graph.HasFact(used), GeneratedOutput: ir.Graph.HasFact(generated)}
		seen[fmt.Sprintf("%s:%s:%s", used.Subject, used.Predicate, used.Object)] = struct{}{}
		seen[fmt.Sprintf("%s:%s:%s", generated.Subject, generated.Predicate, generated.Object)] = struct{}{}
	}
	for _, policy := range nativeContractPolicies {
		activity, ok := ir.Graph.NodeByName(ir.Namespace, policy.Activity)
		input, inputOK := ir.Graph.NodeByName(ir.Namespace, policy.InputEntity)
		output, outputOK := ir.Graph.NodeByName(ir.Namespace, policy.OutputEntity)
		if !ok || !inputOK || !outputOK {
			return nil, nil, nil, fmt.Errorf("operation input contract policy fact endpoint is missing for %q", policy.Activity)
		}
		used := semantic.FactKey{Subject: activity.ID, Predicate: semantic.Used, Object: input.ID}
		generated := semantic.FactKey{Subject: output.ID, Predicate: semantic.WasGeneratedBy, Object: activity.ID}
		if !ir.Graph.HasFact(used) || !ir.Graph.HasFact(generated) {
			return nil, nil, nil, fmt.Errorf("operation input contract policy facts are incomplete for %q", policy.Activity)
		}
		policyFacts[policy.Activity] = operationInputContractFacts{UsedInput: ir.Graph.HasFact(used), GeneratedOutput: ir.Graph.HasFact(generated)}
		seen[fmt.Sprintf("%s:%s:%s", used.Subject, used.Predicate, used.Object)] = struct{}{}
		seen[fmt.Sprintf("%s:%s:%s", generated.Subject, generated.Predicate, generated.Object)] = struct{}{}
	}
	if len(ir.Graph.AllFacts()) != len(seen) || len(seen) != (len(nativeOperationInputs)+len(nativeOperationObligations)+len(nativeContractPolicies))*2 {
		return nil, nil, nil, fmt.Errorf("operation input contract has unexpected facts")
	}
	return facts, obligationFacts, policyFacts, nil
}
