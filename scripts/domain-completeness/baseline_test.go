package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiscoverBaselineDownloadsExactPriorDevReceipt(t *testing.T) {
	const (
		repository = "owner/repo"
		headSHA    = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		newerSHA   = "cccccccccccccccccccccccccccccccccccccccc"
		currentSHA = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		runID      = int64(72)
		artifactID = int64(88)
	)
	baseline := Report{
		Schema: ReceiptSchema, ProfileID: ProfileID, SubjectSHA: headSHA,
		Contract:   ContractRef{SemanticHash: "same-profile"},
		Generated:  GeneratedRef{SemanticHash: "same-profile", SemanticsEqual: true},
		Snapshot:   Snapshot{Repository: repository, SubjectSHA: headSHA, WorkflowRun: runID, RunAttempt: 2},
		Dimensions: []Dimension{{ID: "use_case_coverage", MetricID: "metric", Unit: "cases", Status: "PROGRESS", Numerator: 2, Denominator: 4}},
	}
	baseline.Digest, _ = reportDigest(baseline)
	archive := testBaselineArchive(t, baseline)
	incompatible := baseline
	incompatible.SubjectSHA = newerSHA
	incompatible.Snapshot.SubjectSHA = newerSHA
	incompatible.Contract.SemanticHash = "different-profile"
	incompatible.Generated.SemanticHash = "different-profile"
	incompatible.Digest, _ = reportDigest(incompatible)
	incompatibleArchive := testBaselineArchive(t, incompatible)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("authorization header = %q", r.Header.Get("Authorization"))
		}
		switch {
		case r.URL.Path == "/repos/owner/repo/actions/workflows/language-utility-evidence.yml/runs":
			if r.URL.Query().Get("branch") != "dev" || r.URL.Query().Get("event") != "push" || r.URL.Query().Get("status") != "completed" {
				t.Errorf("run query = %s", r.URL.RawQuery)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"total_count":2,"workflow_runs":[{"id":73,"head_branch":"dev","head_sha":"` + newerSHA + `","event":"push","status":"completed","conclusion":"success","run_attempt":1},{"id":72,"head_branch":"dev","head_sha":"` + headSHA + `","event":"push","status":"completed","conclusion":"success","run_attempt":2}]}`))
		case r.URL.Path == "/repos/owner/repo/actions/runs/73/artifacts":
			w.Header().Set("Content-Type", "application/json")
			payload, _ := json.Marshal(map[string]any{"total_count": 1, "artifacts": []baselineArtifact{{
				ID: 89, Name: "language-utility-evidence-" + newerSHA, SizeInBytes: int64(len(incompatibleArchive)), Digest: artifactDigest(incompatibleArchive),
			}}})
			_, _ = w.Write(payload)
		case r.URL.Path == "/repos/owner/repo/actions/runs/72/artifacts":
			w.Header().Set("Content-Type", "application/json")
			payload, _ := json.Marshal(map[string]any{"total_count": 1, "artifacts": []baselineArtifact{{
				ID: artifactID, Name: "language-utility-evidence-" + headSHA, SizeInBytes: int64(len(archive)), Digest: artifactDigest(archive),
			}}})
			_, _ = w.Write(payload)
		case r.URL.Path == "/repos/owner/repo/actions/artifacts/89/zip":
			_, _ = w.Write(incompatibleArchive)
		case r.URL.Path == "/repos/owner/repo/actions/artifacts/88/zip":
			_, _ = w.Write(archive)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	t.Setenv("GITHUB_API_URL", server.URL)
	t.Setenv("GITHUB_TOKEN", "test-token")
	current := Report{
		Schema: ReceiptSchema, ProfileID: ProfileID, SubjectSHA: currentSHA,
		Contract:   ContractRef{SemanticHash: "same-profile"},
		Generated:  GeneratedRef{SemanticHash: "same-profile", SemanticsEqual: true},
		Snapshot:   Snapshot{Repository: repository, SubjectSHA: currentSHA, WorkflowRun: 90, RunAttempt: 1},
		Dimensions: []Dimension{{ID: "use_case_coverage", MetricID: "metric", Unit: "cases", Status: "PROGRESS", Numerator: 3, Denominator: 4}},
	}
	receipt, receiptRaw, artifact, err := discoverBaseline(context.Background(), current)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Report
	if err := json.Unmarshal(receiptRaw, &decoded); err != nil {
		t.Fatal(err)
	}
	if receipt.SubjectSHA != headSHA || receipt.Snapshot.WorkflowRun != runID || receipt.Snapshot.RunAttempt != 2 ||
		receipt.Digest != baseline.Digest || decoded.Digest != baseline.Digest || artifact.ID != artifactID || int(artifact.SizeInBytes) != len(archive) {
		t.Fatalf("selected baseline identity = %#v, artifact=%#v", receipt.Snapshot, artifact)
	}
}

func TestBaselineReceiptRejectsDuplicateOrMissingEntry(t *testing.T) {
	archive := testBaselineArchive(t, Report{})
	if _, err := baselineReceipt(archive); err != nil {
		t.Fatalf("expected one receipt entry: %v", err)
	}
	wrong := bytes.NewBuffer(nil)
	writer := zip.NewWriter(wrong)
	entry, _ := writer.Create("unrelated.json")
	_, _ = entry.Write([]byte(`{}`))
	_ = writer.Close()
	if _, err := baselineReceipt(wrong.Bytes()); err == nil {
		t.Fatal("missing receipt entry was accepted")
	}
}

func testBaselineArchive(t *testing.T, report Report) []byte {
	t.Helper()
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	buffer := bytes.NewBuffer(nil)
	writer := zip.NewWriter(buffer)
	entry, err := writer.Create("domain-completeness/receipt.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}
