package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type independentReceipt struct {
	Schema                     string                      `json:"schema"`
	Decision                   string                      `json:"decision"`
	ProjectionDigest           string                      `json:"projection_digest"`
	DenominatorReconciliations []DenominatorReconciliation `json:"denominator_reconciliations"`
	Predicates                 []PredicateObservation      `json:"predicates"`
}

type fixtureMeasurement struct {
	SourceDigests []SourceDigestComparison
	Changed       []string
	Outputs       []OutputMetadata
	Consumer      independentReceipt
}

func runProof(root, outputDir string) (Evidence, error) {
	before, beforeErr := gitStatus(root)
	evidence, bodyErr := runProofBody(root, outputDir)
	after, afterErr := gitStatus(root)
	evidence.RepositoryNetState = repositoryObservation(before, after, beforeErr, afterErr)
	if beforeErr != nil || afterErr != nil {
		evidence.Decision = "FAIL_CLOSED"
		evidence.Reason = "REPOSITORY_OBSERVATION_UNKNOWN"
	}
	if bodyErr != nil {
		return evidence, bodyErr
	}
	if !allProofsPass(evidence) {
		evidence.Decision = "FAIL_CLOSED"
		evidence.Reason = "MANUAL_SOURCE_REGISTRATION_EDIT_FREE_PROJECTION_PROOF_INCOMPLETE"
	}
	return evidence, nil
}

func runProofBody(root, outputDir string) (Evidence, error) {
	loaded, err := loadManifests(root)
	if err != nil {
		return Evidence{}, err
	}
	baseline, diagnostic := observeBaseline(root)
	if diagnostic != nil {
		return Evidence{}, diagnosticError(diagnostic)
	}
	baseOutputs, _, err := renderOutputs(root, outputDir, loaded)
	if err != nil {
		return Evidence{}, err
	}
	reconciliations, diagnostic := reconcileDenominators(root, loaded)
	if diagnostic != nil {
		return Evidence{}, diagnosticError(diagnostic)
	}
	adoption, positivePredicates, err := runIndependentConsumer(root)
	if err != nil {
		return Evidence{}, err
	}
	fixture, err := measureFixture(root)
	if err != nil {
		return Evidence{}, err
	}
	baseIDs := make([]string, 0, len(loaded))
	for _, item := range loaded {
		baseIDs = append(baseIDs, item.Manifest.StableID)
	}
	sort.Strings(baseIDs)
	evidence := Evidence{
		Schema: evidenceSchema, Decision: "PASS", Reason: "MANUAL_SOURCE_REGISTRATION_EDIT_FREE_PROJECTION_PROVEN",
		BoundedSlice: baseIDs, BaselineTouchpoints: baseline.Observed, BaselineObservation: baseline.Touchpoints,
		Metrics:                    integrationMetrics(len(loaded), baseline.Observed, len(fixture.Changed), adoption, countUnequalSourceDigests(fixture.SourceDigests)),
		MetricDeltas:               metricDeltas(baseline.Observed, len(fixture.Changed), countUnequalSourceDigests(fixture.SourceDigests)),
		DenominatorReconciliations: reconciliations, SourceDigestPreservation: fixture.SourceDigests,
		GeneratedOutputChanges: fixture.Changed, GeneratedOutputChangeCount: len(fixture.Changed), GeneratedOutputDenominator: len(baseOutputs),
		ProductionAdoption: ratioMetric(adoption, 1), GeneratedOutputs: observedOutputs(outputDir), FixtureGeneratedOutputs: fixture.Outputs,
	}
	evidence.ProjectionReplay = projectionReplay(root, outputDir, loaded, baseOutputs)
	evidence.ManifestOrderInvariant = manifestOrderInvariant(root, outputDir, loaded, baseOutputs)
	evidence.SemanticCausality = semanticCausality(root, outputDir, loaded, baseOutputs)
	evidence.CommentInvariant = commentInvariant(root, outputDir, loaded, baseOutputs)
	evidence.NewConceptFixture = passedScenario("new-local-concept-fixture", fmt.Sprintf("temporary local manifest changed %d/%d generated outputs and no existing source bytes", len(fixture.Changed), len(baseOutputs)))
	evidence.FailureContracts = failureContracts(root, loaded, baseOutputs)
	evidence.DenominatorMismatch, evidence.StaleDenominatorReceipt = denominatorMismatchContract(root, loaded)
	negativePredicates, err := independentFailurePredicates(root)
	if err != nil {
		return Evidence{}, err
	}
	evidence.PredicateObservations = append(positivePredicates, negativePredicates...)
	evidence.Strategies = strategyResults(root, outputDir, loaded, evidence)
	evidence.Claims = buildClaims(evidence.PredicateObservations)
	return evidence, nil
}

func projectionReplay(root, outputDir string, loaded []LoadedManifest, base map[string][]byte) ScenarioResult {
	replay, _, err := renderOutputs(root, outputDir, loaded)
	if err != nil {
		return failedScenario("projection-twice-byte-equality", "PASS", err.Error())
	}
	if !sameOutputs(base, replay) {
		return failedScenario("projection-twice-byte-equality", "PASS", "byte-for-byte replay diverged")
	}
	return passedScenario("projection-twice-byte-equality", "two projections are byte-for-byte identical")
}

func manifestOrderInvariant(root, outputDir string, loaded []LoadedManifest, base map[string][]byte) ScenarioResult {
	reordered := append([]LoadedManifest(nil), loaded...)
	for left, right := 0, len(reordered)-1; left < right; left, right = left+1, right-1 {
		reordered[left], reordered[right] = reordered[right], reordered[left]
	}
	projection, _, err := renderOutputs(root, outputDir, reordered)
	if err != nil {
		return failedScenario("manifest-order-invariance", "PASS", err.Error())
	}
	if !sameOutputs(base, projection) {
		return failedScenario("manifest-order-invariance", "PASS", "manifest order changed generated bytes")
	}
	return passedScenario("manifest-order-invariance", "reordering local manifests preserves every output byte")
}

func semanticCausality(root, outputDir string, loaded []LoadedManifest, base map[string][]byte) ScenarioResult {
	mutated, err := renderedMutation(loaded[0], func(manifest *Manifest) { manifest.Concept.PositiveEffect += " [semantic mutation]" })
	if err != nil {
		return failedScenario("semantic-manifest-causality", "PASS", err.Error())
	}
	projection, _, err := renderOutputs(root, outputDir, append([]LoadedManifest{mutated}, loaded[1:]...))
	if err != nil {
		return failedScenario("semantic-manifest-causality", "PASS", err.Error())
	}
	changed := changedOutputNames(base, projection)
	want := []string{"README.md", "catalog.json", "denominator.json", "digest.json", "manifest-digests.json", "projection.json"}
	if !sameStringSet(changed, want) {
		return failedScenario("semantic-manifest-causality", "PASS", fmt.Sprintf("changed outputs=%v, want=%v", changed, want))
	}
	return passedScenario("semantic-manifest-causality", "rendered semantic manifest bytes were decoded before projection and changed semantic surfaces only")
}

func commentInvariant(root, outputDir string, loaded []LoadedManifest, base map[string][]byte) ScenarioResult {
	mutated, err := renderedMutation(loaded[0], func(manifest *Manifest) { manifest.Comments = append(manifest.Comments, "presentation-only comment") })
	if err != nil {
		return failedScenario("comment-only-invariance", "PASS", err.Error())
	}
	projection, _, err := renderOutputs(root, outputDir, append([]LoadedManifest{mutated}, loaded[1:]...))
	if err != nil {
		return failedScenario("comment-only-invariance", "PASS", err.Error())
	}
	changed := changedOutputNames(base, projection)
	want := []string{"digest.json", "manifest-digests.json"}
	if !sameStringSet(changed, want) {
		return failedScenario("comment-only-invariance", "PASS", fmt.Sprintf("changed outputs=%v, want=%v", changed, want))
	}
	return passedScenario("comment-only-invariance", "rendered comment-only manifest bytes changed raw digest while semantic projection remained equal")
}

func renderedMutation(item LoadedManifest, mutate func(*Manifest)) (LoadedManifest, error) {
	clone := cloneLoaded([]LoadedManifest{item})[0]
	mutate(&clone.Manifest)
	raw, err := renderJSON(clone.Manifest)
	if err != nil {
		return LoadedManifest{}, err
	}
	return decodeManifest(clone.SourcePath, raw)
}

func measureFixture(root string) (fixtureMeasurement, error) {
	tempRoot, err := os.MkdirTemp("", "gooo-registry-projection-proof-")
	if err != nil {
		return fixtureMeasurement{}, err
	}
	defer os.RemoveAll(tempRoot)
	if err := copyTree(root, tempRoot); err != nil {
		return fixtureMeasurement{}, err
	}
	tempLoaded, err := loadManifests(tempRoot)
	if err != nil {
		return fixtureMeasurement{}, err
	}
	baseOutputs, _, err := renderOutputs(tempRoot, filepath.Join(tempRoot, "proof-base"), tempLoaded)
	if err != nil {
		return fixtureMeasurement{}, err
	}
	before := sourceDigests(tempRoot)
	fixtureRaw, err := os.ReadFile(filepath.Join(root, "scripts/conflict-free-registry-projection/testdata/new-concept.manifest.json"))
	if err != nil {
		return fixtureMeasurement{}, err
	}
	fixturePath := filepath.Join(tempRoot, filepath.FromSlash(expectedManifestPath("language-registry-projection-fixture")))
	if err := os.MkdirAll(filepath.Dir(fixturePath), 0o755); err != nil {
		return fixtureMeasurement{}, err
	}
	if err := os.WriteFile(fixturePath, fixtureRaw, 0o644); err != nil {
		return fixtureMeasurement{}, err
	}
	withFixture, err := loadManifests(tempRoot)
	if err != nil {
		return fixtureMeasurement{}, err
	}
	afterOutputDir := filepath.Join(tempRoot, defaultOutput)
	afterOutputs, _, err := renderOutputs(tempRoot, afterOutputDir, withFixture)
	if err != nil {
		return fixtureMeasurement{}, err
	}
	if err := writeOutputs(afterOutputDir, afterOutputs); err != nil {
		return fixtureMeasurement{}, err
	}
	after := sourceDigests(tempRoot)
	comparisons := compareSourceDigests(before, after)
	consumer, err := runIndependentConsumerAt(tempRoot)
	if err != nil {
		return fixtureMeasurement{}, err
	}
	return fixtureMeasurement{SourceDigests: comparisons, Changed: changedOutputNames(baseOutputs, afterOutputs), Outputs: outputMetadata(tempRoot, afterOutputDir, afterOutputs), Consumer: consumer}, nil
}

func runIndependentConsumer(root string) (int, []PredicateObservation, error) {
	receipt, err := runIndependentConsumerAt(root)
	if err != nil {
		return 0, nil, err
	}
	for _, predicate := range receipt.Predicates {
		if predicate.ID == "independent-production-adoption" && predicate.Observed {
			return 1, receipt.Predicates, nil
		}
	}
	return 0, receipt.Predicates, nil
}

func runIndependentConsumerAt(root string) (independentReceipt, error) {
	tempOutput, err := os.CreateTemp("", "consumer-projection-")
	if err != nil {
		return independentReceipt{}, err
	}
	outputPath := tempOutput.Name()
	_ = tempOutput.Close()
	defer os.Remove(outputPath)
	tempReceipt, err := os.CreateTemp("", "consumer-receipt-")
	if err != nil {
		return independentReceipt{}, err
	}
	receiptPath := tempReceipt.Name()
	_ = tempReceipt.Close()
	defer os.Remove(receiptPath)
	command := exec.Command("go", "run", "./scripts/conflict-free-registry-projection-consumer", "-root", root, "-output", outputPath, "-receipt", receiptPath, "-check-generated")
	command.Dir = root
	if data, err := command.CombinedOutput(); err != nil {
		return independentReceipt{}, fmt.Errorf("independent consumer failed: %s", strings.TrimSpace(string(data)))
	}
	raw, err := os.ReadFile(receiptPath)
	if err != nil {
		return independentReceipt{}, err
	}
	receipt := independentReceipt{}
	if err := json.Unmarshal(raw, &receipt); err != nil {
		return independentReceipt{}, err
	}
	if receipt.Decision != "PASS" || len(receipt.Predicates) == 0 {
		return independentReceipt{}, fmt.Errorf("independent consumer receipt did not discharge predicates")
	}
	return receipt, nil
}

func failureContracts(root string, loaded []LoadedManifest, base map[string][]byte) []ScenarioResult {
	duplicate := append(cloneLoaded(loaded), cloneLoaded(loaded[:1])...)
	missing := cloneLoaded(loaded[1:])
	required := make([]string, 0, len(loaded))
	for _, item := range loaded {
		required = append(required, item.Manifest.StableID)
	}
	stale := cloneOutputs(base)
	projection := append([]byte(nil), stale["projection.json"]...)
	projection[0] ^= 1
	stale["projection.json"] = projection
	cross := cloneLoaded(loaded[:1])
	cross[0].SourcePath = "examples/wrong-owner/concept.manifest.json"
	missingBinding := cloneLoaded(loaded[:1])
	missingBinding[0].Manifest.MetricBindings = nil
	malformed := cloneLoaded(loaded[:1])
	malformed[0].Manifest.Schema = "gooo/malformed/v1"
	return []ScenarioResult{
		failureScenario("duplicate-stable-id", validateManifests(duplicate, nil), "DUPLICATE_STABLE_ID"),
		failureScenario("missing-manifest", validateManifests(missing, required), "MISSING_MANIFEST"),
		failureScenario("stale-generated-projection", checkRendered(base, stale), "STALE_GENERATED_PROJECTION"),
		failureScenario("cross-directory-manifest", validateManifestInputs(root, cross, nil), "CROSS_DIRECTORY_MANIFEST"),
		failureScenario("missing-binding", validateManifests(missingBinding, nil), "MISSING_METRIC_BINDING"),
		failureScenario("malformed-manifest", validateManifests(malformed, nil), "INVALID_MANIFEST_IDENTITY"),
	}
}

func denominatorMismatchContract(root string, loaded []LoadedManifest) (ScenarioResult, *DenominatorReconciliation) {
	stale := cloneLoaded(loaded)
	for index := range stale {
		if stale[index].Manifest.StableID == "toolchain-conformance" {
			stale[index].Manifest.Denominators[0].Values["cases"] = 160
		}
	}
	receipts, diagnostic := reconcileDenominators(root, stale)
	var mismatch *DenominatorReconciliation
	for index := range receipts {
		if receipts[index].Reason == "DENOMINATOR_SOURCE_MISMATCH" {
			mismatch = &receipts[index]
			break
		}
	}
	return failureScenario("stale-denominator", diagnostic, "DENOMINATOR_SOURCE_MISMATCH"), mismatch
}

func failureScenario(id string, diagnostic *Diagnostic, expectedReason string) ScenarioResult {
	if diagnostic == nil {
		return failedScenario(id, "FAIL_CLOSED", "invalid input was accepted")
	}
	if diagnostic.Decision != "FAIL_CLOSED" || diagnostic.Stage == "" || diagnostic.Step == "" || diagnostic.Reason != expectedReason {
		return failedScenario(id, "FAIL_CLOSED", fmt.Sprintf("diagnostic=%+v", *diagnostic))
	}
	return ScenarioResult{ID: id, Decision: "PASS", Expected: "FAIL_CLOSED", Diagnostic: diagnostic, Detail: "stage, step, and reason were preserved"}
}

func independentFailurePredicates(root string) ([]PredicateObservation, error) {
	type caseDef struct {
		id, reason string
		mutate     func(string) error
	}
	cases := []caseDef{
		{"consumer-malformed-manifest", "MALFORMED_MANIFEST", func(temp string) error {
			path := filepath.Join(temp, "examples/language-syntax-roundtrip/concept.manifest.json")
			return os.WriteFile(path, []byte("{"), 0o644)
		}},
		{"consumer-missing-manifest", "MISSING_MANIFEST", func(temp string) error {
			return os.Rename(filepath.Join(temp, "examples/language-syntax-roundtrip/concept.manifest.json"), filepath.Join(temp, "examples/language-syntax-roundtrip/concept.manifest.json.missing"))
		}},
		{"consumer-cross-directory-manifest", "CROSS_DIRECTORY_MANIFEST", func(temp string) error {
			from := filepath.Join(temp, "examples/language-syntax-roundtrip/concept.manifest.json")
			to := filepath.Join(temp, "examples/wrong-owner/concept.manifest.json")
			if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
				return err
			}
			return os.Rename(from, to)
		}},
		{"consumer-missing-binding", "MISSING_METRIC_BINDING", func(temp string) error {
			return rewriteManifest(temp, "language-syntax-roundtrip", func(manifest *Manifest) { manifest.MetricBindings = nil })
		}},
		{"consumer-stale-denominator", "DENOMINATOR_SOURCE_MISMATCH", func(temp string) error {
			return rewriteManifest(temp, "toolchain-conformance", func(manifest *Manifest) { manifest.Denominators[0].Values["cases"] = 160 })
		}},
		{"consumer-stale-generated-projection", "STALE_GENERATED_PROJECTION", func(temp string) error {
			path := filepath.Join(temp, defaultOutput, "projection.json")
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			data[0] ^= 1
			return os.WriteFile(path, data, 0o644)
		}},
		{"consumer-duplicate-stable-id", "DUPLICATE_STABLE_ID", func(temp string) error {
			source := filepath.Join(temp, "examples/language-syntax-roundtrip/concept.manifest.json")
			data, err := os.ReadFile(source)
			if err != nil {
				return err
			}
			target := filepath.Join(temp, "examples/duplicate-concept/concept.manifest.json")
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			return os.WriteFile(target, data, 0o644)
		}},
	}
	observations := make([]PredicateObservation, 0, len(cases))
	for _, item := range cases {
		temp, err := os.MkdirTemp("", "gooo-consumer-failure-")
		if err != nil {
			return nil, err
		}
		copyErr := copyTree(root, temp)
		if copyErr == nil {
			copyErr = item.mutate(temp)
		}
		var commandOutput string
		if copyErr == nil {
			commandOutput = runConsumerFailureCommand(root, temp)
		}
		observed := copyErr == nil && strings.Contains(commandOutput, item.reason)
		address := filepath.ToSlash(filepath.Join(temp, "examples", item.id))
		targetDigest := digestBytes([]byte(address + "|" + commandOutput))
		observations = append(observations, PredicateObservation{ID: item.id, ObservedPredicate: "independent consumer rejects " + item.reason, TargetAddress: address, TargetDigest: targetDigest, Observed: observed, Decision: "FAIL_CLOSED", Stage: "REGRESSION", Step: "INDEPENDENT_CONSUMER_FAILURE_CONTRACT", Reason: item.reason})
		_ = os.RemoveAll(temp)
	}
	return observations, nil
}

func runConsumerFailureCommand(repoRoot, tempRoot string) string {
	command := exec.Command("go", "run", "./scripts/conflict-free-registry-projection-consumer", "-root", tempRoot, "-check-generated")
	command.Dir = repoRoot
	data, _ := command.CombinedOutput()
	return string(data)
}

func rewriteManifest(root, stableID string, mutate func(*Manifest)) error {
	path := filepath.Join(root, filepath.FromSlash(expectedManifestPath(stableID)))
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	item, err := decodeManifest(expectedManifestPath(stableID), raw)
	if err != nil {
		return err
	}
	mutate(&item.Manifest)
	data, err := renderJSON(item.Manifest)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func strategyResults(root, outputDir string, loaded []LoadedManifest, evidence Evidence) []StrategyResult {
	results := make([]StrategyResult, 0, 3)
	for _, name := range []string{"FOUNDATION", "COHERENCE", "REGRESSION"} {
		selected := true
		for _, item := range loaded {
			if !contains(item.Manifest.VerificationStrategies, name) {
				selected = false
			}
		}
		result := StrategyResult{Name: name, Selected: selected, Decision: "PASS", Reason: "STRATEGY_SELECTED_AND_DISCHARGED"}
		switch name {
		case "FOUNDATION":
			result.Scenarios = []string{"local-manifest-identity", "resource-availability", "three-concept-denominator-reconciliation", "baseline-touchpoints"}
			if !selected || len(evidence.DenominatorReconciliations) < 3 {
				result.Decision, result.Reason = "FAIL_CLOSED", "FOUNDATION_SOURCE_RECONCILIATION_NOT_DISCHARGED"
			}
		case "COHERENCE":
			result.Scenarios = []string{"projection-twice-byte-equality", "manifest-order-invariance", "production-consumer-adoption"}
			if !selected || evidence.ProjectionReplay.Decision != "PASS" || evidence.ManifestOrderInvariant.Decision != "PASS" || evidence.ProductionAdoption.Numerator != 1 {
				result.Decision, result.Reason = "FAIL_CLOSED", "COHERENCE_REPLAY_NOT_DISCHARGED"
			}
		case "REGRESSION":
			result.Scenarios = []string{"duplicate-stable-id", "missing-manifest", "stale-generated-projection", "cross-directory-manifest", "missing-binding", "malformed-manifest", "stale-denominator", "semantic-manifest-causality", "comment-only-invariance"}
			if !selected || evidence.SemanticCausality.Decision != "PASS" || evidence.CommentInvariant.Decision != "PASS" || evidence.DenominatorMismatch.Decision != "PASS" {
				result.Decision, result.Reason = "FAIL_CLOSED", "REGRESSION_INVARIANTS_NOT_DISCHARGED"
			}
			for _, scenario := range evidence.FailureContracts {
				if scenario.Decision != "PASS" {
					result.Decision, result.Reason = "FAIL_CLOSED", "REGRESSION_FAILURE_CONTRACT_NOT_DISCHARGED"
				}
			}
		}
		results = append(results, result)
	}
	return results
}

func buildClaims(observations []PredicateObservation) []Claim {
	claims := make([]Claim, 0, len(observations))
	for _, observation := range observations {
		predicateDigest := digestBytes([]byte(observation.ObservedPredicate + "|" + observation.TargetDigest))
		terminalState := "OPEN"
		terminalReason := observation.Reason
		terminalStage := observation.Stage
		terminalStep := observation.Step
		if observation.Observed && observation.Decision == "PASS" {
			terminalState = "DISCHARGED"
			terminalReason = "independent_consumer_recomputed_predicate"
			terminalStage = observation.Stage
			terminalStep = "INDEPENDENT_CONSUMER_PREDICATE"
		} else if observation.Decision == "FAIL_CLOSED" {
			terminalState = "REFUTED"
		}
		claims = append(claims, Claim{ID: observation.ID, Proposition: observation.ObservedPredicate, ObservedPredicate: observation.ObservedPredicate, TargetAddress: observation.TargetAddress, TargetDigest: observation.TargetDigest, Transitions: []ClaimTransition{{State: "OPEN", ObservedPredicate: observation.ObservedPredicate, PredicateDigest: predicateDigest, TargetAddress: observation.TargetAddress, TargetDigest: observation.TargetDigest, Stage: "FOUNDATION", Step: "CLAIM_OPEN", Reason: "proposition_under_review"}, {State: terminalState, ObservedPredicate: observation.ObservedPredicate, PredicateDigest: predicateDigest, TargetAddress: observation.TargetAddress, TargetDigest: observation.TargetDigest, Stage: terminalStage, Step: terminalStep, Reason: terminalReason}}})
	}
	return claims
}

func allProofsPass(evidence Evidence) bool {
	for _, scenario := range []ScenarioResult{evidence.ProjectionReplay, evidence.ManifestOrderInvariant, evidence.SemanticCausality, evidence.CommentInvariant, evidence.NewConceptFixture, evidence.DenominatorMismatch} {
		if scenario.Decision != "PASS" {
			return false
		}
	}
	if len(evidence.DenominatorReconciliations) < 3 || uniqueStableIDs(evidence.DenominatorReconciliations) != 3 || evidence.StaleDenominatorReceipt == nil || evidence.StaleDenominatorReceipt.Decision != "FAIL_CLOSED" || evidence.StaleDenominatorReceipt.Reason != "DENOMINATOR_SOURCE_MISMATCH" || evidence.GeneratedOutputChangeCount != 6 || evidence.GeneratedOutputDenominator != 8 || evidence.ProductionAdoption.Numerator != 1 || evidence.ProductionAdoption.Denominator != 1 {
		return false
	}
	for _, comparison := range evidence.SourceDigestPreservation {
		if !comparison.Equal {
			return false
		}
	}
	if len(evidence.SourceDigestPreservation) != 12 {
		return false
	}
	if len(evidence.FailureContracts) < 6 {
		return false
	}
	for _, scenario := range evidence.FailureContracts {
		if scenario.Decision != "PASS" || scenario.Diagnostic == nil {
			return false
		}
	}
	for _, predicate := range evidence.PredicateObservations {
		if !predicate.Observed || predicate.TargetAddress == "" || predicate.TargetDigest == "" {
			return false
		}
	}
	for _, strategy := range evidence.Strategies {
		if !strategy.Selected || strategy.Decision != "PASS" {
			return false
		}
	}
	return validateClaims(evidence.Claims) == nil && evidence.RepositoryNetState.NetStateEqual && evidence.RepositoryNetState.NetState == "NET_STATE_EQUAL"
}

func validateClaims(claims []Claim) error {
	propositions, addresses, targets := map[string]struct{}{}, map[string]struct{}{}, map[string]struct{}{}
	for _, item := range claims {
		if item.ID == "" || item.Proposition == "" || item.ObservedPredicate == "" || item.TargetAddress == "" || item.TargetDigest == "" || len(item.Transitions) < 2 {
			return fmt.Errorf("invalid claim shape")
		}
		if _, ok := propositions[item.Proposition]; ok {
			return fmt.Errorf("duplicate claim proposition")
		}
		propositions[item.Proposition] = struct{}{}
		if _, ok := addresses[item.TargetAddress]; ok {
			return fmt.Errorf("duplicate claim target address")
		}
		addresses[item.TargetAddress] = struct{}{}
		if _, ok := targets[item.TargetDigest]; ok {
			return fmt.Errorf("duplicate claim target digest")
		}
		targets[item.TargetDigest] = struct{}{}
		if item.Transitions[0].State != "OPEN" {
			return fmt.Errorf("claim does not begin OPEN")
		}
		last := item.Transitions[len(item.Transitions)-1]
		if last.State != "DISCHARGED" && last.State != "REFUTED" && last.State != "OPEN" {
			return fmt.Errorf("claim does not terminate in a known state")
		}
		for _, transition := range item.Transitions {
			if transition.ObservedPredicate != item.ObservedPredicate || transition.TargetAddress != item.TargetAddress || transition.TargetDigest != item.TargetDigest || transition.PredicateDigest == "" {
				return fmt.Errorf("claim transition is not target-bound")
			}
		}
		if last.State == "DISCHARGED" && last.Reason != "independent_consumer_recomputed_predicate" {
			return fmt.Errorf("claim discharged without independent predicate")
		}
	}
	return nil
}

func cloneLoaded(input []LoadedManifest) []LoadedManifest {
	output := make([]LoadedManifest, 0, len(input))
	for _, item := range input {
		data, _ := json.Marshal(item.Manifest)
		manifest := Manifest{}
		_ = json.Unmarshal(data, &manifest)
		output = append(output, LoadedManifest{Manifest: manifest, SourcePath: item.SourcePath, RawDigest: item.RawDigest})
	}
	return output
}
func cloneOutputs(input map[string][]byte) map[string][]byte {
	output := make(map[string][]byte, len(input))
	for key, value := range input {
		output[key] = append([]byte(nil), value...)
	}
	return output
}
func sameOutputs(left, right map[string][]byte) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if !bytes.Equal(value, right[key]) {
			return false
		}
	}
	return true
}
func changedOutputNames(left, right map[string][]byte) []string {
	keys := map[string]struct{}{}
	for key := range left {
		keys[key] = struct{}{}
	}
	for key := range right {
		keys[key] = struct{}{}
	}
	changed := []string{}
	for key := range keys {
		if !bytes.Equal(left[key], right[key]) {
			changed = append(changed, key)
		}
	}
	sort.Strings(changed)
	return changed
}
func sameStringSet(left, right []string) bool {
	left = sortedStringsCopy(left)
	right = sortedStringsCopy(right)
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
func passedScenario(id, detail string) ScenarioResult {
	return ScenarioResult{ID: id, Decision: "PASS", Expected: "PASS", Detail: detail}
}
func failedScenario(id, expected, detail string) ScenarioResult {
	return ScenarioResult{ID: id, Decision: "FAIL_CLOSED", Expected: expected, Detail: detail}
}
func countUnequalSourceDigests(values []SourceDigestComparison) int {
	count := 0
	for _, item := range values {
		if !item.Equal {
			count++
		}
	}
	return count
}
func uniqueStableIDs(values []DenominatorReconciliation) int {
	seen := map[string]struct{}{}
	for _, item := range values {
		seen[item.StableID] = struct{}{}
	}
	return len(seen)
}

func sourceDigests(root string) []SourceDigestComparison {
	result := make([]SourceDigestComparison, 0, len(baselineTouchpoints))
	for _, item := range baselineTouchpoints {
		data, err := os.ReadFile(joinRoot(root, item.Path))
		digest := "UNKNOWN"
		if err == nil {
			digest = digestBytes(data)
		}
		result = append(result, SourceDigestComparison{Path: item.Path, Before: digest})
	}
	return result
}
func compareSourceDigests(before, after []SourceDigestComparison) []SourceDigestComparison {
	result := make([]SourceDigestComparison, len(before))
	for index := range before {
		result[index] = before[index]
		for _, item := range after {
			if item.Path == before[index].Path {
				result[index].After = item.Before
				result[index].Equal = item.Before == before[index].Before
				break
			}
		}
	}
	return result
}
func copyTree(source, target string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}
		if entry.Name() == ".git" || entry.Name() == ".parallel" {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		destination := filepath.Join(target, relative)
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(destination, data, 0o644)
	})
}

func repositoryObservation(before, after []string, beforeErr, afterErr error) RepositoryObservation {
	observation := RepositoryObservation{BeforePaths: before, AfterPaths: after, NetStateEqual: beforeErr == nil && afterErr == nil && sameStringSet(before, after), NetState: "UNKNOWN", TransientMutation: "UNKNOWN", MutationAuthority: "UNKNOWN"}
	if observation.NetStateEqual {
		observation.NetState = "NET_STATE_EQUAL"
	} else if beforeErr == nil && afterErr == nil {
		observation.NetState = "NET_STATE_CHANGED"
	}
	return observation
}
func gitStatus(root string) ([]string, error) {
	command := exec.Command("git", "-C", root, "status", "--porcelain", "--untracked-files=all")
	data, err := command.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil
	}
	return lines, nil
}
func observedOutputs(outputDir string) []OutputMetadata {
	outputs, err := readGenerated(outputDir)
	if err != nil {
		return nil
	}
	return outputMetadata("", outputDir, outputs)
}
