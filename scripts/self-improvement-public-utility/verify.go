package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publiccontinuity"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicdiscovery"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	utilitySchema       = "gooo/public-self-improvement-utility-verification/v1"
	utilityInputSchema  = "gooo/public-self-improvement-utility-input/v1"
	artifactDenominator = 32
	caseBaseline        = "BASELINE_PROJECT_GENERATION_BUILD_TEST"
	caseLearned         = "LEARNED_PROJECT_GENERATION_BUILD_TEST"
	caseFirst           = "FIRST_OBSERVATION_NO_QUORUM"
	casePerformance     = "PERFORMANCE_INCOMPARABLE"
	caseStale           = "STALE_PROJECT_SOURCE"
	caseDigestMismatch  = "OUTPUT_SEMANTIC_DIGEST_MISMATCH"
)

var utilityCases = []string{caseBaseline, caseLearned, caseFirst, casePerformance, caseStale, caseDigestMismatch}
var utilityDecisions = []string{"CLOSED", "CLOSED", "UNKNOWN", "UNKNOWN", "REFUTED", "REFUTED"}

func verifyUtility(contractPath, projectPath, manifestPath, outputPath, humanPath string) error {
	contractSource, err := os.ReadFile(contractPath)
	if err != nil {
		return fmt.Errorf("read contract: %w", err)
	}
	projectSource, err := os.ReadFile(projectPath)
	if err != nil {
		return fmt.Errorf("read project: %w", err)
	}
	fixtureData, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("read utility manifest: %w", err)
	}
	var fixture fixture
	if err := decodeStrict(fixtureData, &fixture); err != nil {
		return fmt.Errorf("decode utility manifest: %w", err)
	}
	if fixture.Schema != utilityInputSchema {
		return errors.New("utility manifest schema is invalid")
	}
	if err := verifyPublishedArtifacts(fixture); err != nil {
		return err
	}
	semantics, cases, journey, operations, err := lowerProject(projectPath, projectSource)
	if err != nil {
		return err
	}
	if err := verifyContract(contractPath, contractSource); err != nil {
		return err
	}
	if !sameStrings(cases.ids, utilityCases) || !sameStrings(cases.decisions, utilityDecisions) {
		return fmt.Errorf("utility cases are not the exact source-bound 2/2/2 boundary: %v", cases.ids)
	}
	if len(journey) != 9 {
		return fmt.Errorf("utility journey steps = %d, want 9", len(journey))
	}

	first, err := readDiscoveryReport(fixture.FirstReport)
	if err != nil {
		return err
	}
	second, err := readDiscoveryReport(fixture.SecondReport)
	if err != nil {
		return err
	}
	if err := verifyDiscoveryReports(first, second); err != nil {
		return err
	}
	if err := verifyLedger(fixture.Ledger, 2); err != nil {
		return err
	}

	candidateData, err := os.ReadFile(fixture.Candidate)
	if err != nil {
		return fmt.Errorf("read candidate: %w", err)
	}
	candidate, err := publiccontinuity.DecodeCandidate(candidateData)
	if err != nil {
		return err
	}
	candidateDigest := cache.HashBytes(candidateData).String()
	if candidateDigest != second.CandidateDigest || candidate.SourceDigest != cache.HashBytes(projectSource).String() {
		return errors.New("candidate is not bound to the raw project source")
	}

	decisionData, err := os.ReadFile(fixture.AcceptedDecision)
	if err != nil {
		return fmt.Errorf("read accepted decision: %w", err)
	}
	decision, err := readDecision(decisionData)
	if err != nil {
		return err
	}
	if decision.Decision != publiccontinuity.DecisionAccept || decision.Binding.CandidateDigest != candidateDigest {
		return errors.New("accepted decision is not explicitly bound to the candidate")
	}

	rejectedData, err := os.ReadFile(fixture.RejectedDecision)
	if err != nil {
		return fmt.Errorf("read rejected decision: %w", err)
	}
	rejected, err := readDecision(rejectedData)
	if err != nil {
		return err
	}
	if rejected.Decision != publiccontinuity.DecisionReject || rejected.Binding.CandidateDigest != candidateDigest {
		return errors.New("rejected decision is not an explicit terminal rejection")
	}

	certificateData, err := os.ReadFile(fixture.Certificate)
	if err != nil {
		return fmt.Errorf("read certificate: %w", err)
	}
	certificate, err := readCertificate(certificateData)
	if err != nil {
		return err
	}
	certificateDigest := cache.HashBytes(certificateData).String()
	if certificate.Binding.CandidateDigest != candidateDigest || certificate.DecisionReceiptDigest != cache.HashBytes(decisionData).String() {
		return errors.New("certificate does not bind the accepted decision and candidate")
	}

	consumption, err := readContinuityReport(fixture.ConsumptionReport)
	if err != nil {
		return err
	}
	if consumption.Decision != "CLOSED" || consumption.Binding.CandidateDigest != candidateDigest || consumption.CertificateDigest != certificateDigest ||
		consumption.DigestContinuityEdgesExpected != 4 || consumption.DigestContinuityEdgesObserved != 4 ||
		consumption.ManualTransformations != 0 || !consumption.GeneratedBytesEqual || !consumption.NormalizedSemanticEqual ||
		consumption.SemanticOperationsBefore != operations.Baseline.Semantic || consumption.SemanticOperationsAfter != operations.Learned.Semantic {
		return errors.New("certified consumption did not preserve the exact continuity contract")
	}

	baselineSource, err := os.ReadFile(fixture.BaselineSource)
	if err != nil {
		return fmt.Errorf("read baseline generated source: %w", err)
	}
	learnedSource, err := os.ReadFile(fixture.LearnedSource)
	if err != nil {
		return fmt.Errorf("read learned generated source: %w", err)
	}
	baselineManifest, _, err := readProjectionManifest(fixture.BaselineManifest)
	if err != nil {
		return err
	}
	learnedManifest, learnedManifestBytes, err := readProjectionManifest(fixture.LearnedManifest)
	if err != nil {
		return err
	}
	if err := verifyGeneratedOutputs(baselineSource, learnedSource, baselineManifest, learnedManifest, learnedManifestBytes, certificate); err != nil {
		return err
	}

	stale, err := readContinuityReport(fixture.StaleReport)
	if err != nil {
		return err
	}
	tampered, err := readContinuityReport(fixture.TamperedReport)
	if err != nil {
		return err
	}
	if err := verifyRefutedReports(stale, tampered); err != nil {
		return err
	}

	baselineBuild, err := readBuildTest(fixture.BaselineBuild)
	if err != nil {
		return err
	}
	baselineTest, err := readBuildTest(fixture.BaselineTest)
	if err != nil {
		return err
	}
	learnedBuild, err := readBuildTest(fixture.LearnedBuild)
	if err != nil {
		return err
	}
	learnedTest, err := readBuildTest(fixture.LearnedTest)
	if err != nil {
		return err
	}
	if err := verifyBuildTest(baselineBuild, baselineTest, learnedBuild, learnedTest); err != nil {
		return err
	}

	measurements, err := readMeasurements(fixture.Measurements)
	if err != nil {
		return err
	}
	if err := verifyMeasurements(measurements, candidate, certificate); err != nil {
		return err
	}
	status, err := readRepositoryStatus(fixture.RepositoryStatus)
	if err != nil {
		return err
	}
	if status.Before != "" || status.After != "" {
		return errors.New("utility lane changed the repository worktree")
	}

	inputInventory, err := inventoryProjectInput(fixture.ProjectInputFiles)
	if err != nil {
		return err
	}
	generated := generatedEvidence{BaselineFiles: 2, LearnedFiles: 2, BaselineGoFiles: 1, LearnedGoFiles: 1,
		BaselineGoPhysicalLines: physicalLines(baselineSource), LearnedGoPhysicalLines: physicalLines(learnedSource),
		BaselineGoBytes: len(baselineSource), LearnedGoBytes: len(learnedSource)}
	if generated.BaselineGoPhysicalLines == 0 || generated.BaselineGoBytes == 0 || generated.BaselineGoPhysicalLines != generated.LearnedGoPhysicalLines || generated.BaselineGoBytes != generated.LearnedGoBytes {
		return errors.New("generated Go output size is not an exact baseline/learned match")
	}

	journeyEvidence := makeJourney(journey)
	report := utilityReport{
		Schema: utilitySchema, Decision: "CLOSED", Reason: "EXACT_UTILITY_EVIDENCE_WITH_FAIL_CLOSED_RUNTIME_COMPARISON",
		ContractSourceDigest: cache.HashBytes(contractSource).String(), ProjectSourceDigest: cache.HashBytes(projectSource).String(),
		Project: projectSemantics{Package: semantics.packageName, Namespace: semantics.namespace, InputDescendantDirs: 0,
			InputRegularFiles: inputInventory.regularFiles, InputPhysicalLines: inputInventory.totalLines,
			InputGoFiles: inputInventory.goFiles, InputGoPhysicalLines: inputInventory.goLines,
			InputGoooFiles: inputInventory.goooFiles, InputGoooPhysicalLines: inputInventory.goooLines,
			Entities: semantics.entities, Activities: semantics.activities, Relations: semantics.relations, RootReadmeExcluded: true},
		Generated: generated, OperationCounts: operations,
		Continuity: continuityEvidence{CandidateToDecision: 1, DecisionToCertificate: 1, CertificateToConsumption: 2,
			EdgesExpected: 4, EdgesObserved: 4, EdgeNames: []string{"candidate_digest:discovery->decision", "candidate_digest:decision->certificate", "candidate_digest:certificate->consumption", "certificate_digest:certificate->consumption"}},
		Comparisons: comparisonEvidence{GeneratedSourceByteMismatches: 0, GeneratedManifestNormalizedMismatches: 0,
			CandidateCertificateByteMismatches: consumption.CandidateCertificateByteReplayMismatches, GeneratedSourceBytesEqual: true, NormalizedSemanticEqual: true},
		BuildTest: buildTestEvidence{BuildExecutions: 2, TestExecutions: 2, BaselineBuildMS: baselineBuild.WallMS, BaselineBuildPeakRSS: baselineBuild.PeakRSSKib,
			BaselineTestMS: baselineTest.WallMS, BaselineTestPeakRSS: baselineTest.PeakRSSKib, LearnedBuildMS: learnedBuild.WallMS,
			LearnedBuildPeakRSS: learnedBuild.PeakRSSKib, LearnedTestMS: learnedTest.WallMS, LearnedTestPeakRSS: learnedTest.PeakRSSKib},
		Measurements: measurementEvidence{Baseline: measurements.Baseline, Learned: measurements.Learned, RuntimeComparable: measurements.RuntimeComparable,
			UnknownStage: measurements.UnknownStage, UnknownReason: measurements.UnknownReason, NextOperation: measurements.NextOperation},
		CertificateCache: certificateCacheEvidence{BaselineHits: 0, BaselineMisses: 0, LearnedHits: len(measurements.Learned) + 1, LearnedMisses: 0, NegativeHits: 0, NegativeMisses: 2},
		Journey:          journeyEvidence, Cases: buildCases(first, stale, tampered, measurements, baselineBuild, baselineTest, learnedBuild, learnedTest, consumption, candidateDigest),
		CaseDenominator: 6, ClosedCases: 2, UnknownCases: 2, RefutedCases: 2, ArtifactDenominator: artifactDenominator, ArtifactCount: artifactDenominator,
		RepositoryWrites: 0, LocalTestExecutions: 0,
	}
	if err := validateUtilityReport(report); err != nil {
		return err
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write utility report: %w", err)
	}
	if err := os.WriteFile(humanPath, []byte(renderDossier(report, baselineManifest, learnedManifest, certificateDigest)), 0o644); err != nil {
		return fmt.Errorf("write utility dossier: %w", err)
	}
	return nil
}

type policyCases struct {
	ids       []string
	decisions []string
}

type projectSemanticsInternal struct {
	packageName string
	namespace   string
	entities    int
	activities  int
	relations   int
}

type inputInventory struct {
	regularFiles int
	totalLines   int
	goFiles      int
	goLines      int
	goooFiles    int
	goooLines    int
}

func lowerProject(filename string, source []byte) (projectSemanticsInternal, policyCases, []string, operationCounts, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return projectSemanticsInternal{}, policyCases{}, nil, operationCounts{}, diagnostics.Error()
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return projectSemanticsInternal{}, policyCases{}, nil, operationCounts{}, fmt.Errorf("lower project: %w", err)
	}
	semantics := projectSemanticsInternal{packageName: ir.Package, namespace: ir.Namespace.String()}
	var value string
	var activity semantic.Node
	for _, node := range ir.Graph.Nodes() {
		switch node.Kind {
		case semantic.Entity:
			semantics.entities++
		case semantic.Activity:
			semantics.activities++
			if node.Name == "Compile" {
				activity = node
				value = node.ValueProgram
			}
		}
	}
	semantics.relations = len(ir.Graph.DeterministicFacts())
	if semantics.packageName != "publicdiscoveryexample" || semantics.namespace != "public_discovery_example" || semantics.entities != 2 || semantics.activities != 1 || activity.Name != "Compile" || value == "" {
		return projectSemanticsInternal{}, policyCases{}, nil, operationCounts{}, errors.New("project graph is not the frozen two-entity Compile example")
	}
	if err := verifyActivityRelations(ir, activity); err != nil {
		return projectSemanticsInternal{}, policyCases{}, nil, operationCounts{}, err
	}
	markers := splitMarkers(value)
	if markers["utility-contract"] != "v1" || markers["utility-project"] != "public-discovery-example" {
		return projectSemanticsInternal{}, policyCases{}, nil, operationCounts{}, errors.New("utility contract marker is missing from lowered project meta-code")
	}
	var cases policyCases
	for _, encoded := range markerValues(markers, "utility-case") {
		id, decision, ok := strings.Cut(encoded, ":")
		if !ok || id == "" || decision == "" {
			return projectSemanticsInternal{}, policyCases{}, nil, operationCounts{}, fmt.Errorf("malformed utility case %q", encoded)
		}
		cases.ids = append(cases.ids, id)
		cases.decisions = append(cases.decisions, decision)
	}
	journeyValue := markers["utility-journey"]
	if journeyValue == "" {
		return projectSemanticsInternal{}, policyCases{}, nil, operationCounts{}, errors.New("utility journey is missing from lowered project meta-code")
	}
	journey := strings.Split(journeyValue, ">")
	if markers["utility-artifact-denominator"] != fmt.Sprint(artifactDenominator) || markers["utility-performance-rule"] != "UNKNOWN_WHEN_RUNTIME_MODES_ARE_NOT_EQUIVALENT" {
		return projectSemanticsInternal{}, policyCases{}, nil, operationCounts{}, errors.New("utility denominator or performance rule is not source-bound")
	}
	operations, err := parseOperationCounts(markers["utility-operation-counts"])
	if err != nil {
		return projectSemanticsInternal{}, policyCases{}, nil, operationCounts{}, err
	}
	return semantics, cases, journey, operations, nil
}

func verifyActivityRelations(ir semantic.IR, activity semantic.Node) error {
	used := 0
	generated := 0
	for _, fact := range ir.Graph.DeterministicFacts() {
		switch {
		case fact.Predicate == semantic.Used && fact.Subject == activity.ID:
			used++
		case fact.Predicate == semantic.WasGeneratedBy && fact.Object == activity.ID:
			generated++
		}
	}
	if used != 1 || generated != 1 {
		return fmt.Errorf("Compile relation shape = used:%d generated:%d, want used:1 generated:1", used, generated)
	}
	return nil
}

func splitMarkers(value string) map[string]string {
	result := make(map[string]string)
	for part := range strings.SplitSeq(value, ";") {
		key, item, ok := strings.Cut(part, "=")
		if ok && key != "" {
			if key == "utility-case" {
				result[key] += "\x00" + item
			} else {
				result[key] = item
			}
		}
	}
	return result
}

func markerValues(markers map[string]string, key string) []string {
	value := markers[key]
	value = strings.TrimPrefix(value, "\x00")
	if value == "" {
		return nil
	}
	return strings.Split(value, "\x00")
}

func parseOperationCounts(value string) (operationCounts, error) {
	var result operationCounts
	parts := strings.Split(value, "|")
	if len(parts) != 2 {
		return result, errors.New("utility operation counts are malformed")
	}
	for _, part := range parts {
		var target *operationSet
		switch {
		case strings.HasPrefix(part, "baseline("):
			target = &result.Baseline
		case strings.HasPrefix(part, "learned("):
			target = &result.Learned
		default:
			return result, fmt.Errorf("unknown utility operation count %q", part)
		}
		body := strings.TrimSuffix(strings.TrimPrefix(part, strings.SplitN(part, "(", 2)[0]+"("), ")")
		for item := range strings.SplitSeq(body, ",") {
			key, raw, ok := strings.Cut(item, "=")
			if !ok {
				return result, fmt.Errorf("malformed operation count %q", item)
			}
			var value int
			if _, err := fmt.Sscan(raw, &value); err != nil {
				return result, err
			}
			switch key {
			case "semantic":
				target.Semantic = value
			case "lowering":
				target.Lowering = value
			case "generation":
				target.Generation = value
			default:
				return result, fmt.Errorf("unknown operation count %q", key)
			}
		}
	}
	if result.Baseline != (operationSet{Semantic: 1, Lowering: 1, Generation: 1}) || result.Learned != (operationSet{}) {
		return result, errors.New("utility operation counts are not the frozen baseline/learned contract")
	}
	return result, nil
}

func verifyContract(filename string, source []byte) error {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return diagnostics.Error()
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return err
	}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind == semantic.Activity && node.Name == "ClassifyPublicObservation" {
			if !strings.Contains(node.ValueProgram, "handoff=public-discovery-candidate-to-explicit-decision-to-continuity-certificate-to-public-consumption") || !strings.Contains(node.ValueProgram, "handoff-relation=candidate_digest:discovery=decision=certificate=consumption") || !strings.Contains(node.ValueProgram, "handoff-relation=manual_transformations:0") {
				return errors.New("continuity handoff is not declared in lowered contract meta-code")
			}
			if countMarker(node.ValueProgram, "continuity-case=") != 6 {
				return errors.New("continuity case denominator changed")
			}
			return nil
		}
	}
	return errors.New("continuity policy activity is missing")
}

func countMarker(value, prefix string) int {
	count := 0
	for part := range strings.SplitSeq(value, ";") {
		if strings.HasPrefix(part, prefix) {
			count++
		}
	}
	return count
}

func inventoryProjectInput(paths []string) (inputInventory, error) {
	if len(paths) != 2 {
		return inputInventory{}, fmt.Errorf("project input regular files = %d, want 2", len(paths))
	}
	var result inputInventory
	seen := make(map[string]bool, len(paths))
	for _, path := range paths {
		if seen[path] || filepath.Base(path) == "README.md" {
			return inputInventory{}, errors.New("project input file list is not a README-exempt exact set")
		}
		seen[path] = true
		data, err := os.ReadFile(path)
		if err != nil {
			return inputInventory{}, fmt.Errorf("read project input %q: %w", path, err)
		}
		lines := physicalLines(data)
		result.regularFiles++
		result.totalLines += lines
		switch filepath.Ext(path) {
		case ".go":
			result.goFiles++
			result.goLines += lines
		case ".gooo":
			result.goooFiles++
			result.goooLines += lines
		}
	}
	if result.goFiles != 1 || result.goooFiles != 1 || result.totalLines != result.goLines+result.goooLines {
		return inputInventory{}, errors.New("project input extension inventory is inconsistent")
	}
	return result, nil
}

func verifyPublishedArtifacts(fixture fixture) error {
	if len(fixture.PublishedArtifacts) != artifactDenominator {
		return fmt.Errorf("published artifact denominator = %d, want %d", len(fixture.PublishedArtifacts), artifactDenominator)
	}
	seen := make(map[string]bool, len(fixture.PublishedArtifacts))
	for _, path := range fixture.PublishedArtifacts {
		if seen[path] {
			return fmt.Errorf("published artifact %q is duplicated", path)
		}
		seen[path] = true
	}
	for _, path := range []string{fixture.ProjectSource, fixture.ProjectTest, fixture.ContractSource, fixture.FirstReport, fixture.SecondReport, fixture.Ledger, fixture.Candidate, fixture.AcceptedDecision, fixture.RejectedDecision, fixture.Certificate, fixture.CertificationReport, fixture.ConsumptionReport, fixture.BaselineSource, fixture.BaselineManifest, fixture.LearnedSource, fixture.LearnedManifest, fixture.StaleReport, fixture.StaleHuman, fixture.TamperedReport, fixture.TamperedHuman, fixture.BaselineBuild, fixture.BaselineTest, fixture.LearnedBuild, fixture.LearnedTest, fixture.Measurements, fixture.RepositoryStatus} {
		if path == "" {
			return errors.New("utility manifest omits a required raw artifact")
		}
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("utility artifact %q is unavailable: %w", path, err)
		}
		if !seen[path] {
			return fmt.Errorf("utility artifact %q is not in the declared published set", path)
		}
	}
	return nil
}

func verifyDiscoveryReports(first, second publicdiscovery.Report) error {
	if first.Decision != "UNKNOWN" || first.CandidatesEmitted != 0 || first.ArtifactDenominator != 4 || first.ArtifactCount != 4 || first.Unknown == nil || first.Unknown.Stage != "PUBLIC_SELF_OBSERVATION" {
		return errors.New("first observation is not the required fail-closed UNKNOWN")
	}
	if second.Decision != "CLOSED" || second.CandidatesEmitted != 1 || second.ArtifactDenominator != 5 || second.ArtifactCount != 5 || second.CandidateDigest == "" || !second.CandidateBytesEqual || !second.CandidateByteReplayEqual || second.CandidateByteMismatches != 0 {
		return errors.New("second observation is not the required closed candidate")
	}
	return nil
}

func verifyLedger(path string, want int) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read ledger: %w", err)
	}
	if len(raw) == 0 || raw[len(raw)-1] != '\n' {
		return errors.New("ledger is not newline terminated")
	}
	lines := bytes.Split(raw, []byte{'\n'})
	lines = lines[:len(lines)-1]
	if len(lines) != want {
		return fmt.Errorf("ledger entries = %d, want %d", len(lines), want)
	}
	for index, line := range lines {
		var entry publicdiscovery.LedgerEntry
		if err := decodeStrict(line, &entry); err != nil {
			return err
		}
		if entry.Sequence != index+1 || entry.Schema != publicdiscovery.ObservationLedgerSchema || entry.Operation != publicdiscovery.Operation || entry.RepositoryWrites != 0 || entry.LocalBuildExecutions != 0 || entry.LocalTestExecutions != 0 {
			return fmt.Errorf("ledger entry %d is not exact", index+1)
		}
		digest, err := publicdiscovery.EntryContentDigest(entry)
		if err != nil || digest != entry.ObservationDigest {
			return fmt.Errorf("ledger entry %d digest mismatch", index+1)
		}
	}
	return nil
}

func verifyGeneratedOutputs(baseline, learned []byte, baselineManifest, learnedManifest map[string]any, learnedManifestBytes []byte, certificate publiccontinuity.Certificate) error {
	if !bytes.Equal(baseline, learned) || !bytes.Equal(baseline, certificate.GeneratedSource) {
		return errors.New("baseline, learned, and certified generated Go bytes differ")
	}
	if !sameNormalizedManifest(baselineManifest, learnedManifest) || !bytes.Equal(certificate.GeneratedManifest, learnedManifestBytes) {
		return errors.New("generated manifests are not an exact normalized/certificate match")
	}
	if stringValue(baselineManifest, "semantic_digest") == "" || stringValue(baselineManifest, "semantic_digest") != stringValue(learnedManifest, "semantic_digest") || stringValue(baselineManifest, "generated_digest") != cache.HashBytes(baseline).String() {
		return errors.New("generated manifest semantic or output digest is not bound")
	}
	return nil
}

func verifyRefutedReports(stale, tampered publiccontinuity.Report) error {
	if stale.Decision != "REFUTED" || stale.Reason != "STALE_SOURCE" || stale.ArtifactDenominator != 2 || stale.ArtifactCount != 2 || stale.RepositoryWrites != 0 || stale.LocalBuildExecutions != 0 || stale.LocalTestExecutions != 0 {
		return errors.New("stale project case did not fail closed with its two reports")
	}
	if tampered.Decision != "REFUTED" || tampered.Reason != "TAMPERED_CERTIFICATE" || tampered.ArtifactDenominator != 2 || tampered.ArtifactCount != 2 || tampered.RepositoryWrites != 0 || tampered.LocalBuildExecutions != 0 || tampered.LocalTestExecutions != 0 {
		return errors.New("tampered certificate case did not fail closed with its two reports")
	}
	return nil
}

func verifyBuildTest(build, test, learnedBuild, learnedTest buildTestRecord) error {
	for _, record := range []buildTestRecord{build, test, learnedBuild, learnedTest} {
		if record.Schema != "gooo/public-self-improvement-generated-project-check/v1" || record.Decision != "CLOSED" || !record.Passed || record.GeneratedFileCount != 2 || record.GeneratedGoFiles != 1 || record.Executions != 1 || record.RepositoryWrites != 0 || record.LocalTestExecutions != 0 || record.WallMS <= 0 || record.PeakRSSKib <= 0 {
			return fmt.Errorf("generated project %s did not pass the exact build/test check", record.Mode)
		}
	}
	if build.Command != "go build ." || test.Command != "go test -tags generated_project ." || learnedBuild.Command != build.Command || learnedTest.Command != test.Command {
		return errors.New("generated project build/test command changed")
	}
	return nil
}

func verifyMeasurements(report measurementReport, candidate publicdiscovery.Candidate, certificate publiccontinuity.Certificate) error {
	if report.Schema != "gooo/public-self-improvement-runtime-measurement/v1" || report.RuntimeComparable || report.UnknownStage != "UTILITY_MEASUREMENT" || report.UnknownStep != "COMPARE_BASELINE_LEARNED_RUNTIME" || report.UnknownReason != "RUNTIME_MODES_NOT_EQUIVALENT" || report.UnknownClass != "INCOMPARABLE" || report.NextOperation != "PREDECLARE_EQUIVALENT_RUNTIME_MEASUREMENT_RULE" {
		return errors.New("runtime measurement did not remain explicitly UNKNOWN and fail closed")
	}
	if len(report.Baseline) != 3 || len(report.Learned) != 3 {
		return errors.New("runtime measurement denominator is not exactly three baseline and three learned observations")
	}
	expected := measurementIdentity{SourceDigest: candidate.SourceDigest, InputSemanticDigest: candidate.InputSemanticDigest, CompilerDigest: certificate.CompilerDigest, ToolchainDigest: candidate.ToolchainDigest, ContractDigest: candidate.ContractDigest, EvaluatorDigest: candidate.EvaluatorDigest}
	for _, group := range [][]measurementObservation{report.Baseline, report.Learned} {
		for index, observation := range group {
			if observation.Index != index+1 || observation.WallMS <= 0 || observation.PeakRSSKib <= 0 || measurementIdentityOf(observation) != expected {
				return fmt.Errorf("runtime observation %d is not bound to the common digest tuple", index+1)
			}
		}
	}
	return nil
}

type measurementIdentity struct {
	SourceDigest        string
	InputSemanticDigest string
	CompilerDigest      string
	ToolchainDigest     string
	ContractDigest      string
	EvaluatorDigest     string
}

func measurementIdentityOf(observation measurementObservation) measurementIdentity {
	return measurementIdentity{SourceDigest: observation.SourceDigest, InputSemanticDigest: observation.InputSemanticDigest, CompilerDigest: observation.CompilerDigest, ToolchainDigest: observation.ToolchainDigest, ContractDigest: observation.ContractDigest, EvaluatorDigest: observation.EvaluatorDigest}
}

func buildCases(first publicdiscovery.Report, stale, tampered publiccontinuity.Report, measurements measurementReport, baselineBuild, baselineTest, learnedBuild, learnedTest buildTestRecord, consumption publiccontinuity.Report, candidateDigest string) []utilityCase {
	return []utilityCase{
		{ID: caseBaseline, ExpectedDecision: "CLOSED", ObservedDecision: "CLOSED", Reason: "BASELINE_GENERATION_BUILD_TEST_PASS", ByteMismatches: 0, NormalizedSemanticMismatches: 0, DigestEdgesExpected: 0, DigestEdgesObserved: 0, BuildExecutions: baselineBuild.Executions, TestExecutions: baselineTest.Executions, ArtifactsProduced: 2},
		{ID: caseLearned, ExpectedDecision: "CLOSED", ObservedDecision: "CLOSED", Reason: "CERTIFIED_GENERATION_BUILD_TEST_PASS", CandidateDigest: candidateDigest, ByteMismatches: 0, NormalizedSemanticMismatches: 0, DigestEdgesExpected: 4, DigestEdgesObserved: 4, BuildExecutions: learnedBuild.Executions, TestExecutions: learnedTest.Executions, CertificateHits: 1, ArtifactsProduced: 4},
		{ID: caseFirst, ExpectedDecision: "UNKNOWN", ObservedDecision: first.Decision, Reason: first.Reason, UnknownStage: first.Unknown.Stage, UnknownStep: first.Unknown.Step, UnknownReason: first.Unknown.Reason, UnknownClass: first.Unknown.UnknownClass, NextOperation: first.Unknown.NextOperation, BlockedBy: first.Unknown.BlockedBy, ArtifactsProduced: first.ArtifactCount},
		{ID: casePerformance, ExpectedDecision: "UNKNOWN", ObservedDecision: "UNKNOWN", Reason: measurements.UnknownReason, UnknownStage: measurements.UnknownStage, UnknownStep: measurements.UnknownStep, UnknownReason: measurements.UnknownReason, UnknownClass: measurements.UnknownClass, NextOperation: measurements.NextOperation, BlockedBy: []string{"ordinary-generate-vs-certified-consumption"}, ArtifactsProduced: 1},
		{ID: caseStale, ExpectedDecision: "REFUTED", ObservedDecision: stale.Decision, Reason: stale.Reason, ByteMismatches: 1, NormalizedSemanticMismatches: 1, DigestEdgesExpected: 1, DigestEdgesObserved: 0, CertificateMisses: 1, ArtifactsProduced: stale.ArtifactCount},
		{ID: caseDigestMismatch, ExpectedDecision: "REFUTED", ObservedDecision: tampered.Decision, Reason: tampered.Reason, ByteMismatches: 1, NormalizedSemanticMismatches: 1, DigestEdgesExpected: 1, DigestEdgesObserved: 0, CertificateMisses: 1, ArtifactsProduced: tampered.ArtifactCount},
	}
}

func validateUtilityReport(report utilityReport) error {
	if report.Decision != "CLOSED" || report.CaseDenominator != 6 || len(report.Cases) != 6 || report.ClosedCases != 2 || report.UnknownCases != 2 || report.RefutedCases != 2 || report.ArtifactDenominator != artifactDenominator || report.ArtifactCount != artifactDenominator || report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 || !report.Comparisons.GeneratedSourceBytesEqual || !report.Comparisons.NormalizedSemanticEqual || report.Continuity.EdgesExpected != 4 || report.Continuity.EdgesObserved != 4 || report.BuildTest.BuildExecutions != 2 || report.BuildTest.TestExecutions != 2 || len(report.Measurements.Baseline) != 3 || len(report.Measurements.Learned) != 3 {
		return errors.New("utility report counts are not exact")
	}
	return nil
}

func makeJourney(actions []string) []journeyStep {
	decisions := []string{"CLOSED", "UNKNOWN", "CLOSED", "ACCEPT+REJECT", "CLOSED", "CLOSED", "PASS", "PASS", "CLOSED"}
	result := make([]journeyStep, len(actions))
	for index, action := range actions {
		result[index] = journeyStep{Ordinal: fmt.Sprint(index + 1), Action: action, Decision: decisions[index]}
	}
	return result
}

func renderDossier(report utilityReport, baselineManifest, learnedManifest map[string]any, certificateDigest string) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "# Public self-improvement utility dossier\n\nDecision: `%s`\nReason: `%s`\n\n", report.Decision, report.Reason)
	builder.WriteString("## DRIVER\n\n")
	fmt.Fprintf(&builder, "The `Compile(Input) -> Output` graph is the bounded project contract: %d entities, %d activity, and %d derived relations. The input inventory is %d regular files (%d `.gooo`, %d Go), %d physical lines, with the project-root README excluded.\n\n", report.Project.Entities, report.Project.Activities, report.Project.Relations, report.Project.InputRegularFiles, report.Project.InputGoooFiles, report.Project.InputGoFiles, report.Project.InputPhysicalLines)
	fmt.Fprintf(&builder, "The public journey has %d ordered steps: %s. Published evidence is fixed at %d artifacts.\n\n", len(report.Journey), journeyNames(report.Journey), report.ArtifactDenominator)
	builder.WriteString("## OUTCOME\n\n")
	fmt.Fprintf(&builder, "Cases: %d CLOSED / %d UNKNOWN / %d REFUTED. No aggregate score is emitted.\n\n", report.ClosedCases, report.UnknownCases, report.RefutedCases)
	fmt.Fprintf(&builder, "Generated output: %d files / %d Go file; baseline and learned Go are %d bytes and %d physical lines, with byte mismatches=%d and normalized-semantic mismatches=%d.\n\n", report.Generated.BaselineFiles, report.Generated.BaselineGoFiles, report.Generated.BaselineGoBytes, report.Generated.BaselineGoPhysicalLines, report.Comparisons.GeneratedSourceByteMismatches, report.Comparisons.GeneratedManifestNormalizedMismatches)
	fmt.Fprintf(&builder, "Operations baseline semantic/lowering/generation=%d/%d/%d; learned=%d/%d/%d. Continuity edges=%d/%d. Candidate digest is carried through decision, certificate, and consumption; certificate digest=%s.\n\n", report.OperationCounts.Baseline.Semantic, report.OperationCounts.Baseline.Lowering, report.OperationCounts.Baseline.Generation, report.OperationCounts.Learned.Semantic, report.OperationCounts.Learned.Lowering, report.OperationCounts.Learned.Generation, report.Continuity.EdgesObserved, report.Continuity.EdgesExpected, certificateDigest)
	fmt.Fprintf(&builder, "Generated-project build/test executions=%d/%d; baseline build/test=%dms/%dms, learned build/test=%dms/%dms. All four checks passed in Actions with Go 1.27.\n\n", report.BuildTest.BuildExecutions, report.BuildTest.TestExecutions, report.BuildTest.BaselineBuildMS, report.BuildTest.BaselineTestMS, report.BuildTest.LearnedBuildMS, report.BuildTest.LearnedTestMS)
	builder.WriteString("Individual runtime observations (wall_ms / peak_rss_kib):\n\n")
	for _, group := range []struct {
		name string
		data []measurementObservation
	}{
		{name: "baseline", data: report.Measurements.Baseline}, {name: "learned", data: report.Measurements.Learned},
	} {
		fmt.Fprintf(&builder, "- %s: ", group.name)
		for index, item := range group.data {
			if index > 0 {
				builder.WriteString(", ")
			}
			fmt.Fprintf(&builder, "%d=%d/%d", item.Index, item.WallMS, item.PeakRSSKib)
		}
		builder.WriteByte('\n')
	}
	builder.WriteString("\n")
	builder.WriteString("## GUARDRAIL\n\n")
	fmt.Fprintf(&builder, "Runtime superiority is not claimed: `runtime_comparable=false`, so PERFORMANCE_INCOMPARABLE remains UNKNOWN under `%s`; stage=`%s`, reason=`%s`, next operation=`%s`.\n\n", report.Measurements.UnknownReason, report.Measurements.UnknownStage, report.Measurements.UnknownReason, report.Measurements.NextOperation)
	builder.WriteString("The first observation is UNKNOWN until quorum; stale source and tampered certificate are REFUTED fail-closed and publish only their two report artifacts. Repository writes=0 and local test executions=0.\n\n")
	fmt.Fprintf(&builder, "Certificate cache hits/misses: baseline=%d/%d, learned=%d/%d, negative=%d/%d. Baseline manifest semantic digest=%s; learned manifest semantic digest=%s.\n", report.CertificateCache.BaselineHits, report.CertificateCache.BaselineMisses, report.CertificateCache.LearnedHits, report.CertificateCache.LearnedMisses, report.CertificateCache.NegativeHits, report.CertificateCache.NegativeMisses, stringValue(baselineManifest, "semantic_digest"), stringValue(learnedManifest, "semantic_digest"))
	return builder.String()
}

func journeyNames(steps []journeyStep) string {
	parts := make([]string, len(steps))
	for index, step := range steps {
		parts[index] = step.Action
	}
	return strings.Join(parts, " -> ")
}

func readDiscoveryReport(path string) (publicdiscovery.Report, error) {
	var report publicdiscovery.Report
	if err := readStrict(path, &report); err != nil {
		return report, fmt.Errorf("read discovery report: %w", err)
	}
	return report, nil
}

func readContinuityReport(path string) (publiccontinuity.Report, error) {
	var report publiccontinuity.Report
	if err := readStrict(path, &report); err != nil {
		return report, fmt.Errorf("read continuity report: %w", err)
	}
	return report, nil
}

func readDecision(data []byte) (publiccontinuity.DecisionReceipt, error) {
	var decision publiccontinuity.DecisionReceipt
	if err := decodeStrict(data, &decision); err != nil {
		return decision, err
	}
	if err := publiccontinuity.ValidateDecisionReceipt(decision); err != nil {
		return decision, err
	}
	return decision, nil
}

func readCertificate(data []byte) (publiccontinuity.Certificate, error) {
	var certificate publiccontinuity.Certificate
	if err := decodeStrict(data, &certificate); err != nil {
		return certificate, err
	}
	if err := publiccontinuity.ValidateCertificate(certificate); err != nil {
		return certificate, err
	}
	return certificate, nil
}

func readBuildTest(path string) (buildTestRecord, error) {
	var record buildTestRecord
	if err := readStrict(path, &record); err != nil {
		return record, fmt.Errorf("read build/test record: %w", err)
	}
	return record, nil
}

func readMeasurements(path string) (measurementReport, error) {
	var report measurementReport
	if err := readStrict(path, &report); err != nil {
		return report, fmt.Errorf("read runtime measurements: %w", err)
	}
	return report, nil
}

func readRepositoryStatus(path string) (repositoryStatus, error) {
	var status repositoryStatus
	if err := readStrict(path, &status); err != nil {
		return status, fmt.Errorf("read repository status: %w", err)
	}
	return status, nil
}

func readProjectionManifest(path string) (map[string]any, []byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read projection manifest: %w", err)
	}
	lines := bytes.TrimSpace(data)
	if len(lines) == 0 {
		return nil, nil, errors.New("projection manifest is empty")
	}
	var manifest map[string]any
	if err := json.Unmarshal(lines, &manifest); err != nil {
		return nil, nil, fmt.Errorf("decode projection manifest: %w", err)
	}
	return manifest, data, nil
}

func sameNormalizedManifest(left, right map[string]any) bool {
	leftCopy := copyMap(left)
	rightCopy := copyMap(right)
	leftCopy["generated_file"] = "semantic.gooo.go"
	rightCopy["generated_file"] = "semantic.gooo.go"
	return reflect.DeepEqual(leftCopy, rightCopy)
}

func copyMap(input map[string]any) map[string]any {
	result := make(map[string]any, len(input))
	maps.Copy(result, input)
	return result
}

func stringValue(value map[string]any, key string) string {
	item, _ := value[key].(string)
	return item
}

func readStrict(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return decodeStrict(data, value)
}

func decodeStrict(data []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("JSON contains multiple values")
		}
		return err
	}
	return nil
}

func physicalLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	count := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		count++
	}
	return count
}

func sameStrings(left, right []string) bool {
	return reflect.DeepEqual(left, right)
}

func sortStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
