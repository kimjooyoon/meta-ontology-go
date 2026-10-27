package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestCycleCommandPublishesInterruptedObservation(t *testing.T) {
	for _, row := range []struct {
		conclusion, status string
		broken             bool
		exit               int
	}{{"success", "BOUND", false, 0}, {"failure", "BOUND", false, 0},
		{"cancelled", "INTERRUPTED", false, 0}, {"cancelled", "OPEN", true, 1}} {
		t.Run(row.conclusion+row.status, func(t *testing.T) {
			opts, in := cycleFixture()
			if row.broken {
				in.Execution.Value.PlanDigest = "different-plan"
			}
			root := t.TempDir()
			arguments := []string{"-test.run=^TestCycleCommandLifecycleHelper$", "--", "-head-sha", opts.headSHA,
				"-branch", "dev", "-ci-run-id", "42", "-ci-conclusion", row.conclusion, "-check"}
			for _, input := range []struct {
				name  string
				value any
			}{{"metrics", in.Metrics.Value}, {"plan", in.Plan.Value}, {"execution", in.Execution.Value},
				{"receipts", in.Receipts.Value}, {"provenance", in.Provenance.Value}, {"contract", in.Contract.Value}} {
				raw, err := json.Marshal(input.value)
				if err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(root, input.name+".json")
				if err := os.WriteFile(path, raw, 0600); err != nil {
					t.Fatal(err)
				}
				arguments = append(arguments, "-"+input.name, path)
			}
			binary, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, binary, arguments...)
			command.Env = append(os.Environ(), "GOOO_TEST_CYCLE_HELPER=1")
			output, err := command.CombinedOutput()
			if command.ProcessState == nil || command.ProcessState.ExitCode() != row.exit {
				t.Fatal("cycle command outcome", err, string(output))
			}
			var result Envelope
			if err := json.Unmarshal(output, &result); err != nil || result.Status != row.status ||
				result.CIConclusion != row.conclusion || result.PromotionAuthorized {
				t.Fatal("cycle command lost its source outcome", err, string(output))
			}
		})
	}
}

func TestCycleCommandLifecycleHelper(t *testing.T) {
	if os.Getenv("GOOO_TEST_CYCLE_HELPER") != "1" {
		return
	}
	os.Args = append([]string{"cycle"}, os.Args[3:]...)
	flag.CommandLine = flag.NewFlagSet("cycle", flag.ExitOnError)
	main()
	os.Exit(0)
}
