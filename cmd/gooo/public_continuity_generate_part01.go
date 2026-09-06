package main

import (
	"io"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publiccontinuity"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicdiscovery"
)

func runPublicContinuityGenerate(options generateOptions, input generateInput, reader SourceReader, jsonMode bool, stdout, stderr io.Writer, deadline time.Time) int {
	started := time.Now()
	certificate, certificateDigest, reason := loadContinuityCertificate(options, input, reader)
	if reason != "" {
		return writeContinuityGenerationRefuted(options, input, reason, certificateDigest, started, jsonMode, stdout, stderr)
	}
	return writeContinuityGenerationAccepted(options, certificate, certificateDigest, started, jsonMode, stdout, stderr, deadline)
}

func loadContinuityCertificate(options generateOptions, input generateInput, reader SourceReader) (publiccontinuity.Certificate, string, string) {
	data, err := reader.ReadFile(options.continuityCertificateFilename)
	if err != nil {
		return publiccontinuity.Certificate{}, "", "MISSING_CERTIFICATE"
	}
	digest := cache.HashBytes(data).String()
	certificate, err := decodeContinuityCertificate(data)
	if err != nil {
		return publiccontinuity.Certificate{}, digest, "TAMPERED_CERTIFICATE"
	}
	if cache.HashBytes(input.source).String() != certificate.Binding.SourceDigest {
		return certificate, digest, "STALE_SOURCE"
	}
	if certificate.Binding.ToolchainDigest != generation.SemanticRetentionToolchainDigest() {
		return certificate, digest, "MISMATCHED_TOOLCHAIN"
	}
	compilerDigest, err := publiccontinuity.CompilerDigest(reader.ReadFile)
	if err != nil || compilerDigest != certificate.CompilerDigest {
		return certificate, digest, "STALE_COMPILER"
	}
	verifierDigest, err := publiccontinuity.VerifierDigest(reader.ReadFile)
	if err != nil || verifierDigest != certificate.VerifierDigest {
		return certificate, digest, "STALE_VERIFIER"
	}
	if certificate.Binding.ContractDigest != publicdiscovery.PolicySourceDigest() || certificate.Binding.EvaluatorDigest != publicdiscovery.GeneratedEvaluatorDigest() {
		return certificate, digest, "STALE_POLICY"
	}
	return certificate, digest, ""
}
