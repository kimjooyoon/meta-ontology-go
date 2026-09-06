package valueexecution

import (
	"fmt"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// Plan is the compiled execution plan for a value-program graph. Its
// programs and bindings are deliberately private: callers may provide root
// inputs and observe detached evidence, but cannot manufacture execution
// authority by assembling public fields.
type Plan struct {
	Filename            string
	SourceDigest        string
	SemanticFingerprint string
	programs            map[string]Program
	bindings            []bidir.RuntimeBinding
}

// Execution is a detached summary of one plan run. Results contain evidence,
// not ProducedResult handles; the handles never leave the per-run store.
type Execution struct {
	Results     map[string]ResultEvidence `json:"results"`
	ApplyCalls  int                       `json:"apply_calls"`
	Deliveries  int                       `json:"deliveries"`
	Activities  []string                  `json:"activities"`
}

// CompilePlan parses, lowers, validates, and compiles every value activity in
// a source file. Runtime bindings are accepted only here; Compile continues to
// reject them so the existing single-activity API remains explicit.
func CompilePlan(filename string, source []byte) (Plan, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return Plan{}, failAt(ReasonSourceParseFailed, "PARSE", "parse-source", diagnostics.Error().Error())
	}
	document, err := bidir.DocumentFromSyntax(file)
	if err != nil {
		return Plan{}, failAt(ReasonSemanticBindingFailed, "LOWER", "bind-bidir-document", err.Error())
	}
	model, err := bidir.Get(document)
	if err != nil {
		return Plan{}, failAt(ReasonSemanticBindingFailed, "LOWER", "read-bidir-model", err.Error())
	}
	programs := make(map[string]Program)
	activities := 0
	for _, declaration := range document.Declarations {
		if declaration.Kind != bidir.ActivityKind {
			continue
		}
		activities++
		program, err := compileDocumentProgram(filename, source, document, model, declaration.Name)
		if err != nil {
			return Plan{}, err
		}
		if _, exists := programs[declaration.Name]; exists {
			return Plan{}, failAt(ReasonPlanInvalid, "PLAN", "duplicate-activity", declaration.Name)
		}
		programs[declaration.Name] = program
	}
	if activities == 0 {
		return Plan{}, failAt(ReasonPlanInvalid, "PLAN", "require-activity", "source declares no activities")
	}
	if err := validatePlanBindings(programs, model.RuntimeBindings); err != nil {
		return Plan{}, err
	}
	return Plan{Filename: filename, SourceDigest: digestBytes(source), SemanticFingerprint: bidir.SemanticFingerprint(model), programs: programs,
		bindings: append([]bidir.RuntimeBinding(nil), model.RuntimeBindings...)}, nil
}

// Execute runs one isolated plan instance. The map supplies exactly one
// Integer input for each root activity; bound activities must receive their
// input from a validated ProducedResult edge.
func (plan Plan) Execute(rootInputs map[string]int64) (Execution, error) {
	order, incoming, _, err := plan.executionOrder()
	if err != nil {
		return Execution{}, err
	}
	if err := validateExecutionBindings(plan.programs, plan.bindings); err != nil {
		return Execution{}, err
	}
	if err := validateRootInputs(plan.programs, incoming, rootInputs); err != nil {
		return Execution{}, err
	}
	values := make(map[string]ProducedResult, len(plan.programs))
	execution := Execution{Results: make(map[string]ResultEvidence, len(plan.programs)), Activities: append([]string(nil), order...)}
	for _, activity := range order {
		input, hasInput := rootInputs[activity]
		if !hasInput {
			for _, binding := range incoming[activity] {
				result, ok := values[string(binding.Producer.Activity.Name)]
				if !ok {
					return Execution{}, failAt(ReasonPlanExecutionFailed, "EXECUTE", "read-bound-input", activity)
				}
				if err := validateBindingResult(plan.programs[string(binding.Producer.Activity.Name)], plan.programs[activity], binding, result); err != nil {
					return Execution{}, err
				}
				input, err = integerResult(result)
				if err != nil {
					return Execution{}, err
				}
				execution.Deliveries++
				hasInput = true
			}
		}
		if !hasInput {
			return Execution{}, failAt(ReasonExternalInputMissing, "EXECUTE", "require-external-root-input", activity)
		}
		result, err := plan.programs[activity].executeResult([]int64{input}, func() { execution.ApplyCalls++ })
		if err != nil {
			return Execution{}, err
		}
		values[activity] = result
		execution.Results[activity] = result.Evidence()
	}
	return execution, nil
}

func compileDocumentProgram(filename string, source []byte, document bidir.Document, model bidir.Model, activityName string) (Program, error) {
	declaration, ok := activityDeclaration(document, activityName)
	if !ok {
		return Program{}, failAt(ReasonActivityNotFound, "LOWER", "resolve-activity", activityName)
	}
	if len(declaration.Inputs) != 1 || len(declaration.Outputs) != 1 {
		detail := fmt.Sprintf("inputs=%d outputs=%d", len(declaration.Inputs), len(declaration.Outputs))
		return Program{}, failAt(ReasonSignatureArityUnsupported, "TYPECHECK", "bind-operation-signature", detail)
	}
	programText, present := declaration.Attributes[bidir.ActivityValueProgramAttribute]
	if !present || programText == "" {
		return Program{}, failAt(ReasonProgramMissing, "LOWER", "read-computes-program", activityName)
	}
	modelProgram, ok := modelActivityProgram(model, activityName)
	if !ok || modelProgram != programText {
		return Program{}, failAt(ReasonSemanticBindingFailed, "LOWER", "preserve-computes-program", "activity value program was not preserved in the bidir model")
	}
	operationIR, implementation, err := compileOperation(activityName, programText, declaration)
	if err != nil {
		return Program{}, err
	}
	authority, err := newResultAuthority(model, activityName, source, programText, modelProgram, operationIR)
	if err != nil {
		return Program{}, err
	}
	return Program{Activity: activityName, Text: programText, Operation: operationIR, SourceDigest: digestBytes(source),
		SemanticFingerprint: bidir.SemanticFingerprint(model), ModelProgram: modelProgram, implementation: implementation,
		document: document, authority: authority}, nil
}

func validatePlanBindings(programs map[string]Program, bindings []bidir.RuntimeBinding) error {
	seenConsumers := make(map[string]struct{}, len(bindings))
	seenPairs := make(map[string]struct{}, len(bindings))
	for index, binding := range bindings {
		producer := string(binding.Producer.Activity.Name)
		consumer := string(binding.Consumer.Activity.Name)
		if _, ok := programs[producer]; !ok {
			return failAt(ReasonPlanInvalid, "PLAN", "resolve-binding-producer", fmt.Sprintf("binding %d: %s", index, producer))
		}
		if _, ok := programs[consumer]; !ok {
			return failAt(ReasonPlanInvalid, "PLAN", "resolve-binding-consumer", fmt.Sprintf("binding %d: %s", index, consumer))
		}
		key := producer + "\x00" + string(binding.Producer.Port.Name) + "\x00" + consumer + "\x00" + string(binding.Consumer.Port.Name)
		if _, ok := seenPairs[key]; ok {
			return failAt(ReasonPlanInvalid, "PLAN", "reject-binding-duplicate", key)
		}
		seenPairs[key] = struct{}{}
		consumerKey := consumer + "\x00" + string(binding.Consumer.Port.Name)
		if _, ok := seenConsumers[consumerKey]; ok {
			return failAt(ReasonPlanInvalid, "PLAN", "reject-binding-input-conflict", consumerKey)
		}
		seenConsumers[consumerKey] = struct{}{}
	}
	return nil
}

func (plan Plan) executionOrder() ([]string, map[string][]bidir.RuntimeBinding, map[string][]bidir.RuntimeBinding, error) {
	if err := validatePlanBindings(plan.programs, plan.bindings); err != nil {
		return nil, nil, nil, err
	}
	incoming := make(map[string][]bidir.RuntimeBinding, len(plan.programs))
	outgoing := make(map[string][]bidir.RuntimeBinding, len(plan.programs))
	indegree := make(map[string]int, len(plan.programs))
	for activity := range plan.programs {
		indegree[activity] = 0
	}
	for _, binding := range plan.bindings {
		producer, consumer := string(binding.Producer.Activity.Name), string(binding.Consumer.Activity.Name)
		if _, ok := plan.programs[producer]; !ok {
			return nil, nil, nil, failAt(ReasonPlanInvalid, "PLAN", "resolve-binding-producer", producer)
		}
		if _, ok := plan.programs[consumer]; !ok {
			return nil, nil, nil, failAt(ReasonPlanInvalid, "PLAN", "resolve-binding-consumer", consumer)
		}
		incoming[consumer] = append(incoming[consumer], binding)
		outgoing[producer] = append(outgoing[producer], binding)
		indegree[consumer]++
	}
	ready := make([]string, 0, len(indegree))
	for activity, degree := range indegree {
		if degree == 0 {
			ready = append(ready, activity)
		}
	}
	slices.Sort(ready)
	order := make([]string, 0, len(plan.programs))
	for len(ready) > 0 {
		activity := ready[0]
		ready = ready[1:]
		order = append(order, activity)
		for _, binding := range outgoing[activity] {
			consumer := string(binding.Consumer.Activity.Name)
			indegree[consumer]--
			if indegree[consumer] == 0 {
				ready = append(ready, consumer)
			}
		}
		slices.Sort(ready)
	}
	if len(order) != len(plan.programs) {
		return nil, nil, nil, failAt(ReasonPlanInvalid, "PLAN", "reject-binding-cycle", "runtime bindings are cyclic")
	}
	return order, incoming, outgoing, nil
}

func validateExecutionBindings(programs map[string]Program, bindings []bidir.RuntimeBinding) error {
	for index, binding := range bindings {
		producerName := string(binding.Producer.Activity.Name)
		consumerName := string(binding.Consumer.Activity.Name)
		producer, producerOK := programs[producerName]
		consumer, consumerOK := programs[consumerName]
		if !producerOK || !consumerOK {
			return failAt(ReasonPlanInvalid, "PLAN", "resolve-binding-authority", fmt.Sprintf("binding %d references an unknown activity", index))
		}
		if err := producer.validateResultAuthority(); err != nil {
			return failAt(ReasonBindingResultInvalid, "PLAN", "validate-producer-authority", err.Error())
		}
		if err := consumer.validateResultAuthority(); err != nil {
			return failAt(ReasonBindingResultInvalid, "PLAN", "validate-consumer-authority", err.Error())
		}
		if binding.Producer.Port.Name != "result" || binding.Consumer.Port.Name != "input" || binding.Entity == "" ||
			producer.Activity != producerName || consumer.Activity != consumerName ||
			producer.authority.outputEntityID != string(binding.Entity) || len(consumer.Operation.Spec.InputEntities) != 1 ||
			producer.authority.outputEntityName != consumer.Operation.Spec.InputEntities[0] {
			return failAt(ReasonBindingResultInvalid, "PLAN", "validate-binding-authority", fmt.Sprintf("binding %d does not match compiled source and operation specifications", index))
		}
	}
	return nil
}

func validateRootInputs(programs map[string]Program, incoming map[string][]bidir.RuntimeBinding, rootInputs map[string]int64) error {
	for activity := range rootInputs {
		if _, ok := programs[activity]; !ok {
			return failAt(ReasonExternalInputUnexpected, "EXECUTE", "reject-unknown-external-input", activity)
		}
		if len(incoming[activity]) != 0 {
			return failAt(ReasonExternalInputUnexpected, "EXECUTE", "reject-bound-external-input", activity)
		}
	}
	for activity := range programs {
		if len(incoming[activity]) == 0 {
			if _, ok := rootInputs[activity]; !ok {
				return failAt(ReasonExternalInputMissing, "EXECUTE", "require-external-root-input", activity)
			}
		}
	}
	return nil
}

func validateBindingResult(producer, consumer Program, binding bidir.RuntimeBinding, result ProducedResult) error {
	if err := producer.ValidateProducedResult(result); err != nil {
		return failAt(ReasonBindingResultInvalid, "EXECUTE", "validate-bound-result", err.Error())
	}
	if binding.Entity == "" || string(binding.Entity) != producer.authority.outputEntityID || producer.authority.outputEntityName != consumer.Operation.Spec.InputEntities[0] {
		return failAt(ReasonBindingResultInvalid, "TYPECHECK", "validate-bound-result-spec", "result authority does not match the declared source and consumer specification")
	}
	return nil
}

func integerResult(result ProducedResult) (int64, error) {
	value, err := result.Integer()
	if err != nil {
		return 0, failAt(ReasonBindingResultInvalid, "EXECUTE", "read-bound-integer", err.Error())
	}
	return int64(value), nil
}
