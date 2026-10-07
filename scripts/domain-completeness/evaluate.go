package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"unicode"

	"github.com/kimjooyoon/meta-ontology-go/internal/domaincompleteness"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageutility"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	maximumEvidenceFiles = 128
	maximumEvidenceBytes = 25_000_000
)

type inventory struct {
	RepositoryWrites    int  `json:"repository_writes"`
	MutationAuthority   bool `json:"mutation_authority"`
	GenerationAuthority bool `json:"generation_authority"`
	RepairAuthority     bool `json:"repair_authority"`
	MergeAuthority      bool `json:"merge_authority"`
}

type progressReceipt struct {
	Contract struct {
		Digest string `json:"digest"`
	} `json:"contract"`
}

type graphSnapshot struct {
	GraphHash string `json:"graph_hash"`
	Nodes     []struct {
		ID   string `json:"id"`
		Kind string `json:"kind"`
		Name string `json:"name"`
	} `json:"nodes"`
	Relations []map[string]json.RawMessage `json:"relations"`
}

type loadedInputs struct {
	contractRaw    []byte
	contract       languageutility.Contract
	reportRaw      []byte
	report         languageutility.Report
	observationRaw []byte
	observation    languageutility.Observation
	programRaw     []byte
	progressRaw    []byte
	progress       progressReceipt
	inventoryRaw   []byte
	inventory      inventory
	graphRaw       []byte
	graph          graphSnapshot
	finalGraphRaw  []byte
	finalGraph     graphSnapshot
	discovery      capabilityDiscoveryEvidence
	evidenceRefs   []EvidenceRef
	evidenceState  string
	evidenceReason string
	inputBytes     int64
	inputFiles     int
	inputPathSeen  map[string]bool
	issues         []string
}

func evaluate(
	profilePath, contractPath, evidenceDir, baselinePath string, autoBaseline bool, baselineSource baselineArtifact,
	comparisonStatus, comparisonReason, subject string, runID int64, attempt int,
) (Report, []byte, []byte, baselineArtifact, error) {
	profileRaw, err := os.ReadFile(profilePath)
	if err != nil {
		return Report{}, nil, nil, baselineArtifact{}, fmt.Errorf("read profile: %w", err)
	}
	if err := domaincompleteness.VerifyGeneratedProjection(profilePath, profileRaw); err != nil {
		return Report{}, nil, nil, baselineArtifact{}, err
	}
	profileModel, err := compileProfile(profilePath, profileRaw)
	if err != nil {
		return Report{}, nil, nil, baselineArtifact{}, err
	}
	generated := []byte(renderProfile(profileModel))
	generatedSemanticHash, err := semanticHash(profilePath+".generated", generated)
	if err != nil {
		return Report{}, nil, nil, baselineArtifact{}, err
	}
	semanticsEqual := generatedSemanticHash == profileModel.SemanticHash
	profileDigest := digestBytes(profileRaw)
	inputs := loadInputs(contractPath, evidenceDir)
	var baseline Report
	var baselineRaw []byte
	selectedArtifact := baselineSource
	if baselinePath != "" {
		baselineRaw, readErr := os.ReadFile(baselinePath)
		if readErr != nil {
			return Report{}, nil, nil, baselineArtifact{}, fmt.Errorf("read baseline receipt: %w", readErr)
		}
		inputs.inputFiles++
		if selectedArtifact.SizeInBytes > 0 {
			inputs.inputBytes += selectedArtifact.SizeInBytes
		} else {
			inputs.inputBytes += int64(len(baselineRaw))
		}
		if err := json.Unmarshal(baselineRaw, &baseline); err != nil {
			return Report{}, nil, nil, baselineArtifact{}, fmt.Errorf("decode baseline receipt: %w", err)
		}
	}
	inputs.inputFiles++
	inputs.inputBytes += int64(len(profileRaw))
	repositoryWrites := max(inputs.observation.RepositoryWrites, inputs.inventory.RepositoryWrites)
	report := Report{
		Schema: ReceiptSchema, ProfileID: ProfileID, SubjectSHA: subject,
		Decision: "UNKNOWN", Reason: "DOMAIN_EVIDENCE_UNAVAILABLE",
		NextOperation: "COLLECT_DOMAIN_EVIDENCE",
		Contract: ContractRef{
			Path: profilePath, Digest: profileDigest,
			SemanticHash: profileModel.SemanticHash, ReceiptSchema: ReceiptSchema,
		},
		Generated: GeneratedRef{
			Path: "domain-completeness/program.gooo", Digest: digestBytes(generated),
			SemanticHash: generatedSemanticHash, SemanticsEqual: semanticsEqual,
			ReceiptSchema: ReceiptSchema,
		},
		Snapshot: Snapshot{
			Repository: os.Getenv("GITHUB_REPOSITORY"), SubjectSHA: subject,
			WorkflowRun: runID, RunAttempt: attempt, Toolchain: toolchainIdentity(),
			Evaluator: "scripts/domain-completeness@v1",
		},
		RequiredScope: []string{
			"declaration coverage", "generation coverage", "reverse observation coverage",
			"use-case coverage", "boundary coverage", "provenance integrity",
		},
		ExcludedScope: []string{
			"general-purpose language completeness", "external adoption",
			"production readiness", "business correctness beyond observed use cases",
		},
		Investment: Investment{
			Budget: SystemBudget{
				MaximumEvidenceFiles:    maximumEvidenceFiles,
				MaximumEvidenceBytes:    maximumEvidenceBytes,
				MaximumRepositoryWrites: 0,
				MaximumHumanActions:     0,
			},
			Observed: SystemCost{
				EvidenceFiles: inputs.inputFiles, EvidenceBytes: inputs.inputBytes,
				RepositoryWrites: repositoryWrites, HumanActions: 0,
			},
			ComparisonStatus: "UNKNOWN_NO_PROFILE_BOUND_PRIOR_RECEIPT",
		},
	}
	if !semanticsEqual {
		inputs.issues = append(inputs.issues, "generated profile semantic hash differs from source profile")
	}
	if inputs.inputFiles > maximumEvidenceFiles || inputs.inputBytes > maximumEvidenceBytes {
		inputs.issues = append(inputs.issues, "evidence input exceeds the system budget")
	}
	report.Dimensions = measureDimensions(profileModel, inputs, subject, semanticsEqual, runID, attempt)
	report.Summary = summarize(report.Dimensions)
	report.Decision, report.Reason, report.NextOperation, report.FirstUnresolved =
		decide(report.Dimensions, inputs.issues, inputs.evidenceState)
	if baselinePath == "" && autoBaseline {
		candidate, candidateRaw, artifact, discoveryErr := discoverBaseline(context.Background(), report)
		if discoveryErr != nil {
			if discovery, ok := discoveryErr.(baselineDiscoveryError); ok {
				report.Investment.ComparisonStatus = discovery.status
				report.Investment.ComparisonReason = discovery.reason
			} else {
				report.Investment.ComparisonStatus = "UNKNOWN_BASELINE_UNAVAILABLE"
				report.Investment.ComparisonReason = "BASELINE_DISCOVERY_FAILED"
			}
		} else {
			baseline, baselineRaw, selectedArtifact = candidate, candidateRaw, artifact
			inputs.inputFiles++
			inputs.inputBytes += selectedArtifact.SizeInBytes
			report.Investment.Observed.EvidenceFiles = inputs.inputFiles
			report.Investment.Observed.EvidenceBytes = inputs.inputBytes
		}
	}
	comparison := compareReports(report, baseline, baselinePath != "" || selectedArtifact.ID != 0)
	if selectedArtifact.ID != 0 {
		comparison.BaselineArtifactID = selectedArtifact.ID
		comparison.BaselineArtifactDigest = selectedArtifact.Digest
		comparison.BaselineArtifactName = selectedArtifact.Name
		comparison.BaselineArtifactBytes = selectedArtifact.SizeInBytes
	}
	if selectedArtifact.ID == 0 && comparisonStatus != "" {
		comparison.Status = comparisonStatus
		report.Investment.ComparisonReason = comparisonReason
	} else if selectedArtifact.ID == 0 && autoBaseline && baselinePath == "" {
		comparison.Status = report.Investment.ComparisonStatus
	}
	report.Investment.ComparisonStatus = comparison.Status
	if comparison.Status == "COMPARABLE" || comparison.Status == "PARTIAL" {
		report.Investment.Comparison = &comparison
	}
	report.Digest, err = reportDigest(report)
	return report, generated, baselineRaw, selectedArtifact, err
}

func compareReports(current, baseline Report, supplied bool) Comparison {
	result := Comparison{Status: "UNKNOWN_NO_PROFILE_BOUND_PRIOR_RECEIPT", Dimensions: []DimensionDelta{}}
	if !supplied {
		return result
	}
	result.Status = "UNKNOWN_INCOMPATIBLE_BASELINE"
	baselineDigest, err := reportDigest(baseline)
	if err != nil || baseline.Digest == "" || baselineDigest != baseline.Digest ||
		baseline.Schema != current.Schema || baseline.ProfileID != current.ProfileID ||
		baseline.Contract.SemanticHash != current.Contract.SemanticHash ||
		baseline.Generated.SemanticHash != current.Generated.SemanticHash ||
		baseline.Snapshot.Repository != current.Snapshot.Repository ||
		baseline.SubjectSHA == "" || baseline.SubjectSHA == current.SubjectSHA ||
		baseline.Snapshot.Repository == "" || !baseline.Generated.SemanticsEqual || !current.Generated.SemanticsEqual ||
		baseline.Snapshot.SubjectSHA != baseline.SubjectSHA ||
		len(baseline.Dimensions) != len(current.Dimensions) {
		return result
	}
	byID := make(map[string]Dimension, len(baseline.Dimensions))
	for _, dimension := range baseline.Dimensions {
		if _, duplicate := byID[dimension.ID]; duplicate || dimension.ID == "" {
			return result
		}
		byID[dimension.ID] = dimension
	}
	for _, dimension := range current.Dimensions {
		prior, exists := byID[dimension.ID]
		if !exists || prior.MetricID != dimension.MetricID || prior.Unit != dimension.Unit ||
			prior.Denominator != dimension.Denominator || prior.Denominator <= 0 ||
			prior.Numerator < 0 || prior.Numerator > prior.Denominator || prior.UnknownUnits < 0 || prior.RefutedUnits < 0 ||
			dimension.Numerator < 0 || dimension.Numerator > dimension.Denominator {
			return Comparison{Status: "UNKNOWN_INCOMPATIBLE_BASELINE", Dimensions: []DimensionDelta{}}
		}
		axisStatus := "COMPARABLE"
		if prior.UnknownUnits > 0 || dimension.UnknownUnits > 0 || prior.Status == "UNKNOWN" || dimension.Status == "UNKNOWN" ||
			prior.Status == "FAIL_CLOSED" || dimension.Status == "FAIL_CLOSED" {
			axisStatus = "UNKNOWN_UNRESOLVED_EVIDENCE"
		}
		result.Dimensions = append(result.Dimensions, DimensionDelta{
			ID: dimension.ID, Status: axisStatus, NumeratorDelta: dimension.Numerator - prior.Numerator,
			Denominator: dimension.Denominator, BaselineStatus: prior.Status, CurrentStatus: dimension.Status,
		})
	}
	result.Status = "COMPARABLE"
	for _, delta := range result.Dimensions {
		if delta.Status != "COMPARABLE" {
			result.Status = "PARTIAL"
		}
	}
	result.BaselineSubject = baseline.SubjectSHA
	result.BaselineDigest = baseline.Digest
	result.BaselineRunID = baseline.Snapshot.WorkflowRun
	result.BaselineAttempt = baseline.Snapshot.RunAttempt
	return result
}

func loadInputs(contractPath, evidenceDir string) loadedInputs {
	result := loadedInputs{evidenceState: "UNKNOWN", evidenceReason: "EVIDENCE_ARTIFACTS_NOT_READ", inputPathSeen: make(map[string]bool)}
	read := func(path string) []byte {
		value, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		identity, identityErr := filepath.Abs(path)
		if identityErr != nil {
			identity = filepath.Clean(path)
		}
		if !result.inputPathSeen[identity] {
			result.inputPathSeen[identity] = true
			result.inputFiles++
			result.inputBytes += int64(len(value))
		}
		return value
	}
	result.contractRaw = read(contractPath)
	if len(result.contractRaw) > 0 {
		value, err := languageutility.DecodeContract(result.contractRaw)
		if err != nil {
			result.issues = append(result.issues, "invalid language utility contract")
		} else {
			result.contract = value
		}
	}
	result.reportRaw = read(filepath.Join(evidenceDir, "report.json"))
	if len(result.reportRaw) > 0 {
		if err := json.Unmarshal(result.reportRaw, &result.report); err != nil {
			result.issues = append(result.issues, "invalid language utility report")
		}
	}
	result.observationRaw = read(filepath.Join(evidenceDir, "observation.json"))
	if len(result.observationRaw) > 0 {
		value, err := languageutility.DecodeObservation(result.observationRaw)
		if err != nil {
			result.issues = append(result.issues, "invalid language utility observation")
		} else {
			result.observation = value
		}
	}
	result.programRaw = read(filepath.Join(evidenceDir, "program.gooo"))
	result.progressRaw = read(filepath.Join(evidenceDir, "progress-receipt.json"))
	if len(result.progressRaw) > 0 {
		if err := json.Unmarshal(result.progressRaw, &result.progress); err != nil {
			result.issues = append(result.issues, "invalid language utility progress receipt")
		}
	}
	result.inventoryRaw = read(filepath.Join(evidenceDir, "inventory.json"))
	if len(result.inventoryRaw) > 0 {
		if err := json.Unmarshal(result.inventoryRaw, &result.inventory); err != nil {
			result.issues = append(result.issues, "invalid repository inventory")
		}
	}
	result.graphRaw = read(filepath.Join(evidenceDir, "gooo-graph.json"))
	if len(result.graphRaw) > 0 {
		if err := json.Unmarshal(result.graphRaw, &result.graph); err != nil {
			result.issues = append(result.issues, "invalid source graph")
		}
	}
	result.finalGraphRaw = read(filepath.Join(evidenceDir, "final-gooo-graph.json"))
	if len(result.finalGraphRaw) > 0 {
		if err := json.Unmarshal(result.finalGraphRaw, &result.finalGraph); err != nil {
			result.issues = append(result.issues, "invalid generated graph")
		}
	}
	discoveryReportRaw := read(filepath.Join(evidenceDir, capabilityDiscoveryReportPath))
	discoveryReplayRaw := read(filepath.Join(evidenceDir, capabilityDiscoveryReplayPath))
	discoverySourceRaw := read(filepath.Join(evidenceDir, capabilityDiscoverySourceCopy))
	discoveryContractRaw := read(filepath.Join(evidenceDir, capabilityDiscoveryContractCopy))
	result.discovery = validateCapabilityDiscoveryEvidence(
		discoveryReportRaw, discoveryReplayRaw, discoverySourceRaw, discoveryContractRaw,
	)
	if result.discovery.State == "PASS" {
		state, reason := validateCapabilityDiscoveryCells(result.observation, discoveryReportRaw, discoveryReplayRaw)
		if state != "PASS" {
			result.discovery.State, result.discovery.Reason = state, reason
		}
	}
	result.discovery.Refs = discoveryEvidenceRefs(discoveryReportRaw, discoveryReplayRaw, discoverySourceRaw, discoveryContractRaw)
	if result.discovery.State == "FAIL_CLOSED" {
		result.issues = append(result.issues, result.discovery.Reason)
	}
	result.evidenceRefs, result.evidenceState, result.evidenceReason = validateEvidenceReferences(evidenceDir, result.report)
	for _, ref := range result.evidenceRefs {
		read(filepath.Join(evidenceDir, ref.Path))
	}
	return result
}

func validateEvidenceReferences(evidenceDir string, report languageutility.Report) ([]EvidenceRef, string, string) {
	paths := make(map[string]string)
	state := "PASS"
	reason := "ALL_CELL_EVIDENCE_EXACT"
	for _, cell := range report.Cells {
		if cell.State == languageutility.StateClosed && (cell.EvidencePath == "" || cell.EvidenceDigest == "") {
			state = "UNKNOWN"
			reason = "CLOSED_CELL_EVIDENCE_REFERENCE_MISSING"
		}
		if cell.EvidencePath == "" {
			continue
		}
		path := filepath.Clean(cell.EvidencePath)
		if filepath.IsAbs(path) || path == "." || path == ".." || strings.HasPrefix(path, ".."+string(filepath.Separator)) {
			return nil, "FAIL_CLOSED", "EVIDENCE_PATH_ESCAPES_BUNDLE"
		}
		if previous, exists := paths[path]; exists && previous != cell.EvidenceDigest {
			return nil, "FAIL_CLOSED", "EVIDENCE_PATH_HAS_CONFLICTING_DIGESTS"
		}
		paths[path] = cell.EvidenceDigest
	}
	if len(paths) == 0 {
		return nil, "UNKNOWN", "NO_CELL_EVIDENCE_REFERENCES"
	}
	result := make([]EvidenceRef, 0, len(paths))
	for path, expected := range paths {
		if expected == "" {
			state = "UNKNOWN"
			reason = "CELL_EVIDENCE_DIGEST_MISSING"
			result = append(result, EvidenceRef{Role: "language-utility-cell", Path: path})
			continue
		}
		value, err := os.ReadFile(filepath.Join(evidenceDir, path))
		if err != nil {
			state = "UNKNOWN"
			reason = "CELL_EVIDENCE_FILE_MISSING"
			result = append(result, EvidenceRef{Role: "language-utility-cell", Path: path, Digest: expected})
			continue
		}
		actual := digestBytes(value)
		if actual != expected {
			return nil, "FAIL_CLOSED", "CELL_EVIDENCE_DIGEST_MISMATCH"
		}
		result = append(result, EvidenceRef{Role: "language-utility-cell", Path: path, Digest: actual})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result, state, reason
}

func measureDimensions(profile ProfileModel, inputs loadedInputs, subject string, profileSemanticsEqual bool, runID int64, attempt int) []Dimension {
	result := make([]Dimension, 0, len(profile.Dimensions))
	for _, spec := range profile.Dimensions {
		dimension := Dimension{
			ID: spec.ID, MetricID: spec.MetricID, Unit: spec.Unit,
			Status: "UNKNOWN", Evidence: []EvidenceRef{},
		}
		runtime, supported := dimensionRuntimes[spec.MetricID]
		if !supported || runtime.Evaluate == nil {
			dimension = failed(dimension, "PROFILE_METRIC_HANDLER_UNAVAILABLE", "REGISTER_PROFILE_METRIC_HANDLER")
		} else {
			dimension = runtime.Evaluate(spec, profile, inputs, subject, profileSemanticsEqual, runID, attempt)
		}
		result = append(result, dimension)
	}
	return result
}

type dimensionRuntime struct {
	Unit     string
	Evaluate func(dimensionSpec, ProfileModel, loadedInputs, string, bool, int64, int) Dimension
}

var dimensionRuntimes = map[string]dimensionRuntime{
	"gooo.metric.domain-completeness.declaration-coverage.v1": {
		Unit: "declarations",
		Evaluate: func(spec dimensionSpec, _ ProfileModel, inputs loadedInputs, _ string, _ bool, _ int64, _ int) Dimension {
			return measureDeclarationCoverage(spec, inputs.contract, inputs.programRaw, inputs)
		},
	},
	"gooo.metric.domain-completeness.generation-coverage.v1": {
		Unit: "use_cases",
		Evaluate: func(spec dimensionSpec, _ ProfileModel, inputs loadedInputs, _ string, _ bool, _ int64, _ int) Dimension {
			return measureGenerationCoverage(spec, inputs.contract, inputs.report, inputs)
		},
	},
	"gooo.metric.domain-completeness.reverse-observation-coverage.v1": {
		Unit: "observations",
		Evaluate: func(spec dimensionSpec, _ ProfileModel, inputs loadedInputs, _ string, _ bool, _ int64, _ int) Dimension {
			return measureReverseObservationCoverage(spec, inputs)
		},
	},
	"gooo.metric.domain-completeness.use-case-coverage.v1": {
		Unit: "use_cases",
		Evaluate: func(spec dimensionSpec, _ ProfileModel, inputs loadedInputs, _ string, _ bool, _ int64, _ int) Dimension {
			return measureUseCaseCoverage(spec, inputs.contract, inputs.report, inputs)
		},
	},
	"gooo.metric.domain-completeness.boundary-coverage.v1": {
		Unit: "boundaries",
		Evaluate: func(spec dimensionSpec, _ ProfileModel, inputs loadedInputs, _ string, _ bool, _ int64, _ int) Dimension {
			return measureBoundaryCoverage(spec, inputs)
		},
	},
	"gooo.metric.domain-completeness.provenance-integrity.v1": {
		Unit: "bindings",
		Evaluate: func(spec dimensionSpec, profile ProfileModel, inputs loadedInputs, subject string, profileSemanticsEqual bool, runID int64, attempt int) Dimension {
			return measureProvenanceIntegrity(spec, profile, inputs, subject, profileSemanticsEqual, runID, attempt)
		},
	},
}

func measureDeclarationCoverage(spec dimensionSpec, contract languageutility.Contract, program []byte, inputs loadedInputs) Dimension {
	dimension := newDimension(spec, len(contract.UseCases)*len(contract.Stages))
	if len(program) == 0 || dimension.Denominator == 0 {
		return unknown(dimension, "LANGUAGE_UTILITY_PROGRAM_OR_CONTRACT_MISSING", "GENERATE_DECLARED_PROGRAM")
	}
	file, diagnostics := syntax.ParseFile("language-utility-program.gooo", string(program))
	if file == nil || diagnostics.HasErrors() {
		return failed(dimension, "GENERATED_PROGRAM_INVALID", "REPAIR_GENERATED_PROGRAM")
	}
	activities := make(map[string]int)
	for _, declaration := range file.Decls {
		if activity, ok := declaration.(*syntax.ActivityDecl); ok {
			activities[activity.Name]++
		}
	}
	expectedProgram, err := languageutility.GenerateProgram(contract)
	if err != nil {
		return failed(dimension, "CONTRACT_PROGRAM_GENERATION_FAILED", "REPAIR_CONTRACT_GENERATOR")
	}
	if !bytes.Equal(program, []byte(expectedProgram)) {
		return failed(dimension, "GENERATED_PROGRAM_DIFFERS_FROM_CONTRACT", "REGENERATE_DECLARED_PROGRAM")
	}
	for _, useCase := range contract.UseCases {
		for _, stage := range contract.Stages {
			name := "Observe" + symbol(useCase.ID) + symbol(stage.ID)
			if activities[name] == 1 {
				dimension.Numerator++
			}
		}
	}
	dimension.Evidence = []EvidenceRef{{Role: "source-contract", Path: "examples/language-utility/contract.json", Digest: digestBytes(inputs.contractRaw)}}
	dimension.Evidence = append(dimension.Evidence, inputs.discovery.Refs...)
	dimension.Status = classify(dimension.Numerator, dimension.Denominator, 0, false)
	if dimension.Status != "PASS" {
		dimension.FirstUnresolved = &Frontier{Unit: "declared-cell", Stage: "SYNTAX_ACCEPTED", Reason: "GENERATED_ACTIVITY_MISSING", NextOperation: "REGENERATE_DECLARED_PROGRAM"}
	}
	return dimension
}

func measureGenerationCoverage(spec dimensionSpec, contract languageutility.Contract, report languageutility.Report, inputs loadedInputs) Dimension {
	dimension := newDimension(spec, len(contract.UseCases))
	if len(report.Cells) == 0 {
		return unknown(dimension, "UTILITY_REPORT_MISSING", "GENERATE_UTILITY_REPORT")
	}
	for _, useCase := range contract.UseCases {
		if useCase.ID == "capability-discovery" {
			if inputs.discovery.State == "PASS" {
				dimension.UnknownUnits++
			} else if inputs.discovery.State == "FAIL_CLOSED" {
				return failed(dimension, inputs.discovery.Reason, "REPAIR_CAPABILITY_DISCOVERY_RECEIPT")
			} else {
				dimension.UnknownUnits++
			}
			continue
		}
		var found *languageutility.CellResult
		for index := range report.Cells {
			cell := &report.Cells[index]
			if cell.UseCaseID == useCase.ID && cell.StageID == "USER_ARTIFACT_VERIFIED" {
				found = cell
				break
			}
		}
		if found == nil {
			continue
		}
		if found.State == languageutility.StateClosed && found.EvidencePath != "" && found.EvidenceDigest != "" &&
			inputs.evidenceState == "PASS" {
			dimension.Numerator++
		}
	}
	dimension.Evidence = append(dimension.Evidence, inputs.evidenceRefs...)
	dimension.Evidence = append(dimension.Evidence, inputs.discovery.Refs...)
	if inputs.discovery.State != "PASS" || dimension.UnknownUnits > 0 {
		dimension.Status = "UNKNOWN"
		dimension.UnknownUnits = max(dimension.UnknownUnits, 1)
		dimension.FirstUnresolved = &Frontier{Unit: "capability-discovery", Stage: "GENERATION_COVERAGE",
			Reason: "DISCOVERY_RECEIPT_HAS_NO_GENERATED_CODE_ARTIFACT", NextOperation: "GENERATE_A_SOURCE_BOUND_CODE_ARTIFACT"}
		return dimension
	}
	dimension.Status = classify(dimension.Numerator, dimension.Denominator, 0, false)
	if dimension.Status != "PASS" {
		dimension.FirstUnresolved = firstOpenCell(report, contract)
		if dimension.FirstUnresolved == nil {
			dimension.FirstUnresolved = &Frontier{Unit: "generated-artifact", Stage: "USER_ARTIFACT_VERIFIED", Reason: inputs.evidenceReason, NextOperation: "COLLECT_GENERATION_EVIDENCE"}
		}
	}
	return dimension
}

func measureReverseObservationCoverage(spec dimensionSpec, inputs loadedInputs) Dimension {
	dimension := newDimension(spec, 2)
	if len(inputs.programRaw) == 0 || len(inputs.graphRaw) == 0 || len(inputs.finalGraphRaw) == 0 {
		return unknown(dimension, "SOURCE_OR_REVERSE_GRAPH_MISSING", "GENERATE_AND_REOBSERVE_STRUCTURE")
	}
	if inputs.graph.GraphHash == "" || inputs.finalGraph.GraphHash == "" {
		return unknown(dimension, "GRAPH_IDENTITY_MISSING", "OBSERVE_GENERATED_GRAPH")
	}
	dimension.Evidence = []EvidenceRef{
		{Role: "generated-structure", Path: "program.gooo", Digest: digestBytes(inputs.programRaw)},
		{Role: "source-graph", Path: "gooo-graph.json", Digest: digestBytes(inputs.graphRaw)},
		{Role: "reverse-observation", Path: "final-gooo-graph.json", Digest: digestBytes(inputs.finalGraphRaw)},
	}
	dimension.Evidence = append(dimension.Evidence, inputs.discovery.Refs...)
	if !bytes.Equal(inputs.graphRaw, inputs.finalGraphRaw) {
		if !equalGraph(inputs.graph, inputs.finalGraph) {
			return failed(dimension, "REVERSE_GRAPH_MISMATCH", "REGENERATE_OR_REOBSERVE_STRUCTURE")
		}
	}
	if equalGraph(inputs.graph, inputs.finalGraph) {
		dimension.Numerator = 1
		if inputs.discovery.State != "PASS" {
			dimension.UnknownUnits = 1
			dimension.Status = "UNKNOWN"
			dimension.FirstUnresolved = &Frontier{Unit: "capability-discovery", Stage: "REVERSE_OBSERVATION_COVERAGE",
				Reason: inputs.discovery.Reason, NextOperation: "REVERSE_OBSERVE_A_DISCOVERY_BOUND_ARTIFACT"}
			return dimension
		}
		reverse, ok := requireDiscoveryReceiptDimension(inputs.discovery.Report, "reverse_observation_coverage")
		if !ok || reverse.Status != "UNKNOWN" {
			return failed(dimension, "CAPABILITY_DISCOVERY_REVERSE_RECEIPT_CHANGED", "REPAIR_DISCOVERY_REVERSE_EVIDENCE")
		}
		dimension.UnknownUnits = 1
		dimension.Status = "UNKNOWN"
		dimension.FirstUnresolved = &Frontier{Unit: "capability-discovery", Stage: "REVERSE_OBSERVATION_COVERAGE",
			Reason: reverse.Reason, NextOperation: "REVERSE_OBSERVE_A_DISCOVERY_BOUND_ARTIFACT"}
		return dimension
	}
	return failed(dimension, "REVERSE_GRAPH_MISMATCH", "REGENERATE_OR_REOBSERVE_STRUCTURE")
}

func measureUseCaseCoverage(spec dimensionSpec, contract languageutility.Contract, report languageutility.Report, inputs loadedInputs) Dimension {
	dimension := newDimension(spec, len(contract.UseCases))
	if len(report.UseCases) == 0 {
		return unknown(dimension, "USE_CASE_SUMMARY_MISSING", "EVALUATE_DECLARED_USE_CASES")
	}
	dimension.Numerator = report.Summary.CompleteUseCases
	dimension.UnknownUnits = report.Summary.UnknownCells
	dimension.RefutedUnits = report.Summary.RefutedCells
	dimension.Evidence = []EvidenceRef{{Role: "use-case-outcomes", Path: "report.json", Digest: digestBytes(inputs.reportRaw)}}
	dimension.Evidence = append(dimension.Evidence, inputs.discovery.Refs...)
	if dimension.RefutedUnits > 0 {
		return failed(dimension, "USE_CASE_EVIDENCE_REFUTED", "RETAIN_AND_REPAIR_COUNTEREXAMPLE")
	}
	if dimension.UnknownUnits > 0 {
		return unknown(dimension, "USE_CASE_EVIDENCE_UNKNOWN", "RESOLVE_FIRST_UNKNOWN_USE_CASE")
	}
	unknownUnits := 0
	if inputs.discovery.State == "PASS" {
		realUseCase, ok := requireDiscoveryReceiptDimension(inputs.discovery.Report, "real_use_case_coverage")
		if !ok || realUseCase.Status != "UNKNOWN" {
			return failed(dimension, "CAPABILITY_DISCOVERY_USE_CASE_RECEIPT_CHANGED", "REPAIR_DISCOVERY_USE_CASE_EVIDENCE")
		}
		unknownUnits++
		dimension.FirstUnresolved = &Frontier{Unit: "capability-discovery", Stage: "REAL_USE_CASE_COVERAGE",
			Reason: realUseCase.Reason, NextOperation: "BIND_INDEPENDENT_INPUT_AND_EXPECTED_OUTPUT_CASES"}
	} else {
		unknownUnits++
	}
	dimension.UnknownUnits += unknownUnits
	dimension.Status = classify(dimension.Numerator, dimension.Denominator, dimension.UnknownUnits, false)
	if dimension.Status != "PASS" {
		if dimension.FirstUnresolved == nil {
			dimension.FirstUnresolved = firstOpenCell(report, contract)
		}
	}
	return dimension
}

func measureBoundaryCoverage(spec dimensionSpec, inputs loadedInputs) Dimension {
	dimension := newDimension(spec, 6)
	if len(inputs.inventoryRaw) == 0 || len(inputs.observationRaw) == 0 {
		return unknown(dimension, "BOUNDARY_OBSERVATION_MISSING", "COLLECT_BOUNDARY_OBSERVATIONS")
	}
	checks := []bool{
		inputs.inventory.RepositoryWrites == 0,
		!inputs.inventory.MutationAuthority,
		!inputs.inventory.GenerationAuthority,
		!inputs.inventory.RepairAuthority,
		!inputs.inventory.MergeAuthority,
	}
	dimension.Evidence = []EvidenceRef{
		{Role: "repository-boundary", Path: "inventory.json", Digest: digestBytes(inputs.inventoryRaw)},
		{Role: "observer-write-count", Path: "observation.json", Digest: digestBytes(inputs.observationRaw)},
	}
	dimension.Evidence = append(dimension.Evidence, inputs.discovery.Refs...)
	for _, passed := range checks {
		if passed {
			dimension.Numerator++
		} else {
			dimension.RefutedUnits++
		}
	}
	if inputs.discovery.State == "PASS" {
		dimension.Numerator++
	} else if inputs.discovery.State == "FAIL_CLOSED" {
		dimension.RefutedUnits++
	} else {
		dimension.UnknownUnits++
	}
	if dimension.RefutedUnits > 0 || inputs.observation.RepositoryWrites != 0 {
		return failed(dimension, "READ_ONLY_BOUNDARY_REFUTED", "RESTORE_READ_ONLY_OBSERVER_BOUNDARY")
	}
	dimension.Status = classify(dimension.Numerator, dimension.Denominator, dimension.UnknownUnits, false)
	if dimension.Status != "PASS" {
		dimension.FirstUnresolved = &Frontier{Unit: "authority-boundary", Stage: "BOUNDARY_COVERAGE", Reason: "BOUNDARY_CHECK_INCOMPLETE", NextOperation: "COLLECT_BOUNDARY_OBSERVATIONS"}
	}
	return dimension
}

func measureProvenanceIntegrity(spec dimensionSpec, profile ProfileModel, inputs loadedInputs, subject string, profileSemanticsEqual bool, runID int64, attempt int) Dimension {
	dimension := newDimension(spec, 9)
	check := func(ok bool, missing bool) {
		if missing {
			dimension.UnknownUnits++
		} else if ok {
			dimension.Numerator++
		} else {
			dimension.RefutedUnits++
		}
	}
	reportLoaded := len(inputs.reportRaw) > 0
	observationLoaded := len(inputs.observationRaw) > 0
	contractLoaded := len(inputs.contractRaw) > 0
	programLoaded := len(inputs.programRaw) > 0
	check(canonicalSubject(subject) && inputs.report.SubjectSHA == subject && inputs.observation.SubjectSHA == subject, !reportLoaded || !observationLoaded || subject == "")
	check(inputs.progress.Contract.Digest == digestBytes(inputs.contractRaw), !contractLoaded || len(inputs.progressRaw) == 0)
	check(inputs.discovery.State == "PASS", inputs.discovery.State == "UNKNOWN")
	expectedProgram, programErr := languageutility.GenerateProgram(inputs.contract)
	check(programErr == nil && bytes.Equal(inputs.programRaw, []byte(expectedProgram)) &&
		inputs.report.ProgramDigest == digestBytes(inputs.programRaw), !contractLoaded || !programLoaded || !reportLoaded)
	replayed := false
	if contractLoaded && observationLoaded && reportLoaded {
		replay, err := languageutility.Evaluate(inputs.contract, inputs.observation)
		replayed = err == nil && replay.Digest == inputs.report.Digest
		check(replayed, false)
	} else {
		check(false, true)
	}
	check(inputs.evidenceState == "PASS", inputs.evidenceState == "UNKNOWN")
	check(profile.SemanticHash != "" && profileSemanticsEqual && inputs.report.ContractID == inputs.contract.ID, !reportLoaded || !contractLoaded)
	runIdentityMissing := os.Getenv("GITHUB_REPOSITORY") == "" || runID <= 0 || attempt <= 0
	check(!runIdentityMissing, runIdentityMissing)
	check(toolchainIdentity() == "go1.27.1", false)
	dimension.Evidence = []EvidenceRef{
		{Role: "profile-source", Path: "scripts/domain-completeness/profile.gooo", Digest: profile.SourceDigest},
		{Role: "utility-contract", Path: "examples/language-utility/contract.json", Digest: digestBytes(inputs.contractRaw)},
		{Role: "utility-observation", Path: "observation.json", Digest: digestBytes(inputs.observationRaw)},
		{Role: "utility-report", Path: "report.json", Digest: digestBytes(inputs.reportRaw)},
		{Role: "generated-program", Path: "program.gooo", Digest: digestBytes(inputs.programRaw)},
	}
	dimension.Evidence = append(dimension.Evidence, inputs.discovery.Refs...)
	if dimension.RefutedUnits > 0 {
		return failed(dimension, "PROVENANCE_BINDING_MISMATCH", "REGENERATE_EXACT_PROVENANCE_RECEIPT")
	}
	dimension.Status = classify(dimension.Numerator, dimension.Denominator, dimension.UnknownUnits, false)
	if dimension.Status != "PASS" {
		dimension.FirstUnresolved = &Frontier{Unit: "exact-run-binding", Stage: "PROVENANCE_INTEGRITY", Reason: "PROVENANCE_BINDING_UNKNOWN", NextOperation: "COLLECT_EXACT_PROVENANCE"}
	}
	return dimension
}

func newDimension(spec dimensionSpec, denominator int) Dimension {
	return Dimension{
		ID: spec.ID, MetricID: spec.MetricID, Status: "UNKNOWN",
		Denominator: denominator, Unit: spec.Unit, Evidence: []EvidenceRef{},
	}
}

func classify(numerator, denominator, unknown int, refuted bool) string {
	return domaincompleteness.ClassifyDomainCompleteness(
		int64(numerator), int64(denominator), int64(unknown), refuted,
	)
}

func unknown(dimension Dimension, reason, operation string) Dimension {
	dimension.Status = "UNKNOWN"
	if dimension.UnknownUnits == 0 {
		dimension.UnknownUnits = max(dimension.Denominator-dimension.Numerator, 0)
	}
	dimension.FirstUnresolved = &Frontier{Unit: dimension.ID, Stage: dimension.ID, Reason: reason, NextOperation: operation}
	return dimension
}

func failed(dimension Dimension, reason, operation string) Dimension {
	dimension.Status = "FAIL_CLOSED"
	if dimension.RefutedUnits == 0 {
		dimension.RefutedUnits = 1
	}
	dimension.FirstUnresolved = &Frontier{Unit: dimension.ID, Stage: dimension.ID, Reason: reason, NextOperation: operation}
	return dimension
}

func summarize(values []Dimension) Summary {
	result := Summary{DimensionsTotal: len(values)}
	for _, value := range values {
		switch value.Status {
		case "PASS":
			result.Pass++
		case "PROGRESS":
			result.Progress++
		case "UNKNOWN":
			result.Unknown++
		case "FAIL_CLOSED":
			result.FailClosed++
		}
	}
	return result
}

func decide(values []Dimension, issues []string, evidenceState string) (string, string, string, *Frontier) {
	invalidInput := len(issues) > 0 || evidenceState == "FAIL_CLOSED"
	outcome := domaincompleteness.SelectDomainCompletenessOutcome(
		invalidInput,
		hasDimensionStatus(values, "FAIL_CLOSED"),
		hasDimensionStatus(values, "UNKNOWN"),
		hasDimensionStatus(values, "PROGRESS"),
	)
	switch outcome {
	case "FAIL_CLOSED":
		if invalidInput {
			return outcome, "DOMAIN_RECEIPT_INPUT_INVALID", "REPAIR_DOMAIN_RECEIPT_INPUTS", nil
		}
		for _, value := range values {
			if value.Status == outcome {
				operation, frontier := nextOperation(value, "REPAIR_DOMAIN_EVIDENCE")
				return outcome, "DOMAIN_EVIDENCE_CONTRADICTED", operation, frontier
			}
		}
	case "UNKNOWN":
		for _, value := range values {
			if value.Status == outcome {
				operation, frontier := nextOperation(value, "COLLECT_DOMAIN_EVIDENCE")
				return outcome, "DOMAIN_EVIDENCE_INCOMPLETE", operation, frontier
			}
		}
	case "PROGRESS":
		for _, value := range values {
			if value.Status == outcome {
				operation, frontier := nextOperation(value, "RESOLVE_DOMAIN_GAP")
				return outcome, "PROFILE_GAPS_REMAIN", operation, frontier
			}
		}
	case "PASS":
		return outcome, "ALL_REQUIRED_DIMENSIONS_CLOSED", "NO_ACTION", nil
	}
	return "FAIL_CLOSED", "DOMAIN_OUTCOME_UNSUPPORTED", "REPAIR_DOMAIN_PROFILE", nil
}

func hasDimensionStatus(values []Dimension, status string) bool {
	for _, value := range values {
		if value.Status == status {
			return true
		}
	}
	return false
}

func nextOperation(value Dimension, fallback string) (string, *Frontier) {
	if value.FirstUnresolved == nil {
		return fallback, nil
	}
	return value.FirstUnresolved.NextOperation, value.FirstUnresolved
}

func firstOpenCell(report languageutility.Report, contract languageutility.Contract) *Frontier {
	for _, useCase := range contract.UseCases {
		for _, stage := range contract.Stages {
			for _, cell := range report.Cells {
				if cell.UseCaseID == useCase.ID && cell.StageID == stage.ID && cell.State != languageutility.StateClosed {
					return &Frontier{
						Unit: useCase.ID, Stage: stage.ID, Reason: cell.Reason,
						NextOperation: "RESOLVE_" + stage.ID,
					}
				}
			}
		}
	}
	return nil
}

func equalGraph(left, right graphSnapshot) bool {
	leftBytes, leftErr := json.Marshal(left)
	rightBytes, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftBytes, rightBytes)
}

func reportDigest(value Report) (string, error) {
	value.Digest = ""
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return digestBytes(raw), nil
}

func symbol(value string) string {
	parts := strings.FieldsFunc(value, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	var result strings.Builder
	for _, part := range parts {
		runes := []rune(strings.ToLower(part))
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		result.WriteString(string(runes))
	}
	return result.String()
}

func toolchainIdentity() string {
	return runtime.Version()
}
