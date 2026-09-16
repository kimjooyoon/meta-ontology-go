package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

func TestVerifierProcessClassification(t *testing.T) {
	cases := []struct {
		name             string
		mode             string
		missing          bool
		wantReason       string
		wantClass        string
		wantNil          bool
		wantExit         int
		wantDecision     string
		wantExitErr      bool
		wantUnknownClass string
		wantNext         string
	}{
		{name: "success", mode: "success", wantNil: true, wantExit: 0},
		{name: "positive exit", mode: "positive-exit", wantReason: "PROJECTED_COMPILE_OR_TEST_FAILED", wantClass: "KNOWN_CONTRADICTION", wantExit: 7, wantDecision: "REFUTED", wantExitErr: true, wantNext: "report-counterexample"},
		{name: "start unavailable", missing: true, wantReason: "PROJECTED_COMPILE_OR_TEST_UNAVAILABLE", wantClass: "DIRECT_MISSING", wantExit: -1, wantDecision: "UNKNOWN", wantUnknownClass: "DIRECT_MISSING", wantNext: "restore-operation-evidence"},
		{name: "signal termination", mode: "signal", wantReason: "PROJECTED_COMPILE_OR_TEST_INTERRUPTED", wantClass: "DIRECT_MISSING", wantExit: -1, wantDecision: "UNKNOWN", wantExitErr: true, wantUnknownClass: "DIRECT_MISSING", wantNext: "restore-operation-evidence"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.mode == "signal" && runtime.GOOS == "windows" {
				t.Skip("self-termination does not provide a portable signal exit on Windows")
			}
			result, runErr := runVerifierProcessFixture(t, testCase.mode, testCase.missing)
			if result.Observation.ExitCode != testCase.wantExit {
				t.Fatalf("exit code = %d, want %d (err=%v)", result.Observation.ExitCode, testCase.wantExit, runErr)
			}
			if _, got := errors.AsType[*exec.ExitError](runErr); got != testCase.wantExitErr {
				t.Fatalf("ExitError observed = %t, want %t (err=%v)", got, testCase.wantExitErr, runErr)
			}
			failure := classifyVerifierProcess("go-test-projected-workspace", result, runErr)
			if testCase.wantNil {
				if failure != nil || runErr != nil {
					t.Fatalf("success classified as failure=%v err=%v", failure, runErr)
				}
				return
			}
			if failure == nil || failure.reason != testCase.wantReason || failure.class != testCase.wantClass || failure.next != testCase.wantNext {
				t.Fatalf("failure = %+v, want reason=%s class=%s next=%s", failure, testCase.wantReason, testCase.wantClass, testCase.wantNext)
			}
			action := generation.Action{IndicatorID: "verifier-process-fixture"}
			observed := observationFailureFromError(action, failure, result.Observation)
			if observed.Decision != testCase.wantDecision || observed.Stage != "verify-operation" ||
				observed.Step != "go-test-projected-workspace" || observed.Reason != testCase.wantReason ||
				observed.UnknownClass != testCase.wantUnknownClass || observed.NextOperation != testCase.wantNext ||
				observed.BlockedBy == nil || len(observed.BlockedBy) != 0 || observed.ActionIndicatorID != action.IndicatorID {
				t.Fatalf("observation failure = %+v, want decision=%s stage=verify-operation step=go-test-projected-workspace reason=%s unknown_class=%s next=%s blocked_by=[]", observed, testCase.wantDecision, testCase.wantReason, testCase.wantUnknownClass, testCase.wantNext)
			}
		})
	}
}

func runVerifierProcessFixture(t *testing.T, mode string, missing bool) (processResult, error) {
	t.Helper()
	root := t.TempDir()
	descriptor := []string{"<workspace>", "verifier-fixture", mode}
	if missing {
		return runProcess(root, os.Environ(), descriptor, []string{filepath.Join(root, "missing-verifier")})
	}
	environment := append(os.Environ(), "META_EXECUTION_VERIFIER_FIXTURE="+mode)
	return runProcess(root, environment, descriptor, []string{os.Args[0], "-test.run=^TestVerifierProcessFixtureHelper$"})
}

func TestVerifierProcessFixtureHelper(t *testing.T) {
	mode := os.Getenv("META_EXECUTION_VERIFIER_FIXTURE")
	if mode == "" {
		return
	}
	switch mode {
	case "success":
		os.Exit(0)
	case "positive-exit":
		os.Exit(7)
	case "signal":
		process, err := os.FindProcess(os.Getpid())
		if err != nil {
			t.Fatalf("find helper process: %v", err)
		}
		if err := process.Kill(); err != nil {
			t.Fatalf("terminate helper process: %v", err)
		}
		select {}
	default:
		t.Fatalf("unknown helper mode %q", mode)
	}
}

func TestVerifierCacheInvocationEnvironmentPreservesCallerControl(t *testing.T) {
	trace := metaExecutionTrace{state: newMetaExecutionTraceStateWithWriter(&bytes.Buffer{})}
	trace.action.Activity = "CollapseAssignReturn"
	for _, test := range []struct {
		name string
		env  []string
		want []string
	}{
		{"default", []string{"PATH=bin"}, []string{"PATH=bin", "GODEBUG=gocachetest=1"}},
		{"other-flags", []string{"GODEBUG=panicnil=1"}, []string{"GODEBUG=panicnil=1,gocachetest=1"}},
		{"disabled", []string{"GODEBUG=gocachetest=0"}, []string{"GODEBUG=gocachetest=0"}},
		{"enabled", []string{"GODEBUG=gocachetest=1"}, []string{"GODEBUG=gocachetest=1"}},
		{"invalid", []string{"GODEBUG=gocachetest=invalid"}, []string{"GODEBUG=gocachetest=invalid"}},
		{"last-key", []string{"GODEBUG=gocachetest=0", "GODEBUG=panicnil=1"},
			[]string{"GODEBUG=panicnil=1,gocachetest=1", "GODEBUG=panicnil=1,gocachetest=1"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			before := append([]string{}, test.env...)
			got := verifierCacheEnvironment(test.env, &trace)
			if !reflect.DeepEqual(got, test.want) || !reflect.DeepEqual(test.env, before) {
				t.Fatalf("caller environment changed: got=%v input=%v want=%v", got, test.env, test.want)
			}
		})
	}
	t.Run("inherited-parent", func(t *testing.T) {
		t.Setenv("GODEBUG", "panicnil=1")
		t.Setenv("GOOO_CACHE_INVOCATION_PARENT", "present")
		got := verifierCacheEnvironment(nil, &trace)
		if !slices.Contains(got, "GOOO_CACHE_INVOCATION_PARENT=present") ||
			!slices.Contains(got, "GODEBUG=panicnil=1,gocachetest=1") || os.Getenv("GODEBUG") != "panicnil=1" {
			t.Fatalf("inherited environment or parent control changed: %v", got)
		}
	})
}

func TestVerifierCacheInvocationLeavesOtherOperationsUntouched(t *testing.T) {
	environment := []string{"PATH=bin", "GODEBUG=panicnil=1"}
	for _, activity := range []string{"ExtractFunction", "", "UNKNOWN_OPERATION"} {
		trace := metaExecutionTrace{state: newMetaExecutionTraceStateWithWriter(&bytes.Buffer{})}
		trace.action.Activity = activity
		if got := verifierCacheEnvironment(environment, &trace); !reflect.DeepEqual(got, environment) {
			t.Fatalf("unsupported operation %q acquired diagnostics: %v", activity, got)
		}
	}
	for _, trace := range []*metaExecutionTrace{nil, {}} {
		if got := verifierCacheEnvironment(nil, trace); got != nil {
			t.Fatalf("unobserved call changed inherited environment: %v", got)
		}
	}
	trace := metaExecutionTrace{state: newMetaExecutionTraceStateWithWriter(&bytes.Buffer{})}
	trace.action.Activity = "SplitGoDeclarations"
	if got := verifierCacheEnvironment([]string{}, &trace); !reflect.DeepEqual(got, []string{"GODEBUG=gocachetest=1"}) {
		t.Fatalf("split verifier has no diagnostic request: %v", got)
	}
}

func TestVerifierCacheInvocationNativeSourceChangeCannotReuseSuccess(t *testing.T) {
	// Synthetic cache conformance only, not a production speedup or utility claim.
	source := t.TempDir()
	replay, err := newMetaReplayWorkspace()
	if err != nil {
		t.Fatal(err)
	}
	defer replay.close()
	root := replay.path
	write := func(name, content string) {
		if err := os.WriteFile(filepath.Join(source, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module example.com/gooo-verifier-cache-invocation-witness\n\ngo 1.27.0\n")
	write("value.go", "package cachewitness\n\nfunc Value() int { return 42 }\n")
	write("value_test.go", fmt.Sprintf("package cachewitness\n\nimport \"testing\"\n\nfunc TestValue(t *testing.T) {\n"+
		"t.Log(%q)\nif Value() != 42 { t.Fatal(\"value changed\") }\n}\n", root))
	environment := replaceEnvironment(os.Environ(), "GOWORK", "off")
	environment = replaceEnvironment(environment, "GOTOOLCHAIN", "local")
	environment = replaceEnvironment(environment, "GOFLAGS", "")
	environment = replaceEnvironment(environment, "GODEBUG", "panicnil=1")
	trace := metaExecutionTrace{state: newMetaExecutionTraceStateWithWriter(&bytes.Buffer{})}
	trace.action.Activity = "CollapseAssignReturn"
	restoreVerifierReplayFixture(t, replay, source, root)
	first, firstErr := runGoTestObserved(root, environment, ".", &trace, "first")
	restoreVerifierReplayFixture(t, replay, source, root)
	second, secondErr := runGoTestObserved(root, environment, ".", &trace, "replay")
	if firstErr != nil || secondErr != nil || first.Observation.ExitCode != 0 || second.Observation.ExitCode != 0 {
		t.Fatalf("native fixture did not pass: first=%v %s replay=%v %s", firstErr, first.Stderr, secondErr, second.Stderr)
	}
	if strings.Contains(string(first.Stdout), "(cached)") || !strings.Contains(string(second.Stdout), "(cached)") {
		t.Fatalf("native cache control not observed: first=%s replay=%s", first.Stdout, second.Stdout)
	}
	for _, result := range []processResult{first, second} {
		if !bytes.Contains(result.Stderr, []byte("testcache: ")) ||
			result.Observation.RawStderrDigest != digestBytes(result.Stderr) ||
			result.Observation.StderrBytes != len(result.Stderr) {
			t.Fatalf("invocation did not produce source-bound native diagnostics: %#v stderr=%s", result.Observation, result.Stderr)
		}
	}
	write("value.go", "package cachewitness\n\nfunc Value() int { return 41 }\n")
	restoreVerifierReplayFixture(t, replay, source, root)
	changed, changedErr := runGoTestObserved(root, environment, ".", &trace, "changed-source")
	failure := classifyVerifierProcess("synthetic-cache-witness", changed, changedErr)
	if changedErr == nil || changed.Observation.ExitCode <= 0 || failure == nil ||
		failure.reason != "PROJECTED_COMPILE_OR_TEST_FAILED" || failure.class != "KNOWN_CONTRADICTION" ||
		strings.Contains(string(changed.Stdout), "(cached)") {
		t.Fatalf("changed source reused success: result=%#v error=%v failure=%#v", changed, changedErr, failure)
	}
}

func restoreVerifierReplayFixture(t *testing.T, replay *metaReplayWorkspace, source, expectedRoot string) {
	t.Helper()
	root, err := replay.restore(source)
	if err != nil || root != expectedRoot {
		t.Fatalf("replay address or restoration changed: root=%q expected=%q error=%v", root, expectedRoot, err)
	}
}

func TestMetaReplayWorkspaceRestoresPristineSource(t *testing.T) {
	source := t.TempDir()
	const original = "package fixture\n"
	if err := os.WriteFile(filepath.Join(source, "input.go"), []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	replay, err := newMetaReplayWorkspace()
	if err != nil {
		t.Fatal(err)
	}
	defer replay.close()
	root := replay.path
	restoreVerifierReplayFixture(t, replay, source, root)
	for _, name := range []string{"input.go", "residue.go"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("mutated"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	restoreVerifierReplayFixture(t, replay, source, root)
	data, err := os.ReadFile(filepath.Join(root, "input.go"))
	if err != nil || string(data) != original {
		t.Fatalf("prior candidate survived restoration: %q (%v)", data, err)
	}
	if _, err := os.Stat(filepath.Join(root, "residue.go")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("prior candidate residue survived restoration: %v", err)
	}
	data, err = os.ReadFile(filepath.Join(source, "input.go"))
	if err != nil || string(data) != original {
		t.Fatalf("input source was changed: %q (%v)", data, err)
	}
}

func TestMetaReplayWorkspaceSeparatesInvocationsAndRevokesClosedLease(t *testing.T) {
	first, err := newMetaReplayWorkspace()
	if err != nil {
		t.Fatal(err)
	}
	defer first.close()
	second, err := newMetaReplayWorkspace()
	if err != nil {
		t.Fatal(err)
	}
	defer second.close()
	if first.path == second.path {
		t.Fatal("independent invocations acquired the same workspace")
	}
	first.close()
	for _, replay := range []*metaReplayWorkspace{nil, {}, first} {
		if _, err := replay.restore(t.TempDir()); err == nil {
			t.Fatal("unallocated or closed workspace acquired restore authority")
		}
	}
}

func TestMetaReplayWorkspaceRejectsAliasedSourceWithoutDeletingIt(t *testing.T) {
	replay, err := newMetaReplayWorkspace()
	if err != nil {
		t.Fatal(err)
	}
	defer replay.close()
	path := filepath.Join(replay.path, "keep")
	if err := os.WriteFile(path, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := replay.restore(replay.path); err == nil {
		t.Fatal("aliased input was admitted")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "original" {
		t.Fatalf("rejected aliased source was deleted: %q (%v)", data, err)
	}
}
