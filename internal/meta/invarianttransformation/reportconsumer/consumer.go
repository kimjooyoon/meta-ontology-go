package reportconsumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/judge"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
)

type entry struct {
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

// Bind is the independent report consumer boundary. It reads the source and
// both raw repository-entry artifacts, binds them to the exact witness run,
// and only then accepts the report through the independent judge.
func Bind(reportRaw, beforeRaw, afterRaw, source []byte, headSHA string) (model.Report, error) {
	var report model.Report
	if err := json.Unmarshal(reportRaw, &report); err != nil {
		return model.Report{}, fmt.Errorf("REPORT_BINDING/parse-report/REPORT_JSON_INVALID: %w", err)
	}
	if report.HeadSHA != headSHA || !model.ValidHead(headSHA) || report.ExecutionID == "" {
		return model.Report{}, fmt.Errorf("REPORT_BINDING/bind-execution-id/EXECUTION_ID_OR_HEAD_MISMATCH")
	}
	if report.Digest == "" || report.Digest != model.SealReport(report).Digest {
		return model.Report{}, fmt.Errorf("REPORT_BINDING/parse-report/REPORT_DIGEST_INVALID")
	}
	before, beforeEntries, err := readSnapshot(beforeRaw, report.HeadSHA, report.ExecutionID, "before")
	if err != nil {
		return model.Report{}, err
	}
	after, afterEntries, err := readSnapshot(afterRaw, report.HeadSHA, report.ExecutionID, "after")
	if err != nil {
		return model.Report{}, err
	}
	if !reflect.DeepEqual(beforeEntries, afterEntries) {
		return model.Report{}, fmt.Errorf("REPOSITORY_OBSERVATION/adjudicate/NET_REPOSITORY_CONTENT_STATE_CHANGED")
	}
	originalDigest := report.Digest
	observation := model.RepositoryObservation{
		Before: before, After: after, Observed: true, State: model.RepositoryNetContentStateUnchanged,
		ExecutionID: report.ExecutionID, WitnessReportDigest: originalDigest,
	}
	for _, item := range report.Cases {
		if item.Receipt.ExecutionID != report.ExecutionID {
			return model.Report{}, fmt.Errorf("REPORT_BINDING/bind-receipt/EXECUTION_ID_MISMATCH")
		}
		observation.WitnessReceiptDigests = append(observation.WitnessReceiptDigests, item.Receipt.Digest)
		for _, effect := range item.Receipt.Effects {
			if effect.ExecutionID != report.ExecutionID || effect.Artifact.ExecutionID != report.ExecutionID {
				return model.Report{}, fmt.Errorf("REPORT_BINDING/bind-effect/EXECUTION_ID_MISMATCH")
			}
			observation.WitnessEffectDigests = append(observation.WitnessEffectDigests, effect.ExecutionReceiptDigest)
			observation.WitnessArtifactDigests = append(observation.WitnessArtifactDigests, effect.Artifact.ContentDigest)
		}
	}
	report.RepositoryObservation = observation
	report.Summary.RepositoryNetStatusObserved = true
	report.Summary.RepositoryNetStatusUnchanged = true
	report.Summary.RepositoryNetContentState = model.RepositoryNetContentStateUnchanged
	report.Summary.RepositoryNetSnapshotObservations = 1
	report.Summary.RepositoryNetSnapshotDenominator = 1
	report.Summary.RepositoryPathAuthorization = false
	report.Summary.RepositoryActualOrTransientWrites = model.UnknownEffectScope
	report.Summary.RepositoryWrites = -1
	report.Summary.AmbientProcessAuthority = model.UnknownEffectScope
	report.Indicators = judge.Indicators(report.Summary)
	report = model.SealReport(report)
	if err := judge.ValidateReport(report, source); err != nil {
		return model.Report{}, fmt.Errorf("REPORT_BINDING/independent-judge/%w", err)
	}
	return report, nil
}

func readSnapshot(raw []byte, headSHA, executionID, label string) (model.RepositorySnapshot, []entry, error) {
	var snapshot model.RepositorySnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return model.RepositorySnapshot{}, nil, fmt.Errorf("REPOSITORY_SNAPSHOT/%s/parse-metadata/SNAPSHOT_JSON_INVALID: %w", label, err)
	}
	if snapshot.Schema != model.RepositorySnapshotSchema || snapshot.HeadSHA != headSHA || snapshot.ExecutionID != executionID {
		return model.RepositorySnapshot{}, nil, fmt.Errorf("REPOSITORY_BINDING/compare-execution-id/EXECUTION_ID_MISMATCH")
	}
	if !filepath.IsAbs(snapshot.EntriesPath) || snapshot.EntriesPath == "" || !allowedSnapshotPath(snapshot.EntriesPath) || !model.ValidDigest(snapshot.EntriesDigest) || !model.ValidDigest(snapshot.PathDigest) || snapshot.EntryCount < 0 {
		return model.RepositorySnapshot{}, nil, fmt.Errorf("REPOSITORY_SNAPSHOT/%s/validate-metadata/SNAPSHOT_SUMMARY_INVALID", label)
	}
	entriesRaw, err := os.ReadFile(snapshot.EntriesPath)
	if err != nil {
		return model.RepositorySnapshot{}, nil, fmt.Errorf("REPOSITORY_SNAPSHOT/%s/load-raw-entries/RAW_ENTRIES_UNAVAILABLE: %w", label, err)
	}
	if model.DigestBytes(entriesRaw) != snapshot.EntriesDigest {
		return model.RepositorySnapshot{}, nil, fmt.Errorf("REPOSITORY_SNAPSHOT/%s/load-raw-entries/RAW_ENTRIES_DIGEST_MISMATCH", label)
	}
	var entries []entry
	if err := json.Unmarshal(entriesRaw, &entries); err != nil {
		return model.RepositorySnapshot{}, nil, fmt.Errorf("REPOSITORY_SNAPSHOT/%s/reconstruct-entries/RAW_ENTRIES_JSON_INVALID: %w", label, err)
	}
	canonical, err := json.Marshal(entries)
	if err != nil || !bytes.Equal(entriesRaw, append(canonical, '\n')) {
		return model.RepositorySnapshot{}, nil, fmt.Errorf("REPOSITORY_SNAPSHOT/%s/reconstruct-entries/RAW_ENTRY_CANONICAL_BYTES_MISMATCH", label)
	}
	if len(entries) != snapshot.EntryCount {
		return model.RepositorySnapshot{}, nil, fmt.Errorf("REPOSITORY_SNAPSHOT/%s/reconstruct-entries/ENTRY_COUNT_MISMATCH", label)
	}
	for index, item := range entries {
		if item.Path == "" || filepath.IsAbs(filepath.FromSlash(item.Path)) || item.Path != filepath.ToSlash(filepath.Clean(filepath.FromSlash(item.Path))) || strings.HasPrefix(item.Path, "../") || item.Path == ".." || !model.ValidDigest(item.Digest) {
			return model.RepositorySnapshot{}, nil, fmt.Errorf("REPOSITORY_SNAPSHOT/%s/reconstruct-entries/RAW_ENTRY_DIGEST_INVALID", label)
		}
		if index > 0 && entries[index-1].Path >= item.Path {
			return model.RepositorySnapshot{}, nil, fmt.Errorf("REPOSITORY_SNAPSHOT/%s/reconstruct-entries/RAW_ENTRY_ORDER_INVALID", label)
		}
	}
	if model.Digest(entries) != snapshot.PathDigest {
		return model.RepositorySnapshot{}, nil, fmt.Errorf("REPOSITORY_SNAPSHOT/%s/reconstruct-entries/PATH_CONTENT_DIGEST_MISMATCH", label)
	}
	return snapshot, entries, nil
}

func allowedSnapshotPath(path string) bool {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	root := snapshotTempRoot()
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return false
	}
	canonicalRoot, err = filepath.Abs(canonicalRoot)
	if err != nil || !withinPath(canonicalRoot, absolute) {
		return false
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return false
	}
	resolved, err = filepath.Abs(resolved)
	return err == nil && withinPath(canonicalRoot, resolved)
}

func snapshotTempRoot() string {
	if root := os.Getenv("RUNNER_TEMP"); root != "" {
		return root
	}
	return os.TempDir()
}

func withinPath(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != "." && relative != ".." && !filepath.IsAbs(relative) && !strings.HasPrefix(relative, ".."+string(os.PathSeparator))
}

// EqualRawEntries is intentionally exported for CI-side negative fixtures and
// compares reconstructed path/content entries, not only summary digests.
func EqualRawEntries(left, right []byte) error {
	var a, b []entry
	if err := json.Unmarshal(left, &a); err != nil {
		return fmt.Errorf("REPOSITORY_SNAPSHOT/compare-entries/LEFT_INVALID")
	}
	if err := json.Unmarshal(right, &b); err != nil {
		return fmt.Errorf("REPOSITORY_SNAPSHOT/compare-entries/RIGHT_INVALID")
	}
	sort.Slice(a, func(i, j int) bool { return a[i].Path < a[j].Path })
	sort.Slice(b, func(i, j int) bool { return b[i].Path < b[j].Path })
	if !reflect.DeepEqual(a, b) {
		return fmt.Errorf("REPOSITORY_OBSERVATION/compare-entries/CONTENT_ENTRY_MISMATCH")
	}
	return nil
}
