package reportconsumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	SemanticSourceDigest string `json:"semantic_source_digest"`
}

// ArtifactSemanticProjection is an independent observation of the generated
// artifact. Raw path, bytes, and execution identity are retained so the
// semantic projection remains provenance-bound; only the explicitly dynamic
// provenance fields may be normalized for replay comparison.
type ArtifactSemanticProjection struct {
	Schema                      string                `json:"schema"`
	HeadSHA                     string                `json:"head_sha"`
	CaseID                      string                `json:"case_id"`
	Path                        string                `json:"path"`
	RawDigest                   string                `json:"raw_digest"`
	RawSize                     int                   `json:"raw_size"`
	ExecutionID                 string                `json:"execution_id"`
	SourceDigest                string                `json:"source_digest"`
	SubjectSHA                  string                `json:"subject_sha"`
	ObservedAuthorizationDigest string                `json:"observed_authorization_digest"`
	ExpectedAuthorizationDigest string                `json:"expected_authorization_digest"`
	EffectDigest                string                `json:"effect_digest"`
	Semantic                    ArtifactSemanticValue `json:"semantic"`
	CanonicalSemanticBytes      string                `json:"canonical_semantic_bytes"`
	SemanticDigest              string                `json:"semantic_digest"`
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
		projection.SourceDigest != report.SourceDigest || projection.Semantic.SemanticSourceDigest != report.SemanticSourceDigest ||
		projection.SubjectSHA != headSHA || projection.ExecutionID != report.ExecutionID {
		return ArtifactSemanticProjection{}, fmt.Errorf("ARTIFACT_OBSERVATION/bind-effect/ARTIFACT_SEMANTIC_PROVENANCE_MISMATCH")
	}
	projection.ExpectedAuthorizationDigest = approved.Artifact.AuthorizationDigest
	if projection.ObservedAuthorizationDigest != projection.ExpectedAuthorizationDigest {
		return ArtifactSemanticProjection{}, fmt.Errorf("ARTIFACT_OBSERVATION/bind-effect/ARTIFACT_AUTHORIZATION_DIGEST_MISMATCH")
	}
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
		SemanticSourceDigest: fields.SemanticSourceDigest,
	}
	canonical, err := json.Marshal(semantic)
	if err != nil {
		return ArtifactSemanticProjection{}, fmt.Errorf("ARTIFACT_SEMANTIC_PROJECTION/encode/CANONICAL_SEMANTIC_BYTES_INVALID: %w", err)
	}
	return ArtifactSemanticProjection{
		Schema: artifactSemanticProjectionSchema, HeadSHA: headSHA, CaseID: fields.CaseID, Path: path,
		RawDigest: model.DigestBytes(raw), RawSize: len(raw), ExecutionID: fields.ExecutionID,
		SourceDigest: fields.SourceDigest, SubjectSHA: fields.SubjectSHA, ObservedAuthorizationDigest: fields.AuthorizationDigest,
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

// ValidateArtifactProjectionBinding proves that an expected projection was
// produced from the same baseline artifact observation. Semantic equality is
// intentionally separate from this exact raw/provenance binding check.
func ValidateArtifactProjectionBinding(expected, actual ArtifactSemanticProjection) error {
	if !reflect.DeepEqual(expected, actual) {
		return fmt.Errorf("ARTIFACT_SEMANTIC_PROJECTION/bind-expected/EXPECTED_PROJECTION_NOT_BASELINE_BOUND")
	}
	return nil
}

// CompareArtifactProvenance checks the raw and execution-bound identity that
// semantic comparison intentionally excludes.
func CompareArtifactProvenance(expected, actual ArtifactSemanticProjection) error {
	if expected.HeadSHA != actual.HeadSHA || expected.CaseID != actual.CaseID || expected.Path != actual.Path ||
		expected.RawDigest != actual.RawDigest || expected.RawSize != actual.RawSize || expected.ExecutionID != actual.ExecutionID ||
		expected.SourceDigest != actual.SourceDigest || expected.SubjectSHA != actual.SubjectSHA ||
		expected.ObservedAuthorizationDigest != actual.ObservedAuthorizationDigest ||
		expected.ExpectedAuthorizationDigest != actual.ExpectedAuthorizationDigest || expected.EffectDigest != actual.EffectDigest {
		return fmt.Errorf("ARTIFACT_OBSERVATION/compare/ARTIFACT_PROVENANCE_MISMATCH")
	}
	return nil
}

// DecodeArtifactSemanticProjection rejects unknown fields and trailing JSON so
// an expected projection cannot silently carry an unreviewed authority field.
func DecodeArtifactSemanticProjection(raw []byte) (ArtifactSemanticProjection, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var projection ArtifactSemanticProjection
	if err := decoder.Decode(&projection); err != nil {
		return ArtifactSemanticProjection{}, fmt.Errorf("ARTIFACT_SEMANTIC_PROJECTION/parse/EXPECTED_PROJECTION_JSON_INVALID: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return ArtifactSemanticProjection{}, fmt.Errorf("ARTIFACT_SEMANTIC_PROJECTION/parse/EXPECTED_PROJECTION_TRAILING_JSON")
		}
		return ArtifactSemanticProjection{}, fmt.Errorf("ARTIFACT_SEMANTIC_PROJECTION/parse/EXPECTED_PROJECTION_TRAILING_JSON: %w", err)
	}
	return projection, nil
}

const closureReceiptSchema = "gooo/invariant-transformation-closure-receipt/v2"

const (
	closureMetricArtifactBytesFirst          = "artifact-bytes/first"
	closureMetricArtifactBytesSecond         = "artifact-bytes/second"
	closureMetricSemanticEquality            = "semantic-equality/pair"
	closureMetricRawProvenanceFirst          = "raw-provenance/first"
	closureMetricRawProvenanceSecond         = "raw-provenance/second"
	closureMetricAuthorizationFirst          = "authorization/first"
	closureMetricAuthorizationSecond         = "authorization/second"
	closureMetricOutputTamper                = "output-tamper"
	closureMetricAuthorizationTamper         = "authorization-tamper"
	closureMetricCommentOnly                 = "comment-only/semantic-preservation"
	closureMetricFinalClosure                = "final-closure"
	closureMetricInventoryExpected           = 11
	closureMetricArtifactBytesExpected       = 2
	closureMetricSemanticEqualityExpected    = 1
	closureMetricRawProvenanceExpected       = 2
	closureMetricAuthorizationExpected       = 2
	closureMetricOutputTamperExpected        = 1
	closureMetricAuthorizationTamperExpected = 1
	closureMetricCommentOnlyExpected         = 1
	closureMetricFinalClosureExpected        = 1
)

const (
	closureProducerExecutor     = model.ExecutorID
	closureArtifactConsumer     = "invarianttransformation.artifact-semantic-consumer"
	closureInterventionConsumer = "invarianttransformation.intervention-consumer"
	closureClosureConsumer      = "invarianttransformation.closure-consumer"
)

var closureMetricIDs = []string{
	closureMetricArtifactBytesFirst,
	closureMetricArtifactBytesSecond,
	closureMetricSemanticEquality,
	closureMetricRawProvenanceFirst,
	closureMetricRawProvenanceSecond,
	closureMetricAuthorizationFirst,
	closureMetricAuthorizationSecond,
	closureMetricOutputTamper,
	closureMetricAuthorizationTamper,
	closureMetricCommentOnly,
	closureMetricFinalClosure,
}

// ClosureMetricEvidence is one addressed relation in the fixed closure
// inventory. Its observed digest commits a relation-specific canonical payload;
// scalar counters are derived from these rows and never act as evidence.
type ClosureMetricEvidence struct {
	MetricID               string `json:"metric_id"`
	Occurrence             string `json:"occurrence"`
	Address                string `json:"address"`
	Producer               string `json:"producer"`
	IndependentConsumer    string `json:"independent_consumer"`
	MetaOperation          string `json:"meta_operation"`
	ProofChoice            string `json:"proof_choice"`
	TargetAddress          string `json:"target_address"`
	TargetDigest           string `json:"target_digest"`
	ObservedEvidenceDigest string `json:"observed_evidence_digest"`
	Decision               string `json:"decision"`
	Resolution             string `json:"resolution"`
	Stage                  string `json:"stage"`
	Step                   string `json:"step"`
	Reason                 string `json:"reason"`
}

type closureMetricDigestPayload struct {
	MetricID            string `json:"metric_id"`
	Occurrence          string `json:"occurrence"`
	Address             string `json:"address"`
	Producer            string `json:"producer"`
	IndependentConsumer string `json:"independent_consumer"`
	MetaOperation       string `json:"meta_operation"`
	ProofChoice         string `json:"proof_choice"`
	TargetAddress       string `json:"target_address"`
	TargetDigest        string `json:"target_digest"`
	Decision            string `json:"decision"`
	Resolution          string `json:"resolution"`
	Stage               string `json:"stage"`
	Step                string `json:"step"`
	Reason              string `json:"reason"`
	Evidence            any    `json:"evidence"`
}

// TypedTamperReceipt records the exact single semantic/provenance field that
// a negative artifact fixture changed, while retaining both raw byte digests.
type TypedTamperReceipt struct {
	Schema              string `json:"schema"`
	Kind                string `json:"kind"`
	CaseID              string `json:"case_id"`
	ExecutionID         string `json:"execution_id"`
	BaselinePath        string `json:"baseline_path"`
	TamperedPath        string `json:"tampered_path"`
	BaselineRawDigest   string `json:"baseline_raw_digest"`
	TamperedRawDigest   string `json:"tampered_raw_digest"`
	ChangedField        string `json:"changed_field"`
	BaselineValue       string `json:"baseline_value"`
	TamperedValue       string `json:"tampered_value"`
	SemanticDigestEqual bool   `json:"semantic_digest_equal"`
	Rejected            bool   `json:"rejected"`
	Decision            string `json:"decision"`
	Resolution          string `json:"resolution"`
	Stage               string `json:"stage"`
	Step                string `json:"step"`
	Reason              string `json:"reason"`
	EvidenceDigest      string `json:"evidence_digest"`
	Digest              string `json:"digest"`
}

type ArtifactClosureEvidence struct {
	Path                        string `json:"path"`
	RawDigest                   string `json:"raw_digest"`
	RawSize                     int    `json:"raw_size"`
	ExecutionID                 string `json:"execution_id"`
	SourceDigest                string `json:"source_digest"`
	SubjectSHA                  string `json:"subject_sha"`
	ObservedAuthorizationDigest string `json:"observed_authorization_digest"`
	ExpectedAuthorizationDigest string `json:"expected_authorization_digest"`
	EffectDigest                string `json:"effect_digest"`
	SemanticDigest              string `json:"semantic_digest"`
	CanonicalSemanticBytes      string `json:"canonical_semantic_bytes"`
}

// ClosureReceipt is the final, independently reconstructed gate. A bound
// report remains preliminary until this receipt binds the exact 11-row metric
// inventory, actual artifact observations, intervention receipt, and tamper
// receipts.
type ClosureReceipt struct {
	Schema                                 string                  `json:"schema"`
	HeadSHA                                string                  `json:"head_sha"`
	PreliminaryDecision                    string                  `json:"preliminary_decision"`
	PreliminaryResolution                  string                  `json:"preliminary_resolution"`
	PreliminaryReason                      string                  `json:"preliminary_reason"`
	PreliminaryDecisionScope               string                  `json:"preliminary_decision_scope"`
	FirstReportDigest                      string                  `json:"first_report_digest"`
	SecondReportDigest                     string                  `json:"second_report_digest"`
	FirstArtifact                          ArtifactClosureEvidence `json:"first_artifact"`
	SecondArtifact                         ArtifactClosureEvidence `json:"second_artifact"`
	OutputTamperReceipt                    TypedTamperReceipt      `json:"output_tamper_receipt"`
	AuthorizationTamperReceipt             TypedTamperReceipt      `json:"authorization_tamper_receipt"`
	MetricEvidence                         []ClosureMetricEvidence `json:"metric_evidence"`
	ExpectedMetricEvidence                 int                     `json:"expected_metric_evidence"`
	ObservedMetricEvidence                 int                     `json:"observed_metric_evidence"`
	MetricInventoryDigest                  string                  `json:"metric_inventory_digest"`
	ArtifactBytesObserved                  int                     `json:"artifact_bytes_observed"`
	ExpectedArtifactBytes                  int                     `json:"expected_artifact_bytes"`
	RawProvenanceBindings                  int                     `json:"raw_provenance_bindings"`
	ExpectedRawProvenanceBindings          int                     `json:"expected_raw_provenance_bindings"`
	AuthorizationBindings                  int                     `json:"authorization_bindings"`
	ExpectedAuthorizationBindings          int                     `json:"expected_authorization_bindings"`
	SemanticEqualityObserved               int                     `json:"semantic_equality_observed"`
	ExpectedSemanticEquality               int                     `json:"expected_semantic_equality"`
	OutputSemanticTamperRejected           int                     `json:"output_semantic_tamper_rejected"`
	ExpectedOutputSemanticTamperRejections int                     `json:"expected_output_semantic_tamper_rejections"`
	AuthorizationTamperRejected            int                     `json:"authorization_tamper_rejected"`
	ExpectedAuthorizationTamperRejections  int                     `json:"expected_authorization_tamper_rejections"`
	CommentOnlyPreservationObserved        int                     `json:"comment_only_preservation_observed"`
	ExpectedCommentOnlyPreservation        int                     `json:"expected_comment_only_preservation"`
	FinalClosureGateObserved               int                     `json:"final_closure_gate_observed"`
	ExpectedFinalClosureGate               int                     `json:"expected_final_closure_gate"`
	Decision                               string                  `json:"decision"`
	Resolution                             string                  `json:"resolution"`
	Reason                                 string                  `json:"reason"`
	Digest                                 string                  `json:"digest"`
}

type closureMetricCounts struct {
	ArtifactBytesObserved, RawProvenanceBindings, AuthorizationBindings                 int
	SemanticEqualityObserved, OutputSemanticTamperRejected, AuthorizationTamperRejected int
	CommentOnlyPreservationObserved, FinalClosureGateObserved                           int
}

// Close reconstructs both reports and both artifact projections from the
// source boundary, then verifies the intervention consumer receipt and typed
// negative artifact cases. The final PASS is produced only after the exact
// eleven addressed evidence rows have been derived and validated.
func Close(firstReport, secondReport model.Report, firstExpected, secondExpected ArtifactSemanticProjection, source []byte, headSHA, outputTamperPath, authorizationTamperPath string, interventionReportRaw, interventionConsumerRaw []byte) (ClosureReceipt, error) {
	if !model.ValidHead(headSHA) || firstReport.DecisionScope != model.PreliminaryDecisionScope || secondReport.DecisionScope != model.PreliminaryDecisionScope {
		return ClosureReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/bind-reports/PRELIMINARY_SCOPE_INVALID")
	}
	firstActual, err := ArtifactProjection(firstReport, source, headSHA)
	if err != nil {
		return ClosureReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/reconstruct-first/%w", err)
	}
	secondActual, err := ArtifactProjection(secondReport, source, headSHA)
	if err != nil {
		return ClosureReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/reconstruct-second/%w", err)
	}
	if err := ValidateArtifactProjectionBinding(firstExpected, firstActual); err != nil {
		return ClosureReceipt{}, err
	}
	if err := ValidateArtifactProjectionBinding(secondExpected, secondActual); err != nil {
		return ClosureReceipt{}, err
	}
	semanticEqual := CompareArtifactSemanticProjection(firstActual, secondActual) == nil
	if !semanticEqual {
		return ClosureReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/compare-semantic/SEMANTIC_REPLAY_MISMATCH")
	}
	outputTamper, err := observeOutputTamper(firstActual, outputTamperPath, headSHA)
	if err != nil {
		return ClosureReceipt{}, err
	}
	authorizationTamper, err := observeAuthorizationTamper(firstActual, authorizationTamperPath, headSHA)
	if err != nil {
		return ClosureReceipt{}, err
	}
	commentMetric, err := observeCommentOnlyMetric(interventionReportRaw, interventionConsumerRaw, source, headSHA)
	if err != nil {
		return ClosureReceipt{}, err
	}
	rows := []ClosureMetricEvidence{
		newClosureMetric(closureMetricArtifactBytesFirst, "first", closureArtifactAddress(closureMetricArtifactBytesFirst, firstActual), closureProducerExecutor, closureArtifactConsumer, "observe-artifact-bytes", model.ProofCoherence, closureArtifactAddress(closureMetricArtifactBytesFirst, firstActual), firstActual.RawDigest, artifactBytesPayload(firstActual), "ARTIFACT_OBSERVATION", "read-artifact-bytes", "ARTIFACT_BYTES_OBSERVED"),
		newClosureMetric(closureMetricArtifactBytesSecond, "second", closureArtifactAddress(closureMetricArtifactBytesSecond, secondActual), closureProducerExecutor, closureArtifactConsumer, "observe-artifact-bytes", model.ProofCoherence, closureArtifactAddress(closureMetricArtifactBytesSecond, secondActual), secondActual.RawDigest, artifactBytesPayload(secondActual), "ARTIFACT_OBSERVATION", "read-artifact-bytes", "ARTIFACT_BYTES_OBSERVED"),
		newClosureMetric(closureMetricSemanticEquality, "first<->second", "semantic:"+firstActual.CaseID+":first<->second", "invarianttransformation.report-consumer", closureArtifactConsumer, "compare-semantic-digest-pair", model.ProofRegression, "semantic:"+firstActual.CaseID+":first<->second", model.Digest([]string{firstActual.SemanticDigest, secondActual.SemanticDigest}), struct {
			First  string `json:"first"`
			Second string `json:"second"`
		}{firstActual.SemanticDigest, secondActual.SemanticDigest}, "ARTIFACT_REPLAY", "compare-semantic-projection", "SEMANTIC_DIGEST_PAIR_EQUAL"),
		newClosureMetric(closureMetricRawProvenanceFirst, "first", closureArtifactAddress(closureMetricRawProvenanceFirst, firstActual), closureProducerExecutor, closureArtifactConsumer, "bind-raw-provenance", model.ProofCoherence, closureArtifactAddress(closureMetricRawProvenanceFirst, firstActual), rawProvenanceTargetDigest(firstActual), rawProvenancePayload(firstActual), "ARTIFACT_OBSERVATION", "bind-raw-provenance", "RAW_PROVENANCE_BOUND"),
		newClosureMetric(closureMetricRawProvenanceSecond, "second", closureArtifactAddress(closureMetricRawProvenanceSecond, secondActual), closureProducerExecutor, closureArtifactConsumer, "bind-raw-provenance", model.ProofCoherence, closureArtifactAddress(closureMetricRawProvenanceSecond, secondActual), rawProvenanceTargetDigest(secondActual), rawProvenancePayload(secondActual), "ARTIFACT_OBSERVATION", "bind-raw-provenance", "RAW_PROVENANCE_BOUND"),
		newClosureMetric(closureMetricAuthorizationFirst, "first", closureArtifactAddress(closureMetricAuthorizationFirst, firstActual), closureProducerExecutor, closureArtifactConsumer, "bind-authorization-receipt", model.ProofCoherence, closureArtifactAddress(closureMetricAuthorizationFirst, firstActual), firstActual.ExpectedAuthorizationDigest, authorizationPayload(firstActual), "ARTIFACT_AUTHORIZATION", "compare-observed-authorization", "AUTHORIZATION_BOUND"),
		newClosureMetric(closureMetricAuthorizationSecond, "second", closureArtifactAddress(closureMetricAuthorizationSecond, secondActual), closureProducerExecutor, closureArtifactConsumer, "bind-authorization-receipt", model.ProofCoherence, closureArtifactAddress(closureMetricAuthorizationSecond, secondActual), secondActual.ExpectedAuthorizationDigest, authorizationPayload(secondActual), "ARTIFACT_AUTHORIZATION", "compare-observed-authorization", "AUTHORIZATION_BOUND"),
		newClosureMetric(closureMetricOutputTamper, "output-only", outputTamperReceiptAddress(outputTamper), "ci-artifact-tamper-fixture", closureArtifactConsumer, "reject-typed-output-tamper", model.ProofCoherence, outputTamperReceiptAddress(outputTamper), outputTamper.EvidenceDigest, outputTamper, outputTamper.Stage, outputTamper.Step, outputTamper.Reason),
		newClosureMetric(closureMetricAuthorizationTamper, "authorization-only", authorizationTamperReceiptAddress(authorizationTamper), "ci-artifact-tamper-fixture", closureArtifactConsumer, "reject-typed-authorization-tamper", model.ProofCoherence, authorizationTamperReceiptAddress(authorizationTamper), authorizationTamper.EvidenceDigest, authorizationTamper, authorizationTamper.Stage, authorizationTamper.Step, authorizationTamper.Reason),
		commentMetric,
	}
	precedingDigest := sortedMetricEvidenceDigest(rows)
	rows = append(rows, newClosureMetric(closureMetricFinalClosure, "inventory-1..10", "closure:metric-inventory", closureClosureConsumer, closureArtifactConsumer, "validate-closure-metric-inventory", model.ProofCoherence, "closure:metric-inventory", precedingDigest, struct {
		EvidenceDigests []string `json:"evidence_digests"`
		InventoryDigest string   `json:"inventory_digest"`
	}{sortedMetricEvidence(rows), precedingDigest}, "CLOSURE", "validate-11-metric-inventory", "ALL_PRECEDING_EVIDENCE_BOUND"))
	counts, err := deriveClosureMetricCounts(rows)
	if err != nil {
		return ClosureReceipt{}, err
	}
	decision, resolution, reason := deriveClosureDecision(counts)
	if decision != model.DecisionPass {
		return ClosureReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/adjudicate/%s", reason)
	}
	closure := ClosureReceipt{
		Schema: closureReceiptSchema, HeadSHA: headSHA,
		PreliminaryDecision: firstReport.Decision, PreliminaryResolution: firstReport.Resolution, PreliminaryReason: firstReport.Reason, PreliminaryDecisionScope: firstReport.DecisionScope,
		FirstReportDigest: firstReport.Digest, SecondReportDigest: secondReport.Digest,
		FirstArtifact: artifactClosureEvidence(firstActual), SecondArtifact: artifactClosureEvidence(secondActual), OutputTamperReceipt: outputTamper, AuthorizationTamperReceipt: authorizationTamper,
		MetricEvidence: rows, ExpectedMetricEvidence: closureMetricInventoryExpected, ObservedMetricEvidence: len(rows), MetricInventoryDigest: model.Digest(rows),
		ArtifactBytesObserved: counts.ArtifactBytesObserved, ExpectedArtifactBytes: closureMetricArtifactBytesExpected,
		RawProvenanceBindings: counts.RawProvenanceBindings, ExpectedRawProvenanceBindings: closureMetricRawProvenanceExpected,
		AuthorizationBindings: counts.AuthorizationBindings, ExpectedAuthorizationBindings: closureMetricAuthorizationExpected,
		SemanticEqualityObserved: counts.SemanticEqualityObserved, ExpectedSemanticEquality: closureMetricSemanticEqualityExpected,
		OutputSemanticTamperRejected: counts.OutputSemanticTamperRejected, ExpectedOutputSemanticTamperRejections: closureMetricOutputTamperExpected,
		AuthorizationTamperRejected: counts.AuthorizationTamperRejected, ExpectedAuthorizationTamperRejections: closureMetricAuthorizationTamperExpected,
		CommentOnlyPreservationObserved: counts.CommentOnlyPreservationObserved, ExpectedCommentOnlyPreservation: closureMetricCommentOnlyExpected,
		FinalClosureGateObserved: counts.FinalClosureGateObserved, ExpectedFinalClosureGate: closureMetricFinalClosureExpected,
		Decision: decision, Resolution: resolution, Reason: reason,
	}
	closure.Digest = model.Digest(closure)
	return closure, nil
}

func closureArtifactAddress(metricID string, projection ArtifactSemanticProjection) string {
	return metricID + ":" + projection.Path
}

func deriveClosureDecision(counts closureMetricCounts) (string, string, string) {
	if counts.ArtifactBytesObserved != closureMetricArtifactBytesExpected || counts.RawProvenanceBindings != closureMetricRawProvenanceExpected || counts.AuthorizationBindings != closureMetricAuthorizationExpected || counts.SemanticEqualityObserved != closureMetricSemanticEqualityExpected || counts.OutputSemanticTamperRejected != closureMetricOutputTamperExpected || counts.AuthorizationTamperRejected != closureMetricAuthorizationTamperExpected || counts.CommentOnlyPreservationObserved != closureMetricCommentOnlyExpected || counts.FinalClosureGateObserved != closureMetricFinalClosureExpected {
		return model.DecisionFailClosed, model.ResolutionLower, "CLOSURE_METRIC_INVENTORY_NOT_SATISFIED"
	}
	return model.DecisionPass, model.ResolutionExact, "ALL_CLOSURE_METRIC_EVIDENCE_SATISFIED"
}

func newClosureMetric(id, occurrence, address, producer, consumer, operation, proof, targetAddress, targetDigest string, evidence any, stage, step, reason string) ClosureMetricEvidence {
	row := ClosureMetricEvidence{MetricID: id, Occurrence: occurrence, Address: address, Producer: producer, IndependentConsumer: consumer, MetaOperation: operation, ProofChoice: proof, TargetAddress: targetAddress, TargetDigest: targetDigest, Decision: model.DecisionPass, Resolution: model.ResolutionExact, Stage: stage, Step: step, Reason: reason}
	row.ObservedEvidenceDigest = model.Digest(closureMetricDigestPayload{MetricID: row.MetricID, Occurrence: row.Occurrence, Address: row.Address, Producer: row.Producer, IndependentConsumer: row.IndependentConsumer, MetaOperation: row.MetaOperation, ProofChoice: row.ProofChoice, TargetAddress: row.TargetAddress, TargetDigest: row.TargetDigest, Decision: row.Decision, Resolution: row.Resolution, Stage: row.Stage, Step: row.Step, Reason: row.Reason, Evidence: evidence})
	return row
}

func sortedMetricEvidenceDigest(rows []ClosureMetricEvidence) string {
	digests := make([]string, 0, len(rows))
	for _, row := range rows {
		digests = append(digests, row.ObservedEvidenceDigest)
	}
	sort.Strings(digests)
	return model.Digest(digests)
}

func sortedMetricEvidence(rows []ClosureMetricEvidence) []string {
	digests := make([]string, 0, len(rows))
	for _, row := range rows {
		digests = append(digests, row.ObservedEvidenceDigest)
	}
	sort.Strings(digests)
	return digests
}

func deriveClosureMetricCounts(rows []ClosureMetricEvidence) (closureMetricCounts, error) {
	if err := validateClosureMetricInventory(rows); err != nil {
		return closureMetricCounts{}, err
	}
	counts := closureMetricCounts{}
	for _, row := range rows {
		switch row.MetricID {
		case closureMetricArtifactBytesFirst, closureMetricArtifactBytesSecond:
			counts.ArtifactBytesObserved++
		case closureMetricSemanticEquality:
			counts.SemanticEqualityObserved++
		case closureMetricRawProvenanceFirst, closureMetricRawProvenanceSecond:
			counts.RawProvenanceBindings++
		case closureMetricAuthorizationFirst, closureMetricAuthorizationSecond:
			counts.AuthorizationBindings++
		case closureMetricOutputTamper:
			counts.OutputSemanticTamperRejected++
		case closureMetricAuthorizationTamper:
			counts.AuthorizationTamperRejected++
		case closureMetricCommentOnly:
			counts.CommentOnlyPreservationObserved++
		case closureMetricFinalClosure:
			counts.FinalClosureGateObserved++
		}
	}
	return counts, nil
}

func validateClosureMetricInventory(rows []ClosureMetricEvidence) error {
	seen := map[string]bool{}
	seenAddresses := map[string]bool{}
	allowed := map[string]bool{}
	for _, id := range closureMetricIDs {
		allowed[id] = true
	}
	for _, row := range rows {
		if !allowed[row.MetricID] {
			return fmt.Errorf("ARTIFACT_CLOSURE/metrics/unexpected-metric-row/UNEXPECTED_METRIC_ID:%s", row.MetricID)
		}
		if seen[row.MetricID] {
			return fmt.Errorf("ARTIFACT_CLOSURE/metrics/duplicate-metric-row/DUPLICATE_METRIC_ID:%s", row.MetricID)
		}
		if seenAddresses[row.Address] {
			return fmt.Errorf("ARTIFACT_CLOSURE/metrics/duplicate-metric-row/DUPLICATE_METRIC_ADDRESS:%s", row.Address)
		}
		seen[row.MetricID] = true
		seenAddresses[row.Address] = true
		if row.Occurrence == "" || row.Address == "" || row.Producer == "" || row.IndependentConsumer == "" || row.MetaOperation == "" || row.ProofChoice == "" || row.TargetAddress == "" || !model.ValidDigest(row.TargetDigest) || !model.ValidDigest(row.ObservedEvidenceDigest) || row.Decision != model.DecisionPass || row.Resolution != model.ResolutionExact || row.Stage == "" || row.Step == "" || row.Reason == "" {
			return fmt.Errorf("ARTIFACT_CLOSURE/metrics/validate-row/METRIC_EVIDENCE_INCOMPLETE:%s", row.MetricID)
		}
	}
	for _, id := range closureMetricIDs {
		if !seen[id] {
			return fmt.Errorf("ARTIFACT_CLOSURE/metrics/missing-metric-row/MISSING_METRIC_ID:%s", id)
		}
	}
	if len(rows) != closureMetricInventoryExpected {
		return fmt.Errorf("ARTIFACT_CLOSURE/metrics/inventory-count/METRIC_INVENTORY_COUNT_MISMATCH")
	}
	return nil
}

func artifactBytesPayload(projection ArtifactSemanticProjection) any {
	return struct {
		Relation, Path, RawDigest, ExecutionID, SourceDigest, SubjectSHA string
		RawSize                                                          int
	}{"artifact-bytes", projection.Path, projection.RawDigest, projection.ExecutionID, projection.SourceDigest, projection.SubjectSHA, projection.RawSize}
}

func rawProvenancePayload(projection ArtifactSemanticProjection) any {
	return struct {
		Relation, Path, RawDigest, SourceDigest, SubjectSHA, ExecutionID string
		RawSize                                                          int
	}{"raw-provenance", projection.Path, projection.RawDigest, projection.SourceDigest, projection.SubjectSHA, projection.ExecutionID, projection.RawSize}
}

func rawProvenanceTargetDigest(projection ArtifactSemanticProjection) string {
	return model.Digest(rawProvenancePayload(projection))
}

func authorizationPayload(projection ArtifactSemanticProjection) any {
	return struct {
		Relation, Path, Observed, Expected, Effect string
	}{"authorization", projection.Path, projection.ObservedAuthorizationDigest, projection.ExpectedAuthorizationDigest, projection.EffectDigest}
}

func artifactClosureEvidence(projection ArtifactSemanticProjection) ArtifactClosureEvidence {
	return ArtifactClosureEvidence{Path: projection.Path, RawDigest: projection.RawDigest, RawSize: projection.RawSize, ExecutionID: projection.ExecutionID, SourceDigest: projection.SourceDigest, SubjectSHA: projection.SubjectSHA, ObservedAuthorizationDigest: projection.ObservedAuthorizationDigest, ExpectedAuthorizationDigest: projection.ExpectedAuthorizationDigest, EffectDigest: projection.EffectDigest, SemanticDigest: projection.SemanticDigest, CanonicalSemanticBytes: projection.CanonicalSemanticBytes}
}

func outputTamperReceiptAddress(receipt TypedTamperReceipt) string {
	return "tamper:" + receipt.Kind + ":" + receipt.TamperedRawDigest
}

func authorizationTamperReceiptAddress(receipt TypedTamperReceipt) string {
	return "tamper:" + receipt.Kind + ":" + receipt.TamperedRawDigest
}

func observeOutputTamper(expected ArtifactSemanticProjection, path, headSHA string) (TypedTamperReceipt, error) {
	tampered, err := ProjectArtifactFile(path, headSHA)
	if err != nil {
		return TypedTamperReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/output-tamper/observe/OUTPUT_TAMPER_NOT_OBSERVABLE: %w", err)
	}
	tamperedPath := tampered.Path
	if !allowedSnapshotPath(expected.Path) || !allowedSnapshotPath(tamperedPath) {
		return TypedTamperReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/output-tamper/compare/OUTPUT_TAMPER_PATH_NOT_SAFE")
	}
	tampered.Path = expected.Path
	tampered.ExpectedAuthorizationDigest = expected.ExpectedAuthorizationDigest
	tampered.EffectDigest = expected.EffectDigest
	if !sameArtifactFieldsExcept(expected, tampered, "output") || tampered.Semantic.Output == expected.Semantic.Output || tampered.RawDigest == expected.RawDigest {
		return TypedTamperReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/output-tamper/compare/OUTPUT_TAMPER_NOT_OUTPUT_ONLY")
	}
	if CompareArtifactSemanticProjection(expected, tampered) == nil {
		return TypedTamperReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/output-tamper/adjudicate/OUTPUT_SEMANTIC_TAMPER_ACCEPTED")
	}
	receipt := newTamperReceipt("OUTPUT_ONLY", expected, tampered, tamperedPath, "output", strconv.FormatInt(expected.Semantic.Output, 10), strconv.FormatInt(tampered.Semantic.Output, 10), false, "ARTIFACT_TAMPER", "compare-output-only", "OUTPUT_ONLY_TAMPER_REJECTED")
	return receipt, nil
}

func observeAuthorizationTamper(expected ArtifactSemanticProjection, path, headSHA string) (TypedTamperReceipt, error) {
	tampered, err := ProjectArtifactFile(path, headSHA)
	if err != nil {
		return TypedTamperReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/authorization-tamper/observe/AUTHORIZATION_TAMPER_NOT_OBSERVABLE: %w", err)
	}
	tamperedPath := tampered.Path
	if !allowedSnapshotPath(expected.Path) || !allowedSnapshotPath(tamperedPath) {
		return TypedTamperReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/authorization-tamper/compare/AUTHORIZATION_TAMPER_PATH_NOT_SAFE")
	}
	tampered.Path = expected.Path
	tampered.ExpectedAuthorizationDigest = expected.ExpectedAuthorizationDigest
	tampered.EffectDigest = expected.EffectDigest
	if !sameArtifactFieldsExcept(expected, tampered, "authorization") || tampered.ObservedAuthorizationDigest == expected.ObservedAuthorizationDigest || tampered.RawDigest == expected.RawDigest {
		return TypedTamperReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/authorization-tamper/compare/AUTHORIZATION_TAMPER_NOT_AUTHORIZATION_ONLY")
	}
	if err := CompareArtifactSemanticProjection(expected, tampered); err != nil {
		return TypedTamperReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/authorization-tamper/adjudicate/AUTHORIZATION_TAMPER_CHANGED_SEMANTICS")
	}
	if CompareArtifactProvenance(expected, tampered) == nil {
		return TypedTamperReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/authorization-tamper/adjudicate/AUTHORIZATION_TAMPER_ACCEPTED")
	}
	receipt := newTamperReceipt("AUTHORIZATION_ONLY", expected, tampered, tamperedPath, "authorization", expected.ObservedAuthorizationDigest, tampered.ObservedAuthorizationDigest, true, "ARTIFACT_TAMPER", "compare-authorization-only", "AUTHORIZATION_ONLY_TAMPER_REJECTED")
	return receipt, nil
}

func sameArtifactFieldsExcept(expected, actual ArtifactSemanticProjection, except string) bool {
	if expected.Schema != actual.Schema || expected.HeadSHA != actual.HeadSHA || expected.CaseID != actual.CaseID || expected.Semantic.CaseID != actual.Semantic.CaseID || expected.Semantic.CaseID != expected.CaseID || expected.ExecutionID != actual.ExecutionID || expected.RawSize != actual.RawSize || expected.Semantic.Input != actual.Semantic.Input || expected.Semantic.Operation != actual.Semantic.Operation || expected.SourceDigest != actual.SourceDigest || expected.Semantic.SemanticSourceDigest != actual.Semantic.SemanticSourceDigest || expected.SubjectSHA != actual.SubjectSHA || expected.ExpectedAuthorizationDigest != actual.ExpectedAuthorizationDigest || expected.EffectDigest != actual.EffectDigest || !allowedSnapshotPath(expected.Path) || !allowedSnapshotPath(actual.Path) {
		return false
	}
	if except != "output" && expected.Semantic.Output != actual.Semantic.Output {
		return false
	}
	if except != "authorization" && expected.ObservedAuthorizationDigest != actual.ObservedAuthorizationDigest {
		return false
	}
	return true
}

func newTamperReceipt(kind string, expected, tampered ArtifactSemanticProjection, tamperedPath, field, baselineValue, tamperedValue string, semanticEqual bool, stage, step, reason string) TypedTamperReceipt {
	receipt := TypedTamperReceipt{Schema: "gooo/invariant-transformation-typed-tamper-receipt/v1", Kind: kind, CaseID: expected.CaseID, ExecutionID: expected.ExecutionID, BaselinePath: expected.Path, TamperedPath: tamperedPath, BaselineRawDigest: expected.RawDigest, TamperedRawDigest: tampered.RawDigest, ChangedField: field, BaselineValue: baselineValue, TamperedValue: tamperedValue, SemanticDigestEqual: semanticEqual, Rejected: true, Decision: model.DecisionFailClosed, Resolution: model.ResolutionExact, Stage: stage, Step: step, Reason: reason}
	evidence := receipt
	evidence.EvidenceDigest = ""
	evidence.Digest = ""
	receipt.EvidenceDigest = model.Digest(evidence)
	receipt.Digest = model.Digest(receipt)
	return receipt
}

type closureCoordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type closureTransition struct {
	ClaimID                  string            `json:"claim_id"`
	From                     string            `json:"from"`
	To                       string            `json:"to"`
	Coordinate               closureCoordinate `json:"coordinate"`
	PropositionDigest        string            `json:"proposition_digest"`
	PriorStateDigest         string            `json:"prior_state_digest"`
	EvidenceDigest           string            `json:"evidence_digest"`
	PreviousTransitionDigest string            `json:"previous_transition_digest"`
	CurrentTransitionDigest  string            `json:"current_transition_digest"`
}

type closureInterventionCase struct {
	ID                                        string              `json:"id"`
	Kind                                      string              `json:"kind"`
	SourceEdit                                string              `json:"source_edit"`
	BaselineProjection                        json.RawMessage     `json:"baseline_projection"`
	MutatedProjection                         json.RawMessage     `json:"mutated_projection"`
	BaselineProjectionDigest                  string              `json:"baseline_projection_digest"`
	MutatedProjectionDigest                   string              `json:"mutated_projection_digest"`
	BaselineSourceDigest                      string              `json:"baseline_source_digest"`
	MutatedSourceDigest                       string              `json:"mutated_source_digest"`
	BaselineProvenanceDigest                  string              `json:"baseline_provenance_digest"`
	MutatedProvenanceDigest                   string              `json:"mutated_provenance_digest"`
	ProvenanceDigestChanged                   bool                `json:"provenance_digest_changed"`
	BaselineSemanticDigest                    string              `json:"baseline_semantic_digest"`
	MutatedSemanticDigest                     string              `json:"mutated_semantic_digest"`
	SemanticDigestEqual                       bool                `json:"semantic_digest_equal"`
	BaselineReceiptDigest                     string              `json:"baseline_receipt_digest"`
	MutatedReceiptDigest                      string              `json:"mutated_receipt_digest"`
	BaselineReceiptDecision                   string              `json:"baseline_receipt_decision"`
	MutatedReceiptDecision                    string              `json:"mutated_receipt_decision"`
	BaselineJudgment                          json.RawMessage     `json:"baseline_judgment"`
	MutatedJudgment                           json.RawMessage     `json:"mutated_judgment"`
	BaselineEvidence                          json.RawMessage     `json:"baseline_evidence"`
	MutatedEvidence                           json.RawMessage     `json:"mutated_evidence"`
	BaselineClaimTransitions                  []closureTransition `json:"baseline_claim_transitions"`
	MutatedClaimTransitions                   []closureTransition `json:"mutated_claim_transitions"`
	RawSourceDigestChanged                    bool                `json:"raw_source_digest_changed"`
	ReceiptChanged                            bool                `json:"receipt_changed"`
	SemanticProjectionEqual                   bool                `json:"semantic_projection_equal"`
	DecisionEqual                             bool                `json:"decision_equal"`
	ResolutionEqual                           bool                `json:"resolution_equal"`
	ReasonEqual                               bool                `json:"reason_equal"`
	DecisionChanged                           bool                `json:"decision_changed"`
	ClaimTransitionsEqual                     bool                `json:"claim_transitions_equal"`
	EffectsEqual                              bool                `json:"effects_equal"`
	ReplayObservationEqual                    bool                `json:"replay_observation_equal"`
	EvidenceObservable                        bool                `json:"evidence_observable"`
	RepositoryWritesNotClaimed                bool                `json:"repository_writes_not_claimed"`
	BaselineRepositoryWrites                  int                 `json:"baseline_repository_writes"`
	MutatedRepositoryWrites                   int                 `json:"mutated_repository_writes"`
	BaselineRepositoryWritesObserved          bool                `json:"baseline_repository_writes_observed"`
	MutatedRepositoryWritesObserved           bool                `json:"mutated_repository_writes_observed"`
	BaselineRepositoryNetStatusUnchanged      bool                `json:"baseline_repository_net_status_unchanged"`
	MutatedRepositoryNetStatusUnchanged       bool                `json:"mutated_repository_net_status_unchanged"`
	BaselineRepositoryActualOrTransientWrites string              `json:"baseline_repository_actual_or_transient_writes"`
	MutatedRepositoryActualOrTransientWrites  string              `json:"mutated_repository_actual_or_transient_writes"`
	BaselineRepositoryMutationAuthorized      bool                `json:"baseline_repository_mutation_authorized"`
	MutatedRepositoryMutationAuthorized       bool                `json:"mutated_repository_mutation_authorized"`
	Claim                                     json.RawMessage     `json:"claim"`
	Satisfied                                 bool                `json:"satisfied"`
}

type closureInterventionReport struct {
	Schema                            string                    `json:"schema"`
	HeadSHA                           string                    `json:"head_sha"`
	SourcePath                        string                    `json:"source_path"`
	SourceDigest                      string                    `json:"source_digest"`
	Denominator                       json.RawMessage           `json:"denominator"`
	CaseCount                         int                       `json:"case_count"`
	Cases                             []closureInterventionCase `json:"cases"`
	EffectGates                       json.RawMessage           `json:"effect_gates"`
	EffectGateDenominator             int                       `json:"effect_gate_denominator"`
	EffectGateSatisfied               int                       `json:"effect_gate_satisfied"`
	Decision                          string                    `json:"decision"`
	Resolution                        string                    `json:"resolution"`
	Reason                            string                    `json:"reason"`
	RepositoryWrites                  int                       `json:"repository_writes"`
	RepositoryMutationAuthorized      bool                      `json:"repository_mutation_authorized"`
	TempArtifactWriteAuthorized       bool                      `json:"temp_artifact_write_authorized"`
	RepositoryNetStatusUnchanged      bool                      `json:"repository_net_status_unchanged"`
	RepositoryActualOrTransientWrites string                    `json:"repository_actual_or_transient_writes"`
	RepositoryNetStatusObserved       bool                      `json:"repository_net_status_observed"`
	ExecutedEffects                   int                       `json:"executed_effects"`
	IndependentlyObservedEffects      int                       `json:"independently_observed_effects"`
	UnknownEffectScopes               int                       `json:"unknown_effect_scopes"`
	RepositoryPathAuthorization       bool                      `json:"repository_path_authorization"`
	AmbientProcessAuthority           string                    `json:"ambient_process_authority"`
	CorrectionCount                   int                       `json:"correction_count"`
	CorrectionDenominator             int                       `json:"correction_denominator"`
	Failure                           json.RawMessage           `json:"failure,omitempty"`
	Digest                            string                    `json:"digest"`
}

type closureCommentOnlyReceipt struct {
	Schema                   string `json:"schema"`
	CaseID                   string `json:"case_id"`
	BaselineRawDigest        string `json:"baseline_raw_digest"`
	MutatedRawDigest         string `json:"mutated_raw_digest"`
	BaselineProvenanceDigest string `json:"baseline_provenance_digest"`
	MutatedProvenanceDigest  string `json:"mutated_provenance_digest"`
	BaselineSemanticDigest   string `json:"baseline_semantic_digest"`
	MutatedSemanticDigest    string `json:"mutated_semantic_digest"`
	SemanticDigestEqual      bool   `json:"semantic_digest_equal"`
	BaselineDecision         string `json:"baseline_decision"`
	MutatedDecision          string `json:"mutated_decision"`
	DecisionEqual            bool   `json:"decision_equal"`
	BaselineTransitionDigest string `json:"baseline_transition_digest"`
	MutatedTransitionDigest  string `json:"mutated_transition_digest"`
	ClaimTransitionsEqual    bool   `json:"claim_transitions_equal"`
	Stage                    string `json:"stage"`
	Step                     string `json:"step"`
	Reason                   string `json:"reason"`
	EvidenceDigest           string `json:"evidence_digest"`
	Digest                   string `json:"digest"`
}

type closureInterventionConsumerReceipt struct {
	Schema                                             string                    `json:"schema"`
	HeadSHA                                            string                    `json:"head_sha"`
	ProducerDependencyImports                          int                       `json:"producer_dependency_imports"`
	AllowedProducerDependencyImports                   int                       `json:"allowed_producer_dependency_imports"`
	ReconstructedCases                                 int                       `json:"reconstructed_cases"`
	ExpectedCases                                      int                       `json:"expected_cases"`
	ActualReplay                                       int                       `json:"actual_replay"`
	ExpectedActualReplay                               int                       `json:"expected_actual_replay"`
	ArtifactEvidence                                   model.ArtifactEvidence    `json:"artifact_evidence"`
	ArtifactObserved                                   bool                      `json:"artifact_observed"`
	CoherentTamperRejected                             int                       `json:"coherent_tamper_rejected"`
	ExpectedCoherentTamperRejections                   int                       `json:"expected_coherent_tamper_rejections"`
	ContentObservationCoherentTamperRejected           int                       `json:"content_observation_coherent_tamper_rejected"`
	ExpectedContentObservationCoherentTamperRejections int                       `json:"expected_content_observation_coherent_tamper_rejections"`
	Decision                                           string                    `json:"decision"`
	Resolution                                         string                    `json:"resolution"`
	Reason                                             string                    `json:"reason"`
	RepositoryNetStatusUnchanged                       bool                      `json:"repository_net_status_unchanged"`
	RepositoryNetStatusObserved                        bool                      `json:"repository_net_status_observed"`
	RepositoryNetState                                 string                    `json:"repository_net_state"`
	RepositoryActualOrTransientWrites                  string                    `json:"repository_actual_or_transient_writes"`
	RepositoryPathAuthorization                        bool                      `json:"repository_path_authorization"`
	AmbientProcessAuthority                            string                    `json:"ambient_process_authority"`
	UnknownEffectScopes                                int                       `json:"unknown_effect_scopes"`
	CommentOnly                                        closureCommentOnlyReceipt `json:"comment_only"`
	Digest                                             string                    `json:"digest"`
}

func observeCommentOnlyMetric(interventionRaw, consumerRaw, source []byte, headSHA string) (ClosureMetricEvidence, error) {
	var intervention closureInterventionReport
	if err := decodeStrict(interventionRaw, &intervention); err != nil {
		return ClosureMetricEvidence{}, fmt.Errorf("ARTIFACT_CLOSURE/comment-only/parse-intervention/INTERVENTION_REPORT_NOT_STRICT: %w", err)
	}
	var consumer closureInterventionConsumerReceipt
	if err := decodeStrict(consumerRaw, &consumer); err != nil {
		return ClosureMetricEvidence{}, fmt.Errorf("ARTIFACT_CLOSURE/comment-only/parse-consumer/INTERVENTION_CONSUMER_RECEIPT_NOT_STRICT: %w", err)
	}
	consumerForDigest := consumer
	consumerForDigest.Digest = ""
	interventionForDigest := intervention
	interventionForDigest.Digest = ""
	if intervention.Schema != "gooo/invariant-transformation-intervention-report/v2" || intervention.HeadSHA != headSHA || intervention.SourcePath != model.SourcePath || intervention.SourceDigest != model.DigestBytes(source) || intervention.CaseCount != 3 || len(intervention.Cases) != 3 || intervention.Decision != model.DecisionPass || intervention.Resolution != model.ResolutionExact || intervention.Reason != "ALL_INTERVENTION_OBSERVATIONS_SATISFIED" || intervention.EffectGateDenominator != 8 || intervention.EffectGateSatisfied != 8 || intervention.CorrectionCount != 12 || intervention.CorrectionDenominator != 12 || !model.ValidDigest(intervention.Digest) || intervention.Digest != model.Digest(interventionForDigest) || consumer.Schema != "gooo/invariant-transformation-intervention-consumer/v2" || consumer.HeadSHA != headSHA || !model.ValidDigest(consumer.Digest) || consumer.Digest != model.Digest(consumerForDigest) || consumer.CommentOnly.CaseID != "nonsemantic-source-intervention" {
		return ClosureMetricEvidence{}, fmt.Errorf("ARTIFACT_CLOSURE/comment-only/bind-receipts/COMMENT_ONLY_RECEIPT_IDENTITY_INVALID")
	}
	expectedCaseIDs := map[string]bool{
		"semantic-expected-intervention":  false,
		"semantic-operation-intervention": false,
		"nonsemantic-source-intervention": false,
	}
	var observed *closureInterventionCase
	for index := range intervention.Cases {
		if _, ok := expectedCaseIDs[intervention.Cases[index].ID]; !ok || expectedCaseIDs[intervention.Cases[index].ID] {
			return ClosureMetricEvidence{}, fmt.Errorf("ARTIFACT_CLOSURE/comment-only/inventory/COMMENT_ONLY_CASE_INVENTORY_INVALID")
		}
		expectedCaseIDs[intervention.Cases[index].ID] = true
		if intervention.Cases[index].ID == "nonsemantic-source-intervention" {
			if observed != nil {
				return ClosureMetricEvidence{}, fmt.Errorf("ARTIFACT_CLOSURE/comment-only/select-case/DUPLICATE_COMMENT_ONLY_CASE")
			}
			observed = &intervention.Cases[index]
		}
	}
	if observed == nil || observed.Kind != "NON_SEMANTIC" || !expectedCaseIDs["semantic-expected-intervention"] || !expectedCaseIDs["semantic-operation-intervention"] || !expectedCaseIDs["nonsemantic-source-intervention"] {
		return ClosureMetricEvidence{}, fmt.Errorf("ARTIFACT_CLOSURE/comment-only/select-case/COMMENT_ONLY_CASE_MISSING")
	}
	consumerComment := consumer.CommentOnly
	consumerEvidence := consumerComment
	consumerEvidence.EvidenceDigest = ""
	consumerEvidence.Digest = ""
	if consumerComment.EvidenceDigest != model.Digest(consumerEvidence) || consumerComment.Digest != commentOnlyDigest(consumerComment) {
		return ClosureMetricEvidence{}, fmt.Errorf("ARTIFACT_CLOSURE/comment-only/bind-consumer/COMMENT_ONLY_CONSUMER_RECEIPT_DIGEST_INVALID")
	}
	baselineTransitionDigest := model.Digest(observed.BaselineClaimTransitions)
	mutatedTransitionDigest := model.Digest(observed.MutatedClaimTransitions)
	if observed.BaselineSourceDigest != consumerComment.BaselineRawDigest || observed.MutatedSourceDigest != consumerComment.MutatedRawDigest || observed.BaselineProvenanceDigest != consumerComment.BaselineProvenanceDigest || observed.MutatedProvenanceDigest != consumerComment.MutatedProvenanceDigest || observed.BaselineSemanticDigest != consumerComment.BaselineSemanticDigest || observed.MutatedSemanticDigest != consumerComment.MutatedSemanticDigest || observed.SemanticDigestEqual != consumerComment.SemanticDigestEqual || observed.BaselineReceiptDecision != consumerComment.BaselineDecision || observed.MutatedReceiptDecision != consumerComment.MutatedDecision || observed.DecisionEqual != consumerComment.DecisionEqual || baselineTransitionDigest != consumerComment.BaselineTransitionDigest || mutatedTransitionDigest != consumerComment.MutatedTransitionDigest || observed.ClaimTransitionsEqual != consumerComment.ClaimTransitionsEqual || consumerComment.Stage != "INTERVENTION" || consumerComment.Step != "compare-nonsemantic-projection-and-decision" || consumerComment.Reason != "NONSEMANTIC_PROJECTION_AND_DECISION_PRESERVED" {
		return ClosureMetricEvidence{}, fmt.Errorf("ARTIFACT_CLOSURE/comment-only/compare-receipts/COMMENT_ONLY_RECEIPTS_DIVERGE")
	}
	if !observed.RawSourceDigestChanged || !observed.ProvenanceDigestChanged || !observed.ReceiptChanged || !observed.SemanticDigestEqual || !observed.SemanticProjectionEqual || !observed.DecisionEqual || !observed.ResolutionEqual || !observed.ReasonEqual || !observed.ClaimTransitionsEqual || !consumerComment.SemanticDigestEqual || !consumerComment.DecisionEqual || !consumerComment.ClaimTransitionsEqual {
		return ClosureMetricEvidence{}, fmt.Errorf("ARTIFACT_CLOSURE/comment-only/adjudicate/COMMENT_ONLY_PRESERVATION_NOT_OBSERVED")
	}
	return newClosureMetric(closureMetricCommentOnly, "baseline<->mutated", "intervention:"+observed.ID, "invarianttransformation.intervention", closureInterventionConsumer, "compare-comment-only-source-provenance", model.ProofCoherence, "intervention:"+observed.ID, model.Digest([]string{intervention.Digest, consumerComment.Digest}), struct {
		ProducerReportDigest, ConsumerReceiptDigest, BaselineRawDigest, MutatedRawDigest, BaselineProvenanceDigest, MutatedProvenanceDigest, BaselineSemanticDigest, MutatedSemanticDigest, BaselineDecision, MutatedDecision, BaselineTransitionDigest, MutatedTransitionDigest string
		SemanticDigestEqual, DecisionEqual, ClaimTransitionsEqual                                                                                                                                                                                                                bool
	}{intervention.Digest, consumerComment.Digest, observed.BaselineSourceDigest, observed.MutatedSourceDigest, observed.BaselineProvenanceDigest, observed.MutatedProvenanceDigest, observed.BaselineSemanticDigest, observed.MutatedSemanticDigest, observed.BaselineReceiptDecision, observed.MutatedReceiptDecision, baselineTransitionDigest, mutatedTransitionDigest, observed.SemanticDigestEqual, observed.DecisionEqual, observed.ClaimTransitionsEqual}, "INTERVENTION", "compare-nonsemantic-projection-and-decision", "NONSEMANTIC_PROJECTION_AND_DECISION_PRESERVED"), nil
}

func commentOnlyDigest(receipt closureCommentOnlyReceipt) string {
	receipt.Digest = ""
	return model.Digest(receipt)
}

func decodeStrict(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON")
		}
		return err
	}
	return nil
}

// DecodeClosureReceipt is the strict wire boundary for the final closure
// artifact. Unknown fields and trailing JSON are rejected before adjudication.
func DecodeClosureReceipt(raw []byte) (ClosureReceipt, error) {
	var receipt ClosureReceipt
	if err := decodeStrict(raw, &receipt); err != nil {
		return ClosureReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/parse/FINAL_CLOSURE_RECEIPT_NOT_STRICT: %w", err)
	}
	return receipt, nil
}

// VerifyClosure independently rebuilds the expected closure and rejects any
// duplicate, missing, replaced, or resealed metric row before comparing it.
func VerifyClosure(receipt ClosureReceipt, firstReport, secondReport model.Report, firstExpected, secondExpected ArtifactSemanticProjection, source []byte, headSHA, outputTamperPath, authorizationTamperPath string, interventionReportRaw, interventionConsumerRaw []byte) error {
	if receipt.Schema != closureReceiptSchema {
		return fmt.Errorf("ARTIFACT_CLOSURE/verify/FINAL_CLOSURE_SCHEMA_INVALID")
	}
	if err := validateClosureMetricInventory(receipt.MetricEvidence); err != nil {
		return err
	}
	expected, err := Close(firstReport, secondReport, firstExpected, secondExpected, source, headSHA, outputTamperPath, authorizationTamperPath, interventionReportRaw, interventionConsumerRaw)
	if err != nil {
		return fmt.Errorf("ARTIFACT_CLOSURE/verify/reconstruct-expected/%w", err)
	}
	if !reflect.DeepEqual(receipt, expected) {
		return fmt.Errorf("ARTIFACT_CLOSURE/verify/compare-resealed/FINAL_CLOSURE_RECEIPT_MISMATCH")
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
