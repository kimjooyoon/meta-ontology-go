package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/promotioncontinuity"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/rollbackfixedpoint"
)

func readAppliedRecoveryFixture(t *testing.T) (rollbackfixedpoint.Report, []byte) {
	t.Helper()
	path := "../../internal/meta/languagereadiness/rollbackfixedpoint/testdata/applied-original-8f45c78e.json.gz"
	if root := os.Getenv("GOOO_ROLLBACK_TEST_DRAFT"); root != "" {
		path = filepath.Join(root, "internal/meta/languagereadiness/rollbackfixedpoint/testdata/applied-original-8f45c78e.json.gz")
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	raw, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	var original rollbackfixedpoint.Report
	if err := json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	return original, raw
}

func appliedContinuityCLIInputs(t *testing.T) ([]string, string) {
	t.Helper()
	original, raw := readAppliedRecoveryFixture(t)
	root := t.TempDir()
	recoveryPath := filepath.Join(root, "recovery.json")
	guardPath := filepath.Join(root, "guard.json")
	outputPath := filepath.Join(root, "continuity.json")
	if err := os.WriteFile(recoveryPath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	g := original.Source.Guard
	// The guard projection is synthetic. Recovery keeps the original public bytes.
	guard := map[string]any{"schema": "gooo/autonomy-guarded-promotion/v1", "decision": g.Decision,
		"reason": g.Reason, "resolution": g.Resolution, "report_digest": g.ReportDigest,
		"source": map[string]any{"current_head_sha": g.HeadSHA},
		"summary": map[string]any{"satisfied": g.Satisfied, "total": g.Total, "unresolved": g.Unresolved,
			"repository_writes": 0, "readiness_promotion_authorized": false, "repository_mutation_authorized": false}}
	guardBytes, err := json.Marshal(guard)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(guardPath, guardBytes, 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"--head", g.HeadSHA, "--guard", guardPath, "--recovery", recoveryPath, "--out", outputPath}
	return args, outputPath
}

func TestCLIRecordsAppliedTerminalWithoutPromotion(t *testing.T) {
	args, outputPath := appliedContinuityCLIInputs(t)
	var output bytes.Buffer
	if err := run(args, &output); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	var report promotioncontinuity.Report
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if !promotioncontinuity.IsKnownNonPromotingTerminal(report) || report.Decision != "FAIL_CLOSED" ||
		report.RepositoryMutationAuthorized || report.Summary.AuthorizedRecoveryRoutes != 0 ||
		report.Source.Recovery.TransformationEffects != 2 || report.Source.Recovery.RecoveredFixedPoints != 0 {
		t.Fatalf("CLI changed non-promoting outcome: %s", data)
	}
	if err := run(args, &output); err != nil {
		t.Fatal(err)
	}
	replayed, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, replayed) {
		t.Fatal("consumer classification did not replay")
	}
}
