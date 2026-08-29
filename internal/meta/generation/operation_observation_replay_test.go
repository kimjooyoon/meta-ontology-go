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

func sealReplayFailure(plan Plan, process ProcessObservation) OperationObservationBundle {
	bundle := OperationObservationBundle{
		BaseSHA:           plan.BaseSHA,
		HeadSHA:           plan.HeadSHA,
		PlanDigest:        plan.PlanDigest,
		ManifestDigest:    "sha256:" + strings.Repeat("6", 64),
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
