package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestVerifierProcessClassification(t *testing.T) {
	cases := []struct {
		name       string
		mode       string
		missing    bool
		wantReason string
		wantClass  string
		wantNil    bool
		wantExit   int
		wantExitErr bool
	}{
		{name: "success", mode: "success", wantNil: true, wantExit: 0},
		{name: "positive exit", mode: "positive-exit", wantReason: "PROJECTED_COMPILE_OR_TEST_FAILED", wantClass: "KNOWN_CONTRADICTION", wantExit: 7, wantExitErr: true},
		{name: "start unavailable", missing: true, wantReason: "PROJECTED_COMPILE_OR_TEST_UNAVAILABLE", wantClass: "DIRECT_MISSING", wantExit: -1},
		{name: "signal termination", mode: "signal", wantReason: "PROJECTED_COMPILE_OR_TEST_INTERRUPTED", wantClass: "DIRECT_MISSING", wantExit: -1, wantExitErr: true},
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
			var exitError *exec.ExitError
			if got := errors.As(runErr, &exitError); got != testCase.wantExitErr {
				t.Fatalf("ExitError observed = %t, want %t (err=%v)", got, testCase.wantExitErr, runErr)
			}
			failure := classifyVerifierProcess("go-test-projected-workspace", result, runErr)
			if testCase.wantNil {
				if failure != nil || runErr != nil {
					t.Fatalf("success classified as failure=%v err=%v", failure, runErr)
				}
				return
			}
			if failure == nil || failure.reason != testCase.wantReason || failure.class != testCase.wantClass || failure.next != "restore-operation-evidence" {
				t.Fatalf("failure = %+v, want reason=%s class=%s next=restore-operation-evidence", failure, testCase.wantReason, testCase.wantClass)
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
