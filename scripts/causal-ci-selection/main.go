package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/causalci"
)

func main() {
	mode := flag.String("mode", "produce", "snapshot, observe, produce, interventions, ci-evidence, or adjudicate")
	baseSHA := flag.String("base-sha", "", "base commit for changed-file observation")
	headSHA := flag.String("head-sha", "", "head commit for changed-file observation")
	sourcePath := flag.String("source", "", "logical Gooo policy source path")
	inputPath := flag.String("input", "", "raw observation JSON")
	outputPath := flag.String("output", "", "output artifact")
	beforeSnapshot := flag.String("before-snapshot", "", "pre-isolation repository snapshot")
	afterSnapshot := flag.String("after-snapshot", "", "post-isolation repository snapshot")
	priorClaims := flag.String("prior-claims", "", "raw prior claim templates")
	semanticSource := flag.String("semantic-source", "", "semantic intervention Gooo source")
	nonsemanticSource := flag.String("nonsemantic-source", "", "nonsemantic intervention Gooo source")
	contradictionSource := flag.String("contradiction-source", "", "contradiction intervention Gooo source")
	plansDir := flag.String("plans-dir", "", "directory for intervention plan receipts")
	adjudicationDir := flag.String("adjudication-dir", "", "directory of consumer adjudication receipts")
	scope := flag.String("scope", "base", "base or interventions")
	goVersionPath := flag.String("go-version", "", "captured go version")
	goEnvPath := flag.String("go-env", "", "captured GOVERSION")
	fixHelpPath := flag.String("fix-help", "", "captured go tool fix help")
	fixHelpStderrPath := flag.String("fix-help-stderr", "", "captured go tool fix help stderr")
	fixStdoutPath := flag.String("fix-stdout", "", "captured go fix -diff stdout")
	fixStderrPath := flag.String("fix-stderr", "", "captured go fix -diff stderr")
	fixExitPath := flag.String("fix-exit", "", "captured go fix -diff exit code")
	processDir := flag.String("process-dir", "", "shell observations for independent consumer processes")
	sourceFiles := flag.String("source-files", "", "variant=source path mappings")
	subjectHeadPath := flag.String("subject-head", "", "captured subject checkout HEAD")
	subjectTreePath := flag.String("subject-tree", "", "captured subject checkout tree")
	objectFormatPath := flag.String("object-format", "", "captured Git object format")
	goModDigestPath := flag.String("go-mod-digest", "", "captured go.mod digest")
	goSumDigestPath := flag.String("go-sum-digest", "", "captured go.sum digest")
	goListPath := flag.String("go-list", "", "captured go list ./... package universe")
	goCWDPath := flag.String("go-cwd", "", "captured command working directory")
	activeFixStdoutPath := flag.String("active-fix-stdout", "", "captured active fixer fixture stdout")
	activeFixStderrPath := flag.String("active-fix-stderr", "", "captured active fixer fixture stderr")
	activeFixExitPath := flag.String("active-fix-exit", "", "captured active fixer fixture exit code")
	ciEvidenceDir := flag.String("ci-evidence-dir", "", "directory of raw CI API/process observations")
	ciEvidenceActualCase := flag.String("ci-evidence-actual-case", "", "actual CI observation case ID")
	ciEvidencePath := flag.String("ci-evidence", "", "adjudicated CI API/process observation")
	flag.Parse()

	var err error
	switch *mode {
	case "snapshot":
		err = snapshotCommand(*outputPath)
	case "observe":
		err = observe(*baseSHA, *headSHA, *sourcePath, *beforeSnapshot, *afterSnapshot, *priorClaims, *outputPath)
	case "produce":
		err = produce(*inputPath, *sourcePath, *outputPath)
	case "interventions":
		err = interventions(*inputPath, *sourcePath, *semanticSource, *nonsemanticSource, *contradictionSource, *outputPath, *plansDir)
	case "ci-evidence":
		err = ciEvidence(*ciEvidenceDir, *ciEvidenceActualCase, *outputPath)
	case "adjudicate":
		err = adjudicate(*scope, *inputPath, *plansDir, *adjudicationDir, *processDir, *sourceFiles, *goVersionPath, *goEnvPath, *fixHelpPath, *fixHelpStderrPath, *fixStdoutPath, *fixStderrPath, *fixExitPath, *subjectHeadPath, *subjectTreePath, *objectFormatPath, *goModDigestPath, *goSumDigestPath, *goListPath, *goCWDPath, *activeFixStdoutPath, *activeFixStderrPath, *activeFixExitPath, *ciEvidencePath, *outputPath)
	default:
		err = fmt.Errorf("unknown mode %q", *mode)
	}
	if err != nil {
		fail(err)
	}
}

func snapshotCommand(outputPath string) error {
	if outputPath == "" {
		return fmt.Errorf("snapshot output is required")
	}
	tracked, err := gitPaths("ls-files", "-z")
	if err != nil {
		return err
	}
	untracked, err := gitPaths("ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return err
	}
	objectFormat := strings.TrimSpace(commandOutputIgnoreError("git", "rev-parse", "--show-object-format"))
	if objectFormat == "" {
		return fmt.Errorf("Git object format is unavailable")
	}
	entries := make([]causalci.RepositoryEntry, 0, len(tracked)+len(untracked))
	seen := map[string]bool{}
	for _, path := range tracked {
		entries = append(entries, repositoryEntry(path, true, objectFormat))
		seen[path] = true
	}
	for _, path := range untracked {
		if !seen[path] {
			entries = append(entries, repositoryEntry(path, false, objectFormat))
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return writeJSON(outputPath, causalci.RepositorySnapshot{Entries: entries, SnapshotDigest: mustJSONDigest(entries)})
}

func repositoryEntry(path string, tracked bool, objectFormat string) causalci.RepositoryEntry {
	entry := causalci.RepositoryEntry{Path: path, Tracked: tracked, ObjectFormat: objectFormat}
	info, err := os.Lstat(path)
	if err != nil {
		entry.Kind, entry.Mode, entry.ContentDigest, entry.ObjectID = "missing", "000000", "MISSING:"+err.Error(), "MISSING:"+err.Error()
		return entry
	}
	switch {
	case info.Mode()&os.ModeSymlink != 0:
		entry.Kind, entry.Mode = "symlink", "120000"
		target, targetErr := os.Readlink(path)
		if targetErr != nil {
			entry.SymlinkTargetDigest = "MISSING:" + targetErr.Error()
			entry.ContentDigest = entry.SymlinkTargetDigest
		} else {
			entry.SymlinkTargetDigest = bytesDigest([]byte(target))
			entry.ContentDigest = bytesDigest([]byte(target))
		}
	case info.IsDir():
		entry.Kind, entry.Mode, entry.ContentDigest = "directory", "040000", "directory"
	default:
		entry.Kind = "file"
		if info.Mode().Perm()&0o111 != 0 {
			entry.Mode = "100755"
		} else {
			entry.Mode = "100644"
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			entry.ContentDigest = "MISSING:" + readErr.Error()
		} else {
			entry.ContentDigest = bytesDigest(data)
		}
	}
	entry.ObjectID = strings.TrimSpace(commandOutputIgnoreError("git", "hash-object", "--", path))
	if entry.ObjectID == "" {
		entry.ObjectID = strings.TrimSpace(commandOutputIgnoreError("git", "rev-parse", "HEAD:"+path))
	}
	if entry.ObjectID == "" {
		entry.ObjectID = "UNAVAILABLE:" + path
	}
	return entry
}

func gitPaths(args ...string) ([]string, error) {
	data, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("git path observation: %w", err)
	}
	parts := strings.Split(strings.TrimSuffix(string(data), "\x00"), "\x00")
	if len(parts) == 1 && parts[0] == "" {
		return []string{}, nil
	}
	return parts, nil
}

func observe(baseSHA, headSHA, sourcePath, beforePath, afterPath, priorPath, outputPath string) error {
	if baseSHA == "" || headSHA == "" || sourcePath == "" || beforePath == "" || afterPath == "" || priorPath == "" || outputPath == "" {
		return fmt.Errorf("observe requires base-sha, head-sha, source, before-snapshot, after-snapshot, prior-claims, and output")
	}
	changed, err := changedFiles(baseSHA, headSHA)
	if err != nil {
		return err
	}
	before, err := readSnapshot(beforePath)
	if err != nil {
		return err
	}
	after, err := readSnapshot(afterPath)
	if err != nil {
		return err
	}
	templates, err := readPriorClaims(priorPath)
	if err != nil {
		return err
	}
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read Gooo source: %w", err)
	}
	checkout, err := commandOutput("git", "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	checkout = strings.TrimSpace(checkout)
	objectFormat := strings.TrimSpace(commandOutputIgnoreError("git", "rev-parse", "--show-object-format"))
	if objectFormat == "" {
		return fmt.Errorf("Git object format is unavailable")
	}
	headObject := objectAt(headSHA, sourcePath)
	if headObject == "" {
		return fmt.Errorf("source is not present at head")
	}
	claims := make([]causalci.PriorClaimObservation, 0, len(changed)*len(templates))
	for _, file := range changed {
		for _, template := range templates {
			claims = append(claims, causalci.PriorClaimObservation{TemplateID: template.TemplateID, InstanceID: causalci.ClaimInstanceID(template.TemplateID, file.Path, template.Proposition), SubjectPath: file.Path, Proposition: template.Proposition, State: template.State, Provenance: template.Provenance})
		}
	}
	value := causalci.Observation{Schema: causalci.ObservationSchema, Repository: "kimjooyoon/meta-ontology-go", BaseSHA: baseSHA, HeadSHA: headSHA, ObservedCheckoutSHA: checkout, SourcePath: sourcePath, ObjectFormat: objectFormat, HeadPathObjectID: headObject, SourceBytesDigest: bytesDigest(source), ChangedFiles: changed, PriorClaims: claims, Isolation: causalci.IsolationObservation{Before: before, After: after}}
	return writeJSON(outputPath, value)
}

type priorClaimTemplate struct {
	Schema string `json:"schema"`
	Claims []struct {
		TemplateID  string `json:"template_id"`
		Proposition string `json:"proposition"`
		State       string `json:"state"`
		Provenance  string `json:"provenance"`
	} `json:"claims"`
}

func readPriorClaims(path string) ([]priorClaimTemplateClaim, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var value priorClaimTemplate
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("decode prior claim templates: %w", err)
	}
	if value.Schema != "gooo/causal-ci-selection-prior-observation/v1" || len(value.Claims) == 0 {
		return nil, fmt.Errorf("malformed prior claim templates")
	}
	result := make([]priorClaimTemplateClaim, 0, len(value.Claims))
	for _, claim := range value.Claims {
		if claim.TemplateID == "" || claim.Proposition == "" || claim.State == "" || claim.Provenance == "" {
			return nil, fmt.Errorf("malformed prior claim template")
		}
		result = append(result, priorClaimTemplateClaim{TemplateID: claim.TemplateID, Proposition: claim.Proposition, State: claim.State, Provenance: claim.Provenance})
	}
	return result, nil
}

type priorClaimTemplateClaim struct{ TemplateID, Proposition, State, Provenance string }

func changedFiles(baseSHA, headSHA string) ([]causalci.ChangedFileObservation, error) {
	raw, err := exec.Command("git", "diff", "--name-status", "--no-renames", "--diff-filter=ACMRTUXB", baseSHA+"..."+headSHA).Output()
	if err != nil {
		return nil, fmt.Errorf("collect PR changed files: %w", err)
	}
	text := strings.TrimSuffix(string(raw), "\n")
	if text == "" {
		return []causalci.ChangedFileObservation{}, nil
	}
	lines := strings.Split(text, "\n")
	result := make([]causalci.ChangedFileObservation, 0, len(lines))
	for _, line := range lines {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("malformed git changed-file line %q", line)
		}
		result = append(result, causalci.ChangedFileObservation{Path: parts[1], Status: parts[0], BeforeObject: objectAt(baseSHA, parts[1]), AfterObject: objectAt(headSHA, parts[1])})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result, nil
}

func objectAt(revision, path string) string {
	return strings.TrimSpace(commandOutputIgnoreError("git", "rev-parse", revision+":"+path))
}

func readSnapshot(path string) (causalci.RepositorySnapshot, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return causalci.RepositorySnapshot{}, err
	}
	var value causalci.RepositorySnapshot
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, fmt.Errorf("decode repository snapshot: %w", err)
	}
	return value, nil
}

func produce(inputPath, sourcePath, outputPath string) error {
	input, source, err := readInputs(inputPath, sourcePath)
	if err != nil {
		return err
	}
	receipt, err := causalci.Evaluate(input, sourcePath, source)
	if err != nil {
		return err
	}
	if err := writeJSON(outputPath, receipt); err != nil {
		return err
	}
	state := receipt.Operation.ObservedRepositoryState
	fmt.Printf("causal CI plan: conformance=%s subjects=%d selected=%d unknown=%d fail_closed=%d checks=%d/%d transitions=%d discharged=%d lower_resolution=%d refuted=%d execution=%s net_state=%s changed_paths=%d changed_contents=%d\n", receipt.Conformance.Decision, receipt.Metrics.SubjectTotal, receipt.Metrics.SelectedSubjectTotal, receipt.Metrics.UnknownSubjectTotal, receipt.Metrics.FailClosedSubjectTotal, receipt.Metrics.SelectedCheckTotal, receipt.Metrics.FullSuiteCheckDenominator, receipt.Metrics.ClaimTransitionTotal, receipt.Metrics.DischargedClaimTotal, receipt.Metrics.LowerResolutionClaimTotal, receipt.Metrics.RefutedClaimTotal, receipt.Execution.Result, state.NetState, state.ChangedPathCount, state.ChangedContentCount)
	return nil
}

func interventions(inputPath, sourcePath, semanticPath, nonsemanticPath, contradictionPath, outputPath, plansDir string) error {
	input, baseSource, err := readInputs(inputPath, sourcePath)
	if err != nil {
		return err
	}
	if plansDir == "" {
		return fmt.Errorf("interventions requires plans-dir")
	}
	variants := []struct{ id, path string }{{"base", sourcePath}, {"semantic", semanticPath}, {"nonsemantic", nonsemanticPath}, {"contradiction", contradictionPath}}
	results := make([]causalci.InterventionResult, 0, len(variants))
	observed := make([]string, 0, len(variants))
	for _, variant := range variants {
		if variant.path == "" {
			return fmt.Errorf("interventions requires all variant source paths")
		}
		variantSource := baseSource
		if variant.path != sourcePath {
			variantSource, err = os.ReadFile(variant.path)
			if err != nil {
				return err
			}
		}
		receipt, evalErr := causalci.EvaluateIntervention(input, sourcePath, variantSource)
		if evalErr != nil {
			return fmt.Errorf("evaluate %s intervention: %w", variant.id, evalErr)
		}
		if err := writeJSON(filepath.Join(plansDir, variant.id+"-receipt.json"), receipt); err != nil {
			return err
		}
		results = append(results, causalci.InterventionResult{ID: variant.id, Source: receipt.Source, Conformance: receipt.Conformance, PlanGate: receipt.PlanGate, PolicyContradictions: receipt.PolicyContradictions, Subjects: receipt.Subjects, ClaimTransitions: receipt.ClaimTransitions, Execution: receipt.Execution, PlanDigest: receipt.PlanDigest})
		observed = append(observed, variant.id)
	}
	expected := []string{"base", "semantic", "nonsemantic", "contradiction"}
	if !sameIDs(expected, observed) {
		return fmt.Errorf("intervention variant inventory mismatch")
	}
	report := causalci.InterventionReport{Schema: causalci.ReportSchema, ObservationDigest: bytesDigest(input), ExpectedVariantIDs: expected, ObservedVariantIDs: observed, Base: results[0], Semantic: results[1], Nonsemantic: results[2], Contradiction: results[3]}
	report.Digest, err = jsonDigestWithoutDigest(report)
	if err != nil {
		return err
	}
	if err := writeJSON(outputPath, report); err != nil {
		return err
	}
	fmt.Printf("intervention plans: variants=%d/%d execution=%s semantic_plan_changed=%t nonsemantic_plan_preserved=%t contradiction=%s\n", len(observed), len(report.ExpectedVariantIDs), report.Base.Execution.Result, report.Base.PlanDigest != report.Semantic.PlanDigest, report.Base.PlanDigest == report.Nonsemantic.PlanDigest, report.Contradiction.Conformance.Decision)
	return nil
}

func ciEvidence(directory, actualCaseID, outputPath string) error {
	if directory == "" || outputPath == "" {
		return fmt.Errorf("ci-evidence requires directory and output")
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	observations := make([]causalci.CIEvidenceObservation, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		raw, readErr := os.ReadFile(filepath.Join(directory, entry.Name()))
		if readErr != nil {
			return readErr
		}
		var observation causalci.CIEvidenceObservation
		decoder := json.NewDecoder(strings.NewReader(string(raw)))
		decoder.DisallowUnknownFields()
		if decodeErr := decoder.Decode(&observation); decodeErr != nil {
			return fmt.Errorf("decode CI evidence fixture %s: %w", entry.Name(), decodeErr)
		}
		var trailing any
		if decodeErr := decoder.Decode(&trailing); decodeErr != io.EOF {
			if decodeErr == nil {
				return fmt.Errorf("CI evidence fixture %s has trailing JSON", entry.Name)
			}
			return fmt.Errorf("CI evidence fixture %s has trailing bytes: %w", entry.Name, decodeErr)
		}
		observations = append(observations, observation)
	}
	result, err := causalci.AdjudicateCIEvidence(observations, actualCaseID)
	if err != nil {
		return err
	}
	if err := writeJSON(outputPath, result); err != nil {
		return err
	}
	actualOutcome := causalci.CIEvidenceOutcomeOpen
	for _, row := range result.Rows {
		if row.CaseID == result.ActualCaseID {
			actualOutcome = row.Outcome
			break
		}
	}
	fmt.Printf("CI causal evidence: cases=%d/%d actual=%s outcome=%s resolution=%s root_cause=%d/%d downstream_missing=%d/%d\n", len(result.ObservedCaseIDs), len(result.ExpectedCaseIDs), result.ActualCaseID, actualOutcome, result.ActualResolution, result.ActualRootCauseNumerator, result.ActualRootCauseDenominator, result.DownstreamMissingNumerator, result.DownstreamMissingDenominator)
	return nil
}

func adjudicate(scope, inputPath, plansDir, adjudicationDir, processDir, sourceFilesRaw, goVersionPath, goEnvPath, fixHelpPath, fixHelpStderrPath, fixStdoutPath, fixStderrPath, fixExitPath, subjectHeadPath, subjectTreePath, objectFormatPath, goModDigestPath, goSumDigestPath, goListPath, goCWDPath, activeFixStdoutPath, activeFixStderrPath, activeFixExitPath, ciEvidencePath, outputPath string) error {
	if inputPath == "" || plansDir == "" || adjudicationDir == "" || outputPath == "" {
		return fmt.Errorf("adjudicate requires input, plans-dir, adjudication-dir, and output")
	}
	ids := []string{"base"}
	if scope == "interventions" {
		ids = []string{"base", "semantic", "nonsemantic", "contradiction"}
	} else if scope != "base" {
		return fmt.Errorf("adjudicate scope must be base or interventions")
	}
	input, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}
	var observation causalci.Observation
	if err := json.Unmarshal(input, &observation); err != nil {
		return err
	}
	sources, err := parseSourceFiles(sourceFilesRaw)
	if err != nil {
		return err
	}
	adjudications := make([]causalci.ConsumerAdjudication, 0, len(ids))
	processEvidence := make([]causalci.ProcessExecutionEvidence, 0, len(ids))
	sourcePlanBinding := make([]causalci.SourcePlanBindingEvidence, 0, len(ids))
	observed := make([]string, 0, len(ids))
	sourceBindingNumerator := 0
	sourceReconstructionNumerator := 0
	for _, id := range ids {
		planRaw, err := os.ReadFile(filepath.Join(plansDir, id+"-receipt.json"))
		if err != nil {
			return err
		}
		var plan causalci.Receipt
		if err := json.Unmarshal(planRaw, &plan); err != nil {
			return err
		}
		raw, err := os.ReadFile(filepath.Join(adjudicationDir, id+".json"))
		if err != nil {
			return err
		}
		var value causalci.ConsumerAdjudication
		decoder := json.NewDecoder(strings.NewReader(string(raw)))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&value); err != nil {
			return err
		}
		if value.Schema != causalci.ProcessObservationSchema || value.VariantID != id || value.PlanReceiptDigest != plan.Digest || value.InputDigest != bytesDigest(input) || value.ResultDigest == "" || value.ConsumerIdentity == "" || value.LogicalSourcePath == "" || value.BindingMode == "" || digestConsumerAdjudication(value) != value.Digest {
			return fmt.Errorf("invalid adjudication evidence for %s", id)
		}
		sourcePath, exists := sources[id]
		if !exists {
			return fmt.Errorf("missing source mapping for %s", id)
		}
		actualSource, err := os.ReadFile(sourcePath)
		if err != nil {
			return err
		}
		actualSourceDigest := bytesDigest(actualSource)
		actualObjectID := causalci.GitBlobObjectIDForFormat(actualSource, plan.Source.ObjectFormat)
		expectedBindingMode := "INTERVENTION"
		if id == "base" {
			expectedBindingMode = "HEAD"
		}
		headBindingOK := id != "base" || (observation.ObservedCheckoutSHA == observation.HeadSHA && plan.Source.ObservedCheckoutSHA == observation.ObservedCheckoutSHA && plan.Source.HeadPathObjectID == actualObjectID)
		bindingOK := value.LogicalSourcePath == plan.Source.Path && value.LogicalSourcePath == observation.SourcePath && value.BindingMode == plan.Source.BindingMode && value.BindingMode == expectedBindingMode && value.SourceBytesDigest == actualSourceDigest && plan.Source.RawDigest == actualSourceDigest && plan.Source.SourceBytesDigest == actualSourceDigest && plan.Source.ActualSourceObjectID == actualObjectID && plan.Source.ObjectFormat != "" && plan.Source.ObjectFormat == observation.ObjectFormat && headBindingOK
		sourcePlanBinding = append(sourcePlanBinding, causalci.SourcePlanBindingEvidence{VariantID: id, PlanReceiptDigest: plan.Digest, ExpectedSourceRawDigest: plan.Source.RawDigest, ExpectedSourceBytesDigest: plan.Source.SourceBytesDigest, ExpectedSourceObjectID: plan.Source.ActualSourceObjectID, ExpectedObjectFormat: plan.Source.ObjectFormat, ActualConsumerSourceBytesDigest: value.SourceBytesDigest, ActualConsumerSourceObjectID: actualObjectID, LogicalSourcePath: value.LogicalSourcePath, BindingMode: value.BindingMode, ExpectedBindingMode: expectedBindingMode, Exact: bindingOK})
		if bindingOK {
			sourceBindingNumerator++
		}
		osExit, stdout, stderr, err := readProcessEvidence(processDir, id)
		if err != nil {
			return err
		}
		processEvidence = append(processEvidence, causalci.ProcessExecutionEvidence{VariantID: id, SelfReportedExitCode: value.ExitCode, SelfReportedResult: value.Result, ObservedOSExitCode: osExit, ObservedStdoutDigest: bytesDigest(stdout), ObservedStdoutBytes: len(stdout), ObservedStderrDigest: bytesDigest(stderr), ObservedStderrBytes: len(stderr), ResultDigest: value.ResultDigest})
		processOK := processObservationConformant(value, osExit, stderr)
		if processOK && bindingOK && plan.PlanGate.Decision == causalci.PlanGatePass {
			sourceReconstructionNumerator++
		}
		adjudications = append(adjudications, value)
		observed = append(observed, id)
	}
	goEvidence, err := readGoEvidence(goVersionPath, goEnvPath, fixHelpPath, fixHelpStderrPath, fixStdoutPath, fixStderrPath, fixExitPath, subjectHeadPath, subjectTreePath, objectFormatPath, goModDigestPath, goSumDigestPath, goListPath, goCWDPath, activeFixStdoutPath, activeFixStderrPath, activeFixExitPath)
	if err != nil {
		return err
	}
	goEvidence.Conformant = goEvidence.Conformant && goEvidence.SubjectHeadSHA == observation.HeadSHA && goEvidence.ObjectFormat == observation.ObjectFormat
	if ciEvidencePath == "" {
		return fmt.Errorf("ci-evidence is required")
	}
	ciEvidenceRaw, err := os.ReadFile(ciEvidencePath)
	if err != nil {
		return err
	}
	var ciEvidence causalci.CIEvidenceAdjudication
	ciDecoder := json.NewDecoder(strings.NewReader(string(ciEvidenceRaw)))
	ciDecoder.DisallowUnknownFields()
	if err := ciDecoder.Decode(&ciEvidence); err != nil {
		return fmt.Errorf("decode CI evidence adjudication: %w", err)
	}
	var trailingCIEvidence any
	if decodeErr := ciDecoder.Decode(&trailingCIEvidence); decodeErr != io.EOF {
		return fmt.Errorf("CI evidence adjudication has trailing data")
	}
	if err := causalci.ValidateCIEvidenceAdjudication(ciEvidence); err != nil {
		return err
	}
	if ciEvidence.Schema != causalci.CIEvidenceAdjudicationSchema || ciEvidence.Scope != causalci.CIEvidenceAdjudicationScope || ciEvidence.ActualRootCauseDenominator != 1 || ciEvidence.ActualResolution != causalci.CIEvidenceResolutionLowered || ciEvidence.ActualCoordinate.Stage != "proposal-promotion" || ciEvidence.ActualCoordinate.Step != "fetch-github-evidence" || (ciEvidence.ActualCoordinate.Reason != causalci.CIEvidenceReasonPaginationIncomplete && ciEvidence.ActualCoordinate.Reason != causalci.CIEvidenceReasonCurrentUnavailable) || ciEvidence.CurrentPermissionNumerator != 0 || ciEvidence.CurrentPermissionDenominator != 1 {
		return fmt.Errorf("invalid CI causal evidence adjudication")
	}
	decision := causalci.ConformanceFailClosed
	coordinate := causalci.Coordinate{Stage: "ADJUDICATION", Step: "compare-process-observations", Reason: "PROCESS_OBSERVATION_NOT_CONFORMANT"}
	if sourceReconstructionNumerator == len(ids) && sourceBindingNumerator == len(ids) && sameIDs(ids, observed) && goEvidence.Conformant {
		decision = causalci.ConformancePass
		coordinate = causalci.Coordinate{Stage: "ADJUDICATION", Step: "compare-process-observations", Reason: "PROCESS_OBSERVATION_AND_PLAN_RECONSTRUCTION_OBSERVED"}
	}
	receipt := causalci.AdjudicationReceipt{Schema: causalci.PlanAdjudicationSchema, Scope: causalci.PlanAdjudicationScope, ObservationDigest: bytesDigest(input), ExpectedVariantIDs: ids, ObservedVariantIDs: observed, Adjudications: adjudications, ProcessEvidence: processEvidence, SourcePlanBinding: sourcePlanBinding, SourcePlanBindingNumer: sourceBindingNumerator, SourcePlanBindingDenom: len(ids), SourceReconstructionNumer: sourceReconstructionNumerator, SourceReconstructionDenom: len(ids), Decision: decision, Coordinate: coordinate, SelectedCheckExecutionSchema: causalci.SelectedCheckExecutionSchema, SelectedCheckExecution: causalci.ExecutionUnknown, SelectedCheckExecutionNumer: 0, SelectedCheckExecutionDenom: 1, GoRuntime: goEvidence, CIEvidence: ciEvidence}
	receipt.Digest, err = jsonDigestWithoutAdjudicationDigest(receipt)
	if err != nil {
		return err
	}
	if err := writeJSON(outputPath, receipt); err != nil {
		return err
	}
	fmt.Printf("adjudication: decision=%s source_plan_binding=%d/%d process_observation=%d/%d selected_check_execution=%s/%d/%d go_runtime_conformant=%t\n", receipt.Decision, receipt.SourcePlanBindingNumer, receipt.SourcePlanBindingDenom, receipt.SourceReconstructionNumer, receipt.SourceReconstructionDenom, receipt.SelectedCheckExecution, receipt.SelectedCheckExecutionNumer, receipt.SelectedCheckExecutionDenom, receipt.GoRuntime.Conformant)
	return nil
}

func processObservationConformant(value causalci.ConsumerAdjudication, osExit int, stderr []byte) bool {
	return value.ExitCode == 0 && value.Result == causalci.ExecutionPass && osExit == value.ExitCode && len(stderr) == 0 && value.ConsumerIdentity == "gooo://consumer/causal-ci-selection/process"
}

func parseSourceFiles(raw string) (map[string]string, error) {
	result := map[string]string{}
	if raw == "" {
		return result, fmt.Errorf("source-files is required")
	}
	for _, item := range strings.Split(raw, ",") {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("malformed source mapping %q", item)
		}
		if _, exists := result[parts[0]]; exists {
			return nil, fmt.Errorf("duplicate source mapping %q", parts[0])
		}
		result[parts[0]] = parts[1]
	}
	return result, nil
}

func readProcessEvidence(processDir, id string) (int, []byte, []byte, error) {
	if processDir == "" {
		return 0, nil, nil, fmt.Errorf("process-dir is required")
	}
	exitRaw, err := os.ReadFile(filepath.Join(processDir, id+"-os-exit.txt"))
	if err != nil {
		return 0, nil, nil, err
	}
	exitCode, err := strconv.Atoi(strings.TrimSpace(string(exitRaw)))
	if err != nil {
		return 0, nil, nil, err
	}
	stdout, err := os.ReadFile(filepath.Join(processDir, id+".stdout"))
	if err != nil {
		return 0, nil, nil, err
	}
	stderr, err := os.ReadFile(filepath.Join(processDir, id+".stderr"))
	if err != nil {
		return 0, nil, nil, err
	}
	return exitCode, stdout, stderr, nil
}

func readGoEvidence(versionPath, envPath, helpPath, helpStderrPath, stdoutPath, stderrPath, exitPath, subjectHeadPath, subjectTreePath, objectFormatPath, goModDigestPath, goSumDigestPath, goListPath, goCWDPath, activeFixStdoutPath, activeFixStderrPath, activeFixExitPath string) (causalci.GoRuntimeEvidence, error) {
	paths := []string{versionPath, envPath, helpPath, helpStderrPath, stdoutPath, stderrPath, exitPath, subjectHeadPath, subjectTreePath, objectFormatPath, goModDigestPath, goSumDigestPath, goListPath, goCWDPath, activeFixStdoutPath, activeFixStderrPath, activeFixExitPath}
	for _, path := range paths {
		if path == "" {
			return causalci.GoRuntimeEvidence{}, fmt.Errorf("complete Go 1.27 evidence is required")
		}
	}
	version, err := os.ReadFile(versionPath)
	if err != nil {
		return causalci.GoRuntimeEvidence{}, err
	}
	env, err := os.ReadFile(envPath)
	if err != nil {
		return causalci.GoRuntimeEvidence{}, err
	}
	help, err := os.ReadFile(helpPath)
	if err != nil {
		return causalci.GoRuntimeEvidence{}, err
	}
	helpStderr, err := os.ReadFile(helpStderrPath)
	if err != nil {
		return causalci.GoRuntimeEvidence{}, err
	}
	stdout, err := os.ReadFile(stdoutPath)
	if err != nil {
		return causalci.GoRuntimeEvidence{}, err
	}
	stderr, err := os.ReadFile(stderrPath)
	if err != nil {
		return causalci.GoRuntimeEvidence{}, err
	}
	exitRaw, err := os.ReadFile(exitPath)
	if err != nil {
		return causalci.GoRuntimeEvidence{}, err
	}
	exitCode, err := strconv.Atoi(strings.TrimSpace(string(exitRaw)))
	if err != nil {
		return causalci.GoRuntimeEvidence{}, err
	}
	activeStdout, err := os.ReadFile(activeFixStdoutPath)
	if err != nil {
		return causalci.GoRuntimeEvidence{}, err
	}
	activeStderr, err := os.ReadFile(activeFixStderrPath)
	if err != nil {
		return causalci.GoRuntimeEvidence{}, err
	}
	activeExitRaw, err := os.ReadFile(activeFixExitPath)
	if err != nil {
		return causalci.GoRuntimeEvidence{}, err
	}
	activeExit, err := strconv.Atoi(strings.TrimSpace(string(activeExitRaw)))
	if err != nil {
		return causalci.GoRuntimeEvidence{}, err
	}
	readTrimmed := func(path string) (string, error) {
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			return "", readErr
		}
		return strings.TrimSpace(string(raw)), nil
	}
	subjectHead, err := readTrimmed(subjectHeadPath)
	if err != nil {
		return causalci.GoRuntimeEvidence{}, err
	}
	subjectTree, err := readTrimmed(subjectTreePath)
	if err != nil {
		return causalci.GoRuntimeEvidence{}, err
	}
	objectFormat, err := readTrimmed(objectFormatPath)
	if err != nil {
		return causalci.GoRuntimeEvidence{}, err
	}
	goModDigest, err := readTrimmed(goModDigestPath)
	if err != nil {
		return causalci.GoRuntimeEvidence{}, err
	}
	goSumDigest, err := readTrimmed(goSumDigestPath)
	if err != nil {
		return causalci.GoRuntimeEvidence{}, err
	}
	commandCWD, err := readTrimmed(goCWDPath)
	if err != nil {
		return causalci.GoRuntimeEvidence{}, err
	}
	goListRaw, err := os.ReadFile(goListPath)
	if err != nil {
		return causalci.GoRuntimeEvidence{}, err
	}
	packages := make([]string, 0)
	for _, line := range strings.Split(string(goListRaw), "\n") {
		if value := strings.TrimSpace(line); value != "" {
			packages = append(packages, value)
		}
	}
	packages = sortedUnique(packages)
	packageDigest := mustJSONDigest(packages)
	required := []string{"atomictypes", "embedlit", "slicesbackward", "unsafefuncs", "waitgroupgo"}
	inventory := parseFixerInventory(string(help), required)
	inventorySet := map[string]struct{}{}
	for _, id := range inventory {
		inventorySet[id] = struct{}{}
	}
	satisfied := 0
	for _, id := range required {
		if _, exists := inventorySet[id]; exists {
			satisfied++
		}
	}
	removedID := "fmtappendf"
	_, removedPresent := inventorySet[removedID]
	activeObservedDiff := activeExit != 0 && len(activeStdout) > 0
	activeConformant := activeObservedDiff && len(activeStderr) == 0
	fixStderrAllowed := len(stderr) == 0
	conformant := exactGoVersion(string(version)) && strings.TrimSpace(string(env)) == "go1.27.0" && len(helpStderr) == 0 && satisfied == len(required) && !removedPresent && exitCode == 0 && len(stdout) == 0 && fixStderrAllowed && subjectHead != "" && subjectTree != "" && (objectFormat == "sha1" || objectFormat == "sha256") && goModDigest != "" && goSumDigest != "" && commandCWD != "" && len(packages) > 0 && activeConformant
	return causalci.GoRuntimeEvidence{ExpectedVersion: "go1.27.0", GoVersion: strings.TrimSpace(string(version)), GoEnvGOVERSION: strings.TrimSpace(string(env)), FixerInventory: inventory, FixHelpDigest: bytesDigest(help), FixHelpStderrDigest: bytesDigest(helpStderr), FixHelpStderrBytes: len(helpStderr), FixDiffStdoutDigest: bytesDigest(stdout), FixDiffStderrDigest: bytesDigest(stderr), FixDiffStdoutBytes: len(stdout), FixDiffStderrBytes: len(stderr), FixDiffStderrAllowed: fixStderrAllowed, FixDiffExitCode: exitCode, FixHelpCommandArgv: []string{"go", "tool", "fix", "help"}, FixDiffCommandArgv: []string{"go", "fix", "-diff", "./..."}, CommandCWD: commandCWD, SubjectHeadSHA: subjectHead, SubjectTreeID: subjectTree, ObjectFormat: objectFormat, GoModDigest: goModDigest, GoSumDigest: goSumDigest, PackageUniverseCount: len(packages), PackageUniverseNumerator: len(packages), PackageUniverseDenominator: len(packages), PackageUniverseDigest: packageDigest, PackageListCommandArgv: []string{"go", "list", "./..."}, RequiredFixers: required, RequiredFixersSatisfied: satisfied, RequiredFixersDenominator: len(required), RemovedFixerID: removedID, RemovedFixerPresent: removedPresent, RemovedFixerNumerator: boolInt(removedPresent), RemovedFixerDenominator: 1, ActiveFixFixture: causalci.FixFixtureEvidence{CommandArgv: []string{"go", "fix", "-diff", "./internal/meta/causalci/testdata/causal-ci-selection-fix-fixture"}, ExitCode: activeExit, StdoutDigest: bytesDigest(activeStdout), StdoutBytes: len(activeStdout), StderrDigest: bytesDigest(activeStderr), StderrBytes: len(activeStderr), ExpectedDiff: true, ObservedDiff: activeObservedDiff, Numerator: boolInt(activeObservedDiff), Denominator: 1, Conformant: activeConformant}, Conformant: conformant}, nil
}

func exactGoVersion(raw string) bool {
	fields := strings.Fields(raw)
	return len(fields) >= 4 && fields[0] == "go" && fields[1] == "version" && fields[2] == "go1.27.0" && fields[3] != ""
}

func parseFixerInventory(help string, required []string) []string {
	seen := map[string]struct{}{}
	for _, line := range strings.Split(help, "\n") {
		if !strings.HasPrefix(line, "    ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		id := strings.Trim(fields[0], "`'\"(),:;[]")
		if isFixerID(id) {
			seen[id] = struct{}{}
		}
	}
	// Keep the required list in the function contract: it is the policy's
	// gate, while the returned inventory remains the complete parsed catalog.
	_ = required
	result := make([]string, 0, len(seen))
	for id := range seen {
		result = append(result, id)
	}
	return sortedUnique(result)
}

func isFixerID(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') {
			return false
		}
	}
	return true
}

func sortedUnique(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	unique := result[:0]
	for _, value := range result {
		if value == "" || (len(unique) > 0 && unique[len(unique)-1] == value) {
			continue
		}
		unique = append(unique, value)
	}
	return unique
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func readInputs(inputPath, sourcePath string) ([]byte, []byte, error) {
	if inputPath == "" || sourcePath == "" {
		return nil, nil, fmt.Errorf("input and source are required")
	}
	input, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, nil, err
	}
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return nil, nil, err
	}
	return input, source, nil
}

func writeJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeBytes(path, append(raw, '\n'))
}

func writeBytes(path string, raw []byte) error {
	if path == "" {
		return fmt.Errorf("output is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func commandOutput(name string, args ...string) (string, error) {
	raw, err := exec.Command(name, args...).Output()
	if err != nil {
		return "", fmt.Errorf("%s: %w", name, err)
	}
	return string(raw), nil
}
func commandOutputIgnoreError(name string, args ...string) string {
	raw, _ := exec.Command(name, args...).Output()
	return string(raw)
}
func mustJSONDigest(value any) string { raw, _ := json.Marshal(value); return bytesDigest(raw) }
func jsonDigestWithoutDigest(value causalci.InterventionReport) (string, error) {
	value.Digest = ""
	return jsonDigest(value)
}
func jsonDigestWithoutAdjudicationDigest(value causalci.AdjudicationReceipt) (string, error) {
	value.Digest = ""
	return jsonDigest(value)
}
func digestConsumerAdjudication(value causalci.ConsumerAdjudication) string {
	value.Digest = ""
	raw, _ := json.Marshal(value)
	return bytesDigest(raw)
}
func jsonDigest(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return bytesDigest(raw), nil
}
func bytesDigest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func sameIDs(expected, observed []string) bool {
	if len(expected) != len(observed) {
		return false
	}
	left, right := append([]string(nil), expected...), append([]string(nil), observed...)
	sort.Strings(left)
	sort.Strings(right)
	for i := range left {
		if left[i] == "" || left[i] != right[i] || (i > 0 && left[i] == left[i-1]) {
			return false
		}
	}
	return true
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
