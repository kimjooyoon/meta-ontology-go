package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/causalci"
)

func main() {
	mode := flag.String("mode", "produce", "snapshot, observe, produce, interventions, or adjudicate")
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
	case "adjudicate":
		err = adjudicate(*scope, *inputPath, *plansDir, *adjudicationDir, *goVersionPath, *goEnvPath, *fixHelpPath, *fixHelpStderrPath, *fixStdoutPath, *fixStderrPath, *fixExitPath, *outputPath)
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
	entries := make([]causalci.RepositoryEntry, 0, len(tracked)+len(untracked))
	seen := map[string]bool{}
	for _, path := range tracked {
		entries = append(entries, repositoryEntry(path, true))
		seen[path] = true
	}
	for _, path := range untracked {
		if !seen[path] {
			entries = append(entries, repositoryEntry(path, false))
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return writeJSON(outputPath, causalci.RepositorySnapshot{Entries: entries, SnapshotDigest: mustJSONDigest(entries)})
}

func repositoryEntry(path string, tracked bool) causalci.RepositoryEntry {
	data, err := os.ReadFile(path)
	if err != nil {
		return causalci.RepositoryEntry{Path: path, Tracked: tracked, ContentDigest: "MISSING:" + err.Error()}
	}
	return causalci.RepositoryEntry{Path: path, Tracked: tracked, ContentDigest: bytesDigest(data)}
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
	value := causalci.Observation{Schema: causalci.ObservationSchema, Repository: "kimjooyoon/meta-ontology-go", BaseSHA: baseSHA, HeadSHA: headSHA, ObservedCheckoutSHA: checkout, SourcePath: sourcePath, HeadPathObjectID: headObject, SourceBytesDigest: bytesDigest(source), ChangedFiles: changed, PriorClaims: claims, Isolation: causalci.IsolationObservation{Before: before, After: after}}
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
		results = append(results, causalci.InterventionResult{ID: variant.id, Source: receipt.Source, Conformance: receipt.Conformance, Subjects: receipt.Subjects, ClaimTransitions: receipt.ClaimTransitions, Execution: receipt.Execution, PlanDigest: receipt.PlanDigest})
		observed = append(observed, variant.id)
	}
	report := causalci.InterventionReport{Schema: causalci.ReportSchema, ObservationDigest: bytesDigest(input), ExpectedVariantIDs: []string{"base", "semantic", "nonsemantic", "contradiction"}, ObservedVariantIDs: observed, Base: results[0], Semantic: results[1], Nonsemantic: results[2], Contradiction: results[3]}
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

func adjudicate(scope, inputPath, plansDir, adjudicationDir, goVersionPath, goEnvPath, fixHelpPath, fixHelpStderrPath, fixStdoutPath, fixStderrPath, fixExitPath, outputPath string) error {
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
	adjudications := make([]causalci.ConsumerAdjudication, 0, len(ids))
	observed := make([]string, 0, len(ids))
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
		if value.Schema != causalci.AdjudicationSchema || value.VariantID != id || value.PlanReceiptDigest != plan.Digest || value.InputDigest != bytesDigest(input) || value.OutputDigest == "" || value.ConsumerIdentity == "" || digestConsumerAdjudication(value) != value.Digest {
			return fmt.Errorf("invalid adjudication evidence for %s", id)
		}
		adjudications = append(adjudications, value)
		observed = append(observed, id)
	}
	goEvidence, err := readGoEvidence(goVersionPath, goEnvPath, fixHelpPath, fixHelpStderrPath, fixStdoutPath, fixStderrPath, fixExitPath)
	if err != nil {
		return err
	}
	passCount := 0
	for _, value := range adjudications {
		if value.ExitCode == 0 && value.Result == causalci.ExecutionPass {
			passCount++
		}
	}
	decision := causalci.ConformanceFailClosed
	coordinate := causalci.Coordinate{Stage: "ADJUDICATION", Step: "compare-consumer-results", Reason: "CONSUMER_EXECUTION_NOT_COMPLETE"}
	if passCount == len(ids) && sameIDs(ids, observed) && goEvidence.Conformant {
		decision = causalci.ConformancePass
		coordinate = causalci.Coordinate{Stage: "ADJUDICATION", Step: "compare-consumer-results", Reason: "INDEPENDENT_CONSUMER_EXECUTION_OBSERVED"}
	}
	receipt := causalci.AdjudicationReceipt{Schema: causalci.AdjudicationSchema, Scope: causalci.ReceiptScope, ObservationDigest: bytesDigest(input), ExpectedVariantIDs: ids, ObservedVariantIDs: observed, Adjudications: adjudications, SourceReconstructionNumer: passCount, SourceReconstructionDenom: len(ids), Decision: decision, Coordinate: coordinate, GoRuntime: goEvidence}
	receipt.Digest, err = jsonDigestWithoutAdjudicationDigest(receipt)
	if err != nil {
		return err
	}
	if err := writeJSON(outputPath, receipt); err != nil {
		return err
	}
	fmt.Printf("adjudication: decision=%s source_reconstruction=%d/%d go_runtime_conformant=%t\n", receipt.Decision, receipt.SourceReconstructionNumer, receipt.SourceReconstructionDenom, receipt.GoRuntime.Conformant)
	return nil
}

func readGoEvidence(versionPath, envPath, helpPath, helpStderrPath, stdoutPath, stderrPath, exitPath string) (causalci.GoRuntimeEvidence, error) {
	paths := []string{versionPath, envPath, helpPath, helpStderrPath, stdoutPath, stderrPath, exitPath}
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
	inventory := make([]string, 0)
	for _, line := range strings.Split(strings.TrimSpace(string(help)), "\n") {
		if strings.TrimSpace(line) != "" {
			inventory = append(inventory, strings.TrimSpace(line))
		}
	}
	return causalci.GoRuntimeEvidence{ExpectedVersion: "go1.27.0", GoVersion: strings.TrimSpace(string(version)), GoEnvGOVERSION: strings.TrimSpace(string(env)), FixerInventory: inventory, FixHelpDigest: bytesDigest(help), FixHelpStderrDigest: bytesDigest(helpStderr), FixHelpStderrBytes: len(helpStderr), FixDiffStdoutDigest: bytesDigest(stdout), FixDiffStderrDigest: bytesDigest(stderr), FixDiffStdoutBytes: len(stdout), FixDiffStderrBytes: len(stderr), FixDiffExitCode: exitCode, Conformant: strings.Contains(string(version), "go1.27.0") && strings.TrimSpace(string(env)) == "go1.27.0" && exitCode == 0 && len(stdout) == 0}, nil
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
