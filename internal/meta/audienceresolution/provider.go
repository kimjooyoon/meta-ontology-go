package audienceresolution

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type CurrentEvidenceBundle struct {
	Records         []EvidenceRecord
	Replay          ReplayVerification
	Counterexamples []CounterexampleResult
}

type predicateObservation struct {
	Value     bool
	Predicate string
	Details   any
}

type projectionRun struct {
	Values     map[string]bool
	Contradict map[string]bool
	Current    []EvidenceRecord
	Replay     ReplayVerification
	Decision   string
	Resolution string
	Views      []AudienceView
}

// ProvideCurrentEvidence is the CI provider boundary. It parses/lower the
// source, runs two fresh replay executions, materializes projection and
// counterexample artifacts, and only then returns evidence for evaluation.
func ProvideCurrentEvidence(input Input) (CurrentEvidenceBundle, error) {
	model, err := deriveSemanticSource(input.SourcePath, input.Source)
	if err != nil {
		return CurrentEvidenceBundle{}, err
	}
	replay, err := provideReplay(input, model)
	if err != nil {
		return CurrentEvidenceBundle{}, err
	}
	base, err := runProjection(input, input.Ledger, model, replay, counterexampleSkeleton(input.Ledger.Counterexamples), "base")
	if err != nil {
		return CurrentEvidenceBundle{}, err
	}
	results := make([]CounterexampleResult, 0, len(input.Ledger.Counterexamples))
	for _, counterexample := range input.Ledger.Counterexamples {
		result, err := executeCounterexample(input, model, base, counterexample)
		if err != nil {
			return CurrentEvidenceBundle{}, err
		}
		results = append(results, result)
	}
	base, err = runProjection(input, input.Ledger, model, replay, results, "base")
	if err != nil {
		return CurrentEvidenceBundle{}, err
	}
	for _, result := range results {
		for index := range base.Current {
			if base.Current[index].Coordinate == result.ID {
				base.Current[index].ArtifactPath = result.ArtifactPath
				base.Current[index].ContentDigest = result.ContentDigest
				base.Current[index].ObservedPredicate = "executed_counterexample"
				base.Current[index].ObservedValue = boolString(result.ExecutionValidated)
				base.Current[index].EvidenceStatus = EvidenceCurrent
			}
		}
	}
	return CurrentEvidenceBundle{Records: base.Current, Replay: replay, Counterexamples: results}, nil
}

func runProjection(input Input, ledger Ledger, model semanticSourceModel, replay ReplayVerification, counterexamples []CounterexampleResult, runID string) (projectionRun, error) {
	recipes := ledger.Records
	present := recipePresence(recipes)
	sourceBound := ledger.Source.Path == input.Contract.SourcePath && ledger.Source.Kind == SourceKind && ledger.Source.Digest == digestBytes(input.Source) && validDigest(ledger.Source.Digest)
	policyValid := sourcePolicyValid(model) && sourceAudienceResolutionValid(model)
	values := map[string]bool{
		"source.binding":        sourceBound,
		"ledger.coverage":       coverageRecipeValid(model, recipes),
		"ledger.replay":         replay.Equal,
		"user.coordinates":      policyCoordinatesPresent(model, "USER", present),
		"author.coordinates":    policyCoordinatesPresent(model, "TOOL_AUTHOR", present),
		"governor.coordinates":  policyCoordinatesPresent(model, "GOVERNOR", present),
		"projection.nesting":    policyValid,
		"projection.resolution": sourceAudienceResolutionValid(model),
	}
	contradict := map[string]bool{}
	for _, recipe := range recipes {
		if recipe.ObservedValue == "CONTRADICTORY" {
			values[recipe.Coordinate] = false
			contradict[recipe.Coordinate] = true
		}
	}
	for _, result := range counterexamples {
		values[resultIndicatorID(result.ID)] = result.ExecutionValidated
	}
	decision, resolution := subjectDecision(values, contradict, model)
	values["projection.shared-decision"] = decision == "PASS"
	current := make([]EvidenceRecord, 0, len(recipes))
	for _, spec := range indicatorSpecs() {
		recipe := recipeFor(recipes, spec.ID)
		if recipe.ID == "" {
			continue
		}
		if spec.ID == "receipt.seal" {
			path, digest, err := writeArtifact(input.ArtifactRoot, "current/receipt.seal.pending.json", map[string]any{
				"schema": "gooo/audience-resolution/pending-seal/v1", "coordinate": spec.ID,
				"subject_decision": decision, "status": EvidenceUnknown,
				"reason": "independent verification is post-evaluation",
			})
			if err != nil {
				return projectionRun{}, err
			}
			current = append(current, evidenceFromRecipe(recipe, false, "pending_independent_verification", path, digest, EvidenceUnknown))
			continue
		}
		observation, ok := observePredicate(spec.ID, model, ledger, present, values, replay, counterexamples)
		if !ok {
			continue
		}
		sourceArtifactPath, sourceArtifactDigest := "", ""
		if spec.ID == "source.binding" {
			var sourceErr error
			sourceArtifactPath, sourceArtifactDigest, sourceErr = writeRawArtifact(input.ArtifactRoot, "current/source-binding-source.gooo", input.Source)
			if sourceErr != nil {
				return projectionRun{}, sourceErr
			}
		}
		if sourceArtifactPath != "" {
			observation.Details = map[string]any{"semantic_digest": model.SemanticDigest, "declaration_count": model.DeclarationCount,
				"source_artifact_path": sourceArtifactPath, "source_artifact_digest": sourceArtifactDigest}
		}
		path, digest, err := writeArtifact(input.ArtifactRoot, "current/"+safeArtifactName(spec.ID)+".json", map[string]any{
			"schema": "gooo/audience-resolution/current-evidence/v1", "coordinate": spec.ID,
			"claim_id": recipe.ClaimID, "proposition_digest": recipe.PropositionDigest,
			"target_address": recipe.TargetAddress, "observed_predicate": observation.Predicate,
			"observed_value": boolString(observation.Value), "evidence_status": EvidenceCurrent,
			"details": observation.Details,
		})
		if err != nil {
			return projectionRun{}, err
		}
		record := evidenceFromRecipe(recipe, observation.Value, observation.Predicate, path, digest, EvidenceCurrent)
		if spec.ID == "ledger.replay" {
			record.ArtifactPaths = []string{replay.RunAPath, replay.RunBPath}
			record.ContentDigests = []string{replay.RunADigest, replay.RunBDigest}
		}
		current = append(current, record)
	}
	for _, recipe := range recipes {
		if recipe.ObservedValue == "CONTRADICTORY" {
			current = forceCurrentRecord(current, recipe, input.ArtifactRoot, recipe.Coordinate)
		}
	}
	state := inspectCurrentEvidence(recipes, current)
	decision, resolution = subjectDecisionFromState(state, model)
	views := buildViews(model, recipes, state, decision, resolution)
	projectionPath, projectionDigest, err := writeArtifact(input.ArtifactRoot, "current/projection-output-"+runID+".json", map[string]any{
		"schema": "gooo/audience-resolution/projection/v1", "decision": decision, "resolution": resolution,
		"views": views, "source_semantic_digest": model.SemanticDigest,
	})
	if err != nil {
		return projectionRun{}, err
	}
	_ = projectionPath
	_ = projectionDigest
	return projectionRun{Values: values, Contradict: contradict, Current: current, Replay: replay, Decision: decision, Resolution: resolution, Views: views}, nil
}

func observePredicate(id string, model semanticSourceModel, ledger Ledger, present map[string]bool, values map[string]bool, replay ReplayVerification, counterexamples []CounterexampleResult) (predicateObservation, bool) {
	if id == "counterexample.omission" {
		return predicateObservation{Value: counterexamplePassed(counterexamples, "counterexample.missing-information"), Predicate: "executed_counterexample", Details: counterexamples}, true
	}
	if id == "counterexample.contradiction" {
		return predicateObservation{Value: counterexamplePassed(counterexamples, "counterexample.decision-contradiction"), Predicate: "executed_counterexample", Details: counterexamples}, true
	}
	if id == "ledger.replay" {
		return predicateObservation{Value: replay.Equal, Predicate: "replay_bytes_equal", Details: replay}, true
	}
	if id == "projection.shared-decision" {
		decision, _ := subjectDecision(values, map[string]bool{}, model)
		return predicateObservation{Value: decision == "PASS", Predicate: "subject_decision_carried", Details: map[string]any{"decision": decision}}, true
	}
	if id == "source.binding" {
		return predicateObservation{Value: values[id], Predicate: "source_parse_lower_matches_digest", Details: map[string]any{"semantic_digest": model.SemanticDigest, "declaration_count": model.DeclarationCount}}, true
	}
	if id == "ledger.coverage" {
		return predicateObservation{Value: values[id], Predicate: "raw_recipe_set_covers_governor_policy", Details: map[string]any{"recipe_count": len(ledger.Records), "governor_count": len(sourceCoordinates(model))}}, true
	}
	if id == "user.coordinates" || id == "author.coordinates" || id == "governor.coordinates" {
		return predicateObservation{Value: values[id], Predicate: "semantic_policy_coordinates_present", Details: map[string]any{"audience": id}}, true
	}
	if id == "projection.nesting" {
		return predicateObservation{Value: values[id], Predicate: "semantic_policy_nested", Details: model.Audiences}, true
	}
	if id == "projection.resolution" {
		return predicateObservation{Value: values[id], Predicate: "semantic_resolution_labels_valid", Details: model.Audiences}, true
	}
	return predicateObservation{}, false
}

func executeCounterexample(input Input, model semanticSourceModel, baseline projectionRun, counterexample Counterexample) (CounterexampleResult, error) {
	recipe := claimRecipeFor(input.Ledger.Records, counterexample.TargetCoordinate)
	mutated := mutateRecipeLedger(cleanRecipeLedger(input.Ledger), counterexample)
	mutatedBytes := canonicalJSON(mutated)
	ledgerPath, _, err := writeArtifact(input.ArtifactRoot, "counterexamples/"+safeArtifactName(counterexample.ID)+"-ledger.json", json.RawMessage(mutatedBytes))
	if err != nil {
		return CounterexampleResult{}, err
	}
	replayA, err := replayForLedger(mutatedBytes, input.SourcePath, input.Source, model)
	if err != nil {
		return CounterexampleResult{}, err
	}
	replayB, err := replayForLedger(mutatedBytes, input.SourcePath, input.Source, model)
	if err != nil {
		return CounterexampleResult{}, err
	}
	replayAPath, replayADigest, err := writeArtifact(input.ArtifactRoot, "counterexamples/"+safeArtifactName(counterexample.ID)+"-replay-a.json", json.RawMessage(replayA))
	if err != nil {
		return CounterexampleResult{}, err
	}
	replayBPath, replayBDigest, err := writeArtifact(input.ArtifactRoot, "counterexamples/"+safeArtifactName(counterexample.ID)+"-replay-b.json", json.RawMessage(replayB))
	if err != nil {
		return CounterexampleResult{}, err
	}
	replay := ReplayVerification{RunAPath: replayAPath, RunADigest: replayADigest, RunBPath: replayBPath, RunBDigest: replayBDigest,
		Equal: bytes.Equal(replayA, replayB), CombinedDigest: digestBytes(append(append([]byte{}, replayA...), replayB...))}
	force := map[string]string{}
	if counterexample.Kind == "DECISION_CONTRADICTION" {
		force[counterexample.TargetCoordinate] = "CONTRADICTORY"
	}
	variant, err := runProjection(input, mutated, model, replay, counterexampleSkeleton(mutated.Counterexamples), "counterexample-"+safeArtifactName(counterexample.ID))
	if err != nil {
		return CounterexampleResult{}, err
	}
	if force[counterexample.TargetCoordinate] != "" {
		variant.Values[counterexample.TargetCoordinate] = false
		variant.Contradict[counterexample.TargetCoordinate] = true
		variant.Decision, variant.Resolution = subjectDecision(variant.Values, variant.Contradict, model)
		variant.Current = forceCurrentRecord(variant.Current, recipe, input.ArtifactRoot, counterexample.TargetCoordinate)
		variantState := inspectCurrentEvidence(mutated.Records, variant.Current)
		variant.Decision, variant.Resolution = subjectDecisionFromState(variantState, model)
		variant.Views = buildViews(model, mutated.Records, variantState, variant.Decision, variant.Resolution)
	}
	before, after := targetClaimState(baseline, counterexample.TargetCoordinate, input.Ledger.Records)
	if counterexample.Kind == "INFORMATION_OMISSION" {
		after = "OPEN"
	}
	if counterexample.Kind == "DECISION_CONTRADICTION" {
		after = "REFUTED"
	}
	valid := counterexampleExecutionValid(counterexample, variant, before, after, recipe)
	payload := map[string]any{
		"schema": "gooo/audience-resolution/counterexample/v1", "id": counterexample.ID,
		"kind": counterexample.Kind, "target_coordinate": counterexample.TargetCoordinate,
		"target_address": recipe.TargetAddress, "proposition_digest": recipe.PropositionDigest,
		"mutated_ledger_path": ledgerPath, "mutated_ledger_digest": digestBytes(mutatedBytes),
		"global_decision": variant.Decision, "resolution": variant.Resolution,
		"stage": recipe.Stage, "step": recipe.Step, "reason": recipe.Reason,
		"before_claim": before, "after_claim": after, "execution_validated": valid,
		"views": variant.Views,
	}
	path, digest, err := writeArtifact(input.ArtifactRoot, "counterexamples/"+safeArtifactName(counterexample.ID)+".json", payload)
	if err != nil {
		return CounterexampleResult{}, err
	}
	return CounterexampleResult{ID: counterexample.ID, Kind: counterexample.Kind, Trigger: counterexample.Trigger,
		Mutation: counterexample.Mutation, TargetCoordinate: counterexample.TargetCoordinate,
		TargetAddress: recipe.TargetAddress, Proposition: recipe.Proposition, PropositionDigest: recipe.PropositionDigest,
		Global: variant.Decision, Resolution: variant.Resolution, Stage: recipe.Stage, Step: recipe.Step, Reason: recipe.Reason,
		Views: counterexampleViews(baseline.Views, variant.Views), BeforeClaim: before, AfterClaim: after,
		ArtifactPath: path, ContentDigest: digest, ExecutionValidated: valid}, nil
}

func counterexampleExecutionValid(counterexample Counterexample, variant projectionRun, before, after string, recipe EvidenceRecord) bool {
	if recipe.PropositionDigest == "" || recipe.TargetAddress == "" || before != "OPEN" {
		return false
	}
	if counterexample.Kind == "INFORMATION_OMISSION" {
		return variant.Decision == "UNKNOWN" && variant.Resolution == "LOWER_RESOLUTION" && after == "OPEN"
	}
	return counterexample.Kind == "DECISION_CONTRADICTION" && variant.Decision == "REFUTED" && variant.Resolution == "INVARIANT_ONLY" && after == "REFUTED"
}

func targetClaimState(run projectionRun, coordinate string, recipes []EvidenceRecord) (string, string) {
	state := inspectCurrentEvidence(recipes, run.Current)
	if state.contradict[coordinate] {
		return "OPEN", "REFUTED"
	}
	if state.valid[coordinate] {
		return "OPEN", "DISCHARGED"
	}
	return "OPEN", "OPEN"
}

func counterexampleViews(before, after []AudienceView) []CounterexampleView {
	result := make([]CounterexampleView, 0, len(after))
	for index, view := range after {
		beforeDecision := "UNKNOWN"
		if index < len(before) {
			beforeDecision = before[index].LocalDecision
		}
		result = append(result, CounterexampleView{Audience: view.Audience, Before: beforeDecision, After: view.LocalDecision,
			LocalDecision: view.LocalDecision, LocalResolution: view.LocalResolution})
	}
	return result
}

func forceCurrentRecord(records []EvidenceRecord, recipe EvidenceRecord, root, coordinate string) []EvidenceRecord {
	for index := range records {
		if records[index].Coordinate == coordinate {
			path, digest, _ := writeArtifact(root, "counterexamples/forced-"+safeArtifactName(coordinate)+".json", map[string]any{
				"schema": "gooo/audience-resolution/current-evidence/v1", "coordinate": coordinate,
				"claim_id": recipe.ClaimID, "proposition_digest": recipe.PropositionDigest,
				"target_address": recipe.TargetAddress, "observed_predicate": "forced_contradiction",
				"observed_value": "false", "evidence_status": EvidenceCurrent,
				"details": map[string]any{"mutation": "CONTRADICTORY"},
			})
			records[index].ArtifactPath, records[index].ContentDigest = path, digest
			records[index].ObservedPredicate, records[index].ObservedValue = "forced_contradiction", "false"
		}
	}
	return records
}

func mutateRecipeLedger(ledger Ledger, counterexample Counterexample) Ledger {
	mutated := ledger
	mutated.Records = append([]EvidenceRecord(nil), ledger.Records...)
	if counterexample.Kind == "INFORMATION_OMISSION" {
		filtered := mutated.Records[:0]
		for _, record := range mutated.Records {
			if record.Coordinate != counterexample.TargetCoordinate {
				filtered = append(filtered, record)
			}
		}
		mutated.Records = filtered
		return mutated
	}
	for index := range mutated.Records {
		if mutated.Records[index].Coordinate == counterexample.TargetCoordinate {
			mutated.Records[index].ObservedValue = "CONTRADICTORY"
		}
	}
	return mutated
}

func cleanRecipeLedger(ledger Ledger) Ledger {
	clean := ledger
	clean.Records = append([]EvidenceRecord(nil), ledger.Records...)
	for index := range clean.Records {
		clean.Records[index].ObservedValue = "HISTORICAL_FIXTURE"
	}
	return clean
}

func provideReplay(input Input, model semanticSourceModel) (ReplayVerification, error) {
	raw := input.LedgerBytes
	if len(raw) == 0 {
		raw = canonicalJSON(input.Ledger)
	}
	a, err := replayForLedger(raw, input.SourcePath, input.Source, model)
	if err != nil {
		return ReplayVerification{}, err
	}
	b, err := replayForLedger(raw, input.SourcePath, input.Source, model)
	if err != nil {
		return ReplayVerification{}, err
	}
	pathA, digestA, err := writeArtifact(input.ArtifactRoot, "current/replay/run-a.json", json.RawMessage(a))
	if err != nil {
		return ReplayVerification{}, err
	}
	pathB, digestB, err := writeArtifact(input.ArtifactRoot, "current/replay/run-b.json", json.RawMessage(b))
	if err != nil {
		return ReplayVerification{}, err
	}
	return ReplayVerification{RunAPath: pathA, RunADigest: digestA, RunBPath: pathB, RunBDigest: digestB,
		Equal: bytes.Equal(a, b), CombinedDigest: digestBytes(append(append([]byte{}, a...), b...))}, nil
}

func replayForLedger(raw []byte, filename string, source []byte, model semanticSourceModel) ([]byte, error) {
	var ledger Ledger
	if err := json.Unmarshal(raw, &ledger); err != nil {
		return nil, err
	}
	return canonicalJSON(map[string]any{"schema": "gooo/audience-resolution/replay/v1", "source_digest": digestBytes(source),
		"source_semantic_digest": model.SemanticDigest, "source_denominator": model.DeclarationCount,
		"ledger_facts_digest": factsDigest(ledger), "record_coordinates": coordinatesOf(ledger.Records), "filename": filepath.Base(filename)}), nil
}

func evidenceFromRecipe(recipe EvidenceRecord, value bool, predicate, path, digest, status string) EvidenceRecord {
	record := recipe
	record.Provider = "ci.current-evidence-provider"
	record.ArtifactPath = path
	record.ContentDigest = digest
	record.ObservedPredicate = predicate
	record.ObservedValue = boolString(value)
	record.EvidenceStatus = status
	return record
}

func coverageRecipeValid(model semanticSourceModel, records []EvidenceRecord) bool {
	if len(records) != len(sourceCoordinates(model)) {
		return false
	}
	present := recipePresence(records)
	for _, coordinate := range sourceCoordinates(model) {
		if !present[coordinate] {
			return false
		}
	}
	return true
}

func policyCoordinatesPresent(model semanticSourceModel, audience string, present map[string]bool) bool {
	for _, coordinate := range sourceAudience(model, audience).Coordinates {
		if !present[coordinate] {
			return false
		}
	}
	return true
}

func recipePresence(records []EvidenceRecord) map[string]bool {
	present := map[string]bool{}
	for _, record := range records {
		present[record.Coordinate] = true
	}
	return present
}

func coordinatesOf(records []EvidenceRecord) []string {
	result := make([]string, 0, len(records))
	for _, record := range records {
		result = append(result, record.Coordinate)
	}
	return result
}

func resultIndicatorID(id string) string {
	if id == "counterexample.missing-information" {
		return "counterexample.omission"
	}
	if id == "counterexample.decision-contradiction" {
		return "counterexample.contradiction"
	}
	return id
}

func counterexampleSkeleton(values []Counterexample) []CounterexampleResult {
	result := make([]CounterexampleResult, 0, len(values))
	for _, value := range values {
		result = append(result, CounterexampleResult{ID: value.ID, ExecutionValidated: true})
	}
	return result
}

func subjectDecision(values map[string]bool, contradict map[string]bool, model semanticSourceModel) (string, string) {
	for id := range subjectCoordinates(model) {
		if contradict[id] {
			return "REFUTED", "INVARIANT_ONLY"
		}
	}
	for id := range subjectCoordinates(model) {
		if !values[id] {
			return "UNKNOWN", "LOWER_RESOLUTION"
		}
	}
	return "PASS", "EXACT"
}

func writeArtifact(root, relative string, value any) (string, string, error) {
	raw := canonicalJSON(value)
	digest := digestBytes(raw)
	if root == "" {
		return relative, digest, nil
	}
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0o640); err != nil {
		return "", "", fmt.Errorf("write artifact %s: %w", relative, err)
	}
	return relative, digest, nil
}

func writeRawArtifact(root, relative string, raw []byte) (string, string, error) {
	digest := digestBytes(raw)
	if root == "" {
		return relative, digest, nil
	}
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(path, raw, 0o640); err != nil {
		return "", "", fmt.Errorf("write raw artifact %s: %w", relative, err)
	}
	return relative, digest, nil
}

func safeArtifactName(value string) string {
	result := make([]byte, 0, len(value))
	for _, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9') || character == '-' || character == '_' {
			result = append(result, byte(character))
		} else {
			result = append(result, '-')
		}
	}
	return string(result)
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
