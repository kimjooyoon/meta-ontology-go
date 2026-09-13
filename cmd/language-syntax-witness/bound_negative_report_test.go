package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

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
