package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
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

// Capture the actual read closure, rather than guessing which catalog files matter.
type nativeSyntaxInputs struct {
	fs.FS
	opened map[string]bool
}

func (inputs nativeSyntaxInputs) Open(name string) (fs.File, error) {
	file, err := inputs.FS.Open(name)
	if err == nil {
		inputs.opened[name] = true
	}
	return file, err
}

func nativeBoundSyntaxFixture(t *testing.T) (config, languagesyntax.Report) {
	t.Helper()
	inputs := nativeSyntaxInputs{FS: os.DirFS("../.."), opened: map[string]bool{}}
	registry, err := fs.ReadFile(inputs, "examples/language-syntax-roundtrip/corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	artifact := languageconcept.BuildArtifact(inputs)
	head := strings.Repeat("0", 40)
	baseline := languagesyntax.Evaluate(inputs, head, registry, artifact)
	if err := languagesyntax.Validate(baseline, head); err != nil {
		t.Fatal(err)
	}
	if baseline.Decision != languagesyntax.DecisionPass || !baseline.Source.ConceptBound ||
		!baseline.Source.ObservationKnown || baseline.Summary.Satisfied != languagesyntax.FixedTotal {
		t.Fatalf("original corpus is not a bound positive control: %+v", baseline)
	}
	work := t.TempDir()
	cfg := config{root: filepath.Join(work, "project"), head: head,
		registry: filepath.Join(work, "registry.json"),
		concept:  filepath.Join(work, "concept.json"),
		output:   filepath.Join(work, "report.json")}
	nativeCopySyntaxInputs(t, inputs, cfg.root)
	if err := os.WriteFile(cfg.registry, registry, 0600); err != nil {
		t.Fatal(err)
	}
	nativeWriteSyntaxConcept(t, cfg, artifact)
	return cfg, baseline
}

func nativeCopySyntaxInputs(t *testing.T, inputs nativeSyntaxInputs, root string) {
	t.Helper()
	names := make([]string, 0, len(inputs.opened))
	for name := range inputs.opened {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		info, err := fs.Stat(inputs.FS, name)
		if err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(root, filepath.FromSlash(name))
		if info.IsDir() {
			if err := os.MkdirAll(target, 0700); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if !info.Mode().IsRegular() {
			t.Fatal(fmt.Errorf("unsupported syntax input: %s", name))
		}
		raw, err := fs.ReadFile(inputs.FS, name)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func nativeWriteSyntaxConcept(t *testing.T, cfg config, artifact languageconcept.Artifact) {
	t.Helper()
	raw, err := json.Marshal(artifact)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfg.concept, raw, 0600); err != nil {
		t.Fatal(err)
	}
}

func nativeReadBoundSyntaxReport(t *testing.T, cfg config) (languagesyntax.Report, []byte) {
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
		t.Fatal(err)
	}
	return report, raw
}

func nativeBoundSyntaxCounterexample(t *testing.T, ctx context.Context, binary string) {
	t.Helper()
	cfg, original := nativeBoundSyntaxFixture(t)
	var stdout bytes.Buffer
	if err := produce(cfg, &stdout); err != nil {
		t.Fatalf("copied read closure lost its positive control: %v", err)
	}
	baseline, _ := nativeReadBoundSyntaxReport(t, cfg)
	if !reflect.DeepEqual(original, baseline) {
		t.Fatal("copied read closure changed the original source-bound report")
	}
	target := filepath.Join(cfg.root, "examples/billing/main.gooo")
	before, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	invalid, err := os.ReadFile(filepath.Join(cfg.root, "examples/language-syntax-roundtrip/unknown-keyword.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, invalid, 0600); err != nil {
		t.Fatal(err)
	}
	nativeWriteSyntaxConcept(t, cfg, languageconcept.BuildArtifact(os.DirFS(cfg.root)))
	producer := nativeSyntaxFailure(t, ctx, binary, cfg, "SYNTAX_ROUNDTRIP_MISMATCH")
	report, first := nativeReadBoundSyntaxReport(t, cfg)
	nativeAssertBoundSyntaxMismatch(t, baseline, report)
	replay := nativeSyntaxFailure(t, ctx, binary, cfg, "SYNTAX_ROUNDTRIP_MISMATCH")
	_, repeated := nativeReadBoundSyntaxReport(t, cfg)
	if !bytes.Equal(first, repeated) {
		t.Fatal("bound syntax counterexample changed on producer replay")
	}
	consumerCfg := cfg
	consumerCfg.output, consumerCfg.check = "", cfg.output
	consumer := nativeSyntaxFailure(t, ctx, binary, consumerCfg, "SYNTAX_ROUNDTRIP_MISMATCH")
	_, consumed := nativeReadBoundSyntaxReport(t, cfg)
	if !bytes.Equal(first, consumed) {
		t.Fatal("consumer rewrote the bound syntax counterexample")
	}
	after, err := os.ReadFile(target)
	if err != nil || !bytes.Equal(after, invalid) {
		t.Fatalf("syntax observation rewrote the fixture: %v", err)
	}
	nativeLogBoundSyntaxWitness(t, report, first, before, after, []int{producer, replay, consumer})
}

func nativeAssertBoundSyntaxMismatch(t *testing.T, baseline, report languagesyntax.Report) {
	t.Helper()
	if report.Decision != languagesyntax.DecisionClosed || report.Resolution != languagesyntax.ResolutionExact ||
		report.Reason != "SYNTAX_ROUNDTRIP_MISMATCH" || !report.Source.ConceptBound ||
		!report.Source.ObservationKnown || report.RepositoryWrites != 0 || report.MutationAuthorized ||
		report.Summary.Total != languagesyntax.FixedTotal || report.Summary.NotSatisfied != 1 ||
		report.Summary.Satisfied != languagesyntax.FixedTotal-1 || report.Summary.Unresolved != 0 ||
		len(report.Source.UnregisteredGooo) != 0 || len(report.Source.MissingRegistered) != 0 {
		t.Fatalf("known counterexample was hidden or lost source binding: %+v", report)
	}
	if report.Source.RegistryDigest != baseline.Source.RegistryDigest ||
		report.Source.CorpusDigest == baseline.Source.CorpusDigest ||
		len(report.Cases) != len(baseline.Cases) {
		t.Fatal("counterexample lost registry continuity or actual changed-input identity")
	}
	failed := 0
	for index, item := range report.Cases {
		if !reflect.DeepEqual(item.Definition, baseline.Cases[index].Definition) {
			t.Fatal("counterexample changed the evaluator's case definitions")
		}
		if item.Status == "SATISFIED" {
			continue
		}
		failed++
		if item.Definition.ID != "billing" || item.Definition.Kind != languagesyntax.KindValid ||
			item.Definition.Path != "examples/billing/main.gooo" ||
			item.Definition.ExpectedDecision != languagesyntax.DecisionPass ||
			item.Status != "NOT_SATISFIED" || item.Evidence.ObservedDecision != languagesyntax.DecisionClosed {
			t.Fatalf("wrong counterexample attribution: %+v", item)
		}
		evidence, err := json.Marshal(item.Evidence)
		if err != nil || !bytes.Contains(evidence, []byte("parse.unexpected-token")) ||
			!bytes.Contains(evidence, []byte("examples/billing/main.gooo:4:1")) {
			t.Fatalf("counterexample lost its original location and diagnostic: %s (%v)", evidence, err)
		}
	}
	if failed != 1 {
		t.Fatalf("expected exactly one concrete counterexample, got %d", failed)
	}
}

func nativeLogBoundSyntaxWitness(t *testing.T, report languagesyntax.Report, raw, before, after []byte, exits []int) {
	t.Helper()
	witness := map[string]any{"kind": "SYNTHETIC_SOURCE_BOUND_SYNTAX_COUNTEREXAMPLE"}
	witness["supplied_head_is_synthetic"] = true
	witness["input_scope"] = "CAPTURED_READ_CLOSURE_WITH_POSITIVE_CONTROL"
	witness["modified_path"] = "examples/billing/main.gooo"
	witness["source_before_sha256"] = fmt.Sprintf("sha256:%x", sha256.Sum256(before))
	witness["source_after_sha256"] = fmt.Sprintf("sha256:%x", sha256.Sum256(after))
	witness["report_sha256"] = fmt.Sprintf("sha256:%x", sha256.Sum256(raw))
	witness["producer_replay_consumer_exit_codes"] = exits
	witness["registry_digest"] = report.Source.RegistryDigest
	witness["report"] = json.RawMessage(raw)
	encoded, err := json.Marshal(witness)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("SYNTAX_BOUND_NEGATIVE_NATIVE_WITNESS=%s", encoded)
}
