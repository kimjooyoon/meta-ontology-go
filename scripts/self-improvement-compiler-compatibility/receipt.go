package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/compatibilitypolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/compilercompatibility"
)

type generateSummary struct {
	SemanticHash string `json:"semantic_hash"`
}

type receiptContext struct {
	policy        compatibilitypolicy.Policy
	authorization compilercompatibility.Authorization
	input         []byte
	summary       generateSummary
	source        []byte
	manifest      []byte
}

func runReceipt(contractPath, inputPath, outputRoot, role, candidateID, compilerDigest, testResult, authorizationPath, outputPath string) error {
	if err := validateReceiptArguments(contractPath, inputPath, outputRoot, role, candidateID, compilerDigest, testResult, authorizationPath, outputPath); err != nil {
		return err
	}
	context, err := loadReceiptContext(contractPath, inputPath, outputRoot, role, candidateID, authorizationPath)
	if err != nil {
		return err
	}
	receipt, err := buildExecutionReceipt(context, role, candidateID, compilerDigest, testResult)
	if err != nil {
		return err
	}
	return writeJSON(outputPath, receipt)
}

func validateReceiptArguments(contractPath, inputPath, outputRoot, role, candidateID, compilerDigest, testResult, authorizationPath, outputPath string) error {
	if contractPath == "" || inputPath == "" || outputRoot == "" || role == "" || candidateID == "" || compilerDigest == "" || testResult == "" || authorizationPath == "" || outputPath == "" {
		return errors.New("receipt requires contract, input, output-root, role, candidate-stable-id, compiler-digest, test-result, authorization, and output")
	}
	return nil
}

func loadReceiptContext(contractPath, inputPath, outputRoot, role, candidateID, authorizationPath string) (receiptContext, error) {
	policy, err := loadPolicy(contractPath)
	if err != nil {
		return receiptContext{}, err
	}
	var authorization compilercompatibility.Authorization
	if _, err := readStrict(authorizationPath, &authorization); err != nil {
		return receiptContext{}, err
	}
	if err := compilercompatibility.ValidateAuthorization(authorization); err != nil {
		return receiptContext{}, fmt.Errorf("authorization: %w", err)
	}
	input, err := os.ReadFile(inputPath)
	if err != nil {
		return receiptContext{}, fmt.Errorf("read input: %w", err)
	}
	if authorization.CandidateStableID != candidateID || authorization.SubjectDigest != cache.HashBytes(input).String() || authorization.SuccessorCompilerDigest == "" {
		return receiptContext{}, errors.New("authorization does not bind the receipt subject")
	}
	summary, source, manifest, err := loadGeneratedReceipt(outputRoot)
	if err != nil {
		return receiptContext{}, err
	}
	if role != "predecessor" && role != "successor" {
		return receiptContext{}, fmt.Errorf("receipt role %q is invalid", role)
	}
	manifest, err = normalizeManifestGeneratedFile(manifest)
	if err != nil {
		return receiptContext{}, fmt.Errorf("normalize generated manifest: %w", err)
	}
	return receiptContext{policy: policy, authorization: authorization, input: input, summary: summary, source: source, manifest: manifest}, nil
}

func loadGeneratedReceipt(outputRoot string) (generateSummary, []byte, []byte, error) {
	var summary generateSummary
	summaryData, err := os.ReadFile(filepath.Join(outputRoot, "generate.json"))
	if err != nil {
		return summary, nil, nil, err
	}
	if err := json.Unmarshal(summaryData, &summary); err != nil {
		return summary, nil, nil, err
	}
	if summary.SemanticHash == "" {
		return summary, nil, nil, errors.New("generate summary has no semantic hash")
	}
	source, err := os.ReadFile(filepath.Join(outputRoot, "semantic.gooo.go"))
	if err != nil {
		return summary, nil, nil, fmt.Errorf("read generated source: %w", err)
	}
	manifest, err := os.ReadFile(filepath.Join(outputRoot, "semantic.gooo.manifest.jsonl"))
	if err != nil {
		return summary, nil, nil, fmt.Errorf("read generated manifest: %w", err)
	}
	return summary, source, manifest, nil
}

func buildExecutionReceipt(context receiptContext, role, candidateID, compilerDigest, testResult string) (compilercompatibility.ExecutionReceipt, error) {
	if context.authorization.SuccessorCompilerDigest != compilerDigest && role == "successor" {
		return compilercompatibility.ExecutionReceipt{}, errors.New("successor compiler digest differs from caller-owned authorization")
	}
	inputDigest := cache.HashBytes(context.input).String()
	receipt := compilercompatibility.ExecutionReceipt{
		Schema: compilercompatibility.ConsumptionSchema, Role: role, CandidateStableID: candidateID,
		SubjectDigest: inputDigest, SourceDigest: inputDigest, SemanticIRDigest: context.summary.SemanticHash,
		GeneratedOutputDigest: cache.HashBytes(context.source).String(), GeneratedManifestDigest: cache.HashBytes(context.manifest).String(),
		GeneratedSource: context.source, GeneratedManifest: context.manifest, PolicyDigest: context.policy.SourceDigest,
		PolicyEvaluatorDigest: context.policy.EvaluatorDigest, PolicyResult: compilercompatibility.DecisionClosed,
		CompilerImplementationDigest: compilerDigest, GoToolchainDigest: compilercompatibility.CurrentToolchainDigest(),
		TestContractDigest: compilercompatibility.TestContractDigest(), TestContractResult: testResult,
		AuthorizationDigest: context.authorization.AuthorizationID,
	}
	if err := compilercompatibility.ValidateExecutionReceipt(receipt); err != nil {
		return compilercompatibility.ExecutionReceipt{}, fmt.Errorf("execution receipt: %w", err)
	}
	return receipt, nil
}

func runCertify(contractPath, predecessorPath, successorPath, authorizationPath, outputPath string) error {
	if contractPath == "" || predecessorPath == "" || successorPath == "" || authorizationPath == "" || outputPath == "" {
		return errors.New("certify requires contract, predecessor, successor, authorization, and output")
	}
	policy, err := loadPolicy(contractPath)
	if err != nil {
		return err
	}
	var predecessor, successor compilercompatibility.ExecutionReceipt
	if _, err := readStrict(predecessorPath, &predecessor); err != nil {
		return err
	}
	if _, err := readStrict(successorPath, &successor); err != nil {
		return err
	}
	var authorization compilercompatibility.Authorization
	if _, err := readStrict(authorizationPath, &authorization); err != nil {
		return err
	}
	certificate, err := compilercompatibility.BuildCertificate(predecessor, successor, authorization, policy)
	if err != nil {
		return fmt.Errorf("build compatibility certificate: %w", err)
	}
	return writeJSON(outputPath, certificate)
}

func normalizeManifestGeneratedFile(data []byte) ([]byte, error) {
	key := []byte(`"generated_file":`)
	index := bytes.Index(data, key)
	if index < 0 {
		return nil, errors.New("generated manifest has no generated_file field")
	}
	valueStart := index + len(key)
	decoder := json.NewDecoder(bytes.NewReader(data[valueStart:]))
	var value string
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	valueBytes, err := json.Marshal("semantic.gooo.go")
	if err != nil {
		return nil, err
	}
	valueEnd := valueStart + int(decoder.InputOffset())
	normalized := make([]byte, 0, len(data))
	normalized = append(normalized, data[:valueStart]...)
	normalized = append(normalized, valueBytes...)
	normalized = append(normalized, data[valueEnd:]...)
	return normalized, nil
}

func readCertificate(path string) ([]byte, compilercompatibility.Certificate, error) {
	var certificate compilercompatibility.Certificate
	data, err := readStrict(path, &certificate)
	if err != nil {
		return nil, certificate, err
	}
	if err := compilercompatibility.ValidateCertificate(certificate); err != nil {
		return nil, certificate, err
	}
	return data, certificate, nil
}

func readConsumption(path string) (compilercompatibility.ConsumptionReport, error) {
	var report compilercompatibility.ConsumptionReport
	if path == "" {
		return report, errors.New("compatibility consumption report is required")
	}
	if _, err := readStrict(path, &report); err != nil {
		return report, err
	}
	return report, nil
}

func decodeGenerateOutput(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var summary generateSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return "", err
	}
	return summary.SemanticHash, nil
}
