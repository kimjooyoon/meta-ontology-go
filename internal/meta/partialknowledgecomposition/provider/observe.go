package provider

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	observationSchema = "gooo.partial-knowledge.recipe/v3"
	rawEvidenceSchema = "gooo/partial-knowledge/raw-evidence/v3"
	providerName      = "ci-partial-knowledge-observer"
	sourcePath        = "examples/partial-knowledge-composition/main.gooo"
)

var headPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

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
	Recipes          []Recipe
}

var allowedRecipeFields = map[string]struct{}{
	"case": {}, "producer": {}, "consumer": {}, "meta_operation": {}, "proof_choice": {},
	"left.operation": {}, "left.required": {}, "left.observation_recipe": {}, "left.dependency_recipe": {}, "left.invariant_capability": {},
	"right.operation": {}, "right.required": {}, "right.observation_recipe": {}, "right.dependency_recipe": {}, "right.invariant_capability": {},
}

func Observe(input Input) (RawEvidenceReceipt, error) {
	if input.Repository == "" || !headPattern.MatchString(input.HeadSHA) || input.SourcePath != sourcePath || len(input.Source) == 0 {
		return RawEvidenceReceipt{}, fmt.Errorf("observation identity is malformed")
	}
	model, err := parseSource(input.SourcePath, input.Source)
	if err != nil {
		return RawEvidenceReceipt{}, err
	}
	sourceDigest := digestBytes(input.Source)
	workspace, err := observeWorkspace(input)
	if err != nil {
		return RawEvidenceReceipt{}, err
	}
	workspaceDigest := workspace.EvidenceDigest
	cases := make([]RawCase, 0, len(model.Recipes))
	for _, recipe := range model.Recipes {
		cases = append(cases, RawCase{
			ID: recipe.ID, SourceActivity: recipe.SourceActivity, SourceActivityID: recipe.SourceActivityID,
			Producer: recipe.Producer, Consumer: recipe.Consumer, MetaOperation: recipe.MetaOperation,
			ProofChoice: recipe.ProofChoice,
			Left:        observeOperand(recipe.Left, sourceDigest, model.SemanticIRDigest, workspaceDigest, workspace.RepositoryWrites),
			Right:       observeOperand(recipe.Right, sourceDigest, model.SemanticIRDigest, workspaceDigest, workspace.RepositoryWrites),
		})
	}
	authority := CapabilityObservation{
		Name: "promotion-permission", Available: false, State: "UNKNOWN", Resolution: "LOWER_RESOLUTION",
		Stage: "CAPABILITY_OBSERVATION", Step: "CHECK_PROMOTION_PERMISSION",
		Reason: "PROMOTION_PERMISSION_EVIDENCE_NOT_SUPPLIED",
	}
	authority.EvidenceDigest = digestValue(authority)
	receipt := RawEvidenceReceipt{
		Schema: rawEvidenceSchema, Repository: input.Repository, HeadSHA: input.HeadSHA,
		SourcePath: input.SourcePath, SourceDigest: sourceDigest, SemanticIRDigest: model.SemanticIRDigest,
		SourceCases: len(model.Recipes), SourceCasesTotal: len(activityNames), Provider: providerName,
		Cases: cases, Workspace: workspace, Authority: authority,
	}
	receipt.Digest = receiptDigest(receipt)
	return receipt, nil
}

func parseSource(sourcePath string, source []byte) (sourceModel, error) {
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if file == nil || diagnostics.HasErrors() {
		return sourceModel{}, fmt.Errorf("observer source syntax diagnostics: %d", len(diagnostics))
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return sourceModel{}, fmt.Errorf("observer lower source: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return sourceModel{}, fmt.Errorf("observer validate semantic IR: %w", err)
	}
	activities := make([]*syntax.ActivityDecl, 0, len(activityNames))
	for _, declaration := range file.Decls {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if ok && strings.HasPrefix(activity.Name, "Observe") {
			activities = append(activities, activity)
		}
	}
	if len(activities) != len(activityNames) {
		return sourceModel{}, fmt.Errorf("observer source recipe count = %d, want %d", len(activities), len(activityNames))
	}
	model := sourceModel{SemanticIRDigest: ir.StableHash(), Recipes: make([]Recipe, 0, len(activityNames))}
	for index, expectedName := range activityNames {
		activity := activities[index]
		if activity.Name != expectedName || !activity.ValueProgramPresent || activity.ValueProgram == "" {
			return sourceModel{}, fmt.Errorf("observer source recipe %d is not computed", index+1)
		}
		node, ok := ir.Graph.NodeByName(ir.Namespace, activity.Name)
		if !ok || node.Kind != semantic.Activity || node.ValueProgram != activity.ValueProgram {
			return sourceModel{}, fmt.Errorf("observer lowering lost recipe %q", activity.Name)
		}
		recipe, err := parseRecipe(activity.Name, node.ID.String(), activity.ValueProgram)
		if err != nil {
			return sourceModel{}, fmt.Errorf("observer recipe %q: %w", activity.Name, err)
		}
		if recipe.ID != caseIDs[index] || recipe.Producer != "partial-knowledge-producer" || recipe.Consumer != "partial-knowledge-composition-consumer" || recipe.MetaOperation != metaOperations[index] || recipe.ProofChoice != proofChoices[index] {
			return sourceModel{}, fmt.Errorf("observer recipe %q metadata is not fixed", activity.Name)
		}
		model.Recipes = append(model.Recipes, recipe)
	}
	return model, nil
}

func parseRecipe(activityName, activityID, program string) (Recipe, error) {
	parts := strings.Split(program, "|")
	if len(parts) < 2 || parts[0] != observationSchema {
		return Recipe{}, fmt.Errorf("recipe schema is not %q", observationSchema)
	}
	values := make(map[string]string, len(parts)-1)
	for _, part := range parts[1:] {
		key, value, ok := strings.Cut(part, "=")
		if !ok || key == "" || strings.TrimSpace(key) != key {
			return Recipe{}, fmt.Errorf("recipe field %q is malformed", part)
		}
		if key == "observed" || key == "observed_available" || key == "invariant_evidence" || key == "state" || key == "decision" || key == "resolution" {
			return Recipe{}, fmt.Errorf("recipe contains an observation result or conclusion label")
		}
		if _, ok := allowedRecipeFields[key]; !ok {
			return Recipe{}, fmt.Errorf("recipe field %q is not a permitted recipe declaration", key)
		}
		if _, exists := values[key]; exists {
			return Recipe{}, fmt.Errorf("recipe field %q is duplicated", key)
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
		return Recipe{}, err
	}
	producer, err := get("producer")
	if err != nil {
		return Recipe{}, err
	}
	consumer, err := get("consumer")
	if err != nil {
		return Recipe{}, err
	}
	metaOperation, err := get("meta_operation")
	if err != nil {
		return Recipe{}, err
	}
	proof, err := get("proof_choice")
	if err != nil {
		return Recipe{}, err
	}
	if proof != "FOUNDATION" && proof != "COHERENCE" && proof != "REGRESSION" {
		return Recipe{}, fmt.Errorf("proof choice %q is invalid", proof)
	}
	left, err := parseRecipeOperand(values, "left")
	if err != nil {
		return Recipe{}, err
	}
	right, err := parseRecipeOperand(values, "right")
	if err != nil {
		return Recipe{}, err
	}
	return Recipe{ID: caseID, SourceActivity: activityName, SourceActivityID: activityID, Producer: producer, Consumer: consumer, MetaOperation: metaOperation, ProofChoice: proof, Left: left, Right: right}, nil
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

func observeWorkspace(input Input) (WorkspaceObservation, error) {
	beforeTracked, err := snapshotLines(input.BeforeTracked)
	if err != nil {
		return WorkspaceObservation{}, fmt.Errorf("read tracked pre-snapshot: %w", err)
	}
	beforeUntracked, err := snapshotLines(input.BeforeUntracked)
	if err != nil {
		return WorkspaceObservation{}, fmt.Errorf("read untracked pre-snapshot: %w", err)
	}
	beforeStatus, err := snapshotLines(input.BeforeStatus)
	if err != nil {
		return WorkspaceObservation{}, fmt.Errorf("read status pre-snapshot: %w", err)
	}
	afterTracked, err := snapshotLines(input.AfterTracked)
	if err != nil {
		return WorkspaceObservation{}, fmt.Errorf("read tracked post-snapshot: %w", err)
	}
	afterUntracked, err := snapshotLines(input.AfterUntracked)
	if err != nil {
		return WorkspaceObservation{}, fmt.Errorf("read untracked post-snapshot: %w", err)
	}
	afterStatus, err := snapshotLines(input.AfterStatus)
	if err != nil {
		return WorkspaceObservation{}, fmt.Errorf("read status post-snapshot: %w", err)
	}
	before := newSnapshot(beforeTracked, beforeUntracked, beforeStatus)
	after := newSnapshot(afterTracked, afterUntracked, afterStatus)
	changed := changedPaths(before, after)
	workspace := WorkspaceObservation{
		Before: before, After: after, ChangedPaths: changed, RepositoryWrites: len(changed),
		Stage: "CI_OBSERVATION", Step: "SNAPSHOT_TRACKED_AND_UNTRACKED", Reason: "PRE_POST_SNAPSHOT_COMPARED",
	}
	if len(changed) == 0 {
		workspace.Reason = "PRE_POST_SNAPSHOT_EQUAL"
	}
	workspace.EvidenceDigest = digestValue(workspace)
	return workspace, nil
}

func snapshotLines(raw []byte) ([]string, error) {
	values := map[string]struct{}{}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			values[line] = struct{}{}
		}
	}
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result, nil
}

func newSnapshot(tracked, untracked, status []string) Snapshot {
	snapshot := Snapshot{Tracked: tracked, Untracked: untracked, Status: status}
	snapshot.Digest = digestValue(struct {
		Tracked   []string `json:"tracked"`
		Untracked []string `json:"untracked"`
		Status    []string `json:"status"`
	}{snapshot.Tracked, snapshot.Untracked, snapshot.Status})
	return snapshot
}

func changedPaths(before, after Snapshot) []string {
	beforeSet := snapshotSet(before)
	afterSet := snapshotSet(after)
	values := map[string]struct{}{}
	for path := range beforeSet {
		if _, ok := afterSet[path]; !ok {
			values[path] = struct{}{}
		}
	}
	for path := range afterSet {
		if _, ok := beforeSet[path]; !ok {
			values[path] = struct{}{}
		}
	}
	result := make([]string, 0, len(values))
	for path := range values {
		result = append(result, path)
	}
	sort.Strings(result)
	return result
}

func snapshotSet(snapshot Snapshot) map[string]struct{} {
	values := map[string]struct{}{}
	for _, path := range snapshot.Tracked {
		values["tracked:"+path] = struct{}{}
	}
	for _, path := range snapshot.Untracked {
		values["untracked:"+path] = struct{}{}
	}
	for _, path := range snapshot.Status {
		values["status:"+path] = struct{}{}
	}
	return values
}

func observeOperand(recipe RecipeOperand, sourceDigest, semanticDigest, workspaceDigest string, repositoryWrites int) Evidence {
	evidence := Evidence{
		Operation: recipe.Operation, Required: recipe.Required, Stage: "CI_OBSERVATION", Step: "OBSERVE_RECIPE",
		Provenance: EvidenceProvenance{Provider: providerName, SourcePath: sourcePath, SourceDigest: sourceDigest, SemanticIRDigest: semanticDigest, WorkspaceSnapshotDigest: workspaceDigest},
	}
	switch recipe.ObservationRecipe {
	case "exact":
		evidence.Observed = recipe.Required
		evidence.ObservedAvailable = true
		evidence.Reason = "REQUIRED_VALUE_OBSERVED"
	case "missing":
		evidence.Reason = "REQUIRED_VALUE_NOT_OBSERVED"
	case "dependency":
		evidence.Dependency = upstreamClaim(recipe.DependencyRecipe, sourceDigest, semanticDigest, workspaceDigest)
		evidence.Reason = "UPSTREAM_CLAIM_UNRESOLVED"
	case "invariant":
		if repositoryWrites == 0 {
			evidence.Observed = recipe.Required
			evidence.ObservedAvailable = true
			evidence.InvariantEvidence = recipe.InvariantCapability
			evidence.Reason = "PRE_POST_SNAPSHOT_EQUAL"
		} else {
			evidence.Reason = "INVARIANT_NOT_OBSERVED"
		}
	}
	evidence.EvidenceDigest = digestValue(evidence)
	return evidence
}

func upstreamClaim(target, sourceDigest, semanticDigest, workspaceDigest string) *UpstreamClaim {
	claim := &UpstreamClaim{
		ClaimID:     "upstream/" + target,
		Proposition: "upstream evidence for " + target + " is available",
		Predicate:   "upstream-evidence-available", State: "OPEN", Resolution: "LOWER_RESOLUTION",
		Stage: "UPSTREAM_OBSERVATION", Step: "OBSERVE_DEPENDENCY_CLAIM",
		Reason: "UPSTREAM_EVIDENCE_UNAVAILABLE", RawSourceDigest: sourceDigest,
		SemanticDigest: semanticDigest, WorkspaceSnapshotDigest: workspaceDigest,
		TargetOperation: "bind-" + target, TargetOutput: "ObservationReceipt",
	}
	claim.PropositionDigest = digestValue(claim.Proposition)
	claim.EvidenceDigest = digestValue(struct {
		ClaimID                 string `json:"claim_id"`
		Proposition             string `json:"proposition"`
		PropositionDigest       string `json:"proposition_digest"`
		Predicate               string `json:"predicate"`
		State                   string `json:"state"`
		Resolution              string `json:"resolution"`
		Stage                   string `json:"stage"`
		Step                    string `json:"step"`
		Reason                  string `json:"reason"`
		RawSourceDigest         string `json:"raw_source_digest"`
		SemanticDigest          string `json:"semantic_digest"`
		WorkspaceSnapshotDigest string `json:"workspace_snapshot_digest"`
		TargetOperation         string `json:"target_operation"`
		TargetOutput            string `json:"target_output"`
	}{claim.ClaimID, claim.Proposition, claim.PropositionDigest, claim.Predicate, claim.State, claim.Resolution, claim.Stage, claim.Step, claim.Reason, claim.RawSourceDigest, claim.SemanticDigest, claim.WorkspaceSnapshotDigest, claim.TargetOperation, claim.TargetOutput})
	return claim
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) string {
	raw, _ := json.Marshal(value)
	return digestBytes(raw)
}

func receiptDigest(receipt RawEvidenceReceipt) string {
	receipt.Digest = ""
	return digestValue(receipt)
}
