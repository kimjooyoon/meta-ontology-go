//go:build unix

package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestPackageExecuteInterruptRetainsTiming(t *testing.T) {
	t.Setenv("GOOO_INTERRUPT_TEST_TIMING", "1")
	TestPackageExecuteInterruptJoinsOwnedTool(t)
}

func checkPackageInterruptTiming(t *testing.T, raw []byte, receipt packageExecutionReceipt) {
	t.Helper()
	if os.Getenv("GOOO_INTERRUPT_TEST_TIMING") != "1" {
		if len(raw) != 0 {
			t.Fatal("default interrupted execution gained a diagnostic")
		}
		return
	}
	var diagnostic packageWallDiagnostic
	if err := json.Unmarshal(raw, &diagnostic); err != nil || diagnostic.Decision != "FAIL_CLOSED" ||
		diagnostic.Operation != "execute" || diagnostic.Wall.Status != "OBSERVED" ||
		diagnostic.ManifestDigest != receipt.ManifestDigest || diagnostic.CasesDigest != receipt.CasesDigest {
		t.Fatalf("interruption lost its current diagnostic bindings: decode=%v diagnostic=%+v", err, diagnostic)
	}
	for _, phase := range diagnostic.Wall.Phases {
		if phase.Name == "toolchain_bind" && phase.Outcome == "failed" {
			return
		}
	}
	t.Fatal("cancellation discarded the timing context or lost the interrupted tool phase")
}
