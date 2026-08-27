package reportconsumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/judge"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
)

type entry struct {
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

const artifactSemanticProjectionSchema = "gooo/invariant-transformation-artifact-semantic-projection/v1"

// ArtifactSemanticValue contains only fields whose meaning is supplied by the
// artifact bytes. Execution, authorization, and filesystem identity are kept
// in ArtifactSemanticProjection as provenance and are deliberately not part of
// this semantic value.
type ArtifactSemanticValue struct {
	CaseID               string `json:"case_id"`
	Input                int64  `json:"input"`
	Operation            string `json:"operation"`
	Output               int64  `json:"output"`
	SourceDigest         string `json:"source_digest"`
	SemanticSourceDigest string `json:"semantic_source_digest"`
	SubjectSHA           string `json:"subject_sha"`
}

// ArtifactSemanticProjection is an independent observation of the generated
// artifact. Raw path, bytes, and execution identity are retained so the
// semantic projection remains provenance-bound; only the explicitly dynamic
// provenance fields may be normalized for replay comparison.
type ArtifactSemanticProjection struct {
	Schema                 string                `json:"schema"`
	HeadSHA                string                `json:"head_sha"`
	CaseID                 string                `json:"case_id"`
	Path                   string                `json:"path"`
	RawDigest              string                `json:"raw_digest"`
	RawSize                int                   `json:"raw_size"`
	ExecutionID            string                `json:"execution_id"`
	AuthorizationDigest    string                `json:"authorization_digest,omitempty"`
	EffectDigest           string                `json:"effect_digest,omitempty"`
	Semantic               ArtifactSemanticValue `json:"semantic"`
	CanonicalSemanticBytes string                `json:"canonical_semantic_bytes"`
	SemanticDigest         string                `json:"semantic_digest"`
}

type artifactFields struct {
	CaseID               string
	ExecutionID          string
	Input                int64
	Operation            string
	Output               int64
	SourceDigest         string
	SemanticSourceDigest string
	AuthorizationDigest  string
	SubjectSHA           string
}

// ArtifactProjection first validates the independently consumed report and
// then binds an independently parsed artifact file to the approved effect.
// It never obtains artifact meaning from producer code or report metadata.
func ArtifactProjection(report model.Report, source []byte, headSHA string) (ArtifactSemanticProjection, error) {
	if err := judge.ValidateReport(report, source); err != nil {
		return ArtifactSemanticProjection{}, fmt.Errorf("ARTIFACT_OBSERVATION/validate-report/REPORT_NOT_INDEPENDENTLY_VALID: %w", err)
	}
	var approved *model.Effect
	var approvedReceipt *model.Receipt
	for index := range report.Cases {
		for effectIndex := range report.Cases[index].Receipt.Effects {
			effect := &report.Cases[index].Receipt.Effects[effectIndex]
			if effect.Kind != model.EffectApproved {
				continue
			}
			if approved != nil {
				return ArtifactSemanticProjection{}, fmt.Errorf("ARTIFACT_OBSERVATION/select-approved-effect/MULTIPLE_APPROVED_ARTIFACTS")
			}
			approved = effect
			approvedReceipt = &report.Cases[index].Receipt
		}
	}
	if approved == nil || approvedReceipt == nil {
		return ArtifactSemanticProjection{}, fmt.Errorf("ARTIFACT_OBSERVATION/select-approved-effect/APPROVED_ARTIFACT_MISSING")
	}
	projection, err := ProjectArtifactFile(approved.Artifact.Path, headSHA)
	if err != nil {
		return ArtifactSemanticProjection{}, err
	}
	if projection.CaseID != approved.CaseID || projection.ExecutionID != approved.ExecutionID ||
		projection.RawDigest != approved.Artifact.ContentDigest || projection.RawSize != approved.Artifact.Size ||
		projection.Semantic.CaseID != approvedReceipt.CaseID || projection.Semantic.Input != approvedReceipt.Evidence.InputValue ||
		projection.Semantic.Operation != approvedReceipt.Evidence.CandidateOperation || projection.Semantic.Output != approvedReceipt.Evidence.CandidateResult ||
		projection.Semantic.SourceDigest != report.SourceDigest || projection.Semantic.SemanticSourceDigest != report.SemanticSourceDigest ||
		projection.Semantic.SubjectSHA != headSHA || projection.ExecutionID != report.ExecutionID {
		return ArtifactSemanticProjection{}, fmt.Errorf("ARTIFACT_OBSERVATION/bind-effect/ARTIFACT_SEMANTIC_PROVENANCE_MISMATCH")
	}
	projection.AuthorizationDigest = approved.Artifact.AuthorizationDigest
	projection.EffectDigest = approved.ExecutionReceiptDigest
	return projection, nil
}

// ProjectArtifactFile reads and parses the actual artifact bytes. It is
// intentionally usable by a separate command for tamper regressions without
// importing the producer or trusting an expected artifact digest.
func ProjectArtifactFile(path, headSHA string) (ArtifactSemanticProjection, error) {
	if !model.ValidHead(headSHA) || path == "" || !allowedSnapshotPath(path) {
		return ArtifactSemanticProjection{}, fmt.Errorf("ARTIFACT_OBSERVATION/read/ARTIFACT_PATH_OUTSIDE_SAFE_TEMP_ROOT")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return ArtifactSemanticProjection{}, fmt.Errorf("ARTIFACT_OBSERVATION/read/ARTIFACT_BYTES_UNAVAILABLE: %w", err)
	}
	fields, err := parseArtifactBytes(raw)
	if err != nil {
		return ArtifactSemanticProjection{}, fmt.Errorf("ARTIFACT_OBSERVATION/parse/ARTIFACT_SEMANTIC_BYTES_INVALID: %w", err)
	}
	if !model.ValidExecutionID(headSHA, fields.ExecutionID) || fields.CaseID == "" || !model.ValidDigest(fields.SourceDigest) || !model.ValidDigest(fields.SemanticSourceDigest) || !model.ValidDigest(fields.AuthorizationDigest) || fields.SubjectSHA != headSHA {
		return ArtifactSemanticProjection{}, fmt.Errorf("ARTIFACT_OBSERVATION/parse/ARTIFACT_PROVENANCE_INVALID")
	}
	semantic := ArtifactSemanticValue{
		CaseID: fields.CaseID, Input: fields.Input, Operation: fields.Operation, Output: fields.Output,
		SourceDigest: fields.SourceDigest, SemanticSourceDigest: fields.SemanticSourceDigest, SubjectSHA: fields.SubjectSHA,
	}
	canonical, err := json.Marshal(semantic)
	if err != nil {
		return ArtifactSemanticProjection{}, fmt.Errorf("ARTIFACT_SEMANTIC_PROJECTION/encode/CANONICAL_SEMANTIC_BYTES_INVALID: %w", err)
	}
	return ArtifactSemanticProjection{
		Schema: artifactSemanticProjectionSchema, HeadSHA: headSHA, CaseID: fields.CaseID, Path: path,
		RawDigest: model.DigestBytes(raw), RawSize: len(raw), ExecutionID: fields.ExecutionID,
		Semantic: semantic, CanonicalSemanticBytes: string(canonical), SemanticDigest: model.DigestBytes(canonical),
	}, nil
}

func parseArtifactBytes(raw []byte) (artifactFields, error) {
	lines := strings.Split(string(raw), "\n")
	if len(lines) != 11 || lines[0] != "gooo bounded transformation artifact" || lines[len(lines)-1] != "" {
		return artifactFields{}, fmt.Errorf("unexpected artifact line framing")
	}
	keys := []string{"case", "execution", "input", "operation", "output", "source", "semantic-source", "authorization", "subject"}
	values := make(map[string]string, len(keys))
	for index, key := range keys {
		actualKey, value, ok := strings.Cut(lines[index+1], "=")
		if !ok || actualKey != key || value == "" || values[key] != "" {
			return artifactFields{}, fmt.Errorf("invalid %s field", key)
		}
		values[key] = value
	}
	input, err := strconv.ParseInt(values["input"], 10, 64)
	if err != nil {
		return artifactFields{}, fmt.Errorf("invalid input: %w", err)
	}
	output, err := strconv.ParseInt(values["output"], 10, 64)
	if err != nil {
		return artifactFields{}, fmt.Errorf("invalid output: %w", err)
	}
	return artifactFields{
		CaseID: values["case"], ExecutionID: values["execution"], Input: input, Operation: values["operation"], Output: output,
		SourceDigest: values["source"], SemanticSourceDigest: values["semantic-source"], AuthorizationDigest: values["authorization"], SubjectSHA: values["subject"],
	}, nil
}

// CompareArtifactSemanticProjection compares semantic meaning and fixed
// provenance only. Path, raw bytes, execution ID, authorization digest, and
// effect digest are allowed to vary between executions and remain available in
// each source projection for binding/audit.
func CompareArtifactSemanticProjection(expected, actual ArtifactSemanticProjection) error {
	if expected.Schema != artifactSemanticProjectionSchema || actual.Schema != artifactSemanticProjectionSchema ||
		expected.HeadSHA != actual.HeadSHA || expected.CaseID != actual.CaseID ||
		!reflect.DeepEqual(expected.Semantic, actual.Semantic) || expected.CanonicalSemanticBytes != actual.CanonicalSemanticBytes ||
		expected.SemanticDigest != actual.SemanticDigest {
		return fmt.Errorf("ARTIFACT_SEMANTIC_PROJECTION/compare/ARTIFACT_SEMANTIC_MISMATCH")
	}
	return nil
}

// Bind is the independent report consumer boundary. It reads the source and
// both raw repository-entry artifacts, binds them to the exact witness run,
// and only then accepts the report through the independent judge.
func Bind(reportRaw, beforeRaw, afterRaw, source []byte, headSHA string) (model.Report, error) {
	var report model.Report
	if err := json.Unmarshal(reportRaw, &report); err != nil {
		return model.Report{}, fmt.Errorf("REPORT_BINDING/parse-report/REPORT_JSON_INVALID: %w", err)
	}
	if report.HeadSHA != headSHA || !model.ValidHead(headSHA) || !model.ValidExecutionID(headSHA, report.ExecutionID) {
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
	report.Summary.RepositoryNetContentObserved = true
	report.Summary.RepositoryNetContentUnchanged = true
	report.Summary.RepositoryNetStatusObserved = false
	report.Summary.RepositoryNetStatusUnchanged = false
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
	if snapshot.Schema != model.RepositorySnapshotSchema || snapshot.HeadSHA != headSHA || !model.ValidExecutionID(headSHA, snapshot.ExecutionID) || snapshot.ExecutionID != executionID {
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
