package partialknowledgecomposition

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type sourceModel struct {
	SemanticIRDigest string
	Cases            []Case
}

var allowedRecipeFields = map[string]struct{}{
	"case": {}, "producer": {}, "consumer": {}, "meta_operation": {}, "proof_choice": {},
	"left.operation": {}, "left.required": {}, "left.observation_recipe": {}, "left.dependency_recipe": {}, "left.invariant_capability": {},
	"right.operation": {}, "right.required": {}, "right.observation_recipe": {}, "right.dependency_recipe": {}, "right.invariant_capability": {},
}

func parseSource(sourcePath string, source []byte) (sourceModel, error) {
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if file == nil || diagnostics.HasErrors() {
		return sourceModel{}, fmt.Errorf("source syntax diagnostics prevent recipe reconstruction: %d", len(diagnostics))
	}
	if file.Package == nil || file.Package.Name != "partialknowledgecomposition" ||
		file.Namespace == nil || file.Namespace.Name != "partialknowledgecomposition" {
		return sourceModel{}, fmt.Errorf("source package or namespace is not bound")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return sourceModel{}, fmt.Errorf("lower partial-knowledge recipe source: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return sourceModel{}, fmt.Errorf("validate partial-knowledge semantic IR: %w", err)
	}

	entities := map[string]bool{}
	for _, declaration := range file.Decls {
		if entity, ok := declaration.(*syntax.EntityDecl); ok {
			entities[entity.Name] = true
		}
	}
	for _, name := range []string{"MetaValue", "ObservationReceipt", "DirectUnknown", "DependencyBlocked", "InvariantOnly", "MixedUnresolved", "Foundation", "Coherence", "Regression"} {
		if !entities[name] {
			return sourceModel{}, fmt.Errorf("source vocabulary is missing entity %q", name)
		}
	}

	activities := make([]*syntax.ActivityDecl, 0, len(fixedActivityNames))
	for _, declaration := range file.Decls {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || !strings.HasPrefix(activity.Name, "Observe") {
			continue
		}
		activities = append(activities, activity)
	}
	if len(activities) != FixedDenominator {
		return sourceModel{}, fmt.Errorf("source observation recipe count = %d, want %d", len(activities), FixedDenominator)
	}

	model := sourceModel{SemanticIRDigest: ir.StableHash(), Cases: make([]Case, 0, FixedDenominator)}
	for index, expectedName := range fixedActivityNames {
		activity := activities[index]
		if activity.Name != expectedName || !activity.ValueProgramPresent || activity.ValueProgram == "" {
			return sourceModel{}, fmt.Errorf("source observation %d is not the expected computed recipe", index+1)
		}
		node, ok := ir.Graph.NodeByName(ir.Namespace, activity.Name)
		if !ok || node.Kind != semantic.Activity || node.ValueProgram != activity.ValueProgram {
			return sourceModel{}, fmt.Errorf("semantic lowering lost computed recipe %q", activity.Name)
		}
		parsed, err := parseObservation(activity.Name, node.ID.String(), activity.ValueProgram)
		if err != nil {
			return sourceModel{}, fmt.Errorf("observation recipe %q: %w", activity.Name, err)
		}
		if parsed.ID != fixedCaseIDs[index] || parsed.Producer != Producer || parsed.Consumer != Consumer || parsed.MetaOperation != fixedMetaOperations[index] || parsed.ProofChoice != fixedProofChoices[index] {
			return sourceModel{}, fmt.Errorf("observation recipe %q metadata is not the fixed corpus binding", activity.Name)
		}
		model.Cases = append(model.Cases, parsed)
	}
	return model, nil
}

func parseObservation(activityName, activityID, program string) (Case, error) {
	parts := strings.Split(program, "|")
	if len(parts) < 2 || parts[0] != ObservationSchema {
		return Case{}, fmt.Errorf("computed recipe schema is not %q", ObservationSchema)
	}
	values := make(map[string]string, len(parts)-1)
	for _, part := range parts[1:] {
		key, value, ok := strings.Cut(part, "=")
		if !ok || key == "" {
			return Case{}, fmt.Errorf("computed recipe field %q is malformed", part)
		}
		if strings.TrimSpace(key) != key {
			return Case{}, fmt.Errorf("computed recipe field %q has surrounding whitespace", part)
		}
		if key == "observed" || key == "observed_available" || key == "invariant_evidence" || key == "state" || key == "decision" || key == "resolution" {
			return Case{}, fmt.Errorf("computed recipe contains an observation result or conclusion label")
		}
		if _, ok := allowedRecipeFields[key]; !ok {
			return Case{}, fmt.Errorf("computed recipe field %q is not a permitted recipe declaration", key)
		}
		if _, exists := values[key]; exists {
			return Case{}, fmt.Errorf("computed recipe field %q is duplicated", key)
		}
		values[key] = value
	}
	get := func(key string) (string, error) {
		value, ok := values[key]
		if !ok {
			return "", fmt.Errorf("computed recipe field %q is missing", key)
		}
		return value, nil
	}
	caseID, err := get("case")
	if err != nil {
		return Case{}, err
	}
	producer, err := get("producer")
	if err != nil {
		return Case{}, err
	}
	consumer, err := get("consumer")
	if err != nil {
		return Case{}, err
	}
	metaOperation, err := get("meta_operation")
	if err != nil {
		return Case{}, err
	}
	proofRaw, err := get("proof_choice")
	if err != nil {
		return Case{}, err
	}
	proof := ProofChoice(proofRaw)
	if !validProofChoice(proof) {
		return Case{}, fmt.Errorf("proof choice %q is invalid", proofRaw)
	}
	left, err := parseRecipeOperand(values, "left")
	if err != nil {
		return Case{}, err
	}
	right, err := parseRecipeOperand(values, "right")
	if err != nil {
		return Case{}, err
	}
	return Case{ID: caseID, SourceActivity: activityName, SourceActivityID: activityID, Producer: producer, Consumer: consumer, MetaOperation: metaOperation, ProofChoice: proof, Left: left, Right: right}, nil
}

func parseRecipeOperand(values map[string]string, prefix string) (RecipeOperand, error) {
	get := func(suffix string) (string, error) {
		value, ok := values[prefix+"."+suffix]
		if !ok {
			return "", fmt.Errorf("computed recipe field %q is missing", prefix+"."+suffix)
		}
		return value, nil
	}
	operation, err := get("operation")
	if err != nil {
		return RecipeOperand{}, err
	}
	required, err := get("required")
	if err != nil {
		return RecipeOperand{}, err
	}
	recipe, err := get("observation_recipe")
	if err != nil {
		return RecipeOperand{}, err
	}
	dependency, err := get("dependency_recipe")
	if err != nil {
		return RecipeOperand{}, err
	}
	invariant, err := get("invariant_capability")
	if err != nil {
		return RecipeOperand{}, err
	}
	if operation == "" || required == "" || dependency == "" || invariant == "" {
		return RecipeOperand{}, fmt.Errorf("%s recipe identity is malformed", prefix)
	}
	if recipe != "exact" && recipe != "missing" && recipe != "dependency" && recipe != "invariant" {
		return RecipeOperand{}, fmt.Errorf("%s observation recipe %q is invalid", prefix, recipe)
	}
	if recipe == "dependency" {
		if dependency == "none" || invariant != "none" {
			return RecipeOperand{}, fmt.Errorf("%s dependency recipe is not connected", prefix)
		}
	} else if dependency != "none" {
		return RecipeOperand{}, fmt.Errorf("%s non-dependent recipe carries a dependency", prefix)
	}
	if recipe == "invariant" {
		if invariant == "none" {
			return RecipeOperand{}, fmt.Errorf("%s invariant recipe lacks a capability", prefix)
		}
	} else if invariant != "none" {
		return RecipeOperand{}, fmt.Errorf("%s non-invariant recipe carries a capability", prefix)
	}
	return RecipeOperand{Operation: operation, Required: required, ObservationRecipe: recipe, DependencyRecipe: dependency, InvariantCapability: invariant}, nil
}
