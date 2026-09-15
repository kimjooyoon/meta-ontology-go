package generation

import (
	"strings"
	"testing"
)

func TestOperationObservationReplayBindsCommandAndCanonicalProcess(t *testing.T) {
	plan := actionableReceiptPlan()
	base := ProcessObservation{
		Command:      []string{"go", "run", "<workspace>", "--plan", "meta-execution-function-plan.json"},
		ExitCode:     1,
		StdoutBytes:  2,
		StdoutDigest: "sha256:" + strings.Repeat("1", 64),
		StderrBytes:  3,
		StderrDigest: "sha256:" + strings.Repeat("2", 64),
	}
	withRawDigest := base
	withRawDigest.RawStdoutDigest = "sha256:" + strings.Repeat("3", 64)
	withRawDigest.RawStderrDigest = "sha256:" + strings.Repeat("4", 64)
	first := sealReplayFailure(plan, base)
	second := sealReplayFailure(plan, withRawDigest)
	if first.ReplayDigest != second.ReplayDigest {
		t.Fatal("raw process digest drift changed replay projection")
	}

	canonicalDrift := withRawDigest
	canonicalDrift.StderrDigest = "sha256:" + strings.Repeat("5", 64)
	if first.ReplayDigest == sealReplayFailure(plan, canonicalDrift).ReplayDigest {
		t.Fatal("canonical process digest drift was omitted from replay projection")
	}

	commandDrift := base
	commandDrift.Command = []string{"go", "run", "<workspace>", "--different-option"}
	if first.ReplayDigest == sealReplayFailure(plan, commandDrift).ReplayDigest {
		t.Fatal("command descriptor drift was omitted from replay projection")
	}
}

func TestValidateObservationBundleRejectsNonCanonicalCommandPath(t *testing.T) {
	for _, command := range [][]string{
		{"/tmp/actual-workspace", "go"},
		{"go", "--root=/tmp/actual-workspace"},
		{"go", `C:\actual-workspace`},
		{"go", `--root=C:\actual-workspace`},
	} {
		bundle, plan, manifest := validationFailureBundle(command)
		if err := ValidateObservationBundle(bundle, plan, manifest); err == nil {
			t.Fatalf("non-canonical command was accepted: %v", command)
		}
	}
}

func TestValidateObservationBundleAcceptsCanonicalCommand(t *testing.T) {
	command := []string{"go", "run", "./scripts/source-splitter", "-root", "<workspace>", "-subject", "fixture.go:1:Selected"}
	bundle, plan, manifest := validationFailureBundle(command)
	if err := ValidateObservationBundle(bundle, plan, manifest); err != nil {
		t.Fatalf("canonical command was rejected: %v", err)
	}
}

func validationFailureBundle(command []string) (OperationObservationBundle, Plan, ExecutionManifest) {
	plan := actionableReceiptPlan()
	manifest := BuildExecutionManifest(plan)
	process := ProcessObservation{
		Command:         command,
		ExitCode:        -1,
		RawStdoutDigest: "sha256:" + strings.Repeat("0", 64),
		StdoutDigest:    "sha256:" + strings.Repeat("0", 64),
		RawStderrDigest: "sha256:" + strings.Repeat("0", 64),
		StderrDigest:    "sha256:" + strings.Repeat("0", 64),
	}
	failures := make([]ObservationFailure, 0, len(plan.Selected))
	for _, action := range plan.Selected {
		failures = append(failures, ObservationFailure{
			ActionIndicatorID: action.IndicatorID,
			Decision:          "UNKNOWN",
			Stage:             "execute-operation",
			Step:              "run-selected-operation",
			Reason:            "INSTANCE_EVIDENCE_UNAVAILABLE",
			UnknownClass:      ReceiptUnknownClassDirectMissing,
			NextOperation:     "restore-operation-evidence",
			BlockedBy:         []string{},
			Executor:          process,
		})
	}
	bundle := OperationObservationBundle{
		BaseSHA:          plan.BaseSHA,
		HeadSHA:          plan.HeadSHA,
		PlanDigest:       plan.PlanDigest,
		ManifestDigest:   manifest.ManifestDigest,
		Failures:         failures,
		ObservationTotal: len(failures),
	}
	return SealObservationBundle(bundle), plan, manifest
}

func sealReplayFailure(plan Plan, process ProcessObservation) OperationObservationBundle {
	bundle := OperationObservationBundle{
		BaseSHA:        plan.BaseSHA,
		HeadSHA:        plan.HeadSHA,
		PlanDigest:     plan.PlanDigest,
		ManifestDigest: "sha256:" + strings.Repeat("6", 64),
		Failures: []ObservationFailure{{
			ActionIndicatorID: plan.Selected[0].IndicatorID,
			Decision:          "UNKNOWN",
			Stage:             "execute-operation",
			Step:              "run-selected-operation",
			Reason:            "INSTANCE_EVIDENCE_UNAVAILABLE",
			UnknownClass:      ReceiptUnknownClassDirectMissing,
			NextOperation:     "restore-operation-evidence",
			BlockedBy:         []string{},
			Executor:          process,
		}},
		ObservationTotal:  1,
		ReplayComparisons: 0,
	}
	return SealObservationBundle(bundle)
}
