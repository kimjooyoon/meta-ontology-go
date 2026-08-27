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
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/causalci"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/causalci/consumer"
)

func main() {
	mode := flag.String("mode", "produce", "observe, produce, verify, or interventions")
	baseSHA := flag.String("base-sha", "", "base commit for changed-file observation")
	headSHA := flag.String("head-sha", "", "head commit for changed-file observation")
	sourcePath := flag.String("source", "", "logical Gooo policy source path")
	inputPath := flag.String("input", "", "raw observation JSON")
	outputPath := flag.String("output", "", "output artifact")
	receiptPath := flag.String("receipt", "", "receipt JSON for verification")
	beforeStatus := flag.String("before-status", "", "pre-isolation git status snapshot")
	afterStatus := flag.String("after-status", "", "post-isolation git status snapshot")
	priorClaims := flag.String("prior-claims", "", "raw prior claim observation")
	semanticSource := flag.String("semantic-source", "", "semantic intervention Gooo source")
	nonsemanticSource := flag.String("nonsemantic-source", "", "nonsemantic intervention Gooo source")
	contradictionSource := flag.String("contradiction-source", "", "contradiction intervention Gooo source")
	check := flag.Bool("check", false, "run the independent consumer before writing")
	flag.Parse()

	var err error
	switch *mode {
	case "observe":
		err = observe(*baseSHA, *headSHA, *sourcePath, *beforeStatus, *afterStatus, *priorClaims, *outputPath)
	case "produce":
		err = produce(*inputPath, *sourcePath, *outputPath, *check)
	case "verify":
		err = verify(*inputPath, *sourcePath, *receiptPath)
	case "interventions":
		err = interventions(*inputPath, *sourcePath, *semanticSource, *nonsemanticSource, *contradictionSource, *outputPath, *check)
	default:
		err = fmt.Errorf("unknown mode %q", *mode)
	}
	if err != nil {
		fail(err)
	}
}

func observe(baseSHA, headSHA, sourcePath, beforePath, afterPath, priorPath, outputPath string) error {
	if baseSHA == "" || headSHA == "" || sourcePath == "" || beforePath == "" || afterPath == "" || priorPath == "" || outputPath == "" {
		return fmt.Errorf("observe requires base-sha, head-sha, source, before-status, after-status, prior-claims, and output")
	}
	changed, err := changedFiles(baseSHA, headSHA)
	if err != nil {
		return err
	}
	before, err := snapshot(beforePath)
	if err != nil {
		return err
	}
	after, err := snapshot(afterPath)
	if err != nil {
		return err
	}
	prior, err := readPriorClaims(priorPath)
	if err != nil {
		return err
	}
	claims := make([]causalci.PriorClaimObservation, 0, len(changed)*len(prior))
	for _, file := range changed {
		for _, claim := range prior {
			claim.SubjectPath = file.Path
			claims = append(claims, claim)
		}
	}
	value := causalci.Observation{
		Schema: causalci.ObservationSchema, Repository: "kimjooyoon/meta-ontology-go", BaseSHA: baseSHA, HeadSHA: headSHA, SourcePath: sourcePath,
		ChangedFiles: changed, PriorClaims: claims, Isolation: causalci.IsolationObservation{Before: before, After: after},
	}
	return writeJSON(outputPath, value)
}

type priorClaimsObservation struct {
	Schema string `json:"schema"`
	Claims []struct {
		ClaimID    string `json:"claim_id"`
		State      string `json:"state"`
		Provenance string `json:"provenance"`
	} `json:"claims"`
}

func readPriorClaims(path string) ([]causalci.PriorClaimObservation, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read prior claim observation: %w", err)
	}
	var value priorClaimsObservation
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("decode prior claim observation: %w", err)
	}
	if value.Schema != "gooo/causal-ci-selection-prior-observation/v1" || len(value.Claims) == 0 {
		return nil, fmt.Errorf("malformed prior claim observation")
	}
	result := make([]causalci.PriorClaimObservation, 0, len(value.Claims))
	for _, claim := range value.Claims {
		if claim.ClaimID == "" || claim.State == "" || claim.Provenance == "" {
			return nil, fmt.Errorf("malformed prior claim row")
		}
		result = append(result, causalci.PriorClaimObservation{ClaimID: claim.ClaimID, State: claim.State, Provenance: claim.Provenance})
	}
	return result, nil
}

func changedFiles(baseSHA, headSHA string) ([]causalci.ChangedFileObservation, error) {
	command := exec.Command("git", "diff", "--name-status", "--no-renames", "--diff-filter=ACMRTUXB", baseSHA+"..."+headSHA)
	raw, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("collect PR changed files: %w", err)
	}
	lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []causalci.ChangedFileObservation{}, nil
	}
	result := make([]causalci.ChangedFileObservation, 0, len(lines))
	for _, line := range lines {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("malformed git changed-file line %q", line)
		}
		beforeObject := objectAt(baseSHA, parts[1])
		afterObject := objectAt(headSHA, parts[1])
		result = append(result, causalci.ChangedFileObservation{Path: parts[1], Status: parts[0], BeforeObject: beforeObject, AfterObject: afterObject})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result, nil
}

func objectAt(revision, path string) string {
	value, err := exec.Command("git", "rev-parse", revision+":"+path).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(value))
}

func snapshot(path string) (causalci.RepositorySnapshot, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return causalci.RepositorySnapshot{}, fmt.Errorf("read isolation snapshot: %w", err)
	}
	text := strings.TrimSuffix(string(raw), "\n")
	lines := []string{}
	if text != "" {
		lines = strings.Split(text, "\n")
	}
	digest, err := jsonDigest(lines)
	if err != nil {
		return causalci.RepositorySnapshot{}, err
	}
	return causalci.RepositorySnapshot{StatusLines: lines, StatusDigest: digest}, nil
}

func produce(inputPath, sourcePath, outputPath string, check bool) error {
	input, source, err := readInputs(inputPath, sourcePath)
	if err != nil {
		return err
	}
	receipt, err := causalci.Evaluate(input, sourcePath, source)
	if err != nil {
		return err
	}
	receiptRaw, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	receiptRaw = append(receiptRaw, '\n')
	if check {
		if err := consumer.Verify(input, sourcePath, source, receiptRaw); err != nil {
			return fmt.Errorf("independent consumer rejected receipt: %w", err)
		}
	}
	if err := writeBytes(outputPath, receiptRaw); err != nil {
		return err
	}
	fmt.Printf("causal CI plan: conformance=%s subjects=%d selected=%d unknown=%d fail_closed=%d checks=%d/%d transitions=%d discharged=%d lower_resolution=%d refuted=%d source_reconstruction=%d/%d repository_writes=%d mutation_authority=%t\n", receipt.Conformance.Decision, receipt.Metrics.SubjectTotal, receipt.Metrics.SelectedSubjectTotal, receipt.Metrics.UnknownSubjectTotal, receipt.Metrics.FailClosedSubjectTotal, receipt.Metrics.SelectedCheckTotal, receipt.Metrics.FullSuiteCheckDenominator, receipt.Metrics.ClaimTransitionTotal, receipt.Metrics.DischargedClaimTotal, receipt.Metrics.LowerResolutionClaimTotal, receipt.Metrics.RefutedClaimTotal, receipt.Metrics.SourceReconstructionNumer, receipt.Metrics.SourceReconstructionDenom, receipt.Operation.RepositoryWrites, receipt.Operation.MutationAuthority)
	return nil
}

func verify(inputPath, sourcePath, receiptPath string) error {
	input, source, err := readInputs(inputPath, sourcePath)
	if err != nil {
		return err
	}
	receipt, err := os.ReadFile(receiptPath)
	if err != nil {
		return err
	}
	if err := consumer.Verify(input, sourcePath, source, receipt); err != nil {
		return err
	}
	fmt.Println("independent consumer: PASS mode=INDEPENDENT_RECONSTRUCTION")
	return nil
}

func interventions(inputPath, sourcePath, semanticPath, nonsemanticPath, contradictionPath, outputPath string, check bool) error {
	input, baseSource, err := readInputs(inputPath, sourcePath)
	if err != nil {
		return err
	}
	paths := []struct {
		id   string
		path string
	}{
		{id: "base", path: sourcePath}, {id: "semantic", path: semanticPath}, {id: "nonsemantic", path: nonsemanticPath}, {id: "contradiction", path: contradictionPath},
	}
	results := make([]causalci.InterventionResult, 0, len(paths))
	for _, variant := range paths {
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
		receipt, evalErr := causalci.Evaluate(input, sourcePath, variantSource)
		if evalErr != nil {
			return fmt.Errorf("evaluate %s intervention: %w", variant.id, evalErr)
		}
		receiptRaw, err := json.MarshalIndent(receipt, "", "  ")
		if err != nil {
			return err
		}
		if check {
			if err := consumer.Verify(input, sourcePath, variantSource, append(receiptRaw, '\n')); err != nil {
				return fmt.Errorf("independent consumer rejected %s intervention: %w", variant.id, err)
			}
		}
		results = append(results, causalci.InterventionResult{ID: variant.id, Source: receipt.Source, Conformance: receipt.Conformance, Subjects: receipt.Subjects, ClaimTransitions: receipt.ClaimTransitions, PlanDigest: receipt.PlanDigest})
	}
	report := causalci.InterventionReport{Schema: causalci.ReportSchema, ObservationDigest: bytesDigest(input), Base: results[0], Semantic: results[1], Nonsemantic: results[2], Contradiction: results[3], SourceReconstructionNumer: len(results), SourceReconstructionDenom: len(results)}
	report.Digest, err = jsonDigestWithoutDigest(report)
	if err != nil {
		return err
	}
	if err := writeJSON(outputPath, report); err != nil {
		return err
	}
	fmt.Printf("interventions: source_reconstruction=%d/%d semantic_plan_changed=%t nonsemantic_plan_preserved=%t contradiction=%s\n", report.SourceReconstructionNumer, report.SourceReconstructionDenom, report.Base.PlanDigest != report.Semantic.PlanDigest, report.Base.PlanDigest == report.Nonsemantic.PlanDigest, report.Contradiction.Conformance.Decision)
	return nil
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

func jsonDigest(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return bytesDigest(raw), nil
}

func jsonDigestWithoutDigest(value causalci.InterventionReport) (string, error) {
	value.Digest = ""
	return jsonDigest(value)
}

func bytesDigest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
