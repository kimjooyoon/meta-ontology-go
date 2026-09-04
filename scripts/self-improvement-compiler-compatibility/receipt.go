package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/compilercompatibility"
)

type generateSummary struct {
	SemanticHash string `json:"semantic_hash"`
}

func runReceipt(contractPath, inputPath, outputRoot, role, candidateID, compilerDigest, testResult, authorizationPath, outputPath string) error {
	if contractPath == "" || inputPath == "" || outputRoot == "" || role == "" || candidateID == "" || compilerDigest == "" || testResult == "" || authorizationPath == "" || outputPath == "" {
		return errors.New("receipt requires contract, input, output-root, role, candidate-stable-id, compiler-digest, test-result, authorization, and output")
	}
	policy, err := loadPolicy(contractPath)
	if err != nil {
		return err
	}
	var authorization compilercompatibility.Authorization
	_, err = readStrict(authorizationPath, &authorization)
	if err != nil {
		return err
	}
	if err := compilercompatibility.ValidateAuthorization(authorization); err != nil {
		return fmt.Errorf("authorization: %w", err)
	}
	input, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}
	if authorization.CandidateStableID != candidateID || authorization.SubjectDigest != cache.HashBytes(input).String() || authorization.SuccessorCompilerDigest == "" {
		return errors.New("authorization does not bind the receipt subject")
	}
	var summary generateSummary
	if _, err := readStrict(filepath.Join(outputRoot, "generate.json"), &summary); err != nil {
		return err
	}
	if summary.SemanticHash == "" {
		return errors.New("generate summary has no semantic hash")
	}
	source, err := os.ReadFile(filepath.Join(outputRoot, "semantic.gooo.go"))
	if err != nil {
		return fmt.Errorf("read generated source: %w", err)
	}
	manifest, err := os.ReadFile(filepath.Join(outputRoot, "semantic.gooo.manifest.jsonl"))
	if err != nil {
		return fmt.Errorf("read generated manifest: %w", err)
	}
	if role != "predecessor" && role != "successor" {
		return fmt.Errorf("receipt role %q is invalid", role)
	}
	manifest, err = normalizeManifestGeneratedFile(manifest)
	if err != nil {
		return fmt.Errorf("normalize generated manifest: %w", err)
	}
	inputDigest := cache.HashBytes(input).String()
	receipt := compilercompatibility.ExecutionReceipt{
		Schema: compilercompatibility.ConsumptionSchema, Role: role, CandidateStableID: candidateID,
		SubjectDigest: inputDigest, SourceDigest: inputDigest, SemanticIRDigest: summary.SemanticHash,
		GeneratedOutputDigest: cache.HashBytes(source).String(), GeneratedManifestDigest: cache.HashBytes(manifest).String(),
		GeneratedSource: source, GeneratedManifest: manifest, PolicyDigest: policy.SourceDigest,
		PolicyEvaluatorDigest: policy.EvaluatorDigest, PolicyResult: compilercompatibility.DecisionClosed,
		CompilerImplementationDigest: compilerDigest, GoToolchainDigest: compilercompatibility.CurrentToolchainDigest(),
		TestContractDigest: compilercompatibility.TestContractDigest(), TestContractResult: testResult,
		AuthorizationDigest: authorization.AuthorizationID,
	}
	if authorization.SuccessorCompilerDigest != compilerDigest && role == "successor" {
		return errors.New("successor compiler digest differs from caller-owned authorization")
	}
	if err := compilercompatibility.ValidateExecutionReceipt(receipt); err != nil {
		return fmt.Errorf("execution receipt: %w", err)
	}
	return writeJSON(outputPath, receipt)
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
