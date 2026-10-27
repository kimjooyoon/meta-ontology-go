package generation

import (
	"embed"
	"fmt"
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
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
	Name             string
	Activity         string
	InputEntity      string
	OutputEntity     string
	InputKind        sourcepolicy.SubjectKind
	HeaderCapability string
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
	{Name: "eligible-plain-import-group", Activity: "NormalizeEligibleImportGroup", InputEntity: "FileInput", OutputEntity: "ImportNormalizationPolicy", InputKind: sourcepolicy.SubjectKindFile, HeaderCapability: ImportHeaderPlainCapability},
	{Name: "named-alias-import-header", Activity: "NormalizeNamedAliasImportHeader", InputEntity: "FileInput", OutputEntity: "ImportNormalizationPolicy", InputKind: sourcepolicy.SubjectKindFile, HeaderCapability: ImportHeaderNamedAliasCapability},
}

type operationInputContractBinding struct {
	InputSubjectKind sourcepolicy.SubjectKind
	ActivityID       string
}

type operationInputContractFacts struct {
	UsedInput       bool
	GeneratedOutput bool
}

type operationInputContract struct {
	SourceDigest    string
	SemanticDigest  string
	SourceOrigins   sourcebundle.Provenance
	Entities        [len(nativeContractEntities)]contractEntityReference
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
	ActivityID          string `json:"ActivityID,omitempty"`
	InputEntity         string
	OutputEntity        string
	InputEntityID       string
	OutputEntityID      string
	InputSubjectKind    sourcepolicy.SubjectKind
	SourceDigest        string
	SemanticDigest      string
	SourceOrigins       sourcebundle.Provenance
	UsedInputFact       bool
	GeneratedOutputFact bool
	Obligations         []OperationInputContractObligationEvidence
}

type OperationInputContractObligationEvidence struct {
	Name                string
	Activity            string
	InputEntity         string
	OutputEntity        string
	InputEntityID       string
	OutputEntityID      string
	UsedInputFact       bool
	GeneratedOutputFact bool
}

type OperationInputContractPolicyEvidence struct {
	Name                string
	Activity            string
	InputEntity         string
	OutputEntity        string
	InputEntityID       string
	OutputEntityID      string
	InputSubjectKind    sourcepolicy.SubjectKind
	SourceDigest        string
	SemanticDigest      string
	SourceOrigins       sourcebundle.Provenance
	UsedInputFact       bool
	GeneratedOutputFact bool
	HeaderCapability    string
}

// This explicit embedded family can acquire source partitions without another
// Go source list. Every member still participates in exact ABI validation.
//
//go:embed operation-input-contract*.gooo
var operationInputContractFiles embed.FS

func loadOperationInputContract() (operationInputContract, error) {
	return loadOperationInputContractFS(operationInputContractFiles)
}

func loadOperationInputContractFS(files fs.FS) (operationInputContract, error) {
	if files == nil {
		return operationInputContract{}, fmt.Errorf("operation input contract filesystem is required")
	}
	names, err := fs.Glob(files, "operation-input-contract*.gooo")
	if err != nil {
		return operationInputContract{}, err
	}
	sources, err := sourcebundle.Read(files, names)
	if err != nil {
		return operationInputContract{}, err
	}
	return cachedOperationInputContract(sources)
}

func parseOperationInputContractSources(sources []sourcebundle.Source) (operationInputContract, error) {
	bundle, err := sourcebundle.Lower(sources)
	if err != nil {
		return operationInputContract{}, fmt.Errorf("lower operation input contract: %w", err)
	}
	if bundle.Package != "operationinputcontract" || bundle.Namespace != "operationinputcontract" {
		return operationInputContract{}, fmt.Errorf("operation input contract headers are not exact")
	}
	ir := bundle.IR

	activities := make(map[string]*syntax.ActivityDecl)
	entities := make(map[string]*syntax.EntityDecl)
	for _, declaration := range bundle.Declarations {
		switch declaration := declaration.(type) {
		case *syntax.EntityDecl:
			if entities[declaration.Name] != nil {
				return operationInputContract{}, fmt.Errorf("duplicate operation input contract entity %q", declaration.Name)
			}
			entities[declaration.Name] = declaration
		case *syntax.ActivityDecl:
			if activities[declaration.Name] != nil {
				return operationInputContract{}, fmt.Errorf("duplicate operation input contract activity %q", declaration.Name)
			}
			activities[declaration.Name] = declaration
		default:
			return operationInputContract{}, fmt.Errorf("unknown operation input contract declaration")
		}
	}
	expectedCount := len(nativeOperationInputs) + len(nativeOperationObligations) + len(nativeContractPolicies)
	if len(entities) != len(nativeContractEntities) || len(activities) != expectedCount ||
		len(ir.Graph.Nodes()) != len(entities)+len(activities) {
		return operationInputContract{}, fmt.Errorf("operation input contract declaration count is not exact")
	}
	sourceEntities, err := operationContractEntities(ir)
	if err != nil {
		return operationInputContract{}, err
	}
	inputKinds := map[string]sourcepolicy.SubjectKind{
		"FunctionInput": sourcepolicy.SubjectKindFunction, "FileInput": sourcepolicy.SubjectKindFile,
	}
	expectedActivities := make(map[string]bool, expectedCount)
	seen := make(map[semantic.FactKey]struct{}, expectedCount*2)
	bindings := make(map[sourcepolicy.Operation]operationInputContractBinding, len(nativeOperationInputs))
	facts := make(map[sourcepolicy.Operation]operationInputContractFacts, len(nativeOperationInputs))
	obligationFacts := make(map[string]operationInputContractFacts, len(nativeOperationObligations))
	policyFacts := make(map[string]operationInputContractFacts, len(nativeContractPolicies))
	for _, native := range nativeOperationInputs {
		inputKind, known := inputKinds[native.InputEntity]
		if !known || inputKind != native.InputKind {
			return operationInputContract{}, fmt.Errorf("operation input contract activity %q input kind conflicts with native ABI", native.Activity)
		}
		observed, err := operationContractRelation(ir, activities[native.Activity], native.Activity,
			native.InputEntity, native.OutputEntity, seen)
		if err != nil {
			return operationInputContract{}, err
		}
		expectedActivities[native.Activity] = true
		activity, _ := ir.Graph.NodeByName(ir.Namespace, native.Activity) // relation validation resolved this node above
		bindings[native.Operation] = operationInputContractBinding{InputSubjectKind: inputKind, ActivityID: activity.ID.String()}
		facts[native.Operation] = observed
	}
	for _, obligation := range nativeOperationObligations {
		observed, err := operationContractRelation(ir, activities[obligation.Activity], obligation.Activity,
			obligation.InputEntity, obligation.OutputEntity, seen)
		if err != nil {
			return operationInputContract{}, err
		}
		expectedActivities[obligation.Activity] = true
		obligationFacts[obligation.Activity] = observed
	}
	for _, policy := range nativeContractPolicies {
		if policy.HeaderCapability != ImportHeaderPlainCapability && policy.HeaderCapability != ImportHeaderNamedAliasCapability {
			return operationInputContract{}, fmt.Errorf("operation input contract policy %q header capability is not exact", policy.Name)
		}
		inputKind, known := inputKinds[policy.InputEntity]
		if !known || inputKind != policy.InputKind {
			return operationInputContract{}, fmt.Errorf("operation input contract policy %q input kind conflicts with native ABI", policy.Name)
		}
		observed, err := operationContractRelation(ir, activities[policy.Activity], policy.Activity,
			policy.InputEntity, policy.OutputEntity, seen)
		if err != nil {
			return operationInputContract{}, err
		}
		expectedActivities[policy.Activity] = true
		policyFacts[policy.Activity] = observed
	}
	for name := range activities {
		if !expectedActivities[name] {
			return operationInputContract{}, fmt.Errorf("unknown operation input contract activity %q", name)
		}
	}
	if len(ir.Graph.AllFacts()) != len(seen) || len(seen) != expectedCount*2 {
		return operationInputContract{}, fmt.Errorf("operation input contract has unexpected facts")
	}

	return operationInputContract{
		SourceDigest:    bundle.SourceDigest,
		SemanticDigest:  ir.StableHash(),
		SourceOrigins:   bundle.Provenance,
		Entities:        sourceEntities,
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
// semantic digests as well as to the Used/WasGeneratedBy facts. Entity labels
// retain authored names; the accompanying IDs bind their native roles.
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
		input, output := contract.entity(obligation.InputEntity), contract.entity(obligation.OutputEntity)
		obligations = append(obligations, OperationInputContractObligationEvidence{
			Name: obligation.Name, Activity: obligation.Activity, InputEntity: input.Name, OutputEntity: output.Name,
			InputEntityID: input.ID, OutputEntityID: output.ID,
			UsedInputFact: facts.UsedInput, GeneratedOutputFact: facts.GeneratedOutput,
		})
	}
	input, output := contract.entity(native.InputEntity), contract.entity(native.OutputEntity)
	return OperationInputContractEvidence{
		Operation:           native.Operation,
		Activity:            native.Activity,
		InputEntity:         input.Name,
		OutputEntity:        output.Name,
		InputEntityID:       input.ID,
		OutputEntityID:      output.ID,
		InputSubjectKind:    binding.InputSubjectKind,
		SourceDigest:        contract.SourceDigest,
		SemanticDigest:      contract.SemanticDigest,
		SourceOrigins:       contract.SourceOrigins,
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
	return importNormalizationPolicyEvidence("NormalizeEligibleImportGroup")
}

// ImportHeaderNormalizationPolicyEvidence exposes the contract relation that
// authorizes named-alias single-spec headers for generated source parts.
func ImportHeaderNormalizationPolicyEvidence() (OperationInputContractPolicyEvidence, error) {
	return importNormalizationPolicyEvidence("NormalizeNamedAliasImportHeader")
}

func importNormalizationPolicyEvidence(activity string) (OperationInputContractPolicyEvidence, error) {
	contract, err := loadOperationInputContract()
	if err != nil {
		return OperationInputContractPolicyEvidence{}, err
	}
	if len(nativeContractPolicies) != 2 {
		return OperationInputContractPolicyEvidence{}, fmt.Errorf("import normalization policy cardinality is not exact")
	}
	var policy nativeContractPolicy
	found := false
	for _, candidate := range nativeContractPolicies {
		if candidate.Activity != activity {
			continue
		}
		if found {
			return OperationInputContractPolicyEvidence{}, fmt.Errorf("import normalization policy activity is duplicated")
		}
		policy = candidate
		found = true
	}
	if !found {
		return OperationInputContractPolicyEvidence{}, fmt.Errorf("import normalization policy activity is not registered")
	}
	facts, ok := contract.PolicyFacts[policy.Activity]
	if !ok || !facts.UsedInput || !facts.GeneratedOutput {
		return OperationInputContractPolicyEvidence{}, fmt.Errorf("import normalization policy facts are incomplete")
	}
	input, output := contract.entity(policy.InputEntity), contract.entity(policy.OutputEntity)
	return OperationInputContractPolicyEvidence{
		Name: policy.Name, Activity: policy.Activity, InputEntity: input.Name, OutputEntity: output.Name,
		InputEntityID: input.ID, OutputEntityID: output.ID,
		InputSubjectKind: policy.InputKind, SourceDigest: contract.SourceDigest, SemanticDigest: contract.SemanticDigest,
		SourceOrigins: contract.SourceOrigins,
		UsedInputFact: facts.UsedInput, GeneratedOutputFact: facts.GeneratedOutput, HeaderCapability: policy.HeaderCapability,
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
