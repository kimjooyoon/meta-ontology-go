package extractor

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

func TestCallbackExtractionObservationRequiresCI(t *testing.T) {
	t.Setenv("CI", "")
	observation, err := ObserveCallbackExtraction(context.Background(), "", "", "")
	if err == nil || observation.OperationAdmission != "UNKNOWN" || observation.ApplyPermission != "FORBIDDEN" || len(observation.Runs) != 0 {
		t.Fatalf("local execution escaped its boundary: %+v err=%v", observation, err)
	}
}

func TestCallbackExtractionObservesOriginalAndFinalPackageCI(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("final callback package runtime observation is CI-only")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	observation, err := ObserveCallbackExtraction(ctx, callbackPreviewRepositoryRoot(t), callbackPreviewLogicalPath, "func:"+callbackPreviewTarget)
	if err != nil {
		t.Fatalf("observation=%+v err=%v", observation, err)
	}
	if observation.Decision != "OBSERVED" || observation.Scope != "PACKAGE_TEST_EVENTS_ONLY" || len(observation.Runs) != 2 ||
		observation.AttemptedRuns != 2 || observation.CompletedTestRuns != 2 || observation.RequiredTestRuns != 2 ||
		observation.ModuleDigest == "" || observation.ModuleFiles == 0 || observation.ModuleBytes == 0 ||
		observation.SourcePackageDigest == "" || observation.FinalPackageDigest == "" || observation.TestEventDigest == "" ||
		observation.GeneratedFiles != 6 || observation.OperationAdmission != "UNKNOWN" || observation.ApplyPermission != "FORBIDDEN" ||
		observation.DependencyBinding != "UNBOUND" || observation.Frontier.UnknownClass != "UNBOUNDED" {
		t.Fatalf("observation scope or admission changed: %+v", observation)
	}
	record := observation.Record
	if record.Activity != "BindCallbackExtractionObservers" || record.Fields["State"] != "UNKNOWN" ||
		record.Fields["ObservationDecision"] != "OBSERVED" || record.Fields["ObservedCount"] != "2" || record.Fields["RequiredCount"] != "2" ||
		record.Fields["Scope"] != observation.Scope || record.Fields["SourcePackageDigest"] != observation.SourcePackageDigest ||
		record.Fields["FinalPackageDigest"] != observation.FinalPackageDigest || record.Fields["TestEventDigest"] != observation.TestEventDigest {
		t.Fatalf("runtime observation is not Gooo-bound: %+v", record)
	}
	for _, run := range observation.Runs {
		if run.ExitCode != 0 || !run.TestEventsComplete || run.WallMS < 0 || len(run.Events) == 0 ||
			run.StdoutDigest != proofDigest(run.Stdout) || run.StderrDigest != proofDigest(run.Stderr) {
			t.Fatalf("invalid native package execution receipt: %+v", run)
		}
		t.Logf("callback final-package observation variant=%s exit=%d wall_ms=%d terminal_test_events=%d generated_files=%d scope=%s admission=%s",
			run.Variant, run.ExitCode, run.WallMS, len(run.Events), observation.GeneratedFiles, observation.Scope, observation.OperationAdmission)
	}
}

func TestCallbackPackageCommandFailurePreservesOutputCI(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("native compile-failure observation is CI-only")
	}
	directory := t.TempDir()
	for name, source := range map[string]string{
		"go.mod":          "module example.invalid/callback-failure\n\ngo 1.27.0\n",
		"failure_test.go": "package fixture\nimport \"testing\"\nfunc TestRequired(t *testing.T) { missingObserverSymbol() }\n",
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	run, err := runCallbackPackageObservation(ctx, directory, "source", "TestRequired")
	if err == nil || run.ExitCode != 1 || run.TestEventsComplete || len(run.Stdout)+len(run.Stderr) == 0 ||
		!strings.Contains(string(run.Stdout)+string(run.Stderr), "missingObserverSymbol") ||
		run.StdoutDigest != proofDigest(run.Stdout) || run.StderrDigest != proofDigest(run.Stderr) {
		t.Fatalf("native failure output was lost: run=%+v err=%v", run, err)
	}
	raw, err := json.Marshal(run)
	if err != nil {
		t.Fatal(err)
	}
	var decoded CallbackPackageRun
	if err := json.Unmarshal(raw, &decoded); err != nil || !bytes.Equal(decoded.Stdout, run.Stdout) || !bytes.Equal(decoded.Stderr, run.Stderr) {
		t.Fatalf("failure receipt did not preserve output bytes: %v", err)
	}
	frontier := callbackPackageFailureFrontier(run)
	if frontier.State != "UNKNOWN" || frontier.Step != "RUN_SOURCE_PACKAGE" || frontier.UnknownClass != "DIRECT_MISSING" {
		t.Fatalf("compile failure became a runtime counterexample: %+v", frontier)
	}
}

func TestCallbackObservationSeparatesAttemptsFromCompletedTests(t *testing.T) {
	contract, err := generation.LoadCallbackExtractionContract()
	if err != nil {
		t.Fatal(err)
	}
	observation := CallbackExtractionObservation{Runs: []CallbackPackageRun{{Variant: "source", ExitCode: 1}}}
	if err := bindCallbackPackageObservation(&observation, contract); err != nil {
		t.Fatal(err)
	}
	if observation.AttemptedRuns != 1 || observation.CompletedTestRuns != 0 || observation.RequiredTestRuns != 2 ||
		observation.Record.Fields["ObservedCount"] != "0" || observation.Record.Fields["RequiredCount"] != "2" {
		t.Fatalf("failed command counted as completed tests: %+v", observation)
	}
	frontier := callbackPackageFailureFrontier(CallbackPackageRun{Variant: "final", ExitCode: 1,
		Events: []CallbackPackageTestEvent{{Name: "TestRequired", Action: "fail"}}})
	if frontier.State != "REFUTED" || frontier.UnknownClass != "" || frontier.Reason != "PACKAGE_TEST_COUNTEREXAMPLE_OBSERVED" {
		t.Fatalf("observed runtime counterexample was hidden: %+v", frontier)
	}
}

func TestCallbackPackageTestEventsRejectInvalidEvidence(t *testing.T) {
	for name, raw := range map[string]string{
		"missing-required-test": "{\"Action\":\"pass\",\"Test\":\"TestOther\"}\n",
		"unknown-action":        "{\"Action\":\"FIXED_POINT\",\"Test\":\"TestRequired\"}\n",
		"duplicate-terminal":    "{\"Action\":\"pass\",\"Test\":\"TestRequired\"}\n{\"Action\":\"pass\",\"Test\":\"TestRequired\"}\n",
		"malformed-stream":      "not-json",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := callbackPackageTestEvents([]byte(raw), "TestRequired"); err == nil {
				t.Fatal("invalid test-event evidence was accepted")
			}
		})
	}
}
