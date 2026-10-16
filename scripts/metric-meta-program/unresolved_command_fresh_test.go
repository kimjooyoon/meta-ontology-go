package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/unresolvedpolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/projectionbundle"
)

func TestCommandUnresolvedOriginFresh1733(t *testing.T) {
	for _, mode := range []string{"ascii", "unicode", "input-format"} {
		t.Run(mode, func(t *testing.T) {
			operation := "upstream://operation/new"
			if mode == "unicode" {
				operation = "gooo://user/연산/출처"
			}
			strategy, verification := unresolvedInputsFresh1733(t, operation)
			if mode == "input-format" {
				var formatted bytes.Buffer
				if err := json.Indent(&formatted, strategy, "", "  "); err != nil {
					t.Fatal(err)
				}
				strategy = formatted.Bytes()
			}
			work, output := t.TempDir(), t.TempDir()
			planPath, verificationPath := filepath.Join(work, "strategy.json"), filepath.Join(work, "verification.json")
			for name, raw := range map[string][]byte{planPath: strategy, verificationPath: verification} {
				if err := os.WriteFile(name, raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			var stdout, stderr bytes.Buffer
			code := run([]string{"--strategy", planPath, "--strategy-verification", verificationPath, "--out", output}, &stdout, &stderr)
			var refusal unresolvedpolicy.OperationRefusal
			if err := json.Unmarshal(stdout.Bytes(), &refusal); err != nil {
				t.Fatalf("command discarded structured origin: %s; %s; %v", stdout.String(), stderr.String(), err)
			}
			files, err := os.ReadDir(output)
			if err != nil {
				t.Fatal(err)
			}
			if code != 1 || len(files) != 0 || strings.Contains(stdout.String(), "verified:") ||
				!strings.Contains(stderr.String(), "metric-meta-program: compile:") ||
				refusal.Origin.RequestedOperation != operation ||
				refusal.Origin.Inputs.StrategyPayloadDigest != projectionbundle.Digest(strategy) ||
				refusal.Origin.Inputs.VerificationPayloadDigest != projectionbundle.Digest(verification) ||
				refusal.ReadinessCredit || refusal.RepositoryWorkspaceWrites || refusal.PromotionAuthorized {
				t.Fatalf("command changed failure, origin or output authority: %d, %+v, %v", code, refusal, files)
			}
		})
	}
}
