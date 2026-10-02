package bodyexecution

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

func fixture(t *testing.T) ([]byte, pathplan.Document, bodycodegen.Result, []byte) {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/typed-path-compound.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := os.ReadFile("../../examples/body-codegen/typed-path-compound-plan.json")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := DecodePlan(plan)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := bodycodegen.GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "Combined", doc, "")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.MarshalIndent(prior, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	prior, parent, err := DecodeGeneration(raw)
	if err != nil {
		t.Fatal(err)
	}
	return source, doc, prior, parent
}

func dimension(t *testing.T, result Result, id string) completeness.CompletenessDimension {
	t.Helper()
	for _, d := range result.CompletenessReceipt.Dimensions {
		if d.ID == id {
			return d
		}
	}
	t.Fatalf("missing dimension %s", id)
	return completeness.CompletenessDimension{}
}

func verifyReceipt(t *testing.T, result Result) {
	t.Helper()
	raw, err := json.Marshal(result.CompletenessReceipt)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := completeness.Decode(raw); err != nil {
		t.Fatalf("invalid shared receipt: %v\n%s", err, raw)
	}
}

func TestCompiledRuntimeFiniteScoresAndImmutableParent(t *testing.T) {
	source, doc, prior, parent := fixture(t)
	sourceBefore, parentBefore := bytes.Clone(source), bytes.Clone(parent)
	priorBefore, _ := json.Marshal(prior)
	tool := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		tool += ".exe"
	}
	for _, passed := range []int{3, 2, 0} {
		cases := []pathplan.TestCase{{Input: -4, Expected: -21}, {Input: 1, Expected: 10}, {Input: 4, Expected: 25}}
		for i := passed; i < len(cases); i++ {
			cases[i].Expected = 999
		}
		result, err := Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, cases, tool)
		if err != nil {
			t.Fatal(err)
		}
		verifyReceipt(t, result)
		d := dimension(t, result, "runtime_finite_accuracy")
		want := "PROGRESS"
		if passed == 3 {
			want = "PASS"
		}
		if d.Numerator != passed || d.Denominator != 3 || d.Status != want {
			t.Fatalf("finite score: %+v", d)
		}
		for _, id := range []string{"runtime_completion", "runtime_source_replay", "runtime_build", "execution_boundary", "runtime_deterministic_replay", "reverse_observation_coverage", "runtime_selection_disjointness"} {
			if dimension(t, result, id).Status != "PASS" {
				t.Fatalf("%s did not pass", id)
			}
		}
		if dimension(t, result, "permission_boundary").Status != "UNKNOWN" {
			t.Fatal("invented permission observation")
		}
		if len(result.Observation.Runs) != 2 || result.Observation.ActivityID != "sample://activity/combined" || result.Observation.ExecutableSHA256 == "" {
			t.Fatal("missing native reverse binding")
		}
		if result.Observation.ParentReceiptSHA256 != digest(parent) || !bytes.Equal(result.ParentReceipt, parent) {
			t.Fatal("parent changed")
		}
		wire, _ := json.Marshal(result)
		var decoded Result
		if json.Unmarshal(wire, &decoded) != nil || !bytes.Equal(decoded.ParentReceipt, parent) {
			t.Fatal("wire lost parent whitespace")
		}
	}
	priorAfter, _ := json.Marshal(prior)
	if !bytes.Equal(source, sourceBefore) || !bytes.Equal(parent, parentBefore) || !bytes.Equal(priorAfter, priorBefore) {
		t.Fatal("caller inputs mutated")
	}
}

func TestRejectedRuntimeNeverStartsTool(t *testing.T) {
	for _, change := range []string{"source", "generated", "selection", "plan", "cases-observation", "parent", "cancel", "empty-suite", "missing-tool"} {
		t.Run(change, func(t *testing.T) {
			source, doc, prior, parent := fixture(t)
			cases := doc.TestCases
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			switch change {
			case "source":
				source = append(source, '\n')
			case "generated":
				prior.Source += "\nfunc injected() {}\n"
			case "selection":
				prior.Report.BodyPaths.Search.Selection.Choices["reference"] = "reference_first"
			case "plan":
				doc.Plan.Decisions[0].Intent += " changed"
			case "cases-observation":
				prior.Report.BodyPaths.NativeCases[0].Actual++
			case "parent":
				parent = bytes.Replace(parent, []byte("gooo/body-codegen-typed-path-v1"), []byte("forged/profile"), 1)
			case "cancel":
				cancel()
			case "empty-suite":
				cases = nil
			}
			result, err := Execute(ctx, "fixture.gooo", source, doc, prior, parent, cases, filepath.Join(t.TempDir(), "missing-go"))
			if err == nil || result.Observation.Toolchain.Started || result.Observation.Build.Started || len(result.Observation.Runs) != 0 {
				t.Fatal("rejected input reached a child")
			}
			if change != "missing-tool" && strings.Contains(err.Error(), "Go tool is unavailable") {
				t.Fatal("invalid observation reached tool lookup")
			}
			verifyReceipt(t, result)
			if result.CompletenessReceipt.Decision != "FAIL_CLOSED" || dimension(t, result, "runtime_finite_accuracy").Status != "UNKNOWN" {
				t.Fatal("unobserved runtime was scored")
			}
		})
	}
}

func TestRuntimeSelectionSuiteIsNotCalledIndependent(t *testing.T) {
	source, doc, prior, parent := fixture(t)
	tool := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		tool += ".exe"
	}
	result, err := Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, tool)
	if err != nil {
		t.Fatal(err)
	}
	verifyReceipt(t, result)
	d := dimension(t, result, "runtime_selection_disjointness")
	if d.Numerator != 0 || d.Status != "PROGRESS" {
		t.Fatal("selection cases became independent", d)
	}
}

func TestRuntimeRecordsBuildFailureWithoutScoringUnobservedCases(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX tool stub; portable process tests cover other platforms")
	}
	source, doc, prior, parent := fixture(t)
	tool := filepath.Join(t.TempDir(), "go-stub")
	stub := "#!/bin/sh\nif [ \"$1\" = version ]; then echo 'go version go1.27.1 " + runtime.GOOS + "/" + runtime.GOARCH + "'; exit 0; fi\nexit 7\n"
	if err := os.WriteFile(tool, []byte(stub), 0700); err != nil {
		t.Fatal(err)
	}
	result, err := Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, tool)
	if err == nil || result.Observation.Stage != "BUILD" || !result.Observation.Build.Started || result.Observation.Build.ExitCode == nil || *result.Observation.Build.ExitCode != 7 {
		t.Fatal("build failure missing", err)
	}
	verifyReceipt(t, result)
	if dimension(t, result, "runtime_completion").Status != "FAIL_CLOSED" || dimension(t, result, "runtime_finite_accuracy").Status != "UNKNOWN" {
		t.Fatal("failure invented runtime outputs")
	}
}

func TestChildEnvironmentDoesNotForwardProviderOrGoOverrides(t *testing.T) {
	t.Setenv("HF_TOKEN", "never-forward")
	t.Setenv("GOFLAGS", "-toolexec=bad")
	t.Setenv("GOOO_LAYA_URL", "http://unused")
	for _, entry := range childEnvironment() {
		if strings.Contains(entry, "never-forward") || strings.HasPrefix(entry, "GOFLAGS=") || strings.HasPrefix(entry, "GOOO_LAYA_URL=") {
			t.Fatal("ambient provider/tool override forwarded")
		}
	}
}

func TestRuntimePreservesUnexportedActivityAndMainPackage(t *testing.T) {
	source, doc, _, _ := fixture(t)
	source = bytes.ReplaceAll(bytes.Replace(source, []byte("package sample"), []byte("package main"), 1), []byte("Combined"), []byte("combined"))
	doc.Plan.Base.Name = "combined"
	prior, err := bodycodegen.GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "combined", doc, "")
	if err != nil {
		t.Fatal(err)
	}
	parent, _ := json.Marshal(prior.Report.CompletenessReceipt)
	tool := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		tool += ".exe"
	}
	result, err := Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, tool)
	if err != nil {
		t.Fatal(err)
	}
	verifyReceipt(t, result)
	if dimension(t, result, "runtime_finite_accuracy").Status != "PASS" {
		t.Fatal("activity name changed")
	}
}
