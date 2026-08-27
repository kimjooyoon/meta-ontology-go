package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
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
	OutputArtifact             consumerOutputArtifact      `json:"output_artifact"`
	BindingOutputReceipts      []BindingOutputReceipt      `json:"binding_output_receipts"`
	outputRaw                  []byte
	rawReceipt                 []byte
}

type consumerOutputArtifact struct {
	Path   string `json:"path"`
	Digest string `json:"digest"`
	Bytes  int    `json:"bytes"`
}

type fixtureMeasurement struct {
	SourceDigests []SourceDigestComparison
	Changed       []string
	Outputs       []OutputMetadata
	Consumer      independentReceipt
}

func runProof(root, outputDir string) (Evidence, error) {
	before, beforeErr := gitRepositorySnapshot(root)
	evidence, bodyErr := runProofBody(root, outputDir)
	after, afterErr := gitRepositorySnapshot(root)
	evidence.RepositoryNetState = repositoryObservation(before, after, beforeErr, afterErr)
	evidence.RepositoryNetStatePredicates = predicateMetric(boolInt(evidence.RepositoryNetState.NetStateEqual), 1, decisionForBool(evidence.RepositoryNetState.NetStateEqual), "REGRESSION", "REPOSITORY_NET_STATE", evidence.RepositoryNetState.NetState)
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
	conformance, positivePredicates, consumerOutput, bindingReceipts, err := runIndependentConsumer(root)
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
	metrics := integrationMetrics(len(loaded), baseline.Observed, len(fixture.Changed), conformance, countUnequalSourceDigests(fixture.SourceDigests))
	metrics.ProducerPackageImports = producerPackageImportMetric(root)
	metrics.RawSourceReconstruction = contextualRatioMetric(conformance, 1, "COHERENCE", "RAW_SOURCE_RECONSTRUCTION", "CONSUMER_REBUILT_FROM_RAW_MANIFESTS")
	metrics.SeparateExecutable = contextualRatioMetric(conformance, 1, "COHERENCE", "SEPARATE_EXECUTABLE", "INDEPENDENT_CONSUMER_PROCESS_EXECUTED")
	metrics.AlgorithmicIndependence = unknownRatioMetric("COHERENCE", "ALGORITHMIC_INDEPENDENCE", "ALGORITHM_DIFFERENCE_NOT_OBSERVED")
	evidence := Evidence{
		Schema: evidenceSchema, Decision: "PASS", Reason: "MANUAL_SOURCE_REGISTRATION_EDIT_FREE_PROJECTION_PROVEN",
		BoundedSlice: baseIDs, BaselineTouchpoints: baseline.Observed, BaselineObservation: baseline.Touchpoints,
		Metrics:                    metrics,
		MetricDeltas:               metricDeltas(baseline.Observed, len(fixture.Changed), countUnequalSourceDigests(fixture.SourceDigests)),
		DenominatorReconciliations: reconciliations, SourceDigestPreservation: fixture.SourceDigests,
		GeneratedOutputChanges: fixture.Changed, GeneratedOutputChangeCount: len(fixture.Changed), GeneratedOutputDenominator: len(baseOutputs),
		ConformanceConsumer: predicateMetric(conformance, 1, decisionForRatio(conformance, 1), "COHERENCE", "INDEPENDENT_CONFORMANCE_CONSUMER", "INDEPENDENT_CONSUMER_RECOMPUTED_PROJECTION"),
		ProductionAdoption:  predicateMetric(0, 1, "UNKNOWN", "COHERENCE", "PRODUCTION_CONSUMER_ADOPTION", "NO_PRODUCTION_CONSUMER_EVIDENCE"),
		UseCaseReceipt:      observeUseCaseReceipt(root),
		GeneratedOutputs:    observedOutputs(outputDir), FixtureGeneratedOutputs: fixture.Outputs,
		ConsumerOutputArtifact: consumerOutput, BindingOutputReceipts: bindingReceipts,
	}
	evidence.ProjectionReplay = projectionReplay(root, outputDir, loaded, baseOutputs)
	evidence.ManifestOrderInvariant = manifestOrderInvariant(root, outputDir, loaded, baseOutputs)
	evidence.SemanticCausality = semanticCausality(root, outputDir, loaded, baseOutputs)
	evidence.SemanticMetricChange = semanticMetricChange(root, loaded)
	evidence.CommentInvariant = commentInvariant(root, outputDir, loaded, baseOutputs)
	evidence.CommentPositionInvariant = commentPositionInvariant(root, loaded)
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
	allExpectedIDs := append(append([]string{}, expectedConformancePredicateIDs...), expectedFailurePredicateIDs...)
	evidence.PredicateInventory = makeInventory(allExpectedIDs, predicateIDs(evidence.PredicateObservations), "REGRESSION", "PREDICATE_ID_INVENTORY", "FIXED_PREDICATE_ID_INVENTORY")
	evidence.FailureInventory = makeInventory(expectedFailurePredicateIDs, selectPredicateIDs(evidence.PredicateObservations, expectedFailurePredicateIDs), "REGRESSION", "FAILURE_ID_INVENTORY", "FIXED_FAILURE_ID_INVENTORY")
	evidence.ClaimInventory = makeInventory(allExpectedIDs, claimIDs(evidence.Claims), "COHERENCE", "CLAIM_ID_INVENTORY", "FIXED_CLAIM_ID_INVENTORY")
	evidence.ProvenanceInventory = makeInventory(expectedFailurePredicateIDs, selectPredicateIDs(evidence.PredicateObservations, expectedFailurePredicateIDs), "REGRESSION", "PROVENANCE_ID_INVENTORY", "FIXED_PROVENANCE_ID_INVENTORY")
	claimCount := countDischargedClaims(evidence.Claims)
	evidence.ClaimTransitions = predicateMetric(claimCount, expectedClaimCount, decisionForRatio(claimCount, expectedClaimCount), "COHERENCE", "CLAIM_TRANSITIONS", "OBSERVED_PREDICATE_TRUTH_RECONSTRUCTED")
	evidence.FailurePredicates = failurePredicateMetric(evidence.PredicateObservations)
	evidence.RepositoryNetStatePredicates = predicateMetric(boolInt(evidence.RepositoryNetState.NetStateEqual), 1, decisionForBool(evidence.RepositoryNetState.NetStateEqual), "REGRESSION", "REPOSITORY_NET_STATE", evidence.RepositoryNetState.NetState)
	evidence.BindingPredicates = bindingPredicateMetric(evidence.PredicateObservations, bindingReceipts)
	evidence.ASTResolvedBindings = predicateMetric(len(bindingReceipts), expectedBindingReceiptCount, decisionForRatio(len(bindingReceipts), expectedBindingReceiptCount), "FOUNDATION", "AST_BINDING_RESOLUTION", "GO_AST_SYMBOL_AND_OUTPUT_RECEIPT_RECONSTRUCTED")
	evidence.MetricOccurrences = metricOccurrenceMetric(root, bindingReceipts)
	evidence.UniqueSemanticRelationDigests = uniqueSemanticRelationMetric(bindingReceipts)
	evidence.OutputRowAddresses = uniqueOutputRowMetric(bindingReceipts, []byte(consumerOutput.RawBytes))
	evidence.ProvenancePredicates = provenancePredicateMetric(negativePredicates)
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

func semanticMetricChange(root string, loaded []LoadedManifest) ScenarioResult {
	item, ok := loadedManifestByID(loaded, "language-syntax-roundtrip")
	if !ok || len(item.Manifest.BindingRegistry) == 0 {
		return failedScenario("semantic-metric-change", "PASS", "bounded syntax binding was unavailable")
	}
	binding := item.Manifest.BindingRegistry[0]
	before, err := resolveBindingRelation(root, binding.RawSourceAddress, binding.RegistrationUseAddress, binding.MetricID)
	if err != nil {
		return failedScenario("semantic-metric-change", "PASS", err.Error())
	}
	tempRoot, err := os.MkdirTemp("", "gooo-binding-metric-change-")
	if err != nil {
		return failedScenario("semantic-metric-change", "PASS", err.Error())
	}
	defer os.RemoveAll(tempRoot)
	if err := copyTree(root, tempRoot); err != nil {
		return failedScenario("semantic-metric-change", "PASS", err.Error())
	}
	path := filepath.Join(tempRoot, filepath.FromSlash("internal/meta/languageconcept/catalog.go"))
	raw, err := os.ReadFile(path)
	if err != nil {
		return failedScenario("semantic-metric-change", "PASS", err.Error())
	}
	changedMetric := binding.MetricID + ".changed"
	mutated := bytes.Replace(raw, []byte(binding.MetricID), []byte(changedMetric), 1)
	if bytes.Equal(raw, mutated) {
		return failedScenario("semantic-metric-change", "PASS", "metric literal was not found")
	}
	if err := os.WriteFile(path, mutated, 0o644); err != nil {
		return failedScenario("semantic-metric-change", "PASS", err.Error())
	}
	after, err := resolveBindingRelation(tempRoot, binding.RawSourceAddress, binding.RegistrationUseAddress, changedMetric)
	if err != nil {
		return failedScenario("semantic-metric-change", "PASS", err.Error())
	}
	if before.SemanticDigest == after.SemanticDigest || before.MetricOccurrenceDigest == after.MetricOccurrenceDigest {
		return failedScenario("semantic-metric-change", "PASS", "metric-specific relation digest did not change")
	}
	return passedScenario("semantic-metric-change", "changing the bound metric literal changed its metric-specific relation digest")
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

func commentPositionInvariant(root string, loaded []LoadedManifest) ScenarioResult {
	item, ok := loadedManifestByID(loaded, "language-syntax-roundtrip")
	if !ok || len(item.Manifest.BindingRegistry) == 0 {
		return failedScenario("comment-position-invariance", "PASS", "bounded syntax binding was unavailable")
	}
	binding := item.Manifest.BindingRegistry[0]
	before, err := resolveBindingRelation(root, binding.RawSourceAddress, binding.RegistrationUseAddress, binding.MetricID)
	if err != nil {
		return failedScenario("comment-position-invariance", "PASS", err.Error())
	}
	tempRoot, err := os.MkdirTemp("", "gooo-binding-comment-shift-")
	if err != nil {
		return failedScenario("comment-position-invariance", "PASS", err.Error())
	}
	defer os.RemoveAll(tempRoot)
	if err := copyTree(root, tempRoot); err != nil {
		return failedScenario("comment-position-invariance", "PASS", err.Error())
	}
	path := filepath.Join(tempRoot, filepath.FromSlash("internal/meta/languageconcept/catalog.go"))
	raw, err := os.ReadFile(path)
	if err != nil {
		return failedScenario("comment-position-invariance", "PASS", err.Error())
	}
	if err := os.WriteFile(path, append([]byte("// comment-only position shift\n"), raw...), 0o644); err != nil {
		return failedScenario("comment-position-invariance", "PASS", err.Error())
	}
	shiftedAddress := strings.Replace(binding.RegistrationUseAddress, "@50:3", "@51:3", 1)
	after, err := resolveBindingRelation(tempRoot, binding.RawSourceAddress, shiftedAddress, binding.MetricID)
	if err != nil {
		return failedScenario("comment-position-invariance", "PASS", err.Error())
	}
	rawBefore := digestBytes([]byte(binding.RegistrationUseAddress))
	rawAfter := digestBytes([]byte(shiftedAddress))
	if rawBefore == rawAfter || before.SemanticDigest != after.SemanticDigest || before.MetricOccurrenceDigest != after.MetricOccurrenceDigest {
		return failedScenario("comment-position-invariance", "PASS", "comment-only position shift changed semantic relation")
	}
	return passedScenario("comment-position-invariance", "raw registration address changed while semantic relation digest remained equal")
}

func loadedManifestByID(loaded []LoadedManifest, stableID string) (LoadedManifest, bool) {
	for _, item := range loaded {
		if item.Manifest.StableID == stableID {
			return item, true
		}
	}
	return LoadedManifest{}, false
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

func runIndependentConsumer(root string) (int, []PredicateObservation, ObservedOutputArtifact, []BindingOutputReceipt, error) {
	receipt, err := runIndependentConsumerAt(root)
	if err != nil {
		return 0, nil, ObservedOutputArtifact{}, nil, err
	}
	for _, predicate := range receipt.Predicates {
		if predicate.ID == "independent-conformance-consumer" && predicate.Observed {
			return 1, receipt.Predicates, ObservedOutputArtifact{ObservedPath: embeddedOutputAddress, Digest: receipt.OutputArtifact.Digest, Bytes: receipt.OutputArtifact.Bytes, RawBytes: string(receipt.outputRaw)}, receipt.BindingOutputReceipts, nil
		}
	}
	return 0, receipt.Predicates, ObservedOutputArtifact{ObservedPath: embeddedOutputAddress, Digest: receipt.OutputArtifact.Digest, Bytes: receipt.OutputArtifact.Bytes, RawBytes: string(receipt.outputRaw)}, receipt.BindingOutputReceipts, nil
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
	outputRaw, err := os.ReadFile(outputPath)
	if err != nil {
		return independentReceipt{}, err
	}
	raw, err := os.ReadFile(receiptPath)
	if err != nil {
		return independentReceipt{}, err
	}
	receipt, err := decodeIndependentReceipt(raw)
	if err != nil {
		return independentReceipt{}, err
	}
	if err := validateIndependentReceipt(root, receipt, outputRaw); err != nil {
		return independentReceipt{}, err
	}
	receipt.outputRaw = outputRaw
	receipt.rawReceipt = raw
	return receipt, nil
}

const consumerReceiptSchema = "gooo/manual-source-registration-edit-free-registry-consumer-receipt/v1"

type receiptValidationError struct{ diagnostic Diagnostic }

func (e receiptValidationError) Error() string { return e.diagnostic.Reason }

func receiptBoundaryError(reason string) error {
	return receiptValidationError{diagnostic: Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "RECEIPT_BOUNDARY", Reason: reason}}
}

func receiptBoundaryDiagnostic(err error) *Diagnostic {
	if typed, ok := err.(receiptValidationError); ok {
		diagnostic := typed.diagnostic
		return &diagnostic
	}
	return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "RECEIPT_BOUNDARY", Reason: "RECEIPT_VALIDATION_FAILED"}
}

func decodeIndependentReceipt(raw []byte) (independentReceipt, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	receipt := independentReceipt{}
	if err := decoder.Decode(&receipt); err != nil {
		if strings.Contains(err.Error(), "unknown field") {
			return independentReceipt{}, receiptBoundaryError("RECEIPT_UNKNOWN_FIELD")
		}
		return independentReceipt{}, receiptBoundaryError("RECEIPT_MALFORMED_JSON")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return independentReceipt{}, receiptBoundaryError("RECEIPT_TRAILING_JSON")
	}
	return receipt, nil
}

func validateIndependentReceipt(root string, receipt independentReceipt, outputRaw []byte) error {
	if receipt.Schema != consumerReceiptSchema {
		return receiptBoundaryError("RECEIPT_SCHEMA_MISMATCH")
	}
	if receipt.Decision != "PASS" {
		return receiptBoundaryError("RECEIPT_DECISION_MISMATCH")
	}
	if receipt.ProjectionDigest != digestBytes(outputRaw) {
		return receiptBoundaryError("RECEIPT_PROJECTION_DIGEST_MISMATCH")
	}
	if len(receipt.Predicates) != expectedConformancePredicateCount || !sameStringSet(predicateIDs(receipt.Predicates), expectedConformancePredicateIDs) || hasDuplicateStrings(predicateIDs(receipt.Predicates)) {
		return receiptBoundaryError("RECEIPT_PREDICATE_INVENTORY_MISMATCH")
	}
	for _, predicate := range receipt.Predicates {
		if !predicate.Observed || predicate.PredicateTruth != "TRUE" || predicate.Decision != "PASS" || predicate.TargetAddress == "" || predicate.TargetDigest == "" {
			return receiptBoundaryError("RECEIPT_PREDICATE_INVENTORY_MISMATCH")
		}
	}
	if receipt.OutputArtifact.Path != embeddedOutputAddress || receipt.OutputArtifact.Bytes != len(outputRaw) || receipt.OutputArtifact.Digest != digestBytes(outputRaw) {
		return receiptBoundaryError("RECEIPT_OUTPUT_ARTIFACT_MISMATCH")
	}
	return validateObservedBindingReceipts(root, receipt.BindingOutputReceipts, outputRaw)
}

func validateObservedBindingReceipts(root string, receipts []BindingOutputReceipt, outputRaw []byte) error {
	if len(receipts) != expectedBindingReceiptCount {
		return receiptBoundaryError("RECEIPT_BINDING_INVENTORY_MISMATCH")
	}
	var projection Projection
	decoder := json.NewDecoder(bytes.NewReader(outputRaw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&projection); err != nil {
		return receiptBoundaryError("RECEIPT_OUTPUT_PROJECTION_MALFORMED")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return receiptBoundaryError("RECEIPT_OUTPUT_PROJECTION_TRAILING_JSON")
	}
	expected := map[string]BindingRegistryEntry{}
	expectedMetricIDs := map[string]struct{}{}
	for _, entry := range projection.Catalog {
		for _, binding := range entry.BindingRegistry {
			address := bindingOutputRowAddress(entry.StableID, binding.MetricID)
			if _, exists := expected[address]; exists {
				return receiptBoundaryError("DUPLICATE_OUTPUT_ROW_ADDRESS")
			}
			if _, exists := expectedMetricIDs[binding.MetricID]; exists {
				return receiptBoundaryError("DUPLICATE_OUTPUT_ROW_METRIC_ID")
			}
			expectedMetricIDs[binding.MetricID] = struct{}{}
			expected[address] = binding
		}
	}
	if len(expected) != expectedBindingReceiptCount {
		return receiptBoundaryError("RECEIPT_OUTPUT_ROW_INVENTORY_MISMATCH")
	}
	seen := map[string]struct{}{}
	seenMetricIDs := map[string]struct{}{}
	expectedRelations := map[string]bindingResolution{}
	expectedPairOwners := map[string]string{}
	for address, row := range expected {
		relation, err := resolveBindingRelation(root, row.RawSourceAddress, row.RegistrationUseAddress, row.MetricID)
		if err != nil {
			return receiptBoundaryError("BINDING_RELATION_RECONSTRUCTION_FAILED")
		}
		expectedRelations[address] = relation
		pair := metricOccurrencePairKey(relation.MetricOccurrenceAddress, relation.MetricOccurrenceDigest)
		if owner, exists := expectedPairOwners[pair]; exists && owner != address {
			return receiptBoundaryError("DUPLICATE_METRIC_OCCURRENCE_PAIR")
		}
		expectedPairOwners[pair] = address
	}
	for _, receipt := range receipts {
		if _, ok := seenMetricIDs[receipt.MetricID]; ok {
			return receiptBoundaryError("DUPLICATE_RECEIPT_METRIC_ID")
		}
		seenMetricIDs[receipt.MetricID] = struct{}{}
		if _, ok := seen[receipt.OutputRowAddress]; ok {
			return receiptBoundaryError("DUPLICATE_RECEIPT_OUTPUT_ROW_ADDRESS")
		}
		seen[receipt.OutputRowAddress] = struct{}{}
		row, ok := expected[receipt.OutputRowAddress]
		if !ok {
			return receiptBoundaryError("BINDING_OUTPUT_ROW_ADDRESS_MISMATCH")
		}
		if row.MetricID != receipt.MetricID || row.RawSourceAddress != receipt.RawSourceAddress || row.RegistrationUseAddress != receipt.RegistrationUseAddress || row.SemanticDigest != receipt.SemanticDigest || row.ConsumerEntryPoint != receipt.ConsumerEntryPoint || receipt.OutputAddress != embeddedOutputAddress || receipt.OutputDigest != digestBytes(outputRaw) || receipt.OutputBytes != len(outputRaw) || receipt.OutputRowDigest != digestJSONValue(row) {
			return receiptBoundaryError("BINDING_OUTPUT_ROW_TRIPLE_MISMATCH")
		}
		relation := expectedRelations[receipt.OutputRowAddress]
		if relation.SemanticDigest != receipt.SemanticDigest {
			return receiptBoundaryError("BINDING_SEMANTIC_DIGEST_MISMATCH")
		}
		addressMismatch := relation.MetricOccurrenceAddress != receipt.MetricOccurrenceAddress
		digestMismatch := relation.MetricOccurrenceDigest != receipt.MetricOccurrenceDigest
		if addressMismatch || digestMismatch {
			switch {
			case addressMismatch && !digestMismatch:
				return receiptBoundaryError("BINDING_OCCURRENCE_ADDRESS_MISMATCH")
			case !addressMismatch && digestMismatch:
				return receiptBoundaryError("BINDING_OCCURRENCE_DIGEST_MISMATCH")
			case occurrencePairBelongsToAnotherMetric(receipt, expectedPairOwners):
				return receiptBoundaryError("BINDING_OCCURRENCE_PAIR_MISMATCH")
			default:
				return receiptBoundaryError("BINDING_OCCURRENCE_RELATION_MISMATCH")
			}
		}
	}
	if len(seen) != len(expected) {
		return receiptBoundaryError("RECEIPT_OUTPUT_ROW_INVENTORY_MISMATCH")
	}
	return nil
}

func occurrencePairBelongsToAnotherMetric(current BindingOutputReceipt, expectedPairOwners map[string]string) bool {
	owner, ok := expectedPairOwners[metricOccurrencePairKey(current.MetricOccurrenceAddress, current.MetricOccurrenceDigest)]
	return ok && owner != current.OutputRowAddress
}

func bindingOutputRowAddress(stableID, metricID string) string {
	encode := base64.RawURLEncoding.EncodeToString
	return embeddedOutputAddress + "#/catalog/" + encode([]byte(stableID)) + "/binding_registry/" + encode([]byte(metricID))
}

func digestJSONValue(value any) string {
	raw, _ := json.Marshal(value)
	return digestBytes(raw)
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
	bindingSelfSearch := cloneLoaded(loaded[:1])
	bindingSelfSearch[0].Manifest.BindingRegistry[0].RawSourceAddress = "docs/language/conflict-free-registry-projection.md#self-search"
	bindingOutputMismatch := cloneLoaded(loaded[:1])
	bindingOutputMismatch[0].Manifest.BindingRegistry[0].ObservedOutputDigest = "sha256:stale"
	return []ScenarioResult{
		failureScenario("duplicate-stable-id", validateManifests(duplicate, nil), "DUPLICATE_STABLE_ID"),
		failureScenario("missing-manifest", validateManifests(missing, required), "MISSING_MANIFEST"),
		failureScenario("stale-generated-projection", checkRendered(base, stale), "STALE_GENERATED_PROJECTION"),
		failureScenario("cross-directory-manifest", validateManifestInputs(root, cross, nil), "CROSS_DIRECTORY_MANIFEST"),
		failureScenario("missing-binding", validateManifests(missingBinding, nil), "MISSING_METRIC_BINDING"),
		failureScenario("malformed-manifest", validateManifests(malformed, nil), "INVALID_MANIFEST_IDENTITY"),
		failureScenario("binding-self-search", validateManifestInputs(root, bindingSelfSearch, nil), "UNTRUSTED_BINDING_SOURCE"),
		failureScenario("binding-output-digest-mismatch", validateManifestInputs(root, bindingOutputMismatch, nil), "BINDING_OUTPUT_DIGEST_MISMATCH"),
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

type consumerCommandResult struct {
	ExitCode       int
	Stdout         []byte
	Stderr         []byte
	DiagnosticJSON []byte
	Diagnostic     *Diagnostic
}

type consumerFailureCase struct {
	id, stage, step, reason, rawPath string
	rawPaths                         []string
	mutate                           func(string) error
}

func independentFailurePredicates(root string) ([]PredicateObservation, error) {
	cases := []consumerFailureCase{
		{id: "consumer-malformed-manifest", stage: "FOUNDATION", step: "DECODE_MANIFEST", reason: "MALFORMED_MANIFEST", rawPath: "examples/language-syntax-roundtrip/concept.manifest.json", mutate: func(temp string) error {
			return os.WriteFile(filepath.Join(temp, "examples/language-syntax-roundtrip/concept.manifest.json"), []byte("{"), 0o644)
		}},
		{id: "consumer-missing-manifest", stage: "FOUNDATION", step: "REQUIRED_MANIFEST", reason: "MISSING_MANIFEST", rawPath: "examples/language-syntax-roundtrip/concept.manifest.json.missing", mutate: func(temp string) error {
			return os.Rename(filepath.Join(temp, "examples/language-syntax-roundtrip/concept.manifest.json"), filepath.Join(temp, "examples/language-syntax-roundtrip/concept.manifest.json.missing"))
		}},
		{id: "consumer-cross-directory-manifest", stage: "FOUNDATION", step: "MANIFEST_OWNERSHIP", reason: "CROSS_DIRECTORY_MANIFEST", rawPath: "examples/wrong-owner/concept.manifest.json", rawPaths: []string{"examples/wrong-owner/concept.manifest.json"}, mutate: func(temp string) error {
			from := filepath.Join(temp, "examples/language-syntax-roundtrip/concept.manifest.json")
			to := filepath.Join(temp, "examples/wrong-owner/concept.manifest.json")
			if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
				return err
			}
			return os.Rename(from, to)
		}},
		{id: "consumer-missing-binding", stage: "FOUNDATION", step: "METRIC_BINDINGS", reason: "MISSING_METRIC_BINDING", rawPath: "examples/language-syntax-roundtrip/concept.manifest.json", mutate: func(temp string) error {
			return rewriteManifest(temp, "language-syntax-roundtrip", func(manifest *Manifest) { manifest.MetricBindings = nil; manifest.BindingRegistry = nil })
		}},
		{id: "consumer-stale-denominator", stage: "FOUNDATION", step: "DENOMINATOR_RECONCILIATION", reason: "DENOMINATOR_SOURCE_MISMATCH", rawPath: "examples/toolchain-conformance/concept.manifest.json", mutate: func(temp string) error {
			return rewriteManifest(temp, "toolchain-conformance", func(manifest *Manifest) { manifest.Denominators[0].Values["cases"] = 160 })
		}},
		{id: "consumer-stale-generated-projection", stage: "REGRESSION", step: "GENERATED_OUTPUT", reason: "STALE_GENERATED_PROJECTION", rawPath: defaultOutput + "/projection.json", mutate: func(temp string) error {
			path := filepath.Join(temp, defaultOutput, "projection.json")
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			data[0] ^= 1
			return os.WriteFile(path, data, 0o644)
		}},
		{id: "consumer-duplicate-stable-id", stage: "FOUNDATION", step: "UNIQUE_STABLE_ID", reason: "DUPLICATE_STABLE_ID", rawPath: "examples/duplicate-concept/concept.manifest.json", mutate: func(temp string) error {
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
		{id: "consumer-binding-self-search", stage: "FOUNDATION", step: "BINDING_REGISTRY", reason: "UNTRUSTED_BINDING_SOURCE", rawPath: "examples/language-syntax-roundtrip/concept.manifest.json", mutate: func(temp string) error {
			return rewriteManifest(temp, "language-syntax-roundtrip", func(manifest *Manifest) {
				manifest.BindingRegistry[0].RawSourceAddress = "docs/language/conflict-free-registry-projection.md#self-search"
			})
		}},
		{id: "consumer-binding-output-digest-mismatch", stage: "FOUNDATION", step: "BINDING_REGISTRY", reason: "BINDING_OUTPUT_DIGEST_MISMATCH", rawPath: "examples/language-syntax-roundtrip/concept.manifest.json", mutate: func(temp string) error {
			return rewriteManifest(temp, "language-syntax-roundtrip", func(manifest *Manifest) { manifest.BindingRegistry[0].ObservedOutputDigest = "sha256:stale" })
		}},
		{id: "consumer-binding-comment-only", stage: "FOUNDATION", step: "BINDING_REGISTRY", reason: "UNTRUSTED_BINDING_SOURCE", rawPath: "examples/language-syntax-roundtrip/concept.manifest.json", rawPaths: []string{"examples/language-syntax-roundtrip/concept.manifest.json", "scripts/conflict-free-registry-projection/testdata/comment-only-binding.go"}, mutate: func(temp string) error {
			path := filepath.Join(temp, "scripts/conflict-free-registry-projection/testdata/comment-only-binding.go")
			if err := os.WriteFile(path, []byte("package testdata\n// gooo.metric.language.syntax-roundtrip-readiness-bps.v1\n"), 0o644); err != nil {
				return err
			}
			return rewriteManifest(temp, "language-syntax-roundtrip", func(manifest *Manifest) {
				manifest.BindingRegistry[0].RawSourceAddress = "scripts/conflict-free-registry-projection/testdata/comment-only-binding.go#CommentOnly"
				manifest.BindingRegistry[0].RegistrationUseAddress = "scripts/conflict-free-registry-projection/testdata/comment-only-binding.go#CommentOnly@2:1"
			})
		}},
		{id: "consumer-binding-unused-string", stage: "FOUNDATION", step: "BINDING_REGISTRY", reason: "UNTRUSTED_BINDING_SOURCE", rawPath: "examples/language-syntax-roundtrip/concept.manifest.json", rawPaths: []string{"examples/language-syntax-roundtrip/concept.manifest.json", "scripts/conflict-free-registry-projection/testdata/unused-binding.go"}, mutate: func(temp string) error {
			path := filepath.Join(temp, "scripts/conflict-free-registry-projection/testdata/unused-binding.go")
			if err := os.WriteFile(path, []byte("package testdata\nvar Unused = \"gooo.metric.language.syntax-roundtrip-readiness-bps.v1\"\n"), 0o644); err != nil {
				return err
			}
			return rewriteManifest(temp, "language-syntax-roundtrip", func(manifest *Manifest) {
				manifest.BindingRegistry[0].RawSourceAddress = "scripts/conflict-free-registry-projection/testdata/unused-binding.go#Unused"
				manifest.BindingRegistry[0].RegistrationUseAddress = "scripts/conflict-free-registry-projection/testdata/unused-binding.go#Unused@2:5"
			})
		}},
		{id: "consumer-binding-cross-package-same-name", stage: "FOUNDATION", step: "BINDING_REGISTRY", reason: "UNTRUSTED_BINDING_SOURCE", rawPath: "examples/language-syntax-roundtrip/concept.manifest.json", rawPaths: []string{"examples/language-syntax-roundtrip/concept.manifest.json", "scripts/conflict-free-registry-projection/testdata/other-package/same-name.go"}, mutate: func(temp string) error {
			path := filepath.Join(temp, "scripts/conflict-free-registry-projection/testdata/other-package/same-name.go")
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(path, []byte("package otherpackage\n\nfunc concept(string) {}\n\nfunc use() {\n\tconcept(\"gooo.metric.language.syntax-roundtrip-readiness-bps.v1\")\n}\n"), 0o644); err != nil {
				return err
			}
			return rewriteManifest(temp, "language-syntax-roundtrip", func(manifest *Manifest) {
				manifest.BindingRegistry[0].RawSourceAddress = "internal/meta/languageconcept/catalog.go#concept"
				manifest.BindingRegistry[0].RegistrationUseAddress = "scripts/conflict-free-registry-projection/testdata/other-package/same-name.go#concept@6:2"
			})
		}},
		{id: "consumer-binding-shadowed-local", stage: "FOUNDATION", step: "BINDING_REGISTRY", reason: "UNTRUSTED_BINDING_SOURCE", rawPath: "examples/language-syntax-roundtrip/concept.manifest.json", rawPaths: []string{"examples/language-syntax-roundtrip/concept.manifest.json", "internal/meta/languageconcept/shadowed-binding.go"}, mutate: func(temp string) error {
			path := filepath.Join(temp, "internal/meta/languageconcept/shadowed-binding.go")
			if err := os.WriteFile(path, []byte("package languageconcept\n\nfunc shadowedBindingUse() {\n\tconcept := func(string) {}\n\tconcept(\"gooo.metric.language.syntax-roundtrip-readiness-bps.v1\")\n}\n"), 0o644); err != nil {
				return err
			}
			return rewriteManifest(temp, "language-syntax-roundtrip", func(manifest *Manifest) {
				manifest.BindingRegistry[0].RawSourceAddress = "internal/meta/languageconcept/catalog.go#concept"
				manifest.BindingRegistry[0].RegistrationUseAddress = "internal/meta/languageconcept/shadowed-binding.go#concept@5:2"
			})
		}},
		{id: "consumer-binding-unused-declaration", stage: "FOUNDATION", step: "BINDING_REGISTRY", reason: "UNTRUSTED_BINDING_SOURCE", rawPath: "examples/language-syntax-roundtrip/concept.manifest.json", rawPaths: []string{"examples/language-syntax-roundtrip/concept.manifest.json", "scripts/conflict-free-registry-projection/testdata/unused-declaration.go"}, mutate: func(temp string) error {
			path := filepath.Join(temp, "scripts/conflict-free-registry-projection/testdata/unused-declaration.go")
			if err := os.WriteFile(path, []byte("package testdata\nvar UnusedMetric = \"gooo.metric.language.syntax-roundtrip-readiness-bps.v1\"\n"), 0o644); err != nil {
				return err
			}
			return rewriteManifest(temp, "language-syntax-roundtrip", func(manifest *Manifest) {
				manifest.BindingRegistry[0].RawSourceAddress = "scripts/conflict-free-registry-projection/testdata/unused-declaration.go#UnusedMetric"
				manifest.BindingRegistry[0].RegistrationUseAddress = "scripts/conflict-free-registry-projection/testdata/unused-declaration.go#UnusedMetric@2:5"
			})
		}},
		{id: "consumer-binding-unrelated-use", stage: "FOUNDATION", step: "BINDING_REGISTRY", reason: "UNTRUSTED_BINDING_SOURCE", rawPath: "examples/language-syntax-roundtrip/concept.manifest.json", rawPaths: []string{"examples/language-syntax-roundtrip/concept.manifest.json", "internal/meta/languageconcept/unrelated-binding.go"}, mutate: func(temp string) error {
			path := filepath.Join(temp, "internal/meta/languageconcept/unrelated-binding.go")
			if err := os.WriteFile(path, []byte("package languageconcept\nvar _ = concept\n"), 0o644); err != nil {
				return err
			}
			return rewriteManifest(temp, "language-syntax-roundtrip", func(manifest *Manifest) {
				manifest.BindingRegistry[0].RawSourceAddress = "internal/meta/languageconcept/catalog.go#concept"
				manifest.BindingRegistry[0].RegistrationUseAddress = "internal/meta/languageconcept/unrelated-binding.go#concept@2:9"
			})
		}},
		{id: "consumer-binding-unresolved-import", stage: "LOWER_RESOLUTION", step: "BINDING_PACKAGE_TYPE_CHECK", reason: "PACKAGE_TYPE_CHECK_FAILED", rawPath: "examples/language-syntax-roundtrip/concept.manifest.json", rawPaths: []string{"internal/meta/languageconcept/unresolved-import-binding.go"}, mutate: func(temp string) error {
			path := filepath.Join(temp, "internal/meta/languageconcept/unresolved-import-binding.go")
			return os.WriteFile(path, []byte("package languageconcept\n\nimport _ \"example.invalid/unresolved-binding-import\"\n"), 0o644)
		}},
		{id: "consumer-binding-unrelated-type-error", stage: "LOWER_RESOLUTION", step: "BINDING_PACKAGE_TYPE_CHECK", reason: "PACKAGE_TYPE_CHECK_FAILED", rawPath: "examples/language-syntax-roundtrip/concept.manifest.json", rawPaths: []string{"internal/meta/languageconcept/unrelated-type-error-binding.go"}, mutate: func(temp string) error {
			path := filepath.Join(temp, "internal/meta/languageconcept/unrelated-type-error-binding.go")
			return os.WriteFile(path, []byte("package languageconcept\n\nvar unrelatedBindingTypeError string = 42\n"), 0o644)
		}},
		{id: "consumer-binding-metric-row-swap", stage: "FOUNDATION", step: "BINDING_REGISTRY", reason: "BINDING_SEMANTIC_DIGEST_MISMATCH", rawPath: "examples/language-syntax-roundtrip/concept.manifest.json", mutate: func(temp string) error {
			return rewriteManifest(temp, "language-syntax-roundtrip", func(manifest *Manifest) {
				manifest.BindingRegistry[0].MetricID, manifest.BindingRegistry[1].MetricID = manifest.BindingRegistry[1].MetricID, manifest.BindingRegistry[0].MetricID
			})
		}},
		{id: "consumer-binding-different-metric-literal", stage: "FOUNDATION", step: "BINDING_REGISTRY", reason: "UNTRUSTED_BINDING_SOURCE", rawPath: "examples/language-syntax-roundtrip/concept.manifest.json", mutate: func(temp string) error {
			return rewriteManifest(temp, "language-syntax-roundtrip", func(manifest *Manifest) {
				manifest.BindingRegistry[0].RegistrationUseAddress = "internal/meta/languageconcept/catalog.go#concept@5:3"
			})
		}},
		{id: "consumer-binding-unrelated-call", stage: "FOUNDATION", step: "BINDING_REGISTRY", reason: "UNTRUSTED_BINDING_SOURCE", rawPath: "examples/language-syntax-roundtrip/concept.manifest.json", rawPaths: []string{"examples/language-syntax-roundtrip/concept.manifest.json", "internal/meta/languageconcept/unrelated-binding-call.go"}, mutate: func(temp string) error {
			path := filepath.Join(temp, "internal/meta/languageconcept/unrelated-binding-call.go")
			if err := os.WriteFile(path, []byte("package languageconcept\n\nfunc unrelatedBindingCall() Concept {\n\treturn concept(\"language-syntax-roundtrip\", \"problem\", \"effect\", \"operation\", \"OPERATING\", nil, nil, UseCase{})\n}\n"), 0o644); err != nil {
				return err
			}
			return rewriteManifest(temp, "language-syntax-roundtrip", func(manifest *Manifest) {
				manifest.BindingRegistry[0].RegistrationUseAddress = "internal/meta/languageconcept/unrelated-binding-call.go#concept@4:9"
			})
		}},
	}
	observations := make([]PredicateObservation, 0, len(cases)+7)
	for _, item := range cases {
		temp, err := os.MkdirTemp("", "gooo-consumer-failure-")
		if err != nil {
			return nil, err
		}
		copyErr := copyTree(root, temp)
		if copyErr == nil {
			copyErr = item.mutate(temp)
		}
		result := consumerCommandResult{ExitCode: -1}
		var rawInput []byte
		var rawArtifacts []RawInputArtifact
		if copyErr == nil {
			rawInput, rawArtifacts, copyErr = readFailureInput(temp, item)
			if copyErr == nil {
				result = runConsumerFailureCommand(root, temp)
			}
		}
		observations = append(observations, failurePredicateObservation(item, result, rawInput, rawArtifacts))
		_ = os.RemoveAll(temp)
	}
	receiptObservations, err := receiptBoundaryFailurePredicates(root)
	if err != nil {
		return nil, err
	}
	observations = append(observations, receiptObservations...)
	observations = append(observations, successExitCounterexample())
	return observations, nil
}

func receiptBoundaryFailurePredicates(root string) ([]PredicateObservation, error) {
	base, err := runIndependentConsumerAt(root)
	if err != nil {
		return nil, err
	}
	mutations := []struct {
		id, reason string
		mutate     func(independentReceipt) (independentReceipt, error)
	}{
		{id: "consumer-receipt-occurrence-address-swap", reason: "BINDING_OCCURRENCE_ADDRESS_MISMATCH", mutate: func(receipt independentReceipt) (independentReceipt, error) {
			if len(receipt.BindingOutputReceipts) < 2 {
				return independentReceipt{}, receiptBoundaryError("RECEIPT_BINDING_INVENTORY_MISMATCH")
			}
			receipt.BindingOutputReceipts[0].MetricOccurrenceAddress, receipt.BindingOutputReceipts[1].MetricOccurrenceAddress = receipt.BindingOutputReceipts[1].MetricOccurrenceAddress, receipt.BindingOutputReceipts[0].MetricOccurrenceAddress
			return receipt, nil
		}},
		{id: "consumer-receipt-occurrence-digest-swap", reason: "BINDING_OCCURRENCE_DIGEST_MISMATCH", mutate: func(receipt independentReceipt) (independentReceipt, error) {
			if len(receipt.BindingOutputReceipts) < 2 {
				return independentReceipt{}, receiptBoundaryError("RECEIPT_BINDING_INVENTORY_MISMATCH")
			}
			receipt.BindingOutputReceipts[0].MetricOccurrenceDigest, receipt.BindingOutputReceipts[1].MetricOccurrenceDigest = receipt.BindingOutputReceipts[1].MetricOccurrenceDigest, receipt.BindingOutputReceipts[0].MetricOccurrenceDigest
			return receipt, nil
		}},
		{id: "consumer-receipt-occurrence-pair-swap", reason: "BINDING_OCCURRENCE_PAIR_MISMATCH", mutate: func(receipt independentReceipt) (independentReceipt, error) {
			if len(receipt.BindingOutputReceipts) < 2 {
				return independentReceipt{}, receiptBoundaryError("RECEIPT_BINDING_INVENTORY_MISMATCH")
			}
			receipt.BindingOutputReceipts[0].MetricOccurrenceAddress, receipt.BindingOutputReceipts[1].MetricOccurrenceAddress = receipt.BindingOutputReceipts[1].MetricOccurrenceAddress, receipt.BindingOutputReceipts[0].MetricOccurrenceAddress
			receipt.BindingOutputReceipts[0].MetricOccurrenceDigest, receipt.BindingOutputReceipts[1].MetricOccurrenceDigest = receipt.BindingOutputReceipts[1].MetricOccurrenceDigest, receipt.BindingOutputReceipts[0].MetricOccurrenceDigest
			return receipt, nil
		}},
		{id: "consumer-receipt-output-row-cross-swap", reason: "BINDING_OUTPUT_ROW_TRIPLE_MISMATCH", mutate: func(receipt independentReceipt) (independentReceipt, error) {
			if len(receipt.BindingOutputReceipts) < 2 {
				return independentReceipt{}, receiptBoundaryError("RECEIPT_BINDING_INVENTORY_MISMATCH")
			}
			receipt.BindingOutputReceipts[0].OutputRowAddress, receipt.BindingOutputReceipts[1].OutputRowAddress = receipt.BindingOutputReceipts[1].OutputRowAddress, receipt.BindingOutputReceipts[0].OutputRowAddress
			receipt.BindingOutputReceipts[0].OutputRowDigest, receipt.BindingOutputReceipts[1].OutputRowDigest = receipt.BindingOutputReceipts[1].OutputRowDigest, receipt.BindingOutputReceipts[0].OutputRowDigest
			return receipt, nil
		}},
		{id: "consumer-receipt-unknown-field", reason: "RECEIPT_UNKNOWN_FIELD", mutate: func(receipt independentReceipt) (independentReceipt, error) {
			return receipt, nil
		}},
		{id: "consumer-receipt-duplicate-metric-id", reason: "DUPLICATE_RECEIPT_METRIC_ID", mutate: func(receipt independentReceipt) (independentReceipt, error) {
			if len(receipt.BindingOutputReceipts) < 2 {
				return independentReceipt{}, receiptBoundaryError("RECEIPT_BINDING_INVENTORY_MISMATCH")
			}
			receipt.BindingOutputReceipts[1].MetricID = receipt.BindingOutputReceipts[0].MetricID
			return receipt, nil
		}},
	}
	observations := make([]PredicateObservation, 0, len(mutations))
	for _, mutation := range mutations {
		var raw []byte
		if mutation.id == "consumer-receipt-unknown-field" {
			var object map[string]json.RawMessage
			if err := json.Unmarshal(base.rawReceipt, &object); err != nil {
				return nil, err
			}
			object["unknown_receipt_field"] = json.RawMessage(`"untrusted"`)
			raw, err = json.Marshal(object)
		} else {
			clone, cloneErr := cloneIndependentReceipt(base)
			if cloneErr != nil {
				return nil, cloneErr
			}
			mutated, mutateErr := mutation.mutate(clone)
			if mutateErr != nil {
				return nil, mutateErr
			}
			raw, err = json.Marshal(mutated)
		}
		if err != nil {
			return nil, err
		}
		receipt, decodeErr := decodeIndependentReceipt(raw)
		var diagnostic *Diagnostic
		if decodeErr != nil {
			diagnostic = receiptBoundaryDiagnostic(decodeErr)
		} else if validateErr := validateIndependentReceipt(root, receipt, base.outputRaw); validateErr != nil {
			diagnostic = receiptBoundaryDiagnostic(validateErr)
		}
		if diagnostic == nil {
			return nil, fmt.Errorf("receipt corruption was accepted: %s", mutation.id)
		}
		diagnosticJSON, marshalErr := json.Marshal(diagnostic)
		if marshalErr != nil {
			return nil, marshalErr
		}
		item := consumerFailureCase{id: mutation.id, stage: diagnostic.Stage, step: diagnostic.Step, reason: mutation.reason}
		artifact := RawInputArtifact{Path: "embedded://receipt-boundary/" + mutation.id + ".json", Bytes: raw, Digest: digestBytes(raw)}
		rawInput, marshalErr := json.Marshal([]RawInputArtifact{artifact})
		if marshalErr != nil {
			return nil, marshalErr
		}
		result := consumerCommandResult{ExitCode: 1, DiagnosticJSON: diagnosticJSON, Diagnostic: diagnostic}
		observations = append(observations, failurePredicateObservation(item, result, rawInput, []RawInputArtifact{artifact}))
	}
	return observations, nil
}

func cloneIndependentReceipt(receipt independentReceipt) (independentReceipt, error) {
	raw, err := json.Marshal(receipt)
	if err != nil {
		return independentReceipt{}, err
	}
	return decodeIndependentReceipt(raw)
}

func runConsumerFailureCommand(repoRoot, tempRoot string) consumerCommandResult {
	command := exec.Command("go", "run", "./scripts/conflict-free-registry-projection-consumer", "-root", tempRoot, "-check-generated")
	command.Dir = repoRoot
	return runCommand(command)
}

func runCommand(command *exec.Cmd) consumerCommandResult {
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	exitCode := -1
	if command.ProcessState != nil {
		exitCode = command.ProcessState.ExitCode()
	}
	diagnosticJSON, diagnostic := parseExactDiagnostic(stdout.Bytes(), stderr.Bytes())
	if err == nil && exitCode < 0 {
		exitCode = 0
	}
	return consumerCommandResult{ExitCode: exitCode, Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), DiagnosticJSON: diagnosticJSON, Diagnostic: diagnostic}
}

func parseExactDiagnostic(streams ...[]byte) ([]byte, *Diagnostic) {
	for _, stream := range streams {
		trimmed := bytes.TrimSpace(stream)
		if len(trimmed) == 0 {
			continue
		}
		decoder := json.NewDecoder(bytes.NewReader(trimmed))
		decoder.DisallowUnknownFields()
		diagnostic := Diagnostic{}
		if decoder.Decode(&diagnostic) != nil {
			continue
		}
		if decoder.Decode(&struct{}{}) != io.EOF || diagnostic.Decision == "" || diagnostic.Stage == "" || diagnostic.Step == "" || diagnostic.Reason == "" {
			continue
		}
		return append([]byte(nil), stream...), &diagnostic
	}
	return nil, nil
}

func readFailureInput(temp string, item consumerFailureCase) ([]byte, []RawInputArtifact, error) {
	paths := append([]string(nil), item.rawPaths...)
	if len(paths) == 0 {
		paths = []string{item.rawPath}
	}
	sort.Strings(paths)
	artifacts := make([]RawInputArtifact, 0, len(paths))
	for _, path := range paths {
		data, err := os.ReadFile(filepath.Join(temp, filepath.FromSlash(path)))
		if err != nil {
			return nil, nil, err
		}
		artifacts = append(artifacts, RawInputArtifact{Path: path, Bytes: data, Digest: digestBytes(data)})
	}
	envelope, err := json.Marshal(artifacts)
	if err != nil {
		return nil, nil, err
	}
	return envelope, artifacts, nil
}

func failurePredicateObservation(item consumerFailureCase, result consumerCommandResult, rawInput []byte, rawArtifacts []RawInputArtifact) PredicateObservation {
	stage, step, reason, decision := "UNKNOWN", "UNKNOWN", "UNKNOWN", "UNKNOWN"
	if result.Diagnostic != nil {
		stage, step, reason, decision = result.Diagnostic.Stage, result.Diagnostic.Step, result.Diagnostic.Reason, result.Diagnostic.Decision
	}
	exact := consumerFailureAccepted(result, item)
	truth := "UNKNOWN"
	if result.Diagnostic != nil {
		truth = "FALSE"
		if result.ExitCode != 0 && exact {
			truth = "TRUE"
		}
	}
	diagnosticDigest := ""
	if len(result.DiagnosticJSON) > 0 {
		diagnosticDigest = digestBytes(result.DiagnosticJSON)
	}
	rawDigest := digestBytes(rawInput)
	address := "evidence://consumer-failure/" + item.id
	targetDigest := digestBytes([]byte(address + "|" + rawDigest + "|" + diagnosticDigest + fmt.Sprint(result.ExitCode)))
	return PredicateObservation{ID: item.id, ObservedPredicate: "independent consumer rejects " + item.reason + " with nonzero exit and exact diagnostic fields", TargetAddress: address, TargetDigest: targetDigest, Observed: truth == "TRUE", Decision: decision, PredicateTruth: truth, ExitCode: result.ExitCode, DiagnosticJSON: string(result.DiagnosticJSON), DiagnosticDigest: diagnosticDigest, RawInputDigest: rawDigest, RawInputBytes: string(rawInput), RawInputArtifacts: rawArtifacts, ContentDigest: diagnosticDigest, Stage: stage, Step: step, Reason: reason}
}

func successExitCounterexample() PredicateObservation {
	raw := []byte("{\"decision\":\"FAIL_CLOSED\",\"stage\":\"REGRESSION\",\"step\":\"GENERATED_OUTPUT\",\"reason\":\"STALE_GENERATED_PROJECTION\"}\n")
	result := runCommand(exec.Command("printf", "%s", string(raw)))
	rawArtifacts := []RawInputArtifact{{Path: "embedded://classifier-input/diagnostic.json", Bytes: raw, Digest: digestBytes(raw)}}
	rawInput, _ := json.Marshal(rawArtifacts)
	rawDigest := digestBytes(raw)
	diagnosticDigest := digestBytes(result.DiagnosticJSON)
	address := "evidence://classifier-counterexample/classifier-success-exit-counterexample"
	stage, step, reason := "UNKNOWN", "UNKNOWN", "UNKNOWN"
	if result.Diagnostic != nil {
		stage, step, reason = result.Diagnostic.Stage, result.Diagnostic.Step, result.Diagnostic.Reason
	}
	acceptedAsFailure := consumerFailureAccepted(result, consumerFailureCase{stage: "REGRESSION", step: "GENERATED_OUTPUT", reason: "STALE_GENERATED_PROJECTION"})
	observed := result.ExitCode == 0 && result.Diagnostic != nil && !acceptedAsFailure
	truth := "UNKNOWN"
	if observed {
		truth = "TRUE"
	}
	return PredicateObservation{ID: "classifier-success-exit-counterexample", ObservedPredicate: "the failure classifier rejects diagnostic JSON with a success exit as a failure observation", TargetAddress: address, TargetDigest: digestBytes([]byte(address + "|" + rawDigest + "|" + diagnosticDigest)), Observed: observed, Decision: "PASS", PredicateTruth: truth, ExitCode: result.ExitCode, DiagnosticJSON: string(result.DiagnosticJSON), DiagnosticDigest: diagnosticDigest, RawInputDigest: digestBytes(rawInput), RawInputBytes: string(rawInput), RawInputArtifacts: rawArtifacts, ContentDigest: diagnosticDigest, Stage: stage, Step: step, Reason: reason}
}

func consumerFailureAccepted(result consumerCommandResult, expected consumerFailureCase) bool {
	return result.ExitCode != 0 && result.Diagnostic != nil && result.Diagnostic.Decision == "FAIL_CLOSED" && result.Diagnostic.Stage == expected.stage && result.Diagnostic.Step == expected.step && result.Diagnostic.Reason == expected.reason
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
			result.Scenarios = []string{"projection-twice-byte-equality", "manifest-order-invariance", "independent-conformance-consumer"}
			if !selected || evidence.ProjectionReplay.Decision != "PASS" || evidence.ManifestOrderInvariant.Decision != "PASS" || evidence.ConformanceConsumer.Numerator != 1 {
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
		if observation.PredicateTruth == "TRUE" {
			terminalState = "DISCHARGED"
			terminalReason = "independent_consumer_recomputed_predicate"
			terminalStage = observation.Stage
			terminalStep = "INDEPENDENT_CONSUMER_PREDICATE"
		} else if observation.PredicateTruth == "FALSE" {
			terminalState = "REFUTED"
		}
		claims = append(claims, Claim{ID: observation.ID, Proposition: observation.ObservedPredicate, ObservedPredicate: observation.ObservedPredicate, TargetAddress: observation.TargetAddress, TargetDigest: observation.TargetDigest, Transitions: []ClaimTransition{{State: "OPEN", ObservedPredicate: observation.ObservedPredicate, PredicateTruth: "UNKNOWN", PredicateDigest: predicateDigest, TargetAddress: observation.TargetAddress, TargetDigest: observation.TargetDigest, Stage: "FOUNDATION", Step: "CLAIM_OPEN", Reason: "proposition_under_review"}, {State: terminalState, ObservedPredicate: observation.ObservedPredicate, PredicateTruth: observation.PredicateTruth, PredicateDigest: predicateDigest, TargetAddress: observation.TargetAddress, TargetDigest: observation.TargetDigest, Stage: terminalStage, Step: terminalStep, Reason: terminalReason}}})
	}
	return claims
}

func allProofsPass(evidence Evidence) bool {
	if len(expectedConformancePredicateIDs) != expectedConformancePredicateCount || hasDuplicateStrings(expectedConformancePredicateIDs) || len(expectedFailurePredicateIDs) != expectedFailurePredicateCount || hasDuplicateStrings(expectedFailurePredicateIDs) {
		return false
	}
	for _, scenario := range []ScenarioResult{evidence.ProjectionReplay, evidence.ManifestOrderInvariant, evidence.SemanticCausality, evidence.SemanticMetricChange, evidence.CommentInvariant, evidence.CommentPositionInvariant, evidence.NewConceptFixture, evidence.DenominatorMismatch} {
		if scenario.Decision != "PASS" {
			return false
		}
	}
	if len(evidence.DenominatorReconciliations) < 3 || uniqueStableIDs(evidence.DenominatorReconciliations) != 3 || evidence.StaleDenominatorReceipt == nil || evidence.StaleDenominatorReceipt.Decision != "FAIL_CLOSED" || evidence.StaleDenominatorReceipt.Reason != "DENOMINATOR_SOURCE_MISMATCH" || evidence.GeneratedOutputChangeCount != 6 || evidence.GeneratedOutputDenominator != 8 || evidence.ConformanceConsumer.Numerator != 1 || evidence.ConformanceConsumer.Denominator != 1 || evidence.ProductionAdoption.Numerator != 0 || evidence.ProductionAdoption.Denominator != 1 || evidence.ProductionAdoption.Decision != "UNKNOWN" || evidence.ASTResolvedBindings.Numerator != expectedBindingReceiptCount || evidence.ASTResolvedBindings.Denominator != expectedBindingReceiptCount {
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
	if len(evidence.FailureContracts) != expectedFailureContractCount {
		return false
	}
	if len(evidence.Claims) != expectedClaimCount || len(evidence.PredicateObservations) != expectedConformancePredicateCount+expectedFailurePredicateCount {
		return false
	}
	if evidence.ClaimTransitions.Denominator != expectedClaimCount || evidence.FailurePredicates.Denominator != expectedFailurePredicateCount || evidence.ProvenancePredicates.Denominator != expectedFailurePredicateCount {
		return false
	}
	if len(evidence.BindingOutputReceipts) != expectedBindingReceiptCount || evidence.MetricOccurrences.Numerator != expectedBindingReceiptCount || evidence.MetricOccurrences.Denominator != expectedBindingReceiptCount || evidence.UniqueSemanticRelationDigests.Numerator != expectedBindingReceiptCount || evidence.UniqueSemanticRelationDigests.Denominator != expectedBindingReceiptCount || evidence.OutputRowAddresses.Numerator != expectedBindingReceiptCount || evidence.OutputRowAddresses.Denominator != expectedBindingReceiptCount || evidence.ConsumerOutputArtifact.ObservedPath != embeddedOutputAddress || evidence.ConsumerOutputArtifact.Bytes <= 0 || len(evidence.ConsumerOutputArtifact.RawBytes) != evidence.ConsumerOutputArtifact.Bytes || evidence.ConsumerOutputArtifact.Digest != digestBytes([]byte(evidence.ConsumerOutputArtifact.RawBytes)) {
		return false
	}
	if evidence.Metrics.ProducerPackageImports.Numerator != 0 || evidence.Metrics.ProducerPackageImports.Denominator != 1 || evidence.Metrics.ProducerPackageImports.Decision != "PASS" || evidence.Metrics.RawSourceReconstruction.Numerator != 1 || evidence.Metrics.RawSourceReconstruction.Denominator != 1 || evidence.Metrics.RawSourceReconstruction.Decision != "PASS" || evidence.Metrics.SeparateExecutable.Numerator != 1 || evidence.Metrics.SeparateExecutable.Denominator != 1 || evidence.Metrics.SeparateExecutable.Decision != "PASS" || evidence.Metrics.AlgorithmicIndependence.Numerator != 0 || evidence.Metrics.AlgorithmicIndependence.Denominator != 1 || evidence.Metrics.AlgorithmicIndependence.Decision != "UNKNOWN" {
		return false
	}
	for _, scenario := range evidence.FailureContracts {
		if scenario.Decision != "PASS" || scenario.Diagnostic == nil {
			return false
		}
	}
	for _, predicate := range evidence.PredicateObservations {
		if !predicate.Observed || predicate.PredicateTruth != "TRUE" || predicate.TargetAddress == "" || predicate.TargetDigest == "" {
			return false
		}
	}
	for _, strategy := range evidence.Strategies {
		if !strategy.Selected || strategy.Decision != "PASS" {
			return false
		}
	}
	return validateClaims(evidence.Claims) == nil && evidence.RepositoryNetState.NetStateEqual && evidence.RepositoryNetState.NetState == "NET_STATE_EQUAL" && evidence.RepositoryNetState.NetStatePredicate == "NET_STATE_EQUAL" && evidence.FailurePredicates.Numerator == evidence.FailurePredicates.Denominator && evidence.BindingPredicates.Numerator == evidence.BindingPredicates.Denominator && evidence.ProvenancePredicates.Numerator == evidence.ProvenancePredicates.Denominator && evidence.PredicateInventory.Decision == "PASS" && evidence.FailureInventory.Decision == "PASS" && evidence.ClaimInventory.Decision == "PASS" && evidence.ProvenanceInventory.Decision == "PASS"
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
			if transition.ObservedPredicate != item.ObservedPredicate || transition.TargetAddress != item.TargetAddress || transition.TargetDigest != item.TargetDigest || transition.PredicateDigest == "" || transition.PredicateTruth == "" {
				return fmt.Errorf("claim transition is not target-bound")
			}
		}
		if last.State == "DISCHARGED" && last.Reason != "independent_consumer_recomputed_predicate" {
			return fmt.Errorf("claim discharged without independent predicate")
		}
		if last.State == "DISCHARGED" && last.PredicateTruth != "TRUE" {
			return fmt.Errorf("claim discharged without true observed predicate")
		}
		if last.State == "REFUTED" && last.PredicateTruth != "FALSE" {
			return fmt.Errorf("claim refuted without false observed predicate")
		}
		if last.State == "OPEN" && last.PredicateTruth != "UNKNOWN" {
			return fmt.Errorf("open claim without unknown observed predicate")
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

func repositoryObservation(before, after []RepositoryFileSnapshot, beforeErr, afterErr error) RepositoryObservation {
	beforePaths := snapshotPaths(before)
	afterPaths := snapshotPaths(after)
	equal := beforeErr == nil && afterErr == nil && sameRepositorySnapshots(before, after)
	netState := "UNKNOWN"
	if equal {
		netState = "NET_STATE_EQUAL"
	} else if beforeErr == nil && afterErr == nil {
		netState = "NET_STATE_CHANGED"
	}
	return RepositoryObservation{BeforePaths: beforePaths, AfterPaths: afterPaths, BeforeFiles: before, AfterFiles: after, NetStateEqual: equal, NetState: netState, NetStatePredicate: netState, TransientMutation: "UNKNOWN", MutationAuthority: "UNKNOWN"}
}

func gitRepositorySnapshot(root string) ([]RepositoryFileSnapshot, error) {
	tracked, err := gitFileList(root, "ls-files", "-z")
	if err != nil {
		return nil, err
	}
	untracked, err := gitFileList(root, "ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return nil, err
	}
	paths := map[string]bool{}
	for _, path := range tracked {
		paths[path] = true
	}
	for _, path := range untracked {
		if _, ok := paths[path]; !ok {
			paths[path] = false
		}
	}
	snapshots := make([]RepositoryFileSnapshot, 0, len(paths))
	for path, isTracked := range paths {
		data, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		digest := "MISSING"
		if readErr == nil {
			digest = digestBytes(data)
		}
		snapshots = append(snapshots, RepositoryFileSnapshot{Path: filepath.ToSlash(path), Digest: digest, Tracked: isTracked})
	}
	sort.Slice(snapshots, func(i, j int) bool {
		if snapshots[i].Path == snapshots[j].Path {
			return !snapshots[i].Tracked && snapshots[j].Tracked
		}
		return snapshots[i].Path < snapshots[j].Path
	})
	return snapshots, nil
}

func gitFileList(root string, args ...string) ([]string, error) {
	command := exec.Command("git", append([]string{"-C", root}, args...)...)
	data, err := command.Output()
	if err != nil {
		return nil, err
	}
	parts := bytes.Split(data, []byte{0})
	paths := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(part) > 0 {
			paths = append(paths, filepath.ToSlash(string(part)))
		}
	}
	return paths, nil
}

func sameRepositorySnapshots(left, right []RepositoryFileSnapshot) bool {
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

func snapshotPaths(values []RepositoryFileSnapshot) []string {
	paths := make([]string, 0, len(values))
	for _, value := range values {
		paths = append(paths, value.Path)
	}
	return paths
}

func predicateMetric(numerator, denominator int, decision, stage, step, reason string) PredicateMetric {
	return PredicateMetric{Numerator: numerator, Denominator: denominator, Decision: decision, Stage: stage, Step: step, Reason: reason}
}

func decisionForBool(value bool) string {
	if value {
		return "PASS"
	}
	return "FAIL_CLOSED"
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func countDischargedClaims(claims []Claim) int {
	count := 0
	for _, claim := range claims {
		if len(claim.Transitions) > 0 && claim.Transitions[len(claim.Transitions)-1].State == "DISCHARGED" {
			count++
		}
	}
	return count
}

func failurePredicateMetric(predicates []PredicateObservation) PredicateMetric {
	numerator := 0
	byID := make(map[string]PredicateObservation, len(predicates))
	for _, predicate := range predicates {
		byID[predicate.ID] = predicate
	}
	for _, id := range expectedFailurePredicateIDs {
		if predicate, ok := byID[id]; ok && predicate.PredicateTruth == "TRUE" {
			numerator++
		}
	}
	return predicateMetric(numerator, expectedFailurePredicateCount, decisionForRatio(numerator, expectedFailurePredicateCount), "REGRESSION", "FAILURE_PREDICATES", "EXACT_DIAGNOSTIC_AND_EXIT_CONTRACTS")
}

func bindingPredicateMetric(predicates []PredicateObservation, receipts []BindingOutputReceipt) PredicateMetric {
	for _, predicate := range predicates {
		if predicate.ID == "independent-binding-registry" {
			valid := predicate.PredicateTruth == "TRUE" && len(receipts) == expectedBindingReceiptCount
			return predicateMetric(boolInt(valid), 1, decisionForBool(valid), "FOUNDATION", "BINDING_REGISTRY", "INDEPENDENT_BINDING_RECONSTRUCTION")
		}
	}
	return predicateMetric(0, 1, "FAIL_CLOSED", "FOUNDATION", "BINDING_REGISTRY", "MISSING_BINDING_PREDICATE")
}

func metricOccurrenceMetric(root string, receipts []BindingOutputReceipt) PredicateMetric {
	pairs := map[string]struct{}{}
	addresses := map[string]string{}
	digests := map[string]string{}
	metricIDs := map[string]struct{}{}
	for _, receipt := range receipts {
		if receipt.MetricID == "" || receipt.MetricOccurrenceAddress == "" || receipt.MetricOccurrenceDigest == "" {
			continue
		}
		if _, duplicate := metricIDs[receipt.MetricID]; duplicate {
			continue
		}
		metricIDs[receipt.MetricID] = struct{}{}
		relation, err := resolveBindingRelation(root, receipt.RawSourceAddress, receipt.RegistrationUseAddress, receipt.MetricID)
		if err != nil || relation.MetricOccurrenceAddress != receipt.MetricOccurrenceAddress || relation.MetricOccurrenceDigest != receipt.MetricOccurrenceDigest {
			continue
		}
		if prior, exists := addresses[receipt.MetricOccurrenceAddress]; exists && prior != receipt.MetricOccurrenceDigest {
			continue
		}
		if prior, exists := digests[receipt.MetricOccurrenceDigest]; exists && prior != receipt.MetricOccurrenceAddress {
			continue
		}
		addresses[receipt.MetricOccurrenceAddress] = receipt.MetricOccurrenceDigest
		digests[receipt.MetricOccurrenceDigest] = receipt.MetricOccurrenceAddress
		pairs[metricOccurrencePairKey(receipt.MetricOccurrenceAddress, receipt.MetricOccurrenceDigest)] = struct{}{}
	}
	numerator := len(pairs)
	return predicateMetric(numerator, expectedBindingReceiptCount, decisionForRatio(numerator, expectedBindingReceiptCount), "FOUNDATION", "METRIC_OCCURRENCE", "EXACT_METRIC_OCCURRENCE_PAIRS_RECONSTRUCTED_AND_BIJECTIVE")
}

func uniqueSemanticRelationMetric(receipts []BindingOutputReceipt) PredicateMetric {
	seen := map[string]struct{}{}
	for _, receipt := range receipts {
		if receipt.SemanticDigest != "" {
			seen[receipt.SemanticDigest] = struct{}{}
		}
	}
	return predicateMetric(len(seen), expectedBindingReceiptCount, decisionForRatio(len(seen), expectedBindingReceiptCount), "FOUNDATION", "SEMANTIC_RELATION_UNIQUENESS", "METRIC_SPECIFIC_RELATION_DIGESTS_UNIQUE")
}

func uniqueOutputRowMetric(receipts []BindingOutputReceipt, outputRaw []byte) PredicateMetric {
	projection := Projection{}
	decoder := json.NewDecoder(bytes.NewReader(outputRaw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&projection) != nil {
		return predicateMetric(0, expectedBindingReceiptCount, "FAIL_CLOSED", "COHERENCE", "OUTPUT_ROW_UNIQUENESS", "OUTPUT_PROJECTION_RECONSTRUCTION_FAILED")
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return predicateMetric(0, expectedBindingReceiptCount, "FAIL_CLOSED", "COHERENCE", "OUTPUT_ROW_UNIQUENESS", "OUTPUT_PROJECTION_TRAILING_JSON")
	}
	expected := map[string]BindingRegistryEntry{}
	expectedMetricIDs := map[string]struct{}{}
	for _, entry := range projection.Catalog {
		for _, binding := range entry.BindingRegistry {
			address := bindingOutputRowAddress(entry.StableID, binding.MetricID)
			if _, duplicate := expected[address]; duplicate {
				return predicateMetric(0, expectedBindingReceiptCount, "FAIL_CLOSED", "COHERENCE", "OUTPUT_ROW_UNIQUENESS", "DUPLICATE_OUTPUT_ROW_ADDRESS")
			}
			if _, duplicate := expectedMetricIDs[binding.MetricID]; duplicate {
				return predicateMetric(0, expectedBindingReceiptCount, "FAIL_CLOSED", "COHERENCE", "OUTPUT_ROW_UNIQUENESS", "DUPLICATE_OUTPUT_ROW_METRIC_ID")
			}
			expectedMetricIDs[binding.MetricID] = struct{}{}
			expected[address] = binding
		}
	}
	if len(expected) != expectedBindingReceiptCount {
		return predicateMetric(0, expectedBindingReceiptCount, "FAIL_CLOSED", "COHERENCE", "OUTPUT_ROW_UNIQUENESS", "OUTPUT_ROW_INVENTORY_MISMATCH")
	}
	triples := map[string]struct{}{}
	addressToTriple := map[string]string{}
	tripleToAddress := map[string]string{}
	digestMetricToAddress := map[string]string{}
	seenMetricIDs := map[string]struct{}{}
	for _, receipt := range receipts {
		if _, duplicate := seenMetricIDs[receipt.MetricID]; duplicate {
			continue
		}
		seenMetricIDs[receipt.MetricID] = struct{}{}
		row, ok := expected[receipt.OutputRowAddress]
		if !ok || row.MetricID != receipt.MetricID {
			continue
		}
		rowDigest := digestJSONValue(row)
		if receipt.OutputRowDigest != rowDigest {
			continue
		}
		triple := outputRowTripleKey(receipt.OutputRowAddress, rowDigest, receipt.MetricID)
		if prior, exists := addressToTriple[receipt.OutputRowAddress]; exists && prior != triple {
			continue
		}
		if prior, exists := tripleToAddress[triple]; exists && prior != receipt.OutputRowAddress {
			continue
		}
		digestMetric := rowDigest + "\x00" + receipt.MetricID
		if prior, exists := digestMetricToAddress[digestMetric]; exists && prior != receipt.OutputRowAddress {
			continue
		}
		addressToTriple[receipt.OutputRowAddress] = triple
		tripleToAddress[triple] = receipt.OutputRowAddress
		digestMetricToAddress[digestMetric] = receipt.OutputRowAddress
		triples[triple] = struct{}{}
	}
	numerator := len(triples)
	return predicateMetric(numerator, expectedBindingReceiptCount, decisionForRatio(numerator, expectedBindingReceiptCount), "COHERENCE", "OUTPUT_ROW_UNIQUENESS", "EXACT_OUTPUT_ROW_ADDRESS_DIGEST_METRIC_TRIPLES_RECONSTRUCTED")
}

func metricOccurrencePairKey(address, digest string) string {
	return address + "\x00" + digest
}

func outputRowTripleKey(address, digest, metricID string) string {
	return address + "\x00" + digest + "\x00" + metricID
}

func provenancePredicateMetric(predicates []PredicateObservation) PredicateMetric {
	numerator := 0
	byID := make(map[string]PredicateObservation, len(predicates))
	for _, predicate := range predicates {
		byID[predicate.ID] = predicate
	}
	for _, id := range expectedFailurePredicateIDs {
		if predicate, ok := byID[id]; ok && predicate.ExitCode != -1 && predicate.DiagnosticJSON != "" && predicate.DiagnosticDigest == digestBytes([]byte(predicate.DiagnosticJSON)) && rawInputArtifactsValid(predicate) && predicate.ContentDigest == predicate.DiagnosticDigest {
			numerator++
		}
	}
	return predicateMetric(numerator, expectedFailurePredicateCount, decisionForRatio(numerator, expectedFailurePredicateCount), "REGRESSION", "PROVENANCE_PREDICATES", "EXIT_DIAGNOSTIC_RAW_INPUT_AND_ARTIFACT_PRESERVED")
}

func rawInputArtifactsValid(predicate PredicateObservation) bool {
	if len(predicate.RawInputArtifacts) == 0 || predicate.RawInputBytes == "" {
		return false
	}
	for _, artifact := range predicate.RawInputArtifacts {
		if artifact.Path == "" || artifact.Digest != digestBytes(artifact.Bytes) {
			return false
		}
	}
	envelope, err := json.Marshal(predicate.RawInputArtifacts)
	return err == nil && string(envelope) == predicate.RawInputBytes && predicate.RawInputDigest == digestBytes(envelope)
}

func decisionForRatio(numerator, denominator int) string {
	if denominator > 0 && numerator == denominator {
		return "PASS"
	}
	return "FAIL_CLOSED"
}

func makeInventory(expected, observed []string, stage, step, reason string) InventoryReceipt {
	decision := "PASS"
	if len(expected) != len(observed) || !sameStringSet(expected, observed) || hasDuplicateStrings(observed) {
		decision = "FAIL_CLOSED"
		reason = "FIXED_ID_INVENTORY_MISMATCH"
	}
	return InventoryReceipt{Expected: append([]string(nil), expected...), Observed: append([]string(nil), observed...), Decision: decision, Stage: stage, Step: step, Reason: reason}
}

func predicateIDs(values []PredicateObservation) []string {
	ids := make([]string, 0, len(values))
	for _, value := range values {
		ids = append(ids, value.ID)
	}
	return ids
}

func claimIDs(values []Claim) []string {
	ids := make([]string, 0, len(values))
	for _, value := range values {
		ids = append(ids, value.ID)
	}
	return ids
}

func selectPredicateIDs(values []PredicateObservation, expected []string) []string {
	wanted := make(map[string]struct{}, len(expected))
	for _, id := range expected {
		wanted[id] = struct{}{}
	}
	ids := []string{}
	for _, value := range values {
		if _, ok := wanted[value.ID]; ok {
			ids = append(ids, value.ID)
		}
	}
	return ids
}

func hasDuplicateStrings(values []string) bool {
	seen := map[string]struct{}{}
	for _, value := range values {
		if _, ok := seen[value]; ok {
			return true
		}
		seen[value] = struct{}{}
	}
	return false
}

func observedOutputs(outputDir string) []OutputMetadata {
	outputs, err := readGenerated(outputDir)
	if err != nil {
		return nil
	}
	return outputMetadata("", outputDir, outputs)
}
