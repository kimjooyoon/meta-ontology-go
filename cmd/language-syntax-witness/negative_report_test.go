package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func TestPersistValidatedNegativeReportKeepsFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.json")
	report := languagesyntax.Report{Decision: languagesyntax.DecisionClosed, Reason: "SYNTAX_ROUNDTRIP_MISMATCH"}
	negative := validatedNegativeReportError{decision: report.Decision, reason: report.Reason}
	var stdout bytes.Buffer
	err := persistBuiltReport(path, report, negative, &stdout)
	if err != negative || stdout.Len() != 0 {
		t.Fatalf("negative result became success: err=%v stdout=%s", err, &stdout)
	}
	raw, readErr := os.ReadFile(path)
	if readErr != nil || !bytes.Contains(raw, []byte(report.Reason)) {
		t.Fatalf("negative evidence missing: %s err=%v", raw, readErr)
	}
}

func TestInvalidReportDoesNotReplaceEvidence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.json")
	if err := os.WriteFile(path, []byte("previous"), 0600); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("validation failed")
	var stdout bytes.Buffer
	if err := persistBuiltReport(path, languagesyntax.Report{}, failure, &stdout); err != failure {
		t.Fatalf("validation failure changed: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != "previous" || stdout.Len() != 0 {
		t.Fatalf("invalid evidence wrote output: %s err=%v", raw, err)
	}
}

func TestNegativeReportWriteFailureStaysFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "report.json")
	var stdout bytes.Buffer
	negative := validatedNegativeReportError{decision: "FAIL_CLOSED", reason: "fixture"}
	if err := persistBuiltReport(path, languagesyntax.Report{}, negative, &stdout); err == nil || err == negative || stdout.Len() != 0 {
		t.Fatalf("write failure hidden: err=%v stdout=%s", err, &stdout)
	}
}

func TestSuccessfulPersistenceBranchSerializesExactly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.json")
	report := languagesyntax.Report{Decision: languagesyntax.DecisionPass}
	var stdout bytes.Buffer
	if err := persistBuiltReport(path, report, nil, &stdout); err != nil {
		t.Fatal(err)
	}
	expected, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(raw, append(expected, '\n')) {
		t.Fatalf("successful persistence changed bytes: %s err=%v", raw, err)
	}
	if !strings.Contains(stdout.String(), "decision=PASS") {
		t.Fatalf("successful persistence lost its summary: %s", &stdout)
	}
}

func TestNativeSyntaxNegativeReportCannotBecomeAcceptance(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("public CLI execution belongs to native CI")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	binary := nativeSyntaxBinary(t, ctx)
	t.Run("source-bound-counterexample", func(t *testing.T) {
		nativeBoundSyntaxCounterexample(t, ctx, binary)
	})
	cfg := nativeSyntaxFixture(t)
	producer := nativeSyntaxFailure(t, ctx, binary, cfg, "SYNTAX_ROUNDTRIP_EVIDENCE_UNKNOWN")
	report, raw := nativeSyntaxNegativeReport(t, cfg)
	consumerCfg := cfg
	consumerCfg.output, consumerCfg.check = "", cfg.output
	consumer := nativeSyntaxFailure(t, ctx, binary, consumerCfg, "SYNTAX_ROUNDTRIP_EVIDENCE_UNKNOWN")
	after, err := os.ReadFile(cfg.output)
	if err != nil || !bytes.Equal(raw, after) {
		t.Fatalf("consumer changed negative evidence: %v", err)
	}
	invalidCfg := cfg
	invalidCfg.head = "not-a-commit"
	invalid := nativeSyntaxFailure(t, ctx, binary, invalidCfg, "language syntax report identity mismatch")
	after, err = os.ReadFile(cfg.output)
	if err != nil || !bytes.Equal(raw, after) {
		t.Fatalf("invalid report replaced previous evidence: %v", err)
	}
	outputCfg := cfg
	outputCfg.output = filepath.Dir(cfg.output)
	output := nativeSyntaxFailure(t, ctx, binary, outputCfg, "")
	sum := sha256.Sum256(raw)
	witness := map[string]any{"kind": "SYNTHETIC_UNAVAILABLE_SYNTAX_INPUT"}
	witness["supplied_head_is_synthetic"] = true
	witness["source_meta_binding"] = "UNAVAILABLE_NOT_ASSERTED"
	witness["producer_exit_code"] = producer
	witness["consumer_exit_code"] = consumer
	witness["invalid_report_exit_code"] = invalid
	witness["non_file_output_exit_code"] = output
	witness["report_sha256"] = "sha256:" + hex.EncodeToString(sum[:])
	witness["report"] = json.RawMessage(raw)
	witness["resolution"] = report.Resolution
	encoded, err := json.Marshal(witness)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("SYNTAX_NEGATIVE_NATIVE_WITNESS=%s", encoded)
}

func nativeSyntaxBinary(t *testing.T, ctx context.Context) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "language-syntax-witness")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	command := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build public syntax CLI: %v: %s", err, output)
	}
	return binary
}

func nativeSyntaxFixture(t *testing.T) config {
	t.Helper()
	work := t.TempDir()
	cfg := config{root: filepath.Join(work, "project"), head: strings.Repeat("a", 40),
		registry: filepath.Join(work, "registry.json"),
		concept:  filepath.Join(work, "concept.json"),
		output:   filepath.Join(work, "negative-report.json")}
	if err := os.Mkdir(cfg.root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfg.registry, []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	artifact := languageconcept.Artifact{
		ArtifactDigest: "sha256:" + strings.Repeat("1", 64),
		CatalogDigest:  "sha256:" + strings.Repeat("2", 64)}
	raw, err := json.Marshal(artifact)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfg.concept, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return cfg
}

func nativeSyntaxFailure(t *testing.T, ctx context.Context, binary string, cfg config, want string) int {
	t.Helper()
	args := []string{"-root", cfg.root, "-head", cfg.head, "-registry", cfg.registry,
		"-concept-artifact", cfg.concept}
	if cfg.check != "" {
		args = append(args, "-check", cfg.check)
	} else {
		args = append(args, "-output", cfg.output)
	}
	command := exec.CommandContext(ctx, binary, args...)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 {
		t.Fatalf("expected public CLI exit 1, got %v: %s", err, &stderr)
	}
	if stdout.Len() != 0 || stderr.Len() == 0 || !strings.Contains(stderr.String(), want) {
		t.Fatalf("failure was hidden or changed: stdout=%s stderr=%s", &stdout, &stderr)
	}
	return exit.ExitCode()
}

func nativeSyntaxNegativeReport(t *testing.T, cfg config) (languagesyntax.Report, []byte) {
	t.Helper()
	raw, err := os.ReadFile(cfg.output)
	if err != nil {
		t.Fatal(err)
	}
	var report languagesyntax.Report
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if err := languagesyntax.Validate(report, cfg.head); err != nil {
		t.Fatalf("saved failure is not validated report evidence: %v", err)
	}
	if report.Decision != languagesyntax.DecisionClosed || report.Resolution != languagesyntax.ResolutionLower ||
		report.Reason != "SYNTAX_ROUNDTRIP_EVIDENCE_UNKNOWN" {
		t.Fatalf("unavailable input changed classification: %+v", report)
	}
	if report.Source.ConceptBound || report.Source.ObservationKnown || report.RepositoryWrites != 0 ||
		report.MutationAuthorized || report.Summary.Unresolved != languagesyntax.FixedTotal {
		t.Fatalf("unavailable input invented evidence or effects: %+v", report)
	}
	registrySum := sha256.Sum256([]byte("{"))
	if report.Source.RegistryDigest != "sha256:"+hex.EncodeToString(registrySum[:]) {
		t.Fatal("saved report lost its actual malformed registry identity")
	}
	for _, item := range report.Cases {
		if item.Status != "UNRESOLVED" || item.Evidence.ObservedDecision != "UNKNOWN" {
			t.Fatalf("unavailable case became resolved: %+v", item)
		}
	}
	return report, raw
}
