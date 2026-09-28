package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publiccontinuity"
)

func decodeContinuityCertificate(data []byte) (publiccontinuity.Certificate, error) {
	var certificate publiccontinuity.Certificate
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&certificate); err != nil {
		return certificate, err
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return certificate, errors.New("continuity certificate contains multiple JSON values")
	} else if err != io.EOF {
		return certificate, fmt.Errorf("decode continuity certificate trailer: %w", err)
	}
	if err := publiccontinuity.ValidateCertificate(certificate); err != nil {
		return certificate, err
	}
	return certificate, nil
}

func writeContinuityGenerationRefuted(options generateOptions, input generateInput, reason, certificateDigest string, started time.Time, jsonMode bool, stdout, stderr io.Writer) int {
	report := publiccontinuity.Report{Schema: publiccontinuity.ReportSchema, Lifecycle: "CONSUMPTION", Decision: "REFUTED", Reason: reason,
		CaseID: continuityCaseBindingMismatch, CertificateDigest: certificateDigest, PublicInvocations: 2, LedgerEntries: 2,
		Certificates: 0, DigestContinuityEdgesExpected: 4, DigestContinuityEdgesObserved: 0, ManualTransformations: 0,
		SemanticOperationsBefore: 1, SemanticOperationsAfter: 0, CandidateCertificateByteReplayMismatches: 1,
		GeneratedBytesEqual: false, NormalizedSemanticEqual: false, ArtifactDenominator: 2, ArtifactCount: 2,
		RepositoryWrites: 0, LocalBuildExecutions: 0, LocalTestExecutions: 0, WallMS: continuityWallMS(started), PeakRSSKib: readPeakRSSKib()}
	if len(input.source) > 0 {
		report.Binding.SourceDigest = cache.HashBytes(input.source).String()
	}
	data, err := marshalContinuityJSON(report)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: public continuity generation: encode refuted report: %v\n", err)
		return exitFailure
	}
	human := []byte(renderContinuityReport(report, "The certified handoff was rejected fail-closed; no baseline fallback was used."))
	if err := writeContinuityArtifacts(options.outputDir, []continuityArtifact{
		{name: "continuity-generation-report.json", data: data},
		{name: "continuity-generation-report.md", data: human},
	}); err != nil {
		fmt.Fprintf(stderr, "gooo: public continuity generation: output: %v\n", err)
		return exitFailure
	}
	if jsonMode {
		_, _ = stdout.Write(data)
	} else {
		fmt.Fprintf(stdout, "continuity: REFUTED (%s)\n", filepath.Join(options.outputDir, "continuity-generation-report.md"))
	}
	return exitOK
}

const continuityCaseBindingMismatch = "BINDING_MISMATCH"
