package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

func runBaselineGenerateReport(options generateOptions, input generateInput, jsonMode bool, stdout, stderr io.Writer, deadline time.Time) int {
	started := time.Now()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	artifacts, code := buildGenerateArtifacts(options, input, jsonMode, stdout, stderr, deadline)
	if code != exitOK {
		return code
	}
	if code := writeGenerateArtifacts(artifacts, jsonMode, stdout, stderr); code != exitOK {
		return code
	}
	runtime.ReadMemStats(&after)
	manifestData, err := jsonManifestBytes(artifacts.manifest)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: generate report: %v\n", err)
		return exitFailure
	}
	report := generation.SemanticPublicGenerationReport{
		Schema: generation.SemanticPublicGenerationReportSchema, Lifecycle: generation.SemanticPublicGenerationBaselineLifecycle,
		Decision: "CLOSED", Reason: generation.SemanticPublicGenerationBaselineReason,
		InputSourceDigest: cache.HashBytes(input.source).String(), NormalizedIRDigest: artifacts.ir.StableHash(),
		GeneratedOutputDigest: cache.HashBytes(artifacts.result.Source).String(), GeneratedManifestDigest: cache.HashBytes(manifestData).String(),
		OutputFile: artifacts.output, ManifestFile: artifacts.manifestPath,
		ArtifactCount: generation.SemanticPublicGenerationArtifactsClosed, OutputFileCount: generation.SemanticPublicGenerationArtifactsClosed,
		Metrics: generation.SemanticRetentionRuntimeMetrics{SemanticOperationCount: 1, GeneratedBytesEqual: true, NormalizedSemanticEqual: true,
			WallMS: retentionWallMS(started), PeakRSSKib: readPeakRSSKib(), AllocationCount: int64(after.Mallocs - before.Mallocs), AllocationBytes: int64(after.TotalAlloc - before.TotalAlloc)},
		RepositoryWrites: 0, LocalTestExecutions: 0,
		OutputBytes: int64(len(artifacts.result.Source) + len(manifestData)),
	}
	return writeBaselinePublicReport(options.outputDir, report, started, jsonMode, stdout, stderr)
}

func writeBaselinePublicReport(outputDir string, report generation.SemanticPublicGenerationReport, started time.Time, jsonMode bool, stdout, stderr io.Writer) int {
	if err := ensureOutputDirectory(outputDir); err != nil {
		fmt.Fprintf(stderr, "gooo: generate report: %v\n", err)
		return exitFailure
	}
	jsonPath := filepath.Join(outputDir, "generation-report.json")
	markdownPath := filepath.Join(outputDir, "generation-report.md")
	baseBytes := report.OutputBytes
	var jsonData, markdownData []byte
	for range 3 {
		var err error
		jsonData, err = publicReportJSON(report)
		if err != nil {
			fmt.Fprintf(stderr, "gooo: generate report: %v\n", err)
			return exitFailure
		}
		markdownData = []byte(publicReportMarkdown(report))
		total := baseBytes + int64(len(jsonData)+len(markdownData))
		if report.OutputBytes == total {
			break
		}
		report.OutputBytes = total
	}
	jsonData, err := publicReportJSON(report)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: generate report: %v\n", err)
		return exitFailure
	}
	markdownData = []byte(publicReportMarkdown(report))
	if err := generation.ValidateSemanticPublicGenerationReport(report); err != nil {
		fmt.Fprintf(stderr, "gooo: generate report: %v\n", err)
		return exitFailure
	}
	if err := writeAtomicFiles([]atomicWrite{{path: jsonPath, data: jsonData}, {path: markdownPath, data: markdownData}}); err != nil {
		fmt.Fprintf(stderr, "gooo: generate report output: %v\n", err)
		return exitFailure
	}
	if jsonMode {
		_, _ = stdout.Write(jsonData)
	} else {
		fmt.Fprintf(stdout, "generated: %s\nreport: %s\n", report.OutputFile, markdownPath)
	}
	return exitOK
}

func ensurePublicOutputDirectory(path string) error {
	if err := ensureOutputDirectory(path); err != nil {
		return err
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("inspect caller-owned output: %w", err)
	}
	if len(entries) != 0 {
		return fmt.Errorf("caller-owned output directory must be empty")
	}
	return nil
}

func publicReportJSON(report generation.SemanticPublicGenerationReport) ([]byte, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func publicReportMarkdown(report generation.SemanticPublicGenerationReport) string {
	knowledge := "the ordinary compiler baseline"
	if report.Decision == "CLOSED" && report.Lifecycle == generation.SemanticPublicGenerationRetainedLifecycle {
		knowledge = "the explicitly authorized retained-knowledge certificate"
	}
	output := ""
	if report.OutputFile != "" {
		output = fmt.Sprintf("\nOutput: `%s`\nManifest: `%s`\n", report.OutputFile, report.ManifestFile)
	}
	return fmt.Sprintf("# Gooo public generation\n\nDecision: `%s`\nReason: `%s`\n\nThis invocation used %s. Missing, stale, mismatched, or unauthorized evidence is never used as a baseline fallback.\n%s\nArtifacts: `%d`\nOutput bytes: `%d`\nSemantic operations: `%d`\nCertificate hits: `%d`\nCertificate misses: `%d`\nWall time (ms): `%d`\nPeak RSS (KiB): `%d`\n", report.Decision, report.Reason, knowledge, output, report.ArtifactCount, report.OutputBytes, report.Metrics.SemanticOperationCount, report.Metrics.CertificateHits, report.Metrics.CertificateMisses, report.Metrics.WallMS, report.Metrics.PeakRSSKib)
}
