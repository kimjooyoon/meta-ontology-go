package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

func runPackageConstructionFixture(t *testing.T, args ...string) (packageConstructionReceipt, []byte) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := run(append([]string{"package", "construct", "--json"}, args...), &stdout, &stderr)
	if code != exitOK {
		t.Fatal(code, stderr.String(), stdout.String())
	}
	var r packageConstructionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &r); err != nil || r.Result == nil {
		t.Fatal(err, stdout.String())
	}
	return r, append([]byte(nil), stdout.Bytes()...)
}

func TestPackageConstructionDecisionRetainsUnscoredFault(t *testing.T) {
	r := workspaceexecution.ConstructionResult{
		Construction: bodyexecution.JointConstruction{Decision: "COMPLETE_FINITE"},
		Evaluation: bodyexecution.JointEvaluation{Runtime: bodyexecution.CompositionRuntime{
			Stage: "COMPLETE", FinitePassed: 1, FiniteTotal: 1}},
	}
	if packageConstructionDecision(r) != "COMPLETE_FINITE" {
		t.Fatal("matching evaluation remained partial")
	}
	for _, delivery := range []bodyexecution.CompositionDelivery{
		{Fault: &bodyexecution.CompositionFault{Kind: "ZERO_DIVISOR"}}, {BlockedBy: []string{"upstream"}},
	} {
		r.Evaluation.Runtime.Traces = []bodyexecution.CompositionTrace{{Deliveries: []bodyexecution.CompositionDelivery{delivery}}}
		if packageConstructionDecision(r) != "PARTIAL_FINITE" {
			t.Fatal("unscored failure claimed completion")
		}
	}
}

func TestPackageConstructImportedHelpersAndRelocatedReplay(t *testing.T) {
	root := "../../examples/package-caller-construction/"
	t.Setenv("GOOO_LAYA_URL", "http://127.0.0.1:1/not-used")
	t.Setenv("GOOO_LAYA_API_KEY", "unused")
	r, raw := runPackageConstructionFixture(t, "--construction-cases", root+"construction-cases.json",
		"--cases", root+"evaluation-cases.json", "--attempts", "6", root+"gooo.workspace.json")
	if r.Decision != "COMPLETE_FINITE" || r.Result.Evaluation.Runtime.FinitePassed != 4 ||
		r.Result.ConstructionCases.Cases[0].Expected["app/retry:Main"] == nil {
		t.Fatal(r.Decision, r.Result.Evaluation)
	}
	target := t.TempDir()
	for _, name := range []string{"gooo.workspace.json", "app.gooo.fixture", "budget.gooo.fixture", "evaluation-cases.json"} {
		contents, err := os.ReadFile(root + name)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(target, name), contents, 0600); err != nil {
			t.Fatal(err)
		}
	}
	saved := filepath.Join(target, "saved.json")
	if err := os.WriteFile(saved, raw, 0600); err != nil {
		t.Fatal(err)
	}
	replay, _ := runPackageConstructionFixture(t, "--receipt", saved, "--cases", filepath.Join(target, "evaluation-cases.json"), filepath.Join(target, "gooo.workspace.json"))
	if replay.Decision != "COMPLETE_FINITE" || replay.ReplayedFrom != workspaceDigest(raw) ||
		!replay.Result.Evaluation.ConstructionReplayed || replay.Result.Evaluation.NewModelCalls != 0 {
		t.Fatal(replay)
	}
	// The original receipt retains consumed examples; later evaluation cannot change selection.
	cases, err := os.ReadFile(filepath.Join(target, "evaluation-cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	cases = bytes.Replace(cases, []byte(`"expected":{"app/retry:Main":2}`), []byte(`"expected":{"app/retry:Main":999}`), 1)
	if err := os.WriteFile(filepath.Join(target, "evaluation-cases.json"), cases, 0600); err != nil {
		t.Fatal(err)
	}
	changed, _ := runPackageConstructionFixture(t, "--receipt", saved, "--cases", filepath.Join(target, "evaluation-cases.json"), filepath.Join(target, "gooo.workspace.json"))
	if changed.Decision != "PARTIAL_FINITE" || changed.Result.Evaluation.Runtime.FinitePassed != 3 ||
		changed.Result.Construction.SelectedSource != r.Result.Construction.SelectedSource {
		t.Fatal("evaluation changed construction", changed.Decision)
	}
}

func TestPackageConstructBoundedPartialAndMixedPackages(t *testing.T) {
	root := "../../examples/package-caller-construction/"
	r, _ := runPackageConstructionFixture(t, "--construction-cases", root+"construction-cases.json",
		"--cases", root+"evaluation-cases.json", "--attempts", "5", root+"gooo.workspace.json")
	if r.Decision != "PARTIAL_FINITE" || r.Result.Evaluation.Runtime.FinitePassed != 1 || len(r.Result.Construction.Attempts) != 5 {
		t.Fatal(r.Decision)
	}
	mixed, _ := runPackageConstructionFixture(t, "--construction-cases", root+"mixed-construction-cases.json",
		"--cases", root+"mixed-evaluation-cases.json", "--attempts", "48", root+"mixed.workspace.json")
	if mixed.Decision != "COMPLETE_FINITE" || mixed.Result.Construction.CandidateSpace != "48" ||
		len(mixed.Result.Program.PureCalls.Activities) != 3 || mixed.Result.Evaluation.Runtime.FinitePassed != 4 {
		t.Fatal(mixed.Decision)
	}
}

func TestPackageConstructUsageAndFailureReceipts(t *testing.T) {
	root := "../../examples/package-caller-construction/"
	for _, extras := range [][]string{
		{}, {"--attempts", "0", "--construction-cases", "x"}, {"--attempts", "65", "--construction-cases", "x"},
		{"--receipt", "x", "--model", "x"}, {"--receipt", "x", "--fill-model", "x"},
		{"--receipt", "x", "--construction-cases", "x"}, {"--receipt", "x", "--attempts", "1"},
		{"--receipt", "x", "--receipt", "x"}, {"--receipt", "x", "--entry", "Main"},
		{"--receipt", "x", "--body-plans", "x"},
	} {
		args := append([]string{"--cases", "evaluation.json", "workspace.json"}, extras...)
		if _, _, err := parsePackageConstructArgs(args); err == nil {
			t.Fatal("invalid options accepted", args)
		}
	}
	var stdout, stderr bytes.Buffer
	args := []string{"package", "construct", "--json", "--construction-cases", root + "construction-cases.json",
		"--cases", root + "evaluation-cases.json", "--attempts", "1", "--go", "/missing/go", root + "gooo.workspace.json"}
	if code := run(args, &stdout, &stderr); code != exitFailure {
		t.Fatal("missing tool did not fail", code)
	}
	var r packageConstructionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &r); err != nil || r.Decision != "FAIL_CLOSED" || r.Result == nil ||
		r.Result.Construction.Failure == "" || len(r.Result.Construction.Attempts) != 1 || !strings.Contains(r.Error, "go") {
		t.Fatal(err, stdout.String())
	}
}

func TestPackageConstructionReadsItsDeclaredReceiptSize(t *testing.T) {
	r := packageConstructionReceipt{Schema: "gooo/workspace-caller-construction-receipt/v1",
		Decision: "PARTIAL_FINITE", ManifestDigest: "same-manifest",
		Result: &workspaceexecution.ConstructionResult{Scope: strings.Repeat("x", 17<<20)}}
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "large.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	saved, digest, err := readSavedPackageConstruction(OSFileReader{}, path, "same-manifest")
	if err != nil || saved.Result == nil || len(saved.Result.Scope) != 17<<20 || digest != workspaceDigest(raw) {
		t.Fatal("a valid receipt within the 32 MiB boundary cannot be reopened", err)
	}
}
