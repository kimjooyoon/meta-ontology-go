//go:build unix

package main

import (
	"encoding/json"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestPackageExecuteInterruptJoinsOwnedTool(t *testing.T) {
	if os.Getenv("GOOO_INTERRUPT_TEST_CHILD") == "1" {
		args := []string{"package", "execute", "--json", "--cases", "../../examples/assembly-explainer/cases.json",
			"--go", os.Getenv("GOOO_INTERRUPT_TEST_TOOL"), "../../examples/assembly-explainer/gooo.workspace.json"}
		os.Exit(run(args, os.Stdout, os.Stderr))
	}
	for _, signal := range []os.Signal{os.Interrupt, syscall.SIGTERM} {
		t.Run(signal.String(), func(t *testing.T) {
			fixture := startPackageInterruptFixture(t)
			fixture.awaitTool(t)
			if err := fixture.command.Process.Signal(signal); err != nil {
				t.Fatal(err)
			}
			select {
			case <-fixture.done:
				fixture.joined = true
			case <-time.After(3 * time.Second):
				t.Fatal("CLI did not finish after interruption")
			}
			if fixture.command.ProcessState.ExitCode() != exitFailure || syscall.Kill(fixture.pid, 0) == nil {
				t.Fatalf("interruption lost CLI status or left its tool alive: exit=%d pid=%d", fixture.command.ProcessState.ExitCode(), fixture.pid)
			}
			var receipt packageExecutionReceipt
			if err := json.Unmarshal(fixture.stdout.Bytes(), &receipt); err != nil || receipt.Decision != "FAIL_CLOSED" || receipt.Error == "" {
				t.Fatalf("interruption lost its structured result: decode=%v decision=%s stderr=%s", err, receipt.Decision, fixture.stderr.String())
			}
		})
	}
}
