package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/operationprovenance"
	independent "github.com/kimjooyoon/meta-ontology-go/internal/meta/operationprovenance/verify"
)

func main() {
	mode := flag.String("mode", "build", "build, verify, or interventions")
	sourcePath := flag.String("source", "", "Gooo source")
	receiptPath := flag.String("receipt", "", "receipt JSON for verify")
	consumerSourcePath := flag.String("consumer-source", "", "independent consumer source")
	artifactRoot := flag.String("artifact-root", "", "raw lineage artifact root")
	workspaceRoot := flag.String("workspace-root", ".", "isolated repository workspace to observe")
	outPath := flag.String("out", "", "output JSON")
	flag.Parse()
	if *outPath == "" {
		fail("out is required")
	}

	var value any
	var err error
	switch *mode {
	case "build":
		source := readSource(*sourcePath)
		value, err = operationprovenance.BuildObservedWithArtifacts(source, resolveArtifactRoot(*artifactRoot, *workspaceRoot), *workspaceRoot)
	case "verify":
		if *receiptPath == "" {
			fail("receipt is required for verify")
		}
		payload := readFile(*receiptPath)
		var source []byte
		if *sourcePath != "" {
			source = readFile(*sourcePath)
		}
		var consumerSource []byte
		if *consumerSourcePath != "" {
			consumerSource = readPackageSource(*consumerSourcePath)
		}
		value, err = independent.Verify(payload, source, consumerSource, resolveArtifactRoot(*artifactRoot, *workspaceRoot))
	case "interventions":
		source := readSource(*sourcePath)
		value, err = runInterventions(source, resolveArtifactRoot(*artifactRoot, *workspaceRoot), *workspaceRoot)
	default:
		fail("unknown mode %q", *mode)
	}
	if err != nil {
		fail("%s: %v", *mode, err)
	}
	writeJSON(*outPath, value)
	fmt.Printf("%s: %s\n", *mode, *outPath)
}

func runInterventions(source []byte, artifactRoot, workspaceRoot string) (operationprovenance.InterventionReport, error) {
	marker := []byte("producer=producer:MOP-COHERENCE-001")
	if !bytes.Contains(source, marker) {
		return operationprovenance.InterventionReport{}, fmt.Errorf("semantic intervention marker was not found")
	}
	semanticSource := bytes.Replace(source, marker, []byte("producer=producer:SEMANTIC-INTERVENTION-COHERENCE-001"), 1)
	nonsemanticSource := append(append([]byte(nil), source...), []byte("\n// nonsemantic intervention: presentation changed only\n")...)
	base, err := operationprovenance.BuildObservedWithArtifacts(source, artifactRoot, workspaceRoot)
	if err != nil {
		return operationprovenance.InterventionReport{}, err
	}
	semanticReceipt, err := operationprovenance.BuildObservedWithArtifacts(semanticSource, artifactRoot, workspaceRoot)
	if err != nil {
		return operationprovenance.InterventionReport{}, fmt.Errorf("semantic intervention: %w", err)
	}
	nonsemanticReceipt, err := operationprovenance.BuildObservedWithArtifacts(nonsemanticSource, artifactRoot, workspaceRoot)
	if err != nil {
		return operationprovenance.InterventionReport{}, fmt.Errorf("nonsemantic intervention: %w", err)
	}
	baseResult := interventionResult("baseline", "baseline", base, base)
	semanticResult := interventionResult("semantic-producer-endpoint-change", "semantic", base, semanticReceipt)
	nonsemanticResult := interventionResult("comment-only", "nonsemantic", base, nonsemanticReceipt)
	semanticResult.Status = interventionStatus(semanticResult, true)
	nonsemanticResult.Status = interventionStatus(nonsemanticResult, false)
	noopIssue, err := operationprovenance.RejectNoopMutation(source, artifactRoot)
	if err != nil {
		return operationprovenance.InterventionReport{}, fmt.Errorf("no-op intervention: %w", err)
	}
	noopResult := operationprovenance.InterventionResult{ID: "noop-relation-removal", Kind: "no-op", ObservedFailure: true, FailureStage: noopIssue.Stage, FailureStep: noopIssue.Step, FailureReason: noopIssue.Reason, Status: "PASS"}
	result := operationprovenance.InterventionReport{Schema: "gooo/meta-operation-provenance-intervention/v2", Base: baseResult, Semantic: semanticResult, Nonsemantic: nonsemanticResult, Noop: noopResult}
	result.Digest = digestJSON(resultWithoutDigest(result))
	return result, nil
}

func interventionStatus(result operationprovenance.InterventionResult, semantic bool) string {
	if semantic && result.SemanticDigestChanged && (result.DecisionChanged || result.TransitionChanged) {
		return "PASS"
	}
	if !semantic && result.RawSourceDigestChanged && !result.SemanticDigestChanged && !result.DecisionChanged && !result.TransitionChanged {
		return "PASS"
	}
	return "FAIL"
}

func interventionResult(id, kind string, base, mutated operationprovenance.Receipt) operationprovenance.InterventionResult {
	decisionChanged := decisionFingerprint(base) != decisionFingerprint(mutated)
	transitionChanged := transitionFingerprint(base) != transitionFingerprint(mutated)
	return operationprovenance.InterventionResult{
		ID: id, Kind: kind, RawSourceDigest: mutated.SourceDigest,
		CanonicalSemanticDigest: mutated.CanonicalSemanticDigest, ReceiptDigest: mutated.Digest,
		DecisionFingerprint: decisionFingerprint(mutated), TransitionFingerprint: transitionFingerprint(mutated),
		RawSourceDigestChanged: mutated.SourceDigest != base.SourceDigest,
		SemanticDigestChanged:  mutated.CanonicalSemanticDigest != base.CanonicalSemanticDigest,
		DecisionChanged:        decisionChanged, TransitionChanged: transitionChanged,
	}
}

func decisionFingerprint(receipt operationprovenance.Receipt) string {
	var builder strings.Builder
	for _, scenario := range receipt.Scenarios {
		builder.WriteString(scenario.ID)
		builder.WriteByte('=')
		builder.WriteString(scenario.ConformanceDecision)
		for _, metric := range scenario.Metrics {
			builder.WriteByte('|')
			builder.WriteString(metric.ID)
			builder.WriteByte('=')
			builder.WriteString(metric.Decision)
		}
		builder.WriteByte('\n')
	}
	return digestString(builder.String())
}

func transitionFingerprint(receipt operationprovenance.Receipt) string {
	var builder strings.Builder
	for _, scenario := range receipt.Scenarios {
		builder.WriteString(scenario.ID)
		for _, metric := range scenario.Metrics {
			builder.WriteByte('|')
			builder.WriteString(metric.ID)
			builder.WriteByte('=')
			builder.WriteString(metric.Transition.Transition)
			builder.WriteByte(':')
			builder.WriteString(metric.Transition.NextClaim)
		}
		builder.WriteByte('\n')
	}
	return digestString(builder.String())
}

func resultWithoutDigest(result operationprovenance.InterventionReport) operationprovenance.InterventionReport {
	result.Digest = ""
	return result
}

func digestString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestJSON(value any) string {
	payload, err := json.Marshal(value)
	if err != nil {
		return digestString(fmt.Sprintf("marshal-error:%v", err))
	}
	return digestString(string(payload))
}

func readSource(path string) []byte {
	if path == "" {
		fail("source is required")
	}
	return readFile(path)
}

func resolveArtifactRoot(path, workspaceRoot string) string {
	if path != "" {
		return path
	}
	return filepath.Join(workspaceRoot, "examples", "meta-operation-provenance", "artifacts")
}

func readPackageSource(path string) []byte {
	info, err := os.Stat(path)
	if err != nil {
		fail("stat consumer source %s: %v", path, err)
	}
	if !info.IsDir() {
		return readFile(path)
	}
	entries := make([]string, 0)
	err = filepath.WalkDir(path, func(current string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			entries = append(entries, current)
		}
		return nil
	})
	if err != nil {
		fail("walk consumer source %s: %v", path, err)
	}
	sort.Strings(entries)
	var builder strings.Builder
	for _, entry := range entries {
		builder.WriteString("// FILE: ")
		builder.WriteString(entry)
		builder.WriteByte('\n')
		builder.Write(readFile(entry))
		builder.WriteByte('\n')
	}
	return []byte(builder.String())
}

func readFile(path string) []byte {
	payload, err := os.ReadFile(path)
	if err != nil {
		fail("read %s: %v", path, err)
	}
	return payload
}

func writeJSON(path string, value any) {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fail("encode output: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fail("create output directory: %v", err)
	}
	if err := os.WriteFile(path, append(payload, '\n'), 0o644); err != nil {
		fail("write output: %v", err)
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "meta-operation-provenance: "+format+"\n", args...)
	os.Exit(1)
}
