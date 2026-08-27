package verify

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const observationSchema = "gooo.partial-knowledge.recipe/v3"

var activityNames = []string{
	"ObserveExactPair", "ObserveDirectUnknown", "ObserveDependencyBlock",
	"ObserveInvariant", "ObserveMixedUnresolved",
}

var caseIDs = []string{
	"exact-pair", "direct-unknown", "dependency-blocked",
	"invariant-preservation", "mixed-unknown-and-blocked",
}

var metaOperations = []string{
	"compose-partial-knowledge", "compose-partial-knowledge", "compose-partial-knowledge", "preserve-known-invariant", "compose-partial-knowledge",
}

var proofChoices = []string{"COHERENCE", "FOUNDATION", "COHERENCE", "FOUNDATION", "REGRESSION"}

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
		return sourceModel{}, fmt.Errorf("independent source syntax diagnostics: %d", len(diagnostics))
	}
	if file.Package == nil || file.Package.Name != "partialknowledgecomposition" || file.Namespace == nil || file.Namespace.Name != "partialknowledgecomposition" {
		return sourceModel{}, fmt.Errorf("independent source package or namespace is not bound")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return sourceModel{}, fmt.Errorf("independent lower recipe source: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return sourceModel{}, fmt.Errorf("independent validate semantic IR: %w", err)
	}
	entities := map[string]bool{}
	for _, declaration := range file.Decls {
		if entity, ok := declaration.(*syntax.EntityDecl); ok {
			entities[entity.Name] = true
		}
	}
	for _, name := range []string{"MetaValue", "ObservationReceipt", "DirectUnknown", "DependencyBlocked", "InvariantOnly", "MixedUnresolved", "Foundation", "Coherence", "Regression"} {
		if !entities[name] {
			return sourceModel{}, fmt.Errorf("independent source vocabulary is missing entity %q", name)
		}
	}
	activities := make([]*syntax.ActivityDecl, 0, len(activityNames))
	for _, declaration := range file.Decls {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if ok && strings.HasPrefix(activity.Name, "Observe") {
			activities = append(activities, activity)
		}
	}
	if len(activities) != len(activityNames) {
		return sourceModel{}, fmt.Errorf("independent source recipe count = %d, want %d", len(activities), len(activityNames))
	}
	model := sourceModel{SemanticIRDigest: ir.StableHash(), Cases: make([]Case, 0, len(activityNames))}
	for index, name := range activityNames {
		activity := activities[index]
		if activity.Name != name || !activity.ValueProgramPresent || activity.ValueProgram == "" {
			return sourceModel{}, fmt.Errorf("independent source recipe %d is not computed", index+1)
		}
		node, ok := ir.Graph.NodeByName(ir.Namespace, activity.Name)
		if !ok || node.Kind != semantic.Activity || node.ValueProgram != activity.ValueProgram {
			return sourceModel{}, fmt.Errorf("independent lowering lost recipe %q", activity.Name)
		}
		current, err := parseObservation(activity.Name, node.ID.String(), activity.ValueProgram)
		if err != nil {
			return sourceModel{}, fmt.Errorf("independent recipe %q: %w", activity.Name, err)
		}
		if current.ID != caseIDs[index] || current.Producer != "partial-knowledge-producer" || current.Consumer != "partial-knowledge-composition-consumer" || current.MetaOperation != metaOperations[index] || current.ProofChoice != proofChoices[index] {
			return sourceModel{}, fmt.Errorf("independent recipe %q metadata is not fixed", activity.Name)
		}
		model.Cases = append(model.Cases, current)
	}
	return model, nil
}

func parseObservation(activityName, activityID, program string) (Case, error) {
	parts := strings.Split(program, "|")
	if len(parts) < 2 || parts[0] != observationSchema {
		return Case{}, fmt.Errorf("recipe schema is not %q", observationSchema)
	}
	values := make(map[string]string, len(parts)-1)
	for _, part := range parts[1:] {
		key, value, ok := strings.Cut(part, "=")
		if !ok || key == "" || strings.TrimSpace(key) != key {
			return Case{}, fmt.Errorf("recipe field %q is malformed", part)
		}
		if key == "observed" || key == "observed_available" || key == "invariant_evidence" || key == "state" || key == "decision" || key == "resolution" {
			return Case{}, fmt.Errorf("recipe contains an observation result or conclusion label")
		}
		if _, ok := allowedRecipeFields[key]; !ok {
			return Case{}, fmt.Errorf("recipe field %q is not a permitted recipe declaration", key)
		}
		if _, exists := values[key]; exists {
			return Case{}, fmt.Errorf("recipe field %q is duplicated", key)
		}
		values[key] = value
	}
	get := func(key string) (string, error) {
		value, ok := values[key]
		if !ok {
			return "", fmt.Errorf("recipe field %q is missing", key)
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
	proof, err := get("proof_choice")
	if err != nil {
		return Case{}, err
	}
	if !validProof(proof) {
		return Case{}, fmt.Errorf("proof choice %q is invalid", proof)
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
			return "", fmt.Errorf("recipe field %q is missing", prefix+"."+suffix)
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

func validProof(value string) bool {
	return value == "FOUNDATION" || value == "COHERENCE" || value == "REGRESSION"
}

func deriveState(value Evidence) string {
	if !value.ObservedAvailable {
		if value.Dependency != nil && (value.Dependency.State == "OPEN" || value.Dependency.State == "UNKNOWN") {
			return dependencyBlocked
		}
		return directUnknown
	}
	if value.InvariantEvidence != "" {
		return invariantOnly
	}
	return exact
}

const (
	exact             = "EXACT"
	directUnknown     = "DIRECT_UNKNOWN"
	dependencyBlocked = "DEPENDENCY_BLOCKED"
	invariantOnly     = "INVARIANT_ONLY"
	mixedUnresolved   = "MIXED_UNRESOLVED"
)
