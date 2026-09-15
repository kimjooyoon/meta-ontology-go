package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publiccontinuity"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicdiscovery"
)

func TestRunGeneratePreservesPreviousGoAndPublishesManifest(t *testing.T) {
	root := t.TempDir()
	sourcePath := filepath.Join(root, "main.gooo")
	if err := os.WriteFile(sourcePath, []byte(validSource), 0o640); err != nil {
		t.Fatal(err)
	}
	firstDir := filepath.Join(root, "first")
	var stdout, stderr bytes.Buffer
	if code := runGenerate([]string{sourcePath, "--out", firstDir}, OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("initial generate = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	firstPath := filepath.Join(firstDir, generatedFileName)
	first, err := os.ReadFile(firstPath)
	if err != nil {
		t.Fatal(err)
	}
	userBody := "return Order{\n\t\t// user-owned\n\t}"
	previous := bytes.Replace(first, []byte("return Order{}"), []byte(userBody), 1)
	if bytes.Equal(first, previous) {
		t.Fatal("fixture did not create a user-owned previous body")
	}
	previousPath := filepath.Join(root, "previous.go")
	if err := os.WriteFile(previousPath, previous, 0o640); err != nil {
		t.Fatal(err)
	}
	secondDir := filepath.Join(root, "second")
	manifestPath := filepath.Join(root, "evidence", "projection.jsonl")
	stdout.Reset()
	stderr.Reset()
	if code := runGenerate([]string{sourcePath, "--out", secondDir, "--previous-go", previousPath, "--manifest", manifestPath}, OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("previous-Go generate = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	second, err := os.ReadFile(filepath.Join(secondDir, generatedFileName))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(second, []byte("// user-owned")) || !bytes.Contains(second, []byte(userBody)) {
		t.Fatalf("previous Go slot was not preserved:\n%s", second)
	}
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest projectionManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("manifest is not JSONL JSON: %v", err)
	}
	if manifest.Schema != projectionManifestSchema || manifest.Status != "pass" || !manifest.PreviousGoProvided || manifest.PreviousGoDigest == "" || !manifest.ProtectedBytesEqual || manifest.ResponseDigest == "" || manifest.EvidenceManifest.PayloadSHA256 == "" {
		t.Fatalf("incomplete previous-Go manifest: %#v", manifest)
	}
	if got, err := os.ReadFile(sourcePath); err != nil || !bytes.Equal(got, []byte(validSource)) {
		t.Fatalf("source was modified: %v", err)
	}
}

func TestRunGenerateContinuityBindsPreviousGoInput(t *testing.T) {
	_, testFilename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("test filename is unavailable")
	}
	reader := rootedContinuityReader{root: filepath.Clean(filepath.Join(filepath.Dir(testFilename), "..", "..")), files: map[string][]byte{}}
	root := t.TempDir()
	sourcePath := filepath.Join(root, "main.gooo")
	certificatePath := filepath.Join(root, "continuity-certificate.json")
	reader.files[sourcePath] = []byte(validSource)
	certificateData := continuityCertificateFixture(t, reader, []byte(validSource))
	reader.files[certificatePath] = certificateData

	positiveDir := filepath.Join(root, "positive")
	var stdout, stderr bytes.Buffer
	if code := runGenerate([]string{sourcePath, "--continuity-certificate", certificatePath, "--out", positiveDir}, reader, SyntaxSourceParser{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("empty previous-Go continuity generate = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	positiveOutput, err := os.ReadFile(filepath.Join(positiveDir, generatedFileName))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(positiveOutput, []byte("package generated\n")) {
		t.Fatalf("empty previous-Go continuity output changed: %q", positiveOutput)
	}
	positiveReport, err := os.ReadFile(filepath.Join(positiveDir, "continuity-generation-report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var accepted publiccontinuity.Report
	if err := json.Unmarshal(positiveReport, &accepted); err != nil {
		t.Fatal(err)
	}
	if accepted.Decision != "CLOSED" || accepted.CaseID != "COMPLETE_ACCEPTED_CHAIN" {
		t.Fatalf("empty previous-Go continuity report = %#v", accepted)
	}

	previousPath := filepath.Join(root, "different-previous.go")
	reader.files[previousPath] = []byte("package different\n")
	negativeDir := filepath.Join(root, "negative")
	stdout.Reset()
	stderr.Reset()
	if code := runGenerate([]string{sourcePath, "--continuity-certificate", certificatePath, "--previous-go", previousPath, "--out", negativeDir}, reader, SyntaxSourceParser{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("mismatched previous-Go continuity generate = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	negativeReport, err := os.ReadFile(filepath.Join(negativeDir, "continuity-generation-report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var refuted publiccontinuity.Report
	if err := json.Unmarshal(negativeReport, &refuted); err != nil {
		t.Fatal(err)
	}
	if refuted.Decision != "REFUTED" || refuted.Reason != "BINDING_MISMATCH" || refuted.CaseID != "BINDING_MISMATCH" {
		t.Fatalf("mismatched previous-Go continuity report = %#v", refuted)
	}
	if _, err := os.Stat(filepath.Join(negativeDir, generatedFileName)); !os.IsNotExist(err) {
		t.Fatalf("mismatched previous-Go continuity wrote generated output: %v", err)
	}
}

type rootedContinuityReader struct {
	root  string
	files map[string][]byte
}

func (reader rootedContinuityReader) ReadFile(filename string) ([]byte, error) {
	if data, ok := reader.files[filename]; ok {
		return append([]byte(nil), data...), nil
	}
	return os.ReadFile(filepath.Join(reader.root, filename))
}

func continuityCertificateFixture(t *testing.T, reader SourceReader, source []byte) []byte {
	t.Helper()
	compilerDigest, err := publiccontinuity.CompilerDigest(reader.ReadFile)
	if err != nil {
		t.Fatal(err)
	}
	verifierDigest, err := publiccontinuity.VerifierDigest(reader.ReadFile)
	if err != nil {
		t.Fatal(err)
	}
	contractDigest := publicdiscovery.PolicySourceDigest()
	evaluatorDigest := publicdiscovery.GeneratedEvaluatorDigest()
	previousDigest := cache.HashBytes(nil).String()
	generatedSource := []byte("package generated\n")
	generatedManifest := []byte("{}\n")
	candidate := publicdiscovery.Candidate{
		Schema: publicdiscovery.CandidateSchema, CandidateID: "continuity-test-candidate", Operation: publicdiscovery.Operation,
		Decision: "CLOSED", Reason: publicdiscovery.ReasonClosed, GroupKeyDigest: cache.HashBytes([]byte("group")).String(),
		SourceDigest: cache.HashBytes(source).String(), InputSemanticDigest: cache.HashBytes([]byte("input-semantic")).String(),
		PreviousGoDigest: previousDigest, ToolchainDigest: generation.SemanticRetentionToolchainDigest(), ContractDigest: contractDigest,
		EvaluatorDigest: evaluatorDigest, GeneratedSemanticDigest: cache.HashBytes([]byte("generated-semantic")).String(),
		GeneratedOutputDigest: cache.HashBytes(generatedSource).String(), GeneratedManifestDigest: cache.HashBytes(generatedManifest).String(),
		Quorum: publicdiscovery.Quorum, AuthorizationRequired: true, ProposalRequired: true, CertificateRequired: true,
	}
	candidateData, err := json.Marshal(candidate)
	if err != nil {
		t.Fatal(err)
	}
	candidateDigest := cache.HashBytes(candidateData).String()
	binding := publiccontinuity.BindingFromCandidate(candidate, candidateDigest)
	certificate := publiccontinuity.Certificate{
		Schema: publiccontinuity.CertificateSchema, Mode: publiccontinuity.CertificateMode, ConversionSchema: publiccontinuity.ConversionSchema,
		SourceOperation: publiccontinuity.Operation, TargetOperation: "gooo.generate.public-self-observation-consumption",
		DecisionReceiptDigest: cache.HashBytes([]byte("decision")).String(), Binding: binding,
		ContractSourceDigest: contractDigest, InputSourceDigest: binding.SourceDigest, CompilerDigest: compilerDigest,
		VerifierDigest: verifierDigest, PolicyDigest: contractDigest, EvaluatorDigest: evaluatorDigest,
		GeneratedSource: generatedSource, GeneratedManifest: generatedManifest, GeneratedManifestDigest: binding.GeneratedManifestDigest,
	}
	certificate.CertificateID, err = publiccontinuity.CertificateContentDigest(certificate)
	if err != nil {
		t.Fatal(err)
	}
	if err := publiccontinuity.ValidateCertificate(certificate); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(certificate)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
