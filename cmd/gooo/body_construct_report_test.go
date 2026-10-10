package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"html"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func constructionReportFixture() bodyConstructOutput {
	pass, fail := true, false
	r := bodyexecution.CompositionRuntime{Stage: "COMPLETE", FinitePassed: 1, FiniteTotal: 2,
		Traces: []bodyexecution.CompositionTrace{{CaseIndex: 0, Deliveries: []bodyexecution.CompositionDelivery{
			{ActivityID: "one", Actual: json.RawMessage(`9007199254740993`), Expected: json.RawMessage(`9007199254740993`), Passed: &pass},
			{ActivityID: "two", Actual: json.RawMessage(`9007199254740995`), Expected: json.RawMessage(`9007199254740993`), Passed: &fail},
		}}}}
	c := bodyexecution.JointConstruction{Stage: "COMPLETE", Decision: "PARTIAL_FINITE", StopReason: "PROGRAM_BUDGET",
		ProgramBudget: 2, CandidateSpace: "18446744073709551616", SelectedAttempt: 1,
		Attempts: []bodyexecution.JointAttempt{{Rejection: &bodyexecution.JointCandidateRejection{Reason: "invalid"}},
			{LocalPassed: 3, LocalTotal: 4, Runtime: r}}}
	return bodyConstructOutput{GeneratedNow: true, Construction: c,
		Evaluation: bodyexecution.JointEvaluation{Runtime: r, InputSeparation: bodyexecution.JointInputSeparation{
			UniqueInputs: 3, ConstructionInputs: 1, OtherInputs: 2, DuplicateRows: 1, Scope: "caller root tuples"}}}
}

func TestConstructionReportFormatsPreserveEvidence(t *testing.T) {
	r := constructionReportFixture()
	before, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	for _, format := range []string{"", "json", "text", "markdown"} {
		var out bytes.Buffer
		if err := writeBodyConstructResult(&out, r, format); err != nil {
			t.Fatal(err)
		}
		if format == "" || format == "json" {
			if !bytes.Equal(bytes.TrimSpace(out.Bytes()), before) {
				t.Fatal("JSON wire shape changed")
			}
			continue
		}
		visible := out.String()
		if format == "markdown" {
			visible = html.UnescapeString(visible)
		}
		for _, want := range []string{"PARTIAL_FINITE", "18446744073709551616", "3/4 (75.00%)", "1/2 (50.00%)",
			"1 mismatched", "0 unobserved", "1 consumed during caller construction", "2 other", "1 duplicate rows",
			"9007199254740995", "9007199254740993", "case index 0", "not assessed", "model-training exposure are unknown"} {
			if !strings.Contains(visible, want) {
				t.Fatalf("%s missing %q: %s", format, want, out.String())
			}
		}
	}
	after, _ := json.Marshal(r)
	if !bytes.Equal(before, after) {
		t.Fatal("rendering changed construction or evaluation")
	}
}

func TestConstructionReportMissingPartialAndFaultedObservations(t *testing.T) {
	r := constructionReportFixture().Evaluation.Runtime
	r.Stage, r.FiniteTotal = "FAILED", 6
	r.Traces[0].Deliveries = append(r.Traces[0].Deliveries,
		bodyexecution.CompositionDelivery{Expected: json.RawMessage(`1`), Fault: &bodyexecution.CompositionFault{Kind: "ZERO_DIVISOR"}},
		bodyexecution.CompositionDelivery{Expected: json.RawMessage(`2`), BlockedBy: []string{"producer"}},
		bodyexecution.CompositionDelivery{Fault: &bodyexecution.CompositionFault{Kind: "ZERO_DIVISOR"}})
	want := "1/6 (16.67%) matched; 1 mismatched, 1 faulted, 1 blocked, 2 unobserved; 1 additional unscored faults/blocks"
	if got := constructionRuntimeShare(r); got != want {
		t.Fatal(got)
	}
	for _, runtime := range []bodyexecution.CompositionRuntime{{}, {Stage: "TOOLCHAIN", FiniteTotal: 3}} {
		if got := constructionRuntimeShare(runtime); !strings.Contains(got, "not measured") || strings.Contains(got, "%") {
			t.Fatal("unobserved runtime presented as measured accuracy", got)
		}
	}
	if got := constructionShare(0, 4); got != "0/4 (0.00%)" {
		t.Fatal("observed zero lost", got)
	}
	r.FinitePassed = 2
	if !strings.Contains(constructionRuntimeShare(r), "inconsistent") {
		t.Fatal("inconsistent counts passed")
	}
}

func TestConstructionReportReplayFailureAndEscapedDetail(t *testing.T) {
	r := bodyConstructOutput{Construction: bodyexecution.JointConstruction{SelectedAttempt: -1, Failure: "a|b\n<script>"},
		Evaluation: bodyexecution.JointEvaluation{ReplayFailure: &bodyexecution.JointReplayFailure{AttemptIndex: 2,
			Stage: "RUNTIME", Runtime: bodyexecution.CompositionRuntime{Failure: "timed out"}}}}
	var text bytes.Buffer
	if err := writeBodyConstructResult(&text, r, "markdown"); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"saved construction replay", "none recorded", "not measured", "false", "attempt 3", "timed out", "&#124;", "&#10;", "&lt;script&gt;"} {
		if !strings.Contains(text.String(), want) {
			t.Fatal(want, text.String())
		}
	}
	if err := writeBodyConstructResult(constructionFailWriter{}, r, "text"); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal("writer error lost", err)
	}
}

type constructionFailWriter struct{}

func (constructionFailWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestConstructionReportFlagAndFailureExit(t *testing.T) {
	base := []string{"--source", "source", "--cases", "cases", "--construction", "saved"}
	for _, format := range []string{"json", "text", "markdown"} {
		if _, err := parseBodyConstruct(append(append([]string{}, base...), "--format", format)); err != nil {
			t.Fatal(err)
		}
	}
	for _, suffix := range [][]string{{"--format", "other"}, {"--format", "text", "--format", "json"}} {
		if _, err := parseBodyConstruct(append(append([]string{}, base...), suffix...)); err == nil {
			t.Fatal("invalid format accepted", suffix)
		}
	}
	fixture := "../../examples/caller-guided-construction/"
	args := []string{"body-construct", "--source", fixture + "main.gooo.fixture", "--entry", "Main",
		"--construction-cases", fixture + "construction-cases.json", "--cases", fixture + "evaluation-cases.json",
		"--attempts", "2", "--format", "text", "--go-bin", filepath.Join(t.TempDir(), "missing-go")}
	var out, diagnostic bytes.Buffer
	if code := run(args, &out, &diagnostic); code != exitFailure || diagnostic.Len() == 0 ||
		!strings.Contains(out.String(), "Gooo construction result") || !strings.Contains(out.String(), "Construction failure") {
		t.Fatal("formatted failure lost exit status or evidence", code, out.String(), diagnostic.String())
	}
}

func TestConstructionReportBoundsFailureDetails(t *testing.T) {
	r := constructionReportFixture()
	delivery := r.Evaluation.Runtime.Traces[0].Deliveries[1]
	delivery.Actual, _ = json.Marshal(strings.Repeat("a", 200))
	for range 10 {
		r.Evaluation.Runtime.Traces[0].Deliveries = append(r.Evaluation.Runtime.Traces[0].Deliveries, delivery)
	}
	var text bytes.Buffer
	if err := writeBodyConstructResult(&text, r, "text"); err != nil {
		t.Fatal(err)
	}
	if strings.Count(text.String(), "Evaluation observation:") != 8 ||
		!strings.Contains(text.String(), "3; use --format json") ||
		!strings.Contains(text.String(), "full value in --format json") || strings.Contains(text.String(), "full value in --json") {
		t.Fatal("short report omitted detail count or gave an invalid command", text.String())
	}
}

// Format the retained native observation; never invoke its model or program.
func TestConstructionReportRetainedNativeObservation(t *testing.T) {
	file, err := os.Open("../../internal/meta/languagereadiness/toolchainrelease/testdata/typed-path-construct.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var result bodyConstructOutput
	if err := json.NewDecoder(reader).Decode(&result); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := writeBodyConstructResult(&out, result, "text"); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"COMPLETE_FINITE", "11 / 16", "Selected attempt: \"11\"", "3/3 (100.00%)",
		"0 consumed during caller construction, 3 other", "model-training exposure are unknown"} {
		if !strings.Contains(out.String(), want) {
			t.Fatal("native record summary", want, out.String())
		}
	}
	t.Log("retained native output:\n" + out.String())
}
