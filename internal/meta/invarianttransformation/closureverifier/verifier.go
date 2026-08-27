// Package closureverifier independently reconstructs the bounded closure
// receipt. It deliberately does not import reportconsumer or any producer
// implementation; only the shared model, syntax parser, lowering, judge, and
// digest primitives are used.
package closureverifier

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

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/judge"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const closureSchema = "gooo/invariant-transformation-closure-receipt/v2"

const (
	metricArtifactBytesFirst          = "artifact-bytes/first"
	metricArtifactBytesSecond         = "artifact-bytes/second"
	metricSemanticEquality            = "semantic-equality/pair"
	metricRawProvenanceFirst          = "raw-provenance/first"
	metricRawProvenanceSecond         = "raw-provenance/second"
	metricAuthorizationFirst          = "authorization/first"
	metricAuthorizationSecond         = "authorization/second"
	metricOutputTamper                = "output-tamper"
	metricAuthorizationTamper         = "authorization-tamper"
	metricCommentOnly                 = "comment-only/semantic-preservation"
	metricFinalClosure                = "final-closure"
	metricInventorySize               = 11
	metricArtifactBytesExpected       = 2
	metricSemanticEqualityExpected    = 1
	metricRawProvenanceExpected       = 2
	metricAuthorizationExpected       = 2
	metricOutputTamperExpected        = 1
	metricAuthorizationTamperExpected = 1
	metricCommentOnlyExpected         = 1
	metricFinalClosureExpected        = 1
)

var expectedMetricIDs = []string{
	metricArtifactBytesFirst,
	metricArtifactBytesSecond,
	metricSemanticEquality,
	metricRawProvenanceFirst,
	metricRawProvenanceSecond,
	metricAuthorizationFirst,
	metricAuthorizationSecond,
	metricOutputTamper,
	metricAuthorizationTamper,
	metricCommentOnly,
	metricFinalClosure,
}

type semanticValue struct {
	CaseID               string `json:"case_id"`
	Input                int64  `json:"input"`
	Operation            string `json:"operation"`
	Output               int64  `json:"output"`
	SemanticSourceDigest string `json:"semantic_source_digest"`
}

type projection struct {
	Schema                      string        `json:"schema"`
	HeadSHA                     string        `json:"head_sha"`
	CaseID                      string        `json:"case_id"`
	Path                        string        `json:"path"`
	RawDigest                   string        `json:"raw_digest"`
	RawSize                     int           `json:"raw_size"`
	ExecutionID                 string        `json:"execution_id"`
	SourceDigest                string        `json:"source_digest"`
	SubjectSHA                  string        `json:"subject_sha"`
	ObservedAuthorizationDigest string        `json:"observed_authorization_digest"`
	ExpectedAuthorizationDigest string        `json:"expected_authorization_digest"`
	EffectDigest                string        `json:"effect_digest"`
	Semantic                    semanticValue `json:"semantic"`
	CanonicalSemanticBytes      string        `json:"canonical_semantic_bytes"`
	SemanticDigest              string        `json:"semantic_digest"`
}

type artifactFields struct {
	CaseID, ExecutionID, Operation, SourceDigest, SemanticSourceDigest, AuthorizationDigest, SubjectSHA string
	Input, Output                                                                                       int64
}

type artifactEvidence struct {
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

type tamperReceipt struct {
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

type metricEvidence struct {
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

type metricDigestPayload struct {
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

type closureReceipt struct {
	Schema                                 string           `json:"schema"`
	HeadSHA                                string           `json:"head_sha"`
	PreliminaryDecision                    string           `json:"preliminary_decision"`
	PreliminaryResolution                  string           `json:"preliminary_resolution"`
	PreliminaryReason                      string           `json:"preliminary_reason"`
	PreliminaryDecisionScope               string           `json:"preliminary_decision_scope"`
	FirstReportDigest                      string           `json:"first_report_digest"`
	SecondReportDigest                     string           `json:"second_report_digest"`
	FirstArtifact                          artifactEvidence `json:"first_artifact"`
	SecondArtifact                         artifactEvidence `json:"second_artifact"`
	OutputTamperReceipt                    tamperReceipt    `json:"output_tamper_receipt"`
	AuthorizationTamperReceipt             tamperReceipt    `json:"authorization_tamper_receipt"`
	MetricEvidence                         []metricEvidence `json:"metric_evidence"`
	ExpectedMetricEvidence                 int              `json:"expected_metric_evidence"`
	ObservedMetricEvidence                 int              `json:"observed_metric_evidence"`
	MetricInventoryDigest                  string           `json:"metric_inventory_digest"`
	ArtifactBytesObserved                  int              `json:"artifact_bytes_observed"`
	ExpectedArtifactBytes                  int              `json:"expected_artifact_bytes"`
	RawProvenanceBindings                  int              `json:"raw_provenance_bindings"`
	ExpectedRawProvenanceBindings          int              `json:"expected_raw_provenance_bindings"`
	AuthorizationBindings                  int              `json:"authorization_bindings"`
	ExpectedAuthorizationBindings          int              `json:"expected_authorization_bindings"`
	SemanticEqualityObserved               int              `json:"semantic_equality_observed"`
	ExpectedSemanticEquality               int              `json:"expected_semantic_equality"`
	OutputSemanticTamperRejected           int              `json:"output_semantic_tamper_rejected"`
	ExpectedOutputSemanticTamperRejections int              `json:"expected_output_semantic_tamper_rejections"`
	AuthorizationTamperRejected            int              `json:"authorization_tamper_rejected"`
	ExpectedAuthorizationTamperRejections  int              `json:"expected_authorization_tamper_rejections"`
	CommentOnlyPreservationObserved        int              `json:"comment_only_preservation_observed"`
	ExpectedCommentOnlyPreservation        int              `json:"expected_comment_only_preservation"`
	FinalClosureGateObserved               int              `json:"final_closure_gate_observed"`
	ExpectedFinalClosureGate               int              `json:"expected_final_closure_gate"`
	Decision                               string           `json:"decision"`
	Resolution                             string           `json:"resolution"`
	Reason                                 string           `json:"reason"`
	Digest                                 string           `json:"digest"`
}

// Verify reconstructs the reports, generated artifacts, typed tamper cases,
// intervention evidence, fixed inventory, and all scalar counters without
// calling the producer-side closure implementation.
func Verify(closureRaw, firstReportRaw, secondReportRaw, firstProjectionRaw, secondProjectionRaw, source []byte, headSHA, outputTamperPath, authorizationTamperPath string, interventionReportRaw, interventionConsumerRaw []byte) error {
	var actual closureReceipt
	if err := decodeStrict(closureRaw, &actual); err != nil {
		return fmt.Errorf("ARTIFACT_CLOSURE/verify/parse/FINAL_CLOSURE_RECEIPT_NOT_STRICT: %w", err)
	}
	if actual.Schema != closureSchema || actual.HeadSHA != headSHA || !model.ValidHead(headSHA) {
		return fmt.Errorf("ARTIFACT_CLOSURE/verify/bind-head/HEAD_BINDING_INVALID")
	}
	if err := validateMetricInventory(actual.MetricEvidence); err != nil {
		return err
	}
	var firstReport, secondReport model.Report
	if err := decodeStrict(firstReportRaw, &firstReport); err != nil {
		return fmt.Errorf("ARTIFACT_CLOSURE/verify/parse-first-report/FIRST_REPORT_NOT_STRICT: %w", err)
	}
	if err := decodeStrict(secondReportRaw, &secondReport); err != nil {
		return fmt.Errorf("ARTIFACT_CLOSURE/verify/parse-second-report/SECOND_REPORT_NOT_STRICT: %w", err)
	}
	var firstExpected, secondExpected projection
	if err := decodeStrict(firstProjectionRaw, &firstExpected); err != nil {
		return fmt.Errorf("ARTIFACT_CLOSURE/verify/parse-first-projection/FIRST_PROJECTION_NOT_STRICT: %w", err)
	}
	if err := decodeStrict(secondProjectionRaw, &secondExpected); err != nil {
		return fmt.Errorf("ARTIFACT_CLOSURE/verify/parse-second-projection/SECOND_PROJECTION_NOT_STRICT: %w", err)
	}
	firstActual, err := reconstructArtifact(firstReport, source, headSHA)
	if err != nil {
		return fmt.Errorf("ARTIFACT_CLOSURE/verify/reconstruct-first/%w", err)
	}
	secondActual, err := reconstructArtifact(secondReport, source, headSHA)
	if err != nil {
		return fmt.Errorf("ARTIFACT_CLOSURE/verify/reconstruct-second/%w", err)
	}
	if !reflect.DeepEqual(firstExpected, firstActual) || !reflect.DeepEqual(secondExpected, secondActual) {
		return fmt.Errorf("ARTIFACT_CLOSURE/verify/bind-projection/EXPECTED_PROJECTION_NOT_SOURCE_BOUND")
	}
	expected, err := buildExpectedClosure(firstReport, secondReport, firstActual, secondActual, source, headSHA, outputTamperPath, authorizationTamperPath, interventionReportRaw, interventionConsumerRaw)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(actual, expected) {
		if evidenceDigestOnlyChanged(actual, expected) {
			return fmt.Errorf("ARTIFACT_CLOSURE/verify/metrics/reseal/EVIDENCE_DIGEST_RESEALED_WITHOUT_CANONICAL_PAYLOAD")
		}
		return fmt.Errorf("ARTIFACT_CLOSURE/verify/compare-reconstructed/FINAL_CLOSURE_RECEIPT_MISMATCH")
	}
	return nil
}

func buildExpectedClosure(firstReport, secondReport model.Report, first, second projection, source []byte, headSHA, outputTamperPath, authorizationTamperPath string, interventionReportRaw, interventionConsumerRaw []byte) (closureReceipt, error) {
	if firstReport.DecisionScope != model.PreliminaryDecisionScope || secondReport.DecisionScope != model.PreliminaryDecisionScope {
		return closureReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/verify/bind-reports/PRELIMINARY_SCOPE_INVALID")
	}
	if compareSemantic(first, second) != nil {
		return closureReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/verify/compare-semantic/SEMANTIC_REPLAY_MISMATCH")
	}
	outputTamper, err := observeTamper(first, outputTamperPath, headSHA, "output")
	if err != nil {
		return closureReceipt{}, err
	}
	authorizationTamper, err := observeTamper(first, authorizationTamperPath, headSHA, "authorization")
	if err != nil {
		return closureReceipt{}, err
	}
	commentMetric, err := observeCommentMetric(interventionReportRaw, interventionConsumerRaw, source, headSHA, first)
	if err != nil {
		return closureReceipt{}, err
	}
	rows := []metricEvidence{
		newMetric(metricArtifactBytesFirst, "first", artifactAddress(metricArtifactBytesFirst, first), model.ExecutorID, "invarianttransformation.artifact-semantic-consumer", "observe-artifact-bytes", model.ProofCoherence, artifactAddress(metricArtifactBytesFirst, first), first.RawDigest, artifactBytesPayload(first), "ARTIFACT_OBSERVATION", "read-artifact-bytes", "ARTIFACT_BYTES_OBSERVED"),
		newMetric(metricArtifactBytesSecond, "second", artifactAddress(metricArtifactBytesSecond, second), model.ExecutorID, "invarianttransformation.artifact-semantic-consumer", "observe-artifact-bytes", model.ProofCoherence, artifactAddress(metricArtifactBytesSecond, second), second.RawDigest, artifactBytesPayload(second), "ARTIFACT_OBSERVATION", "read-artifact-bytes", "ARTIFACT_BYTES_OBSERVED"),
		newMetric(metricSemanticEquality, "first<->second", "semantic:"+first.CaseID+":first<->second", "invarianttransformation.report-consumer", "invarianttransformation.artifact-semantic-consumer", "compare-semantic-digest-pair", model.ProofRegression, "semantic:"+first.CaseID+":first<->second", model.Digest([]string{first.SemanticDigest, second.SemanticDigest}), struct {
			First  string `json:"first"`
			Second string `json:"second"`
		}{first.SemanticDigest, second.SemanticDigest}, "ARTIFACT_REPLAY", "compare-semantic-projection", "SEMANTIC_DIGEST_PAIR_EQUAL"),
		newMetric(metricRawProvenanceFirst, "first", artifactAddress(metricRawProvenanceFirst, first), model.ExecutorID, "invarianttransformation.artifact-semantic-consumer", "bind-raw-provenance", model.ProofCoherence, artifactAddress(metricRawProvenanceFirst, first), rawProvenanceTargetDigest(first), rawProvenancePayload(first), "ARTIFACT_OBSERVATION", "bind-raw-provenance", "RAW_PROVENANCE_BOUND"),
		newMetric(metricRawProvenanceSecond, "second", artifactAddress(metricRawProvenanceSecond, second), model.ExecutorID, "invarianttransformation.artifact-semantic-consumer", "bind-raw-provenance", model.ProofCoherence, artifactAddress(metricRawProvenanceSecond, second), rawProvenanceTargetDigest(second), rawProvenancePayload(second), "ARTIFACT_OBSERVATION", "bind-raw-provenance", "RAW_PROVENANCE_BOUND"),
		newMetric(metricAuthorizationFirst, "first", artifactAddress(metricAuthorizationFirst, first), model.ExecutorID, "invarianttransformation.artifact-semantic-consumer", "bind-authorization-receipt", model.ProofCoherence, artifactAddress(metricAuthorizationFirst, first), first.ExpectedAuthorizationDigest, authorizationPayload(first), "ARTIFACT_AUTHORIZATION", "compare-observed-authorization", "AUTHORIZATION_BOUND"),
		newMetric(metricAuthorizationSecond, "second", artifactAddress(metricAuthorizationSecond, second), model.ExecutorID, "invarianttransformation.artifact-semantic-consumer", "bind-authorization-receipt", model.ProofCoherence, artifactAddress(metricAuthorizationSecond, second), second.ExpectedAuthorizationDigest, authorizationPayload(second), "ARTIFACT_AUTHORIZATION", "compare-observed-authorization", "AUTHORIZATION_BOUND"),
		newMetric(metricOutputTamper, "output-only", "tamper:"+outputTamper.Kind+":"+outputTamper.TamperedRawDigest, "ci-artifact-tamper-fixture", "invarianttransformation.artifact-semantic-consumer", "reject-typed-output-tamper", model.ProofCoherence, "tamper:"+outputTamper.Kind+":"+outputTamper.TamperedRawDigest, outputTamper.EvidenceDigest, outputTamper, outputTamper.Stage, outputTamper.Step, outputTamper.Reason),
		newMetric(metricAuthorizationTamper, "authorization-only", "tamper:"+authorizationTamper.Kind+":"+authorizationTamper.TamperedRawDigest, "ci-artifact-tamper-fixture", "invarianttransformation.artifact-semantic-consumer", "reject-typed-authorization-tamper", model.ProofCoherence, "tamper:"+authorizationTamper.Kind+":"+authorizationTamper.TamperedRawDigest, authorizationTamper.EvidenceDigest, authorizationTamper, authorizationTamper.Stage, authorizationTamper.Step, authorizationTamper.Reason),
		commentMetric,
	}
	precedingDigest := sortedEvidenceDigest(rows)
	rows = append(rows, newMetric(metricFinalClosure, "inventory-1..10", "closure:metric-inventory", "invarianttransformation.closure-consumer", "invarianttransformation.artifact-semantic-consumer", "validate-closure-metric-inventory", model.ProofCoherence, "closure:metric-inventory", precedingDigest, struct {
		EvidenceDigests []string `json:"evidence_digests"`
		InventoryDigest string   `json:"inventory_digest"`
	}{sortedEvidence(rows), precedingDigest}, "CLOSURE", "validate-11-metric-inventory", "ALL_PRECEDING_EVIDENCE_BOUND"))
	counts, err := deriveCounts(rows)
	if err != nil {
		return closureReceipt{}, err
	}
	decision, resolution, reason := deriveDecision(counts)
	if decision != model.DecisionPass {
		return closureReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/verify/adjudicate/%s", reason)
	}
	result := closureReceipt{
		Schema: closureSchema, HeadSHA: headSHA,
		PreliminaryDecision: firstReport.Decision, PreliminaryResolution: firstReport.Resolution, PreliminaryReason: firstReport.Reason, PreliminaryDecisionScope: firstReport.DecisionScope,
		FirstReportDigest: firstReport.Digest, SecondReportDigest: secondReport.Digest,
		FirstArtifact: toArtifactEvidence(first), SecondArtifact: toArtifactEvidence(second), OutputTamperReceipt: outputTamper, AuthorizationTamperReceipt: authorizationTamper,
		MetricEvidence: rows, ExpectedMetricEvidence: metricInventorySize, ObservedMetricEvidence: len(rows), MetricInventoryDigest: model.Digest(rows),
		ArtifactBytesObserved: counts.artifactBytes, ExpectedArtifactBytes: metricArtifactBytesExpected,
		RawProvenanceBindings: counts.rawProvenance, ExpectedRawProvenanceBindings: metricRawProvenanceExpected,
		AuthorizationBindings: counts.authorization, ExpectedAuthorizationBindings: metricAuthorizationExpected,
		SemanticEqualityObserved: counts.semanticEquality, ExpectedSemanticEquality: metricSemanticEqualityExpected,
		OutputSemanticTamperRejected: counts.outputTamper, ExpectedOutputSemanticTamperRejections: metricOutputTamperExpected,
		AuthorizationTamperRejected: counts.authorizationTamper, ExpectedAuthorizationTamperRejections: metricAuthorizationTamperExpected,
		CommentOnlyPreservationObserved: counts.commentOnly, ExpectedCommentOnlyPreservation: metricCommentOnlyExpected,
		FinalClosureGateObserved: counts.finalClosure, ExpectedFinalClosureGate: metricFinalClosureExpected,
		Decision: decision, Resolution: resolution, Reason: reason,
	}
	result.Digest = model.Digest(result)
	return result, nil
}

func reconstructArtifact(report model.Report, source []byte, headSHA string) (projection, error) {
	if err := judge.ValidateReport(report, source); err != nil {
		return projection{}, fmt.Errorf("REPORT_NOT_INDEPENDENTLY_VALID: %w", err)
	}
	var approved *model.Effect
	var approvedReceipt *model.Receipt
	for caseIndex := range report.Cases {
		for effectIndex := range report.Cases[caseIndex].Receipt.Effects {
			effect := &report.Cases[caseIndex].Receipt.Effects[effectIndex]
			if effect.Kind != model.EffectApproved {
				continue
			}
			if approved != nil {
				return projection{}, fmt.Errorf("MULTIPLE_APPROVED_ARTIFACTS")
			}
			approved = effect
			approvedReceipt = &report.Cases[caseIndex].Receipt
		}
	}
	if approved == nil || approvedReceipt == nil {
		return projection{}, fmt.Errorf("APPROVED_ARTIFACT_MISSING")
	}
	result, err := readArtifact(approved.Artifact.Path, headSHA)
	if err != nil {
		return projection{}, err
	}
	if result.CaseID != approved.CaseID || result.ExecutionID != approved.ExecutionID || result.RawDigest != approved.Artifact.ContentDigest || result.RawSize != approved.Artifact.Size || result.Semantic.CaseID != approvedReceipt.CaseID || result.Semantic.Input != approvedReceipt.Evidence.InputValue || result.Semantic.Operation != approvedReceipt.Evidence.CandidateOperation || result.Semantic.Output != approvedReceipt.Evidence.CandidateResult || result.SourceDigest != report.SourceDigest || result.Semantic.SemanticSourceDigest != report.SemanticSourceDigest || result.SubjectSHA != headSHA || result.ExecutionID != report.ExecutionID {
		return projection{}, fmt.Errorf("ARTIFACT_SEMANTIC_PROVENANCE_MISMATCH")
	}
	result.ExpectedAuthorizationDigest = approved.Artifact.AuthorizationDigest
	if result.ObservedAuthorizationDigest != result.ExpectedAuthorizationDigest {
		return projection{}, fmt.Errorf("ARTIFACT_AUTHORIZATION_DIGEST_MISMATCH")
	}
	result.EffectDigest = approved.ExecutionReceiptDigest
	return result, nil
}

func readArtifact(path, headSHA string) (projection, error) {
	if !model.ValidHead(headSHA) || path == "" || !allowedTempPath(path) {
		return projection{}, fmt.Errorf("ARTIFACT_OBSERVATION/read/ARTIFACT_PATH_OUTSIDE_SAFE_TEMP_ROOT")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return projection{}, fmt.Errorf("ARTIFACT_OBSERVATION/read/ARTIFACT_BYTES_UNAVAILABLE: %w", err)
	}
	fields, err := parseArtifact(raw)
	if err != nil {
		return projection{}, fmt.Errorf("ARTIFACT_OBSERVATION/parse/ARTIFACT_SEMANTIC_BYTES_INVALID: %w", err)
	}
	if !model.ValidExecutionID(headSHA, fields.ExecutionID) || fields.CaseID == "" || !model.ValidDigest(fields.SourceDigest) || !model.ValidDigest(fields.SemanticSourceDigest) || !model.ValidDigest(fields.AuthorizationDigest) || fields.SubjectSHA != headSHA {
		return projection{}, fmt.Errorf("ARTIFACT_OBSERVATION/parse/ARTIFACT_PROVENANCE_INVALID")
	}
	semantic := semanticValue{CaseID: fields.CaseID, Input: fields.Input, Operation: fields.Operation, Output: fields.Output, SemanticSourceDigest: fields.SemanticSourceDigest}
	canonical, err := json.Marshal(semantic)
	if err != nil {
		return projection{}, err
	}
	return projection{Schema: "gooo/invariant-transformation-artifact-semantic-projection/v1", HeadSHA: headSHA, CaseID: fields.CaseID, Path: path, RawDigest: model.DigestBytes(raw), RawSize: len(raw), ExecutionID: fields.ExecutionID, SourceDigest: fields.SourceDigest, SubjectSHA: fields.SubjectSHA, ObservedAuthorizationDigest: fields.AuthorizationDigest, Semantic: semantic, CanonicalSemanticBytes: string(canonical), SemanticDigest: model.DigestBytes(canonical)}, nil
}

func parseArtifact(raw []byte) (artifactFields, error) {
	lines := strings.Split(string(raw), "\n")
	if len(lines) != 11 || lines[0] != "gooo bounded transformation artifact" || lines[len(lines)-1] != "" {
		return artifactFields{}, fmt.Errorf("unexpected artifact line framing")
	}
	keys := []string{"case", "execution", "input", "operation", "output", "source", "semantic-source", "authorization", "subject"}
	values := make(map[string]string, len(keys))
	for index, key := range keys {
		actual, value, ok := strings.Cut(lines[index+1], "=")
		if !ok || actual != key || value == "" || values[key] != "" {
			return artifactFields{}, fmt.Errorf("invalid %s field", key)
		}
		values[key] = value
	}
	input, err := strconv.ParseInt(values["input"], 10, 64)
	if err != nil {
		return artifactFields{}, err
	}
	output, err := strconv.ParseInt(values["output"], 10, 64)
	if err != nil {
		return artifactFields{}, err
	}
	return artifactFields{CaseID: values["case"], ExecutionID: values["execution"], Input: input, Operation: values["operation"], Output: output, SourceDigest: values["source"], SemanticSourceDigest: values["semantic-source"], AuthorizationDigest: values["authorization"], SubjectSHA: values["subject"]}, nil
}

func compareSemantic(left, right projection) error {
	if left.Schema != right.Schema || left.HeadSHA != right.HeadSHA || left.CaseID != right.CaseID || !reflect.DeepEqual(left.Semantic, right.Semantic) || left.CanonicalSemanticBytes != right.CanonicalSemanticBytes || left.SemanticDigest != right.SemanticDigest {
		return fmt.Errorf("ARTIFACT_SEMANTIC_PROJECTION_MISMATCH")
	}
	return nil
}

func compareFixedArtifactFields(left, right projection, except string) bool {
	if left.Schema != right.Schema || left.HeadSHA != right.HeadSHA || left.CaseID != right.CaseID || left.Semantic.CaseID != right.Semantic.CaseID || left.Semantic.CaseID != left.CaseID || left.ExecutionID != right.ExecutionID || left.RawSize != right.RawSize || left.Semantic.Input != right.Semantic.Input || left.Semantic.Operation != right.Semantic.Operation || left.SourceDigest != right.SourceDigest || left.Semantic.SemanticSourceDigest != right.Semantic.SemanticSourceDigest || left.SubjectSHA != right.SubjectSHA || left.ExpectedAuthorizationDigest != right.ExpectedAuthorizationDigest || left.EffectDigest != right.EffectDigest || !sameTempRoot(left.Path, right.Path) {
		return false
	}
	if except != "output" && left.Semantic.Output != right.Semantic.Output {
		return false
	}
	if except != "authorization" && left.ObservedAuthorizationDigest != right.ObservedAuthorizationDigest {
		return false
	}
	return true
}

func observeTamper(expected projection, path, headSHA, kind string) (tamperReceipt, error) {
	tampered, err := readArtifact(path, headSHA)
	if err != nil {
		return tamperReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/%s-tamper/observe/%s_TAMPER_NOT_OBSERVABLE: %w", kind, strings.ToUpper(kind), err)
	}
	tamperedPath := tampered.Path
	if !allowedTempPath(expected.Path) || !allowedTempPath(tamperedPath) {
		return tamperReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/%s-tamper/compare/%s_TAMPER_PATH_NOT_SAFE", kind, strings.ToUpper(kind))
	}
	// The fixture path is a transport location. Compare a logical artifact
	// path bound to the baseline, while retaining the observed fixture path in
	// the typed receipt and checking it independently above.
	tampered.Path = expected.Path
	// Expected authorization/effect are receipt-bound fields, not bytes in the
	// tamper fixture. Bind them to the independently reconstructed effect before
	// comparing every non-target projection field.
	tampered.ExpectedAuthorizationDigest = expected.ExpectedAuthorizationDigest
	tampered.EffectDigest = expected.EffectDigest
	if kind == "output" {
		if !compareFixedArtifactFields(expected, tampered, "output") || tampered.Semantic.Output == expected.Semantic.Output || tampered.RawDigest == expected.RawDigest {
			return tamperReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/output-tamper/compare/OUTPUT_TAMPER_NOT_OUTPUT_ONLY")
		}
		if compareSemantic(expected, tampered) == nil {
			return tamperReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/output-tamper/adjudicate/OUTPUT_SEMANTIC_TAMPER_ACCEPTED")
		}
		return newTamper("OUTPUT_ONLY", expected, tampered, tamperedPath, "output", strconv.FormatInt(expected.Semantic.Output, 10), strconv.FormatInt(tampered.Semantic.Output, 10), false, "ARTIFACT_TAMPER", "compare-output-only", "OUTPUT_ONLY_TAMPER_REJECTED"), nil
	}
	if !compareFixedArtifactFields(expected, tampered, "authorization") || tampered.ObservedAuthorizationDigest == expected.ObservedAuthorizationDigest || tampered.RawDigest == expected.RawDigest {
		return tamperReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/authorization-tamper/compare/AUTHORIZATION_TAMPER_NOT_AUTHORIZATION_ONLY")
	}
	if err := compareSemantic(expected, tampered); err != nil {
		return tamperReceipt{}, fmt.Errorf("ARTIFACT_CLOSURE/authorization-tamper/adjudicate/AUTHORIZATION_TAMPER_CHANGED_SEMANTICS")
	}
	return newTamper("AUTHORIZATION_ONLY", expected, tampered, tamperedPath, "authorization", expected.ObservedAuthorizationDigest, tampered.ObservedAuthorizationDigest, true, "ARTIFACT_TAMPER", "compare-authorization-only", "AUTHORIZATION_ONLY_TAMPER_REJECTED"), nil
}

func newTamper(kind string, expected, tampered projection, tamperedPath, field, baselineValue, tamperedValue string, semanticEqual bool, stage, step, reason string) tamperReceipt {
	receipt := tamperReceipt{Schema: "gooo/invariant-transformation-typed-tamper-receipt/v1", Kind: kind, CaseID: expected.CaseID, ExecutionID: expected.ExecutionID, BaselinePath: expected.Path, TamperedPath: tamperedPath, BaselineRawDigest: expected.RawDigest, TamperedRawDigest: tampered.RawDigest, ChangedField: field, BaselineValue: baselineValue, TamperedValue: tamperedValue, SemanticDigestEqual: semanticEqual, Rejected: true, Decision: model.DecisionFailClosed, Resolution: model.ResolutionExact, Stage: stage, Step: step, Reason: reason}
	evidence := receipt
	evidence.EvidenceDigest = ""
	evidence.Digest = ""
	receipt.EvidenceDigest = model.Digest(evidence)
	receipt.Digest = model.Digest(receipt)
	return receipt
}

func artifactAddress(metricID string, value projection) string { return metricID + ":" + value.Path }

func artifactBytesPayload(value projection) any {
	return struct {
		Relation, Path, RawDigest, ExecutionID, SourceDigest, SubjectSHA string
		RawSize                                                          int
	}{"artifact-bytes", value.Path, value.RawDigest, value.ExecutionID, value.SourceDigest, value.SubjectSHA, value.RawSize}
}

func rawProvenancePayload(value projection) any {
	return struct {
		Relation, Path, RawDigest, SourceDigest, SubjectSHA, ExecutionID string
		RawSize                                                          int
	}{"raw-provenance", value.Path, value.RawDigest, value.SourceDigest, value.SubjectSHA, value.ExecutionID, value.RawSize}
}

func rawProvenanceTargetDigest(value projection) string {
	return model.Digest(rawProvenancePayload(value))
}

func authorizationPayload(value projection) any {
	return struct {
		Relation, Path, Observed, Expected, Effect string
	}{"authorization", value.Path, value.ObservedAuthorizationDigest, value.ExpectedAuthorizationDigest, value.EffectDigest}
}

func toArtifactEvidence(value projection) artifactEvidence {
	return artifactEvidence{Path: value.Path, RawDigest: value.RawDigest, RawSize: value.RawSize, ExecutionID: value.ExecutionID, SourceDigest: value.SourceDigest, SubjectSHA: value.SubjectSHA, ObservedAuthorizationDigest: value.ObservedAuthorizationDigest, ExpectedAuthorizationDigest: value.ExpectedAuthorizationDigest, EffectDigest: value.EffectDigest, SemanticDigest: value.SemanticDigest, CanonicalSemanticBytes: value.CanonicalSemanticBytes}
}

func newMetric(id, occurrence, address, producer, consumer, operation, proof, targetAddress, targetDigest string, evidence any, stage, step, reason string) metricEvidence {
	row := metricEvidence{MetricID: id, Occurrence: occurrence, Address: address, Producer: producer, IndependentConsumer: consumer, MetaOperation: operation, ProofChoice: proof, TargetAddress: targetAddress, TargetDigest: targetDigest, Decision: model.DecisionPass, Resolution: model.ResolutionExact, Stage: stage, Step: step, Reason: reason}
	row.ObservedEvidenceDigest = model.Digest(metricDigestPayload{MetricID: row.MetricID, Occurrence: row.Occurrence, Address: row.Address, Producer: row.Producer, IndependentConsumer: row.IndependentConsumer, MetaOperation: row.MetaOperation, ProofChoice: row.ProofChoice, TargetAddress: row.TargetAddress, TargetDigest: row.TargetDigest, Decision: row.Decision, Resolution: row.Resolution, Stage: row.Stage, Step: row.Step, Reason: row.Reason, Evidence: evidence})
	return row
}

func sortedEvidence(rows []metricEvidence) []string {
	result := make([]string, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.ObservedEvidenceDigest)
	}
	sort.Strings(result)
	return result
}

func sortedEvidenceDigest(rows []metricEvidence) string { return model.Digest(sortedEvidence(rows)) }

type metricCounts struct {
	artifactBytes, rawProvenance, authorization, semanticEquality, outputTamper, authorizationTamper, commentOnly, finalClosure int
}

func validateMetricInventory(rows []metricEvidence) error {
	seenIDs := map[string]bool{}
	seenAddresses := map[string]bool{}
	allowed := map[string]bool{}
	for _, id := range expectedMetricIDs {
		allowed[id] = true
	}
	for _, row := range rows {
		if !allowed[row.MetricID] {
			return fmt.Errorf("ARTIFACT_CLOSURE/metrics/unexpected-metric-row/UNEXPECTED_METRIC_ID:%s", row.MetricID)
		}
		if seenIDs[row.MetricID] {
			return fmt.Errorf("ARTIFACT_CLOSURE/metrics/duplicate-metric-row/DUPLICATE_METRIC_ID:%s", row.MetricID)
		}
		if seenAddresses[row.Address] {
			return fmt.Errorf("ARTIFACT_CLOSURE/metrics/duplicate-metric-row/DUPLICATE_METRIC_ADDRESS:%s", row.Address)
		}
		seenIDs[row.MetricID], seenAddresses[row.Address] = true, true
		if row.Occurrence == "" || row.Address == "" || row.Producer == "" || row.IndependentConsumer == "" || row.MetaOperation == "" || row.ProofChoice == "" || row.TargetAddress == "" || !model.ValidDigest(row.TargetDigest) || !model.ValidDigest(row.ObservedEvidenceDigest) || row.Decision != model.DecisionPass || row.Resolution != model.ResolutionExact || row.Stage == "" || row.Step == "" || row.Reason == "" {
			return fmt.Errorf("ARTIFACT_CLOSURE/metrics/validate-row/METRIC_EVIDENCE_INCOMPLETE:%s", row.MetricID)
		}
	}
	for _, id := range expectedMetricIDs {
		if !seenIDs[id] {
			return fmt.Errorf("ARTIFACT_CLOSURE/metrics/missing-metric-row/MISSING_METRIC_ID:%s", id)
		}
	}
	if len(rows) != metricInventorySize {
		return fmt.Errorf("ARTIFACT_CLOSURE/metrics/inventory-count/METRIC_INVENTORY_COUNT_MISMATCH")
	}
	return nil
}

func deriveCounts(rows []metricEvidence) (metricCounts, error) {
	if err := validateMetricInventory(rows); err != nil {
		return metricCounts{}, err
	}
	var counts metricCounts
	for _, row := range rows {
		switch row.MetricID {
		case metricArtifactBytesFirst, metricArtifactBytesSecond:
			counts.artifactBytes++
		case metricSemanticEquality:
			counts.semanticEquality++
		case metricRawProvenanceFirst, metricRawProvenanceSecond:
			counts.rawProvenance++
		case metricAuthorizationFirst, metricAuthorizationSecond:
			counts.authorization++
		case metricOutputTamper:
			counts.outputTamper++
		case metricAuthorizationTamper:
			counts.authorizationTamper++
		case metricCommentOnly:
			counts.commentOnly++
		case metricFinalClosure:
			counts.finalClosure++
		}
	}
	return counts, nil
}

func deriveDecision(counts metricCounts) (string, string, string) {
	if counts.artifactBytes != metricArtifactBytesExpected || counts.rawProvenance != metricRawProvenanceExpected || counts.authorization != metricAuthorizationExpected || counts.semanticEquality != metricSemanticEqualityExpected || counts.outputTamper != metricOutputTamperExpected || counts.authorizationTamper != metricAuthorizationTamperExpected || counts.commentOnly != metricCommentOnlyExpected || counts.finalClosure != metricFinalClosureExpected {
		return model.DecisionFailClosed, model.ResolutionLower, "CLOSURE_METRIC_INVENTORY_NOT_SATISFIED"
	}
	return model.DecisionPass, model.ResolutionExact, "ALL_CLOSURE_METRIC_EVIDENCE_SATISFIED"
}

func evidenceDigestOnlyChanged(actual, expected closureReceipt) bool {
	if actual.MetricInventoryDigest == expected.MetricInventoryDigest || actual.Digest == expected.Digest || len(actual.MetricEvidence) != len(expected.MetricEvidence) {
		return false
	}
	changed := 0
	for index := range expected.MetricEvidence {
		left, right := actual.MetricEvidence[index], expected.MetricEvidence[index]
		evidenceChanged := left.ObservedEvidenceDigest != right.ObservedEvidenceDigest
		left.ObservedEvidenceDigest = ""
		right.ObservedEvidenceDigest = ""
		if !reflect.DeepEqual(left, right) {
			return false
		}
		if evidenceChanged {
			changed++
		}
	}
	return changed == 1
}

// --- Source-bound comment-only intervention reconstruction ---

const (
	interventionReportSchema   = "gooo/invariant-transformation-intervention-report/v2"
	interventionConsumerSchema = "gooo/invariant-transformation-intervention-consumer/v2"
	semanticExpectedID         = "semantic-expected-intervention"
	nonSemanticID              = "nonsemantic-source-intervention"
	nonSemanticKind            = "NON_SEMANTIC"
	nonSemanticStep            = "compare-nonsemantic-projection-and-decision"
	nonSemanticReason          = "NONSEMANTIC_PROJECTION_AND_DECISION_PRESERVED"
)

type fixture struct {
	Activity, CaseID, CaseKind                                                                                          string
	Input, CandidateResult, Expected                                                                                    int64
	CandidateOperation, Invariant, InvariantID, DomainID, OperationID, ReplayRecipe, EffectIntent, SemanticSourceDigest string
}

type fixtureProjection struct {
	Activity             string `json:"activity"`
	CaseID               string `json:"case_id"`
	CaseKind             string `json:"case_kind"`
	Input                int64  `json:"input"`
	CandidateOperation   string `json:"candidate_operation"`
	CandidateResult      int64  `json:"candidate_result"`
	Expected             int64  `json:"expected"`
	Invariant            string `json:"invariant"`
	InvariantID          string `json:"invariant_id"`
	DomainID             string `json:"input_domain_id"`
	OperationID          string `json:"operation_id"`
	ReplayRecipe         string `json:"replay_recipe"`
	SemanticSourceDigest string `json:"semantic_source_digest"`
	EffectIntent         string `json:"effect_intent"`
}

type interventionClaim struct {
	ID                string             `json:"id"`
	Status            string             `json:"status"`
	Resolution        string             `json:"resolution"`
	Reason            string             `json:"reason"`
	VerificationCheck string             `json:"verification_check"`
	Coordinate        model.Coordinate   `json:"coordinate"`
	TargetDigest      string             `json:"target_digest"`
	PriorStateDigest  string             `json:"prior_state_digest"`
	EvidenceDigest    string             `json:"evidence_digest"`
	Transitions       []model.Transition `json:"transitions"`
}

type coordinateWire struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type transitionWire struct {
	ClaimID                  string         `json:"claim_id"`
	From                     string         `json:"from"`
	To                       string         `json:"to"`
	Coordinate               coordinateWire `json:"coordinate"`
	PropositionDigest        string         `json:"proposition_digest"`
	PriorStateDigest         string         `json:"prior_state_digest"`
	EvidenceDigest           string         `json:"evidence_digest"`
	PreviousTransitionDigest string         `json:"previous_transition_digest"`
	CurrentTransitionDigest  string         `json:"current_transition_digest"`
}

type interventionCase struct {
	ID                                        string                       `json:"id"`
	Kind                                      string                       `json:"kind"`
	SourceEdit                                string                       `json:"source_edit"`
	BaselineProjection                        fixtureProjection            `json:"baseline_projection"`
	MutatedProjection                         fixtureProjection            `json:"mutated_projection"`
	BaselineProjectionDigest                  string                       `json:"baseline_projection_digest"`
	MutatedProjectionDigest                   string                       `json:"mutated_projection_digest"`
	BaselineSourceDigest                      string                       `json:"baseline_source_digest"`
	MutatedSourceDigest                       string                       `json:"mutated_source_digest"`
	BaselineProvenanceDigest                  string                       `json:"baseline_provenance_digest"`
	MutatedProvenanceDigest                   string                       `json:"mutated_provenance_digest"`
	ProvenanceDigestChanged                   bool                         `json:"provenance_digest_changed"`
	BaselineSemanticDigest                    string                       `json:"baseline_semantic_digest"`
	MutatedSemanticDigest                     string                       `json:"mutated_semantic_digest"`
	SemanticDigestEqual                       bool                         `json:"semantic_digest_equal"`
	BaselineReceiptDigest                     string                       `json:"baseline_receipt_digest"`
	MutatedReceiptDigest                      string                       `json:"mutated_receipt_digest"`
	BaselineReceiptDecision                   string                       `json:"baseline_receipt_decision"`
	MutatedReceiptDecision                    string                       `json:"mutated_receipt_decision"`
	BaselineJudgment                          model.Judgment               `json:"baseline_judgment"`
	MutatedJudgment                           model.Judgment               `json:"mutated_judgment"`
	BaselineEvidence                          model.TransformationEvidence `json:"baseline_evidence"`
	MutatedEvidence                           model.TransformationEvidence `json:"mutated_evidence"`
	BaselineClaimTransitions                  []model.Transition           `json:"baseline_claim_transitions"`
	MutatedClaimTransitions                   []model.Transition           `json:"mutated_claim_transitions"`
	BaselineTransitionDigest                  string                       `json:"baseline_transition_digest"`
	MutatedTransitionDigest                   string                       `json:"mutated_transition_digest"`
	BaselineTransitionStatePathDigest         string                       `json:"baseline_transition_state_path_digest"`
	MutatedTransitionStatePathDigest          string                       `json:"mutated_transition_state_path_digest"`
	TransitionDigestEqual                     bool                         `json:"transition_digest_equal"`
	TransitionDigestChanged                   bool                         `json:"transition_digest_changed"`
	RawSourceDigestChanged                    bool                         `json:"raw_source_digest_changed"`
	ReceiptChanged                            bool                         `json:"receipt_changed"`
	SemanticProjectionEqual                   bool                         `json:"semantic_projection_equal"`
	DecisionEqual                             bool                         `json:"decision_equal"`
	ResolutionEqual                           bool                         `json:"resolution_equal"`
	ReasonEqual                               bool                         `json:"reason_equal"`
	DecisionChanged                           bool                         `json:"decision_changed"`
	TransitionStatePathEqual                  bool                         `json:"transition_state_path_equal"`
	EffectsEqual                              bool                         `json:"effects_equal"`
	ReplayObservationEqual                    bool                         `json:"replay_observation_equal"`
	EvidenceObservable                        bool                         `json:"evidence_observable"`
	RepositoryWritesNotClaimed                bool                         `json:"repository_writes_not_claimed"`
	BaselineRepositoryWrites                  int                          `json:"baseline_repository_writes"`
	MutatedRepositoryWrites                   int                          `json:"mutated_repository_writes"`
	BaselineRepositoryWritesObserved          bool                         `json:"baseline_repository_writes_observed"`
	MutatedRepositoryWritesObserved           bool                         `json:"mutated_repository_writes_observed"`
	BaselineRepositoryNetStatusUnchanged      bool                         `json:"baseline_repository_net_status_unchanged"`
	MutatedRepositoryNetStatusUnchanged       bool                         `json:"mutated_repository_net_status_unchanged"`
	BaselineRepositoryActualOrTransientWrites string                       `json:"baseline_repository_actual_or_transient_writes"`
	MutatedRepositoryActualOrTransientWrites  string                       `json:"mutated_repository_actual_or_transient_writes"`
	BaselineRepositoryMutationAuthorized      bool                         `json:"baseline_repository_mutation_authorized"`
	MutatedRepositoryMutationAuthorized       bool                         `json:"mutated_repository_mutation_authorized"`
	Claim                                     interventionClaim            `json:"claim"`
	Satisfied                                 bool                         `json:"satisfied"`
}

type sliceDenominator struct {
	ID             string `json:"id"`
	CasesTotal     int    `json:"cases_total"`
	CasesSatisfied int    `json:"cases_satisfied"`
	CoverageBPS    int    `json:"coverage_bps"`
}
type fixedDenominator struct {
	ID                      string           `json:"id"`
	CasesTotal              int              `json:"cases_total"`
	SemanticExpectedChange  sliceDenominator `json:"semantic_expected_change"`
	SemanticOperationChange sliceDenominator `json:"semantic_operation_change"`
	NonSemantic             sliceDenominator `json:"nonsemantic_change"`
}
type failure struct {
	CaseID string `json:"case_id"`
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}
type gate struct {
	ID                     string                 `json:"id"`
	Scenario               string                 `json:"scenario"`
	CaseID                 string                 `json:"case_id"`
	SubjectSHA             string                 `json:"subject_sha"`
	Stage                  string                 `json:"stage"`
	Step                   string                 `json:"step"`
	AttemptPath            string                 `json:"attempt_path"`
	TargetPath             string                 `json:"target_path"`
	TargetBeforeExists     bool                   `json:"target_before_exists"`
	TargetAfterExists      bool                   `json:"target_after_exists"`
	TargetBeforeDigest     string                 `json:"target_before_digest"`
	TargetAfterDigest      string                 `json:"target_after_digest"`
	TargetBytesUnchanged   bool                   `json:"target_bytes_unchanged"`
	AuthorizationAttempted bool                   `json:"authorization_attempted"`
	AuthorizationAccepted  bool                   `json:"authorization_accepted"`
	ExecutorAccepted       bool                   `json:"executor_accepted"`
	ArtifactCount          int                    `json:"artifact_count"`
	ArtifactExists         bool                   `json:"artifact_exists"`
	Artifact               model.ArtifactEvidence `json:"artifact"`
	Reason                 string                 `json:"reason"`
	Satisfied              bool                   `json:"satisfied"`
}
type interventionReport struct {
	Schema                            string             `json:"schema"`
	HeadSHA                           string             `json:"head_sha"`
	SourcePath                        string             `json:"source_path"`
	SourceDigest                      string             `json:"source_digest"`
	Denominator                       fixedDenominator   `json:"denominator"`
	CaseCount                         int                `json:"case_count"`
	Cases                             []interventionCase `json:"cases"`
	EffectGates                       []gate             `json:"effect_gates"`
	EffectGateDenominator             int                `json:"effect_gate_denominator"`
	EffectGateSatisfied               int                `json:"effect_gate_satisfied"`
	Decision                          string             `json:"decision"`
	Resolution                        string             `json:"resolution"`
	Reason                            string             `json:"reason"`
	RepositoryWrites                  int                `json:"repository_writes"`
	RepositoryMutationAuthorized      bool               `json:"repository_mutation_authorized"`
	TempArtifactWriteAuthorized       bool               `json:"temp_artifact_write_authorized"`
	RepositoryNetStatusUnchanged      bool               `json:"repository_net_status_unchanged"`
	RepositoryNetState                string             `json:"repository_net_state"`
	RepositoryActualOrTransientWrites string             `json:"repository_actual_or_transient_writes"`
	RepositoryNetStatusObserved       bool               `json:"repository_net_status_observed"`
	ExecutedEffects                   int                `json:"executed_effects"`
	IndependentlyObservedEffects      int                `json:"independently_observed_effects"`
	UnknownEffectScopes               int                `json:"unknown_effect_scopes"`
	RepositoryPathAuthorization       bool               `json:"repository_path_authorization"`
	AmbientProcessAuthority           string             `json:"ambient_process_authority"`
	CorrectionCount                   int                `json:"correction_count"`
	CorrectionDenominator             int                `json:"correction_denominator"`
	Failure                           *failure           `json:"failure,omitempty"`
	Digest                            string             `json:"digest"`
}
type commentReceipt struct {
	Schema                            string `json:"schema"`
	CaseID                            string `json:"case_id"`
	BaselineRawDigest                 string `json:"baseline_raw_digest"`
	MutatedRawDigest                  string `json:"mutated_raw_digest"`
	BaselineProvenanceDigest          string `json:"baseline_provenance_digest"`
	MutatedProvenanceDigest           string `json:"mutated_provenance_digest"`
	BaselineSemanticDigest            string `json:"baseline_semantic_digest"`
	MutatedSemanticDigest             string `json:"mutated_semantic_digest"`
	SemanticDigestEqual               bool   `json:"semantic_digest_equal"`
	BaselineDecision                  string `json:"baseline_decision"`
	MutatedDecision                   string `json:"mutated_decision"`
	DecisionEqual                     bool   `json:"decision_equal"`
	BaselineTransitionDigest          string `json:"baseline_transition_digest"`
	MutatedTransitionDigest           string `json:"mutated_transition_digest"`
	BaselineTransitionStatePathDigest string `json:"baseline_transition_state_path_digest"`
	MutatedTransitionStatePathDigest  string `json:"mutated_transition_state_path_digest"`
	TransitionDigestEqual             bool   `json:"transition_digest_equal"`
	TransitionDigestChanged           bool   `json:"transition_digest_changed"`
	TransitionStatePathEqual          bool   `json:"transition_state_path_equal"`
	Stage                             string `json:"stage"`
	Step                              string `json:"step"`
	Reason                            string `json:"reason"`
	EvidenceDigest                    string `json:"evidence_digest"`
	Digest                            string `json:"digest"`
}
type interventionConsumer struct {
	Schema                                             string                 `json:"schema"`
	HeadSHA                                            string                 `json:"head_sha"`
	ProducerDependencyImports                          int                    `json:"producer_dependency_imports"`
	AllowedProducerDependencyImports                   int                    `json:"allowed_producer_dependency_imports"`
	ReconstructedCases                                 int                    `json:"reconstructed_cases"`
	ExpectedCases                                      int                    `json:"expected_cases"`
	ActualReplay                                       int                    `json:"actual_replay"`
	ExpectedActualReplay                               int                    `json:"expected_actual_replay"`
	ArtifactEvidence                                   model.ArtifactEvidence `json:"artifact_evidence"`
	ArtifactObserved                                   bool                   `json:"artifact_observed"`
	CoherentTamperRejected                             int                    `json:"coherent_tamper_rejected"`
	ExpectedCoherentTamperRejections                   int                    `json:"expected_coherent_tamper_rejections"`
	ContentObservationCoherentTamperRejected           int                    `json:"content_observation_coherent_tamper_rejected"`
	ExpectedContentObservationCoherentTamperRejections int                    `json:"expected_content_observation_coherent_tamper_rejections"`
	Decision                                           string                 `json:"decision"`
	Resolution                                         string                 `json:"resolution"`
	Reason                                             string                 `json:"reason"`
	RepositoryNetStatusUnchanged                       bool                   `json:"repository_net_status_unchanged"`
	RepositoryNetStatusObserved                        bool                   `json:"repository_net_status_observed"`
	RepositoryNetState                                 string                 `json:"repository_net_state"`
	RepositoryActualOrTransientWrites                  string                 `json:"repository_actual_or_transient_writes"`
	RepositoryPathAuthorization                        bool                   `json:"repository_path_authorization"`
	AmbientProcessAuthority                            string                 `json:"ambient_process_authority"`
	UnknownEffectScopes                                int                    `json:"unknown_effect_scopes"`
	CommentOnly                                        commentReceipt         `json:"comment_only"`
	Digest                                             string                 `json:"digest"`
}

type commentSubartifactPayload struct {
	ID                                string `json:"id"`
	Kind                              string `json:"kind"`
	SourceEdit                        string `json:"source_edit"`
	BaselineProjectionDigest          string `json:"baseline_projection_digest"`
	MutatedProjectionDigest           string `json:"mutated_projection_digest"`
	BaselineSourceDigest              string `json:"baseline_source_digest"`
	MutatedSourceDigest               string `json:"mutated_source_digest"`
	BaselineProvenanceDigest          string `json:"baseline_provenance_digest"`
	MutatedProvenanceDigest           string `json:"mutated_provenance_digest"`
	BaselineSemanticDigest            string `json:"baseline_semantic_digest"`
	MutatedSemanticDigest             string `json:"mutated_semantic_digest"`
	BaselineReceiptDigest             string `json:"baseline_receipt_digest"`
	MutatedReceiptDigest              string `json:"mutated_receipt_digest"`
	BaselineReceiptDecision           string `json:"baseline_receipt_decision"`
	MutatedReceiptDecision            string `json:"mutated_receipt_decision"`
	BaselineTransitionDigest          string `json:"baseline_transition_digest"`
	MutatedTransitionDigest           string `json:"mutated_transition_digest"`
	BaselineTransitionStatePathDigest string `json:"baseline_transition_state_path_digest"`
	MutatedTransitionStatePathDigest  string `json:"mutated_transition_state_path_digest"`
	TransitionDigestEqual             bool   `json:"transition_digest_equal"`
	TransitionDigestChanged           bool   `json:"transition_digest_changed"`
	RawSourceDigestChanged            bool   `json:"raw_source_digest_changed"`
	ReceiptChanged                    bool   `json:"receipt_changed"`
	SemanticProjectionEqual           bool   `json:"semantic_projection_equal"`
	DecisionEqual                     bool   `json:"decision_equal"`
	ResolutionEqual                   bool   `json:"resolution_equal"`
	ReasonEqual                       bool   `json:"reason_equal"`
	DecisionChanged                   bool   `json:"decision_changed"`
	TransitionStatePathEqual          bool   `json:"transition_state_path_equal"`
	EffectsEqual                      bool   `json:"effects_equal"`
	ReplayObservationEqual            bool   `json:"replay_observation_equal"`
	EvidenceObservable                bool   `json:"evidence_observable"`
	RepositoryWritesNotClaimed        bool   `json:"repository_writes_not_claimed"`
	ClaimTransitionDigest             string `json:"claim_transition_digest"`
	Satisfied                         bool   `json:"satisfied"`
}

type commentTransitionEvidence struct {
	Scope                        string `json:"scope"`
	SourceAddress                string `json:"source_address"`
	ParserLowererAddress         string `json:"parser_lowerer_address"`
	SemanticDigestAddress        string `json:"semantic_digest_address"`
	DecisionAddress              string `json:"decision_address"`
	ClaimTransitionAddress       string `json:"claim_transition_address"`
	HumanMetricAddress           string `json:"human_metric_address"`
	NonSemanticCaseDigest        string `json:"nonsemantic_case_digest"`
	ConsumerCommentDigest        string `json:"consumer_comment_digest"`
	BaselineRawDigest            string `json:"baseline_raw_digest"`
	MutatedRawDigest             string `json:"mutated_raw_digest"`
	BaselineProjectionDigest     string `json:"baseline_projection_digest"`
	MutatedProjectionDigest      string `json:"mutated_projection_digest"`
	BaselineProvenanceDigest     string `json:"baseline_provenance_digest"`
	MutatedProvenanceDigest      string `json:"mutated_provenance_digest"`
	BaselineSemanticDigest       string `json:"baseline_semantic_digest"`
	MutatedSemanticDigest        string `json:"mutated_semantic_digest"`
	BaselineDecision             string `json:"baseline_decision"`
	MutatedDecision              string `json:"mutated_decision"`
	BaselineDecisionDigest       string `json:"baseline_decision_digest"`
	MutatedDecisionDigest        string `json:"mutated_decision_digest"`
	ClaimTransitionDigest        string `json:"claim_transition_digest"`
	StatePathAddress             string `json:"state_path_address"`
	BaselineStatePathDigest      string `json:"baseline_state_path_digest"`
	MutatedStatePathDigest       string `json:"mutated_state_path_digest"`
	StatePathEqual               bool   `json:"state_path_equal"`
	FullTransitionDigestAddress  string `json:"full_transition_digest_address"`
	BaselineFullTransitionDigest string `json:"baseline_full_transition_digest"`
	MutatedFullTransitionDigest  string `json:"mutated_full_transition_digest"`
	FullTransitionDigestEqual    bool   `json:"full_transition_digest_equal"`
	HumanMetricTargetDigest      string `json:"human_metric_target_digest"`
	SemanticDigestEqual          bool   `json:"semantic_digest_equal"`
	DecisionEqual                bool   `json:"decision_equal"`
	TransitionDigestChanged      bool   `json:"transition_digest_changed"`
}

func commentSubartifactDigest(value *interventionCase) string {
	return model.Digest(commentSubartifactPayload{
		ID: value.ID, Kind: value.Kind, SourceEdit: value.SourceEdit,
		BaselineProjectionDigest: value.BaselineProjectionDigest, MutatedProjectionDigest: value.MutatedProjectionDigest,
		BaselineSourceDigest: value.BaselineSourceDigest, MutatedSourceDigest: value.MutatedSourceDigest,
		BaselineProvenanceDigest: value.BaselineProvenanceDigest, MutatedProvenanceDigest: value.MutatedProvenanceDigest,
		BaselineSemanticDigest: value.BaselineSemanticDigest, MutatedSemanticDigest: value.MutatedSemanticDigest,
		BaselineReceiptDigest: value.BaselineReceiptDigest, MutatedReceiptDigest: value.MutatedReceiptDigest,
		BaselineReceiptDecision: value.BaselineReceiptDecision, MutatedReceiptDecision: value.MutatedReceiptDecision,
		BaselineTransitionDigest: value.BaselineTransitionDigest, MutatedTransitionDigest: value.MutatedTransitionDigest, BaselineTransitionStatePathDigest: value.BaselineTransitionStatePathDigest, MutatedTransitionStatePathDigest: value.MutatedTransitionStatePathDigest,
		TransitionDigestEqual: value.TransitionDigestEqual, TransitionDigestChanged: value.TransitionDigestChanged, RawSourceDigestChanged: value.RawSourceDigestChanged,
		ReceiptChanged: value.ReceiptChanged, SemanticProjectionEqual: value.SemanticProjectionEqual, DecisionEqual: value.DecisionEqual,
		ResolutionEqual: value.ResolutionEqual, ReasonEqual: value.ReasonEqual, DecisionChanged: value.DecisionChanged,
		TransitionStatePathEqual: value.TransitionStatePathEqual, EffectsEqual: value.EffectsEqual,
		ReplayObservationEqual: value.ReplayObservationEqual, EvidenceObservable: value.EvidenceObservable,
		RepositoryWritesNotClaimed: value.RepositoryWritesNotClaimed, ClaimTransitionDigest: model.FullTransitionDigest(value.Claim.Transitions), Satisfied: value.Satisfied,
	})
}

func commentReceiptDigest(value commentReceipt) string { return model.Digest(value) }

func observeCommentMetric(interventionRaw, consumerRaw, source []byte, headSHA string, artifact projection) (metricEvidence, error) {
	var report interventionReport
	if err := decodeStrict(interventionRaw, &report); err != nil {
		return metricEvidence{}, fmt.Errorf("ARTIFACT_CLOSURE/comment-only/parse-intervention/INTERVENTION_REPORT_NOT_STRICT: %w", err)
	}
	var consumer interventionConsumer
	if err := decodeStrict(consumerRaw, &consumer); err != nil {
		return metricEvidence{}, fmt.Errorf("ARTIFACT_CLOSURE/comment-only/parse-consumer/INTERVENTION_CONSUMER_RECEIPT_NOT_STRICT: %w", err)
	}
	if report.Schema != interventionReportSchema || report.HeadSHA != headSHA || report.SourcePath != model.SourcePath || report.SourceDigest != model.DigestBytes(source) || report.RepositoryNetState != model.RepositoryNetStateUnknown || consumer.Schema != interventionConsumerSchema || consumer.HeadSHA != headSHA {
		return metricEvidence{}, fmt.Errorf("ARTIFACT_CLOSURE/comment-only/bind-receipts/INTERVENTION_TOP_RESULT_INVALID")
	}
	var observed *interventionCase
	for index := range report.Cases {
		item := &report.Cases[index]
		if item.ID == nonSemanticID && observed != nil {
			return metricEvidence{}, fmt.Errorf("ARTIFACT_CLOSURE/comment-only/select-case/DUPLICATE_INTERVENTION_CASE_ID:%s", item.ID)
		}
		if item.ID == nonSemanticID {
			observed = item
		}
	}
	if observed == nil || observed.Kind != nonSemanticKind {
		return metricEvidence{}, fmt.Errorf("ARTIFACT_CLOSURE/comment-only/select-case/INTERVENTION_CASE_INVENTORY_INVALID")
	}
	expectedCase, expectedComment, err := expectedCommentEvidence(source, headSHA)
	if err != nil {
		return metricEvidence{}, err
	}
	if !reflect.DeepEqual(*observed, expectedCase) {
		return metricEvidence{}, fmt.Errorf("ARTIFACT_CLOSURE/comment-only/source-reconstruction/COMMENT_INTERVENTION_RESEALED_WITHOUT_SOURCE_MATCH")
	}
	expectedArtifact := model.ArtifactEvidence{Path: artifact.Path, ContentDigest: artifact.RawDigest, Size: artifact.RawSize, CaseID: artifact.CaseID, ExecutionID: artifact.ExecutionID, SubjectSHA: artifact.SubjectSHA, AuthorizationDigest: artifact.ObservedAuthorizationDigest, Producer: model.ProducerID, Executor: model.ExecutorID, Consumer: model.ConsumerID, EffectReceiptDigest: artifact.EffectDigest, RepositoryNetContentObserved: false, RepositoryNetContentUnchanged: false, RepositoryNetStatusObserved: false, RepositoryNetStatusUnchanged: false, RepositoryNetState: model.RepositoryNetStateUnknown}
	if !reflect.DeepEqual(consumer.ArtifactEvidence, expectedArtifact) {
		return metricEvidence{}, fmt.Errorf("ARTIFACT_CLOSURE/comment-only/source-reconstruction/INTERVENTION_CONSUMER_ARTIFACT_NOT_SOURCE_BOUND")
	}
	if !reflect.DeepEqual(consumer.CommentOnly, expectedComment) {
		return metricEvidence{}, fmt.Errorf("ARTIFACT_CLOSURE/comment-only/source-reconstruction/COMMENT_CONSUMER_RESEALED_WITHOUT_SOURCE_MATCH")
	}
	baselineTransitionDigest := model.FullTransitionDigest(observed.BaselineClaimTransitions)
	mutatedTransitionDigest := model.FullTransitionDigest(observed.MutatedClaimTransitions)
	baselineTransitionStatePathDigest := model.TransitionStatePathDigest(observed.BaselineClaimTransitions)
	mutatedTransitionStatePathDigest := model.TransitionStatePathDigest(observed.MutatedClaimTransitions)
	if observed.BaselineTransitionDigest != baselineTransitionDigest || observed.MutatedTransitionDigest != mutatedTransitionDigest || observed.BaselineTransitionStatePathDigest != baselineTransitionStatePathDigest || observed.MutatedTransitionStatePathDigest != mutatedTransitionStatePathDigest || observed.TransitionDigestEqual != (baselineTransitionDigest == mutatedTransitionDigest) || observed.TransitionDigestChanged != (baselineTransitionDigest != mutatedTransitionDigest) {
		return metricEvidence{}, fmt.Errorf("ARTIFACT_CLOSURE/comment-only/compare-transitions/TRANSITION_DIGEST_PROVENANCE_INVALID")
	}
	caseDigest := commentSubartifactDigest(observed)
	commentDigest := commentReceiptDigest(consumer.CommentOnly)
	claimTransitionDigest := model.FullTransitionDigest(observed.Claim.Transitions)
	humanMetricTargetDigest := model.Digest([]string{caseDigest, commentDigest})
	return newMetric(metricCommentOnly, "baseline<->mutated", "intervention:"+observed.ID, "invarianttransformation.intervention", "invarianttransformation.intervention-consumer", "compare-comment-only-source-provenance", model.ProofCoherence, "intervention:"+observed.ID+"/human-metric", model.Digest([]string{caseDigest, commentDigest}), commentTransitionEvidence{
		Scope: "NON_SEMANTIC_SUBARTIFACT_ONLY;SEMANTIC_CASES_OUT_OF_SCOPE", SourceAddress: model.SourcePath, ParserLowererAddress: "source:" + model.SourcePath + "/parser-lowerer", SemanticDigestAddress: "intervention:" + observed.ID + "/semantic-digest", DecisionAddress: "intervention:" + observed.ID + "/decision", ClaimTransitionAddress: "intervention:" + observed.ID + "/claim-transition", HumanMetricAddress: "intervention:" + observed.ID + "/human-metric", NonSemanticCaseDigest: caseDigest, ConsumerCommentDigest: commentDigest, BaselineRawDigest: observed.BaselineSourceDigest, MutatedRawDigest: observed.MutatedSourceDigest, BaselineProjectionDigest: observed.BaselineProjectionDigest, MutatedProjectionDigest: observed.MutatedProjectionDigest, BaselineProvenanceDigest: observed.BaselineProvenanceDigest, MutatedProvenanceDigest: observed.MutatedProvenanceDigest, BaselineSemanticDigest: observed.BaselineSemanticDigest, MutatedSemanticDigest: observed.MutatedSemanticDigest, BaselineDecision: observed.BaselineReceiptDecision, MutatedDecision: observed.MutatedReceiptDecision, BaselineDecisionDigest: model.Digest(observed.BaselineReceiptDecision), MutatedDecisionDigest: model.Digest(observed.MutatedReceiptDecision), ClaimTransitionDigest: claimTransitionDigest, StatePathAddress: "intervention:" + observed.ID + "/transition-state-path", BaselineStatePathDigest: baselineTransitionStatePathDigest, MutatedStatePathDigest: mutatedTransitionStatePathDigest, StatePathEqual: observed.TransitionStatePathEqual, FullTransitionDigestAddress: "intervention:" + observed.ID + "/full-transition-digest", BaselineFullTransitionDigest: baselineTransitionDigest, MutatedFullTransitionDigest: mutatedTransitionDigest, FullTransitionDigestEqual: observed.TransitionDigestEqual, HumanMetricTargetDigest: humanMetricTargetDigest, SemanticDigestEqual: observed.SemanticDigestEqual, DecisionEqual: observed.DecisionEqual, TransitionDigestChanged: observed.TransitionDigestChanged,
	}, "INTERVENTION", nonSemanticStep, nonSemanticReason), nil
}

func expectedCommentEvidence(source []byte, headSHA string) (interventionCase, commentReceipt, error) {
	mutated := append(append([]byte{}, source...), []byte("\n\n// non-semantic intervention: comment and whitespace only\n")...)
	baseFixture, err := parseFixture(source, "preserved-translation")
	if err != nil {
		return interventionCase{}, commentReceipt{}, err
	}
	mutatedFixture, err := parseFixture(mutated, "preserved-translation")
	if err != nil {
		return interventionCase{}, commentReceipt{}, err
	}
	baseReceipt := receiptFromFixture(baseFixture, source, headSHA)
	mutatedReceipt := receiptFromFixture(mutatedFixture, mutated, headSHA)
	baseJudgment, mutatedJudgment := judge.Judge(baseReceipt, source), judge.Judge(mutatedReceipt, mutated)
	baseProjection, mutatedProjection := fixtureToProjection(baseFixture), fixtureToProjection(mutatedFixture)
	baseProjectionDigest, mutatedProjectionDigest := model.Digest(baseProjection), model.Digest(mutatedProjection)
	baseTransitions, mutatedTransitions := transitions(baseReceipt), transitions(mutatedReceipt)
	semanticEqual := baseProjection.SemanticSourceDigest == mutatedProjection.SemanticSourceDigest
	decisionEqual := baseJudgment.Decision == mutatedJudgment.Decision && baseReceipt.Decision == mutatedReceipt.Decision
	resolutionEqual := baseJudgment.Resolution == mutatedJudgment.Resolution && baseReceipt.Resolution == mutatedReceipt.Resolution
	reasonEqual := baseJudgment.Reason == mutatedJudgment.Reason && baseReceipt.Reason == mutatedReceipt.Reason
	statePathEqual := transitionStatePathEqual(baseTransitions, mutatedTransitions)
	baselineTransitionDigest := model.FullTransitionDigest(baseTransitions)
	mutatedTransitionDigest := model.FullTransitionDigest(mutatedTransitions)
	baselineTransitionStatePathDigest := model.TransitionStatePathDigest(baseTransitions)
	mutatedTransitionStatePathDigest := model.TransitionStatePathDigest(mutatedTransitions)
	transitionDigestChanged := baselineTransitionDigest != mutatedTransitionDigest
	replayEqual := baseReceipt.Evidence.ReplayCount == mutatedReceipt.Evidence.ReplayCount && baseReceipt.Evidence.ReplayOperation == mutatedReceipt.Evidence.ReplayOperation && baseReceipt.Evidence.ReplayOutput == mutatedReceipt.Evidence.ReplayOutput && baseReceipt.Evidence.ReplayDigest == mutatedReceipt.Evidence.ReplayDigest && baseReceipt.Evidence.ReplaySemanticDigest == mutatedReceipt.Evidence.ReplaySemanticDigest && baseReceipt.Evidence.ReplayEvidenceDigest == mutatedReceipt.Evidence.ReplayEvidenceDigest
	provenanceChanged := baseReceipt.SourceDigest != mutatedReceipt.SourceDigest
	repositoryWritesNotClaimed := !baseReceipt.RepositoryWritesObserved && !mutatedReceipt.RepositoryWritesObserved && baseReceipt.RepositoryWrites == -1 && mutatedReceipt.RepositoryWrites == -1 && baseReceipt.RepositoryActualOrTransientWrites == model.UnknownEffectScope && mutatedReceipt.RepositoryActualOrTransientWrites == model.UnknownEffectScope && !baseReceipt.RepositoryMutationAuthorized && !mutatedReceipt.RepositoryMutationAuthorized
	evidenceObservable := baseJudgment.Independent && mutatedJudgment.Independent
	status, resolution, reason, satisfied := commentAdjudication([]commentGate{
		{known: true, satisfied: provenanceChanged, reason: "COMMENT_ONLY_RAW_PROVENANCE_NOT_CHANGED"},
		{known: true, satisfied: baseReceipt.SourceDigest != mutatedReceipt.SourceDigest, reason: "COMMENT_ONLY_RAW_SOURCE_DIGEST_NOT_CHANGED"},
		{known: true, satisfied: baseReceipt.Digest != mutatedReceipt.Digest, reason: "COMMENT_ONLY_RECEIPT_NOT_CHANGED"},
		{known: true, satisfied: semanticEqual, reason: "COMMENT_ONLY_SEMANTIC_DIGEST_CHANGED"},
		{known: true, satisfied: reflect.DeepEqual(baseProjection, mutatedProjection), reason: "COMMENT_ONLY_SEMANTIC_PROJECTION_CHANGED"},
		{known: evidenceObservable, satisfied: decisionEqual, reason: "COMMENT_ONLY_DECISION_CHANGED"},
		{known: evidenceObservable, satisfied: resolutionEqual, reason: "COMMENT_ONLY_RESOLUTION_CHANGED"},
		{known: evidenceObservable, satisfied: reasonEqual, reason: "COMMENT_ONLY_REASON_CHANGED"},
		{known: evidenceObservable, satisfied: statePathEqual, reason: "COMMENT_ONLY_TRANSITION_STATE_PATH_CHANGED"},
		{known: evidenceObservable, satisfied: transitionDigestChanged, reason: "COMMENT_ONLY_TRANSITION_DIGEST_NOT_CHANGED"},
		{known: evidenceObservable, satisfied: reflect.DeepEqual(baseReceipt.Effects, mutatedReceipt.Effects), reason: "COMMENT_ONLY_EFFECTS_CHANGED"},
		{known: evidenceObservable, satisfied: replayEqual, reason: "COMMENT_ONLY_REPLAY_OBSERVATION_CHANGED"},
		{known: evidenceObservable, satisfied: repositoryWritesNotClaimed, reason: "COMMENT_ONLY_REPOSITORY_WRITES_OBSERVATION_INVALID"},
		{known: evidenceObservable, satisfied: baseJudgment.Decision == model.DecisionAllowed && mutatedJudgment.Decision == model.DecisionAllowed, reason: "COMMENT_ONLY_JUDGMENT_NOT_ALLOWED"},
	})
	claimCoordinate := model.Coordinate{Stage: "INTERVENTION", Step: nonSemanticStep, Reason: nonSemanticReason}
	claimCoordinate.Reason = reason
	transitionEvidence := model.Digest([]any{baseProjectionDigest, mutatedProjectionDigest, baseJudgment.Decision, mutatedJudgment.Decision, baseJudgment.Resolution, mutatedJudgment.Resolution, baseJudgment.Reason, mutatedJudgment.Reason, baseReceipt.SourceDigest, mutatedReceipt.SourceDigest, provenanceDigest(baseReceipt.SourceDigest, headSHA, nonSemanticID), provenanceDigest(mutatedReceipt.SourceDigest, headSHA, nonSemanticID)})
	transition := model.NewTransition(nonSemanticID+"::claim", model.StatusOpen, status, claimCoordinate, transitionEvidence)
	claim := interventionClaim{ID: nonSemanticID + "::claim", Status: status, Resolution: resolution, Reason: reason, VerificationCheck: "intervention-observation-derived-from-two-independent-receipts", Coordinate: claimCoordinate, TargetDigest: transition.PropositionDigest, PriorStateDigest: transition.PriorStateDigest, EvidenceDigest: transition.EvidenceDigest, Transitions: []model.Transition{transition}}
	comment := commentReceipt{Schema: "gooo/invariant-transformation-comment-only-receipt/v1", CaseID: nonSemanticID, BaselineRawDigest: baseReceipt.SourceDigest, MutatedRawDigest: mutatedReceipt.SourceDigest, BaselineProvenanceDigest: provenanceDigest(baseReceipt.SourceDigest, headSHA, nonSemanticID), MutatedProvenanceDigest: provenanceDigest(mutatedReceipt.SourceDigest, headSHA, nonSemanticID), BaselineSemanticDigest: baseProjection.SemanticSourceDigest, MutatedSemanticDigest: mutatedProjection.SemanticSourceDigest, SemanticDigestEqual: semanticEqual, BaselineDecision: baseJudgment.Decision, MutatedDecision: mutatedJudgment.Decision, DecisionEqual: decisionEqual, BaselineTransitionDigest: baselineTransitionDigest, MutatedTransitionDigest: mutatedTransitionDigest, BaselineTransitionStatePathDigest: baselineTransitionStatePathDigest, MutatedTransitionStatePathDigest: mutatedTransitionStatePathDigest, TransitionDigestEqual: baselineTransitionDigest == mutatedTransitionDigest, TransitionDigestChanged: transitionDigestChanged, TransitionStatePathEqual: statePathEqual, Stage: "INTERVENTION", Step: nonSemanticStep, Reason: reason}
	item := interventionCase{ID: nonSemanticID, Kind: nonSemanticKind, SourceEdit: "comment-and-whitespace-only", BaselineProjection: baseProjection, MutatedProjection: mutatedProjection, BaselineProjectionDigest: baseProjectionDigest, MutatedProjectionDigest: mutatedProjectionDigest, BaselineSourceDigest: baseReceipt.SourceDigest, MutatedSourceDigest: mutatedReceipt.SourceDigest, BaselineProvenanceDigest: provenanceDigest(baseReceipt.SourceDigest, headSHA, nonSemanticID), MutatedProvenanceDigest: provenanceDigest(mutatedReceipt.SourceDigest, headSHA, nonSemanticID), ProvenanceDigestChanged: provenanceChanged, BaselineSemanticDigest: baseProjection.SemanticSourceDigest, MutatedSemanticDigest: mutatedProjection.SemanticSourceDigest, SemanticDigestEqual: semanticEqual, BaselineReceiptDigest: baseReceipt.Digest, MutatedReceiptDigest: mutatedReceipt.Digest, BaselineReceiptDecision: baseReceipt.Decision, MutatedReceiptDecision: mutatedReceipt.Decision, BaselineJudgment: baseJudgment, MutatedJudgment: mutatedJudgment, BaselineEvidence: baseReceipt.Evidence, MutatedEvidence: mutatedReceipt.Evidence, BaselineClaimTransitions: baseTransitions, MutatedClaimTransitions: mutatedTransitions, BaselineTransitionDigest: baselineTransitionDigest, MutatedTransitionDigest: mutatedTransitionDigest, BaselineTransitionStatePathDigest: baselineTransitionStatePathDigest, MutatedTransitionStatePathDigest: mutatedTransitionStatePathDigest, TransitionDigestEqual: baselineTransitionDigest == mutatedTransitionDigest, TransitionDigestChanged: transitionDigestChanged, RawSourceDigestChanged: provenanceChanged, ReceiptChanged: baseReceipt.Digest != mutatedReceipt.Digest, SemanticProjectionEqual: reflect.DeepEqual(baseProjection, mutatedProjection), DecisionEqual: decisionEqual, ResolutionEqual: resolutionEqual, ReasonEqual: reasonEqual, DecisionChanged: !decisionEqual, TransitionStatePathEqual: statePathEqual, EffectsEqual: reflect.DeepEqual(baseReceipt.Effects, mutatedReceipt.Effects), ReplayObservationEqual: replayEqual, EvidenceObservable: evidenceObservable, RepositoryWritesNotClaimed: repositoryWritesNotClaimed, BaselineRepositoryWrites: baseReceipt.RepositoryWrites, MutatedRepositoryWrites: mutatedReceipt.RepositoryWrites, BaselineRepositoryWritesObserved: baseReceipt.RepositoryWritesObserved, MutatedRepositoryWritesObserved: mutatedReceipt.RepositoryWritesObserved, BaselineRepositoryNetStatusUnchanged: baseReceipt.RepositoryNetStatusUnchanged, MutatedRepositoryNetStatusUnchanged: mutatedReceipt.RepositoryNetStatusUnchanged, BaselineRepositoryActualOrTransientWrites: baseReceipt.RepositoryActualOrTransientWrites, MutatedRepositoryActualOrTransientWrites: mutatedReceipt.RepositoryActualOrTransientWrites, BaselineRepositoryMutationAuthorized: baseReceipt.RepositoryMutationAuthorized, MutatedRepositoryMutationAuthorized: mutatedReceipt.RepositoryMutationAuthorized, Claim: claim, Satisfied: satisfied}
	evidence := comment
	evidence.EvidenceDigest = ""
	evidence.Digest = ""
	comment.EvidenceDigest = model.Digest(evidence)
	comment.Digest = model.Digest(comment)
	return item, comment, nil
}

type commentGate struct {
	known     bool
	satisfied bool
	reason    string
}

func commentAdjudication(gates []commentGate) (string, string, string, bool) {
	for _, gate := range gates {
		if gate.known && !gate.satisfied {
			return model.StatusRefuted, model.ResolutionInvariant, gate.reason, false
		}
	}
	for _, gate := range gates {
		if !gate.known {
			return model.StatusOpen, model.ResolutionLower, "INTERVENTION_EVIDENCE_UNOBSERVABLE", false
		}
	}
	return model.StatusDischarged, model.ResolutionExact, nonSemanticReason, true
}

func parseFixture(source []byte, caseID string) (fixture, error) {
	file, diagnostics := syntax.ParseFile(model.SourcePath, string(source))
	if diagnostics.HasErrors() {
		return fixture{}, fmt.Errorf("source syntax invalid: %s", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return fixture{}, err
	}
	semanticSourceDigest := "sha256:" + ir.StableHash()
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || len(activity.Parameters) != 0 || activity.Result.Name != "Transformation" || !activity.ValueProgramPresent {
			continue
		}
		fields, err := parseProgram(activity.ValueProgram)
		if err != nil {
			return fixture{}, err
		}
		if fields["case"] != caseID {
			continue
		}
		input, err := parseInt(fields["input"])
		if err != nil {
			return fixture{}, err
		}
		expected, err := parseInt(fields["expected"])
		if err != nil {
			return fixture{}, err
		}
		candidateResult, err := executeAdd(fields["candidate"], input)
		if err != nil {
			return fixture{}, err
		}
		if fields["invariant"] != "candidate-output-equals-expected" || fields["invariant-id"] != model.InvariantID || fields["domain"] != model.InputDomainID {
			return fixture{}, fmt.Errorf("fixture outside bounded contract")
		}
		return fixture{Activity: activity.Name, CaseID: caseID, CaseKind: fields["kind"], Input: input, CandidateOperation: fields["candidate"], CandidateResult: candidateResult, Expected: expected, Invariant: fields["invariant"], InvariantID: fields["invariant-id"], DomainID: fields["domain"], OperationID: model.Digest([]string{"operation", fields["candidate"]}), ReplayRecipe: fields["replay"], EffectIntent: fields["effect"], SemanticSourceDigest: semanticSourceDigest}, nil
	}
	return fixture{}, fmt.Errorf("fixture case %q missing", caseID)
}

func parseProgram(program string) (map[string]string, error) {
	parts := strings.Split(program, ";")
	if len(parts) != 10 {
		return nil, fmt.Errorf("computes value has %d fields, want 10", len(parts))
	}
	fields := map[string]string{}
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !ok || key == "" || value == "" || fields[key] != "" {
			return nil, fmt.Errorf("invalid or duplicate field %q", part)
		}
		fields[key] = value
	}
	for _, key := range []string{"case", "kind", "input", "candidate", "expected", "invariant", "invariant-id", "domain", "replay", "effect"} {
		if fields[key] == "" {
			return nil, fmt.Errorf("missing field %q", key)
		}
	}
	return fields, nil
}

func parseInt(value string) (int64, error) { return strconv.ParseInt(value, 10, 64) }

func executeAdd(operation string, input int64) (int64, error) {
	name, operandText, ok := strings.Cut(operation, ":")
	if !ok || name != "add" || operandText == "" || strings.Contains(operandText, ":") {
		return 0, fmt.Errorf("operation %q unsupported", operation)
	}
	operand, err := parseInt(operandText)
	if err != nil {
		return 0, err
	}
	const maxInt64 = int64(1<<63 - 1)
	const minInt64 = -maxInt64 - 1
	if (operand > 0 && input > maxInt64-operand) || (operand < 0 && input < minInt64-operand) {
		return 0, fmt.Errorf("operation overflows int64")
	}
	return input + operand, nil
}

func fixtureToProjection(value fixture) fixtureProjection {
	return fixtureProjection{Activity: value.Activity, CaseID: value.CaseID, CaseKind: value.CaseKind, Input: value.Input, CandidateOperation: value.CandidateOperation, CandidateResult: value.CandidateResult, Expected: value.Expected, Invariant: value.Invariant, InvariantID: value.InvariantID, DomainID: value.DomainID, OperationID: value.OperationID, ReplayRecipe: value.ReplayRecipe, SemanticSourceDigest: value.SemanticSourceDigest, EffectIntent: value.EffectIntent}
}

func receiptFromFixture(value fixture, source []byte, headSHA string) model.Receipt {
	sourceDigest := model.DigestBytes(source)
	candidateDigest := model.CandidateDigest(value.CandidateOperation, value.Input, value.CandidateResult)
	before, after, expected := model.SemanticDigest(value.Input), model.SemanticDigest(value.CandidateResult), model.SemanticDigest(value.Expected)
	replayOutput, replayErr := executeAdd(value.ReplayRecipe, value.Input)
	evidence := model.TransformationEvidence{SourceDigest: sourceDigest, SemanticSourceDigest: value.SemanticSourceDigest, CaseStableID: value.CaseID, ActivityStableID: value.Activity, OperationID: value.OperationID, InputDomainID: value.DomainID, InvariantID: value.InvariantID, EffectIntent: value.EffectIntent, InputValue: value.Input, CandidateOperation: value.CandidateOperation, CandidateResult: value.CandidateResult, ExpectedValue: value.Expected, Invariant: value.Invariant, CandidateDigest: candidateDigest, SemanticBeforeDigest: before, SemanticAfterDigest: after, ExpectedSemanticDigest: expected, ReplayRecipe: value.ReplayRecipe, BaselineInputValue: value.Input, BaselineOperation: value.CandidateOperation, BaselineOutput: value.CandidateResult, BaselineDigest: candidateDigest, ReplayCount: 1}
	if replayErr != nil {
		evidence.ReplayFailureStage, evidence.ReplayFailureStep, evidence.ReplayFailureReason = "REGRESSION", "execute-replay", "REGRESSION_REPLAY_RECIPE_UNAVAILABLE"
	} else {
		replayDigest := model.CandidateDigest(value.ReplayRecipe, value.Input, replayOutput)
		evidence.ReplayInputValue, evidence.ReplayOperation, evidence.ReplayOutput = value.Input, value.ReplayRecipe, replayOutput
		evidence.ReplayDigest, evidence.ReplaySemanticDigest = replayDigest, model.SemanticDigest(replayOutput)
		evidence.ReplayEvidenceDigest = model.ReplayDigest(candidateDigest, replayDigest)
		evidence.ReplayCount = 2
		evidence.RegressionWitnessPresent = candidateDigest == replayDigest && after == evidence.ReplaySemanticDigest
	}
	post := model.PostconditionDigest(before, after, expected)
	statuses := map[string]string{"precondition": model.StatusDischarged, "transformation": model.StatusDischarged, "postcondition": model.StatusDischarged, "regression-witness": model.StatusDischarged}
	reasons := map[string]string{"precondition": "EXACT_SOURCE_SNAPSHOT", "transformation": "TRANSFORMATION_OBSERVED", "postcondition": "SEMANTIC_POSTCONDITION_PRESERVED", "regression-witness": "REGRESSION_REPLAY_MATCHED"}
	if value.CandidateResult != value.Expected {
		statuses["postcondition"], reasons["postcondition"] = model.StatusRefuted, "SEMANTIC_POSTCONDITION_REFUTED"
	}
	if evidence.ReplayCount != 2 {
		statuses["regression-witness"], reasons["regression-witness"] = model.StatusOpen, evidence.ReplayFailureReason
	} else if !evidence.RegressionWitnessPresent || value.CandidateResult != value.Expected {
		statuses["regression-witness"], reasons["regression-witness"] = model.StatusRefuted, "REGRESSION_REPLAY_REFUTED"
	}
	claims, values := []model.Claim{}, []model.MetaValue{}
	for _, spec := range model.CanonicalValueSpecs() {
		evidenceDigest := sourceDigest
		if spec.ID == "transformation" {
			evidenceDigest = candidateDigest
		}
		if spec.ID == "postcondition" {
			evidenceDigest = post
		}
		if spec.ID == "regression-witness" {
			evidenceDigest = evidence.ReplayEvidenceDigest
		}
		coordinate := model.Coordinate{Stage: spec.Coordinate.Stage, Step: spec.Coordinate.Step, Reason: reasons[spec.ID]}
		id := value.CaseID + "::" + spec.ID
		transition := model.NewTransition(id, model.StatusOpen, statuses[spec.ID], coordinate, evidenceDigest)
		claim := model.Claim{ID: id, Status: statuses[spec.ID], Reason: reasons[spec.ID], VerificationCheck: spec.VerificationCheck, Coordinate: coordinate, TargetDigest: transition.PropositionDigest, PriorStateDigest: transition.PriorStateDigest, EvidenceDigests: digestList(evidenceDigest), Transitions: []model.Transition{transition}}
		claims = append(claims, claim)
		values = append(values, model.MetaValue{ID: spec.ID, Kind: spec.Kind, Value: statuses[spec.ID], EvidenceDigest: evidenceDigest, Producer: spec.Producer, Consumer: spec.Consumer, MetaOperation: spec.MetaOperation, ProofChoice: spec.ProofChoice, VerificationCheck: spec.VerificationCheck, Coordinate: coordinate})
	}
	decision, resolution, reason := deriveReceipt(claims)
	receipt := model.Receipt{Schema: model.ReceiptSchema, CaseID: value.CaseID, ExecutionID: headSHA + "::preserved-translation", CaseKind: value.CaseKind, ActivityStableID: value.Activity, HeadSHA: headSHA, SourcePath: model.SourcePath, SourceDigest: sourceDigest, SemanticSourceDigest: value.SemanticSourceDigest, ContractDigest: model.ValueContractDigest(), ValidatorContractDigest: model.ValidatorContractDigest(), Producer: model.ProducerID, Consumer: model.ConsumerID, MetaOperation: model.AuthorityOp, ProofChoice: model.ProofRegression, Values: values, Claims: claims, Evidence: evidence, Decision: decision, Resolution: resolution, Reason: reason, Phase: model.ReceiptProvisional, Effects: []model.Effect{}, RepositoryNetStatusObserved: false, RepositoryNetStatusUnchanged: false, RepositoryNetState: model.RepositoryNetStateUnknown, RepositoryActualOrTransientWrites: model.UnknownEffectScope, RepositoryWritesObserved: false, RepositoryWrites: -1, RepositoryMutationAuthorized: false, RepositoryPathAuthorization: false, AmbientProcessAuthority: model.UnknownEffectScope, AuthorityScope: model.AuthorityScope}
	receipt.AuthorizationDigest = model.AuthorizationDigest(receipt)
	return model.SealReceipt(receipt)
}

func deriveReceipt(claims []model.Claim) (string, string, string) {
	for _, claim := range claims {
		if claim.Status == model.StatusRefuted {
			return model.DecisionRefuted, model.ResolutionInvariant, claim.Reason
		}
	}
	for _, claim := range claims {
		if claim.Status == model.StatusOpen {
			return model.DecisionBlocked, model.ResolutionLower, claim.Reason
		}
	}
	return model.DecisionAllowed, model.ResolutionExact, "ALL_INVARIANTS_DISCHARGED"
}
func digestList(value string) []string {
	if value == "" {
		return []string{}
	}
	return []string{value}
}
func transitions(receipt model.Receipt) []model.Transition {
	result := []model.Transition{}
	for _, claim := range receipt.Claims {
		result = append(result, claim.Transitions...)
	}
	return result
}
func transitionWires(claims []model.Claim) []transitionWire {
	result := []transitionWire{}
	for _, claim := range claims {
		for _, transition := range claim.Transitions {
			result = append(result, transitionWire{ClaimID: transition.ClaimID, From: transition.From, To: transition.To, Coordinate: coordinateWire{Stage: transition.Coordinate.Stage, Step: transition.Coordinate.Step, Reason: transition.Coordinate.Reason}, PropositionDigest: transition.PropositionDigest, PriorStateDigest: transition.PriorStateDigest, EvidenceDigest: transition.EvidenceDigest, PreviousTransitionDigest: transition.PreviousTransitionDigest, CurrentTransitionDigest: transition.CurrentTransitionDigest})
		}
	}
	return result
}
func transitionStatePathEqual(left, right []model.Transition) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].ClaimID != right[index].ClaimID || left[index].From != right[index].From || left[index].To != right[index].To || left[index].Coordinate != right[index].Coordinate {
			return false
		}
	}
	return true
}
func provenanceDigest(sourceDigest, headSHA, caseID string) string {
	return model.Digest([]string{"invariant-transformation-source-provenance", sourceDigest, headSHA, caseID})
}

func expectedCommentReceipt(source []byte, headSHA string) (commentReceipt, error) {
	_, receipt, err := expectedCommentEvidence(source, headSHA)
	return receipt, err
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

func allowedTempPath(path string) bool {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	root := os.Getenv("RUNNER_TEMP")
	if root == "" {
		root = os.TempDir()
	}
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return false
	}
	canonicalRoot, err = filepath.Abs(canonicalRoot)
	if err != nil || !within(canonicalRoot, absolute) {
		return false
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false
	}
	resolved, err = filepath.Abs(resolved)
	return err == nil && within(canonicalRoot, resolved)
}
func sameTempRoot(left, right string) bool { return allowedTempPath(left) && allowedTempPath(right) }
func within(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != "." && relative != ".." && !filepath.IsAbs(relative) && !strings.HasPrefix(relative, ".."+string(os.PathSeparator))
}

// ResealEvidenceDigestFixture creates a coherent but source-independent
// metric-evidence reseal for CI. Verify must reject it by recomputing payloads.
func ResealEvidenceDigestFixture(raw []byte) ([]byte, error) {
	var value closureReceipt
	if err := decodeStrict(raw, &value); err != nil {
		return nil, err
	}
	if len(value.MetricEvidence) == 0 {
		return nil, fmt.Errorf("metric evidence is empty")
	}
	value.MetricEvidence[0].ObservedEvidenceDigest = model.Digest([]string{"forged-evidence-payload"})
	value.MetricInventoryDigest = model.Digest(value.MetricEvidence)
	value.Digest = ""
	value.Digest = model.Digest(value)
	return json.MarshalIndent(value, "", "  ")
}

// ResealCommentFixture changes one comment-only source observation in both
// intervention artifacts and reseals their own digests. Source reconstruction
// must still reject the pair.
func ResealCommentFixture(interventionRaw, consumerRaw []byte) ([]byte, []byte, error) {
	var report interventionReport
	var consumer interventionConsumer
	if err := decodeStrict(interventionRaw, &report); err != nil {
		return nil, nil, err
	}
	if err := decodeStrict(consumerRaw, &consumer); err != nil {
		return nil, nil, err
	}
	for index := range report.Cases {
		if report.Cases[index].ID == nonSemanticID {
			report.Cases[index].BaselineSourceDigest = model.Digest([]string{"forged-comment-baseline"})
			report.Cases[index].RawSourceDigestChanged = true
		}
	}
	consumer.CommentOnly.BaselineRawDigest = model.Digest([]string{"forged-comment-baseline"})
	return resealCommentPair(report, consumer)
}

// ResealCommentSemanticDigestFixture changes the reported semantic digest for
// the comment-only mutation and coherently reseals both nested artifacts. The
// independent source reconstruction must reject this semantic intervention.
func ResealCommentSemanticDigestFixture(interventionRaw, consumerRaw []byte) ([]byte, []byte, error) {
	var report interventionReport
	var consumer interventionConsumer
	if err := decodeStrict(interventionRaw, &report); err != nil {
		return nil, nil, err
	}
	if err := decodeStrict(consumerRaw, &consumer); err != nil {
		return nil, nil, err
	}
	for index := range report.Cases {
		if report.Cases[index].ID == nonSemanticID {
			forged := model.Digest([]string{"forged-comment-semantic-digest"})
			report.Cases[index].MutatedSemanticDigest = forged
			report.Cases[index].SemanticDigestEqual = false
		}
	}
	consumer.CommentOnly.MutatedSemanticDigest = model.Digest([]string{"forged-comment-semantic-digest"})
	consumer.CommentOnly.SemanticDigestEqual = false
	return resealCommentPair(report, consumer)
}

// ResealCommentGateFixture flips one calculated preservation gate and reseals
// the report. The independent verifier must derive the gate from source and
// reject the forged result rather than accepting its self-description.
func ResealCommentGateFixture(interventionRaw, consumerRaw []byte) ([]byte, []byte, error) {
	var report interventionReport
	var consumer interventionConsumer
	if err := decodeStrict(interventionRaw, &report); err != nil {
		return nil, nil, err
	}
	if err := decodeStrict(consumerRaw, &consumer); err != nil {
		return nil, nil, err
	}
	for index := range report.Cases {
		if report.Cases[index].ID == nonSemanticID {
			report.Cases[index].SemanticProjectionEqual = false
		}
	}
	return resealCommentPair(report, consumer)
}

// ResealCommentTransitionFixture changes one transition representation in the
// producer report and reseals only that report. The source-bound closure check
// must reject state, path, and full-digest substitutions independently.
func ResealCommentTransitionFixture(interventionRaw []byte, kind string) ([]byte, error) {
	var report interventionReport
	if err := decodeStrict(interventionRaw, &report); err != nil {
		return nil, err
	}
	found := false
	for index := range report.Cases {
		item := &report.Cases[index]
		if item.ID != nonSemanticID {
			continue
		}
		found = true
		if len(item.BaselineClaimTransitions) == 0 {
			return nil, fmt.Errorf("comment-only transition fixture requires a baseline transition")
		}
		switch kind {
		case "state":
			item.BaselineClaimTransitions[0].To = model.StatusRefuted
		case "path":
			item.BaselineClaimTransitions[0].Coordinate.Step = "tampered-transition-path"
		case "digest":
			item.BaselineTransitionDigest = model.Digest([]string{"forged-full-transition-digest"})
		default:
			return nil, fmt.Errorf("unknown comment transition fixture %q", kind)
		}
		break
	}
	if !found {
		return nil, fmt.Errorf("comment-only transition fixture case is missing")
	}
	report.Digest = ""
	report.Digest = model.Digest(report)
	return json.MarshalIndent(report, "", "  ")
}

// ResealSemanticCaseFixture changes only a semantic intervention case and
// reseals the producer report. The comment-only metric intentionally excludes
// that case; CI records this as an explicit out-of-scope invariant, not as
// comment-only evidence.
func ResealSemanticCaseFixture(interventionRaw []byte) ([]byte, error) {
	var report interventionReport
	if err := decodeStrict(interventionRaw, &report); err != nil {
		return nil, err
	}
	for index := range report.Cases {
		if report.Cases[index].ID != semanticExpectedID {
			continue
		}
		projection := report.Cases[index].MutatedProjection
		projection.Expected++
		report.Cases[index].MutatedProjection = projection
		report.Cases[index].MutatedProjectionDigest = model.Digest(projection)
		item := &report.Cases[index]
		transitionEvidence := model.Digest([]any{item.BaselineProjectionDigest, item.MutatedProjectionDigest, item.BaselineJudgment.Decision, item.MutatedJudgment.Decision, item.BaselineJudgment.Resolution, item.MutatedJudgment.Resolution, item.BaselineJudgment.Reason, item.MutatedJudgment.Reason, item.BaselineSourceDigest, item.MutatedSourceDigest, item.BaselineProvenanceDigest, item.MutatedProvenanceDigest})
		transition := model.NewTransition(item.Claim.ID, model.StatusOpen, item.Claim.Status, item.Claim.Coordinate, transitionEvidence)
		item.Claim.TargetDigest = transition.PropositionDigest
		item.Claim.PriorStateDigest = transition.PriorStateDigest
		item.Claim.EvidenceDigest = transition.EvidenceDigest
		item.Claim.Transitions = []model.Transition{transition}
		break
	}
	report.Digest = ""
	report.Digest = model.Digest(report)
	return json.MarshalIndent(report, "", "  ")
}

func resealCommentPair(report interventionReport, consumer interventionConsumer) ([]byte, []byte, error) {
	consumer.CommentOnly.EvidenceDigest = ""
	consumer.CommentOnly.Digest = ""
	consumer.CommentOnly.EvidenceDigest = model.Digest(commentReceipt{Schema: consumer.CommentOnly.Schema, CaseID: consumer.CommentOnly.CaseID, BaselineRawDigest: consumer.CommentOnly.BaselineRawDigest, MutatedRawDigest: consumer.CommentOnly.MutatedRawDigest, BaselineProvenanceDigest: consumer.CommentOnly.BaselineProvenanceDigest, MutatedProvenanceDigest: consumer.CommentOnly.MutatedProvenanceDigest, BaselineSemanticDigest: consumer.CommentOnly.BaselineSemanticDigest, MutatedSemanticDigest: consumer.CommentOnly.MutatedSemanticDigest, SemanticDigestEqual: consumer.CommentOnly.SemanticDigestEqual, BaselineDecision: consumer.CommentOnly.BaselineDecision, MutatedDecision: consumer.CommentOnly.MutatedDecision, DecisionEqual: consumer.CommentOnly.DecisionEqual, BaselineTransitionDigest: consumer.CommentOnly.BaselineTransitionDigest, MutatedTransitionDigest: consumer.CommentOnly.MutatedTransitionDigest, BaselineTransitionStatePathDigest: consumer.CommentOnly.BaselineTransitionStatePathDigest, MutatedTransitionStatePathDigest: consumer.CommentOnly.MutatedTransitionStatePathDigest, TransitionDigestEqual: consumer.CommentOnly.TransitionDigestEqual, TransitionDigestChanged: consumer.CommentOnly.TransitionDigestChanged, TransitionStatePathEqual: consumer.CommentOnly.TransitionStatePathEqual, Stage: consumer.CommentOnly.Stage, Step: consumer.CommentOnly.Step, Reason: consumer.CommentOnly.Reason})
	consumer.CommentOnly.Digest = model.Digest(consumer.CommentOnly)
	consumer.Digest = ""
	consumer.Digest = model.Digest(consumer)
	report.Digest = ""
	report.Digest = model.Digest(report)
	left, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	right, err := json.MarshalIndent(consumer, "", "  ")
	return append(left, '\n'), append(right, '\n'), err
}
