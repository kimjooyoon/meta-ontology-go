package verify

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const observationSchema = "gooo.partial-knowledge.observation/v2"

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

func parseSource(sourcePath string, source []byte) (sourceModel, error) {
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if file == nil || diagnostics.HasErrors() {
		return sourceModel{}, fmt.Errorf("source syntax diagnostics prevent independent reconstruction: %d", len(diagnostics))
	}
	if file.Package == nil || file.Package.Name != "partialknowledgecomposition" || file.Namespace == nil || file.Namespace.Name != "partialknowledgecomposition" {
		return sourceModel{}, fmt.Errorf("source package or namespace is not bound")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return sourceModel{}, fmt.Errorf("independent source lower failed: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return sourceModel{}, fmt.Errorf("independent semantic IR validation failed: %w", err)
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
	activities := make([]*syntax.ActivityDecl, 0, len(activityNames))
	for _, declaration := range file.Decls {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if ok && strings.HasPrefix(activity.Name, "Observe") {
			activities = append(activities, activity)
		}
	}
	if len(activities) != len(activityNames) {
		return sourceModel{}, fmt.Errorf("source observation count = %d, want %d", len(activities), len(activityNames))
	}
	model := sourceModel{SemanticIRDigest: ir.StableHash(), Cases: make([]Case, 0, len(activityNames))}
	for index, name := range activityNames {
		activity := activities[index]
		if activity.Name != name || !activity.ValueProgramPresent || activity.ValueProgram == "" {
			return sourceModel{}, fmt.Errorf("source observation %d is not the expected computed receipt", index+1)
		}
		node, ok := ir.Graph.NodeByName(ir.Namespace, activity.Name)
		if !ok || node.Kind != semantic.Activity || node.ValueProgram != activity.ValueProgram {
			return sourceModel{}, fmt.Errorf("lowered source lost computed receipt %q", activity.Name)
		}
		current, err := parseObservation(activity.Name, node.ID.String(), activity.ValueProgram)
		if err != nil {
			return sourceModel{}, fmt.Errorf("observation %q: %w", activity.Name, err)
		}
		if current.ID != caseIDs[index] || current.Producer != "partial-knowledge-producer" || current.Consumer != "partial-knowledge-composition-consumer" || current.MetaOperation != metaOperations[index] || current.ProofChoice != proofChoices[index] {
			return sourceModel{}, fmt.Errorf("observation %q metadata is not the fixed corpus binding", activity.Name)
		}
		model.Cases = append(model.Cases, current)
	}
	return model, nil
}

func parseObservation(activityName, activityID, program string) (Case, error) {
	parts := strings.Split(program, "|")
	if len(parts) < 2 || parts[0] != observationSchema {
		return Case{}, fmt.Errorf("computed receipt schema is not %q", observationSchema)
	}
	values := make(map[string]string, len(parts)-1)
	for _, part := range parts[1:] {
		key, value, ok := strings.Cut(part, "=")
		if !ok || key == "" {
			return Case{}, fmt.Errorf("computed receipt field %q is malformed", part)
		}
		if key == "state" || key == "decision" || key == "resolution" {
			return Case{}, fmt.Errorf("computed receipt contains a conclusion label")
		}
		if _, exists := values[key]; exists {
			return Case{}, fmt.Errorf("computed receipt field %q is duplicated", key)
		}
		values[key] = value
	}
	get := func(key string) (string, error) {
		value, ok := values[key]
		if !ok {
			return "", fmt.Errorf("computed receipt field %q is missing", key)
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
	left, err := parseEvidence(values, "left")
	if err != nil {
		return Case{}, err
	}
	right, err := parseEvidence(values, "right")
	if err != nil {
		return Case{}, err
	}
	return Case{ID: caseID, SourceActivity: activityName, SourceActivityID: activityID, Producer: producer, Consumer: consumer, MetaOperation: metaOperation, ProofChoice: proof, Left: left, Right: right}, nil
}

func parseEvidence(values map[string]string, prefix string) (Evidence, error) {
	get := func(suffix string) (string, error) {
		value, ok := values[prefix+"."+suffix]
		if !ok {
			return "", fmt.Errorf("computed receipt field %q is missing", prefix+"."+suffix)
		}
		return value, nil
	}
	operation, err := get("operation")
	if err != nil {
		return Evidence{}, err
	}
	required, err := get("required")
	if err != nil {
		return Evidence{}, err
	}
	observed, err := get("observed")
	if err != nil {
		return Evidence{}, err
	}
	availableRaw, err := get("observed_available")
	if err != nil {
		return Evidence{}, err
	}
	available, err := strconv.ParseBool(availableRaw)
	if err != nil {
		return Evidence{}, fmt.Errorf("%s.observed_available is not boolean", prefix)
	}
	dependency, err := get("dependency_claim_id")
	if err != nil {
		return Evidence{}, err
	}
	prior, err := get("prior_state")
	if err != nil {
		return Evidence{}, err
	}
	invariant, err := get("invariant_evidence")
	if err != nil {
		return Evidence{}, err
	}
	value := Evidence{Operation: operation, Required: required, Observed: observed, ObservedAvailable: available, DependencyClaimID: dependency, PriorState: prior, InvariantEvidence: invariant}
	if value.Operation == "" || value.Required == "" || value.PriorState != "OPEN" {
		return Evidence{}, fmt.Errorf("%s evidence identity is malformed", prefix)
	}
	if value.ObservedAvailable {
		if value.Observed != value.Required || value.DependencyClaimID != "" {
			return Evidence{}, fmt.Errorf("%s available evidence is not exact", prefix)
		}
	} else if value.Observed != "" || value.InvariantEvidence != "" {
		return Evidence{}, fmt.Errorf("%s unavailable evidence carries an observation or invariant", prefix)
	}
	if value.InvariantEvidence != "" && !value.ObservedAvailable {
		return Evidence{}, fmt.Errorf("%s invariant evidence is unavailable", prefix)
	}
	return value, nil
}

func applyIntervention(model *sourceModel, mode string) (Intervention, error) {
	if mode == "" {
		mode = "none"
	}
	intervention := Intervention{Mode: mode}
	switch mode {
	case "none":
		intervention.Comment = "no counterfactual mutation"
	case "comment-only":
		intervention.Comment = "source comment only; semantic observations are unchanged"
	case "semantic":
		for index := range model.Cases {
			if model.Cases[index].ID != "direct-unknown" {
				continue
			}
			before := model.Cases[index].Left.ObservedAvailable
			model.Cases[index].Left.ObservedAvailable = true
			model.Cases[index].Left.Observed = model.Cases[index].Left.Required
			intervention.Semantic = true
			intervention.Target = "direct-unknown.left.observed_available"
			intervention.From = strconv.FormatBool(before)
			intervention.To = "true"
			return intervention, nil
		}
		return Intervention{}, fmt.Errorf("semantic intervention target is missing")
	default:
		return Intervention{}, fmt.Errorf("intervention %q is invalid", mode)
	}
	return intervention, nil
}

func deriveState(value Evidence) string {
	if !value.ObservedAvailable {
		if value.DependencyClaimID != "" {
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

func validProof(value string) bool {
	return value == "FOUNDATION" || value == "COHERENCE" || value == "REGRESSION"
}
