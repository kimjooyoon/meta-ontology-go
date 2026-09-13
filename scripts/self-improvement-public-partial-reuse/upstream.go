package main

import (
	"bytes"
	"debug/buildinfo"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicorchestration"
)

type upstreamReport = publicorchestration.Report

type upstreamFailure struct {
	Decision string `json:"decision"`
	Reason string `json:"reason"`
	Unknown *publicorchestration.UnknownState `json:"unknown"`
	Diagnostic string `json:"diagnostic,omitempty"`
}

func (failure *upstreamFailure) Error() string {
	data, _ := json.Marshal(failure)
	return string(data)
}

func upstreamReject(reason string, err error) error {
	failure := &upstreamFailure{Decision: "REFUTED", Reason: reason}
	if err != nil {
		failure.Diagnostic = err.Error()
	}
	return failure
}

func upstreamUnknown(reason, step, class string, err error) error {
	blocked := []string{}
	if class == "DEPENDENCY_BLOCKED" {
		blocked = []string{"orchestration.verification"}
	}
	failure := &upstreamFailure{Decision: "UNKNOWN", Reason: reason, Unknown: &publicorchestration.UnknownState{
		Stage: "UPSTREAM_ORCHESTRATION", Step: step, Reason: reason, UnknownClass: class,
		NextOperation: "PROVIDE_AND_VERIFY_BOUND_ORCHESTRATION_EVIDENCE", BlockedBy: blocked,
	}}
	if err != nil {
		failure.Diagnostic = err.Error()
	}
	return failure
}

func decodeUpstream(data []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("upstream evidence contains trailing JSON")
	}
	return nil
}

func validateUpstreamClaim(report upstreamReport) error {
	if report.Schema != publicorchestration.ReportSchema || report.Operation != publicorchestration.Operation {
		return upstreamReject("UPSTREAM_REPORT_IDENTITY_INVALID", nil)
	}
	if report.Decision == "UNKNOWN" {
		return upstreamUnknown("UPSTREAM_DECISION_UNKNOWN", "READ_REPORT", "DEPENDENCY_BLOCKED", nil)
	}
	if report.Decision != "CLOSED" || report.Unknown != nil || report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 {
		return upstreamReject("UPSTREAM_REPORT_CONTRADICTS_AUTHORIZED_BOUNDARY", nil)
	}
	for _, digest := range upstreamBindings(report) {
		if !cache.Digest(digest).Known() {
			return upstreamUnknown("UPSTREAM_BINDING_MISSING", "READ_REPORT", "DIRECT_MISSING", nil)
		}
	}
	if report.CaseID != publicorchestration.CaseAuthorizedOrchestration || len(report.StatePath) == 0 {
		return upstreamReject("UPSTREAM_AUTHORIZED_PATH_MISSING", nil)
	}
	return nil
}

func upstreamBindings(report upstreamReport) [7]string {
	return [7]string{report.PolicySourceDigest, report.PolicySemanticDigest, report.PolicyEvaluatorDigest,
		report.HandoffDigest, report.AuthorizationDigest, report.CertificateDigest, report.ReceiptDigest}
}

func pinUpstreamVerifier(input runInput, directory string) (string, error) {
	data, err := readRegular(input.OrchestrationVerifier)
	if err != nil {
		return "", upstreamUnknown("UPSTREAM_VERIFIER_NOT_ESTABLISHED", "PIN_VERIFIER", "DEPENDENCY_BLOCKED", err)
	}
	if cache.HashBytes(data).String() != strings.TrimPrefix(input.OrchestrationVerifierDigest, "sha256:") {
		return "", upstreamReject("UPSTREAM_VERIFIER_DIGEST_MISMATCH", nil)
	}
	filename := filepath.Join(directory, "orchestration-verifier")
	if err := writeNew(filename, data, 0o555); err != nil {
		return "", upstreamUnknown("UPSTREAM_VERIFIER_NOT_ESTABLISHED", "PIN_VERIFIER", "DEPENDENCY_BLOCKED", err)
	}
	info, err := buildinfo.ReadFile(filename)
	if err != nil {
		return "", upstreamUnknown("UPSTREAM_VERIFIER_NOT_ESTABLISHED", "PIN_VERIFIER", "DEPENDENCY_BLOCKED", err)
	}
	if err := validateUpstreamVerifierBuild(info, input.ExpectedProducerHead); err != nil {
		return "", err
	}
	return filename, nil
}

func validateUpstreamVerifierBuild(info *debug.BuildInfo, expectedHead string) error {
	settings := map[string]string{}
	for _, setting := range info.Settings {
		settings[setting.Key] = setting.Value
	}
	if info.Path != "github.com/kimjooyoon/meta-ontology-go/scripts/self-improvement-public-orchestration" ||
		info.GoVersion != "go1.27.0" || settings["vcs.revision"] != expectedHead || settings["vcs.modified"] != "false" {
		return upstreamReject("UPSTREAM_VERIFIER_BUILD_IDENTITY_MISMATCH", nil)
	}
	return nil
}

func verifyUpstream(input runInput) (upstreamReport, []byte, error) {
	var report upstreamReport
	data, err := readRegular(input.OrchestrationReport)
	if err != nil {
		return report, nil, upstreamUnknown("UPSTREAM_REPORT_UNAVAILABLE", "READ_REPORT", "DIRECT_MISSING", err)
	}
	if err := decodeUpstream(data, &report); err != nil {
		return report, nil, upstreamReject("UPSTREAM_REPORT_MALFORMED", err)
	}
	if err := validateUpstreamClaim(report); err != nil {
		return report, nil, err
	}
	if input.OrchestrationEvidence == "" || input.OrchestrationVerifier == "" || input.OrchestrationVerifierDigest == "" || input.ExpectedProducerHead == "" {
		return report, nil, upstreamUnknown("UPSTREAM_EVIDENCE_REQUIRED", "LOAD_EVIDENCE", "DIRECT_MISSING", nil)
	}
	directory := filepath.Join(input.Out, "upstream-proof")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return report, nil, upstreamUnknown("UPSTREAM_SCRATCH_UNAVAILABLE", "SNAPSHOT_EVIDENCE", "DIRECT_MISSING", err)
	}
	snapshot, err := snapshotUpstream(input.OrchestrationEvidence, directory)
	if err != nil {
		var pathError *os.PathError
		if errors.As(err, &pathError) {
			return report, nil, upstreamUnknown("UPSTREAM_MATERIAL_UNAVAILABLE", "SNAPSHOT_EVIDENCE", "DIRECT_MISSING", err)
		}
		return report, nil, upstreamReject("UPSTREAM_MANIFEST_INVALID", err)
	}
	if !bytes.Equal(data, snapshot.Resume) {
		return report, nil, upstreamReject("UPSTREAM_REPORT_NOT_BOUND_TO_BUNDLE", nil)
	}
	verifier, err := pinUpstreamVerifier(input, directory)
	if err != nil {
		return report, nil, err
	}
	verifiedPath := filepath.Join(directory, "verification.json")
	started := time.Now()
	command := exec.Command(verifier, "-mode", "verify", "-evidence-manifest", snapshot.Manifest,
		"-output", verifiedPath, "-human-output", filepath.Join(directory, "verification.md"))
	output, err := command.CombinedOutput()
	if err != nil {
		return report, nil, upstreamUnknown("UPSTREAM_VERIFICATION_NOT_ESTABLISHED", "VERIFY_EVIDENCE", "DEPENDENCY_BLOCKED", fmt.Errorf("%w: %s", err, output))
	}
	verifiedBytes, err := readRegular(verifiedPath)
	if err != nil {
		return report, nil, upstreamUnknown("UPSTREAM_VERIFICATION_OUTPUT_MISSING", "COMPARE_VERIFICATION", "DEPENDENCY_BLOCKED", err)
	}
	var verified upstreamReport
	if err := decodeUpstream(verifiedBytes, &verified); err != nil {
		return report, nil, upstreamReject("UPSTREAM_VERIFICATION_OUTPUT_MALFORMED", err)
	}
	if err := validateUpstreamClaim(verified); err != nil {
		return report, nil, err
	}
	if upstreamBindings(report) != upstreamBindings(verified) {
		return report, nil, upstreamReject("UPSTREAM_REPORT_CONTRADICTS_RECONSTRUCTION", nil)
	}
	witness := map[string]any{
		"schema": "gooo/partial-reuse-upstream-admission/v1", "decision": "CLOSED",
		"scope": "PINNED_VERIFIER_RAW_BUNDLE_RECONSTRUCTION", "cross_workflow_admission": "UNKNOWN",
		"report_digest": cache.HashBytes(data).String(), "manifest_digest": snapshot.OriginalDigest,
		"verification_digest": cache.HashBytes(verifiedBytes).String(), "member_digests": snapshot.Digests,
		"verifier_digest": strings.TrimPrefix(input.OrchestrationVerifierDigest, "sha256:"),
		"verifier_source_head": input.ExpectedProducerHead, "policy_source_digest": verified.PolicySourceDigest,
		"wall_ms": time.Since(started).Milliseconds(), "repository_mutation_authorized": false,
	}
	if err := writeJSON(filepath.Join(input.Out, "upstream-admission.json"), witness); err != nil {
		return report, nil, err
	}
	encoded, _ := json.Marshal(witness)
	fmt.Printf("UPSTREAM_EVIDENCE_ADMISSION=%s\n", encoded)
	return report, data, nil
}
