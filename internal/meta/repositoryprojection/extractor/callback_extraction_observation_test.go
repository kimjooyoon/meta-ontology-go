package extractor

import (
	"context"
	"os"
	"testing"
	"time"
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
		if run.ExitCode != 0 || run.WallMS < 0 || run.StdoutDigest == "" || run.StderrDigest == "" || len(run.Events) == 0 {
			t.Fatalf("invalid native package execution receipt: %+v", run)
		}
		t.Logf("callback final-package observation variant=%s exit=%d wall_ms=%d terminal_test_events=%d generated_files=%d scope=%s admission=%s",
			run.Variant, run.ExitCode, run.WallMS, len(run.Events), observation.GeneratedFiles, observation.Scope, observation.OperationAdmission)
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
