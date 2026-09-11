package publictestreuse

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

const (
	ReceiptSchema            = "gooo/public-test-reuse-receipt/v1"
	ReportSchema             = "gooo/public-test-reuse-report/v1"
	OriginalOperation        = Operation + ".baseline"
	ReplayOperation          = Operation + ".replay"
	ReceiptReason            = ReasonBaseline
	ReuseReason              = ReasonReuse
	UnknownStage             = "TEST_REUSE"
	UnknownClass             = "INCOMPLETE_EVIDENCE"
	UnknownAuthorization     = "AUTHORIZE_REUSE"
	UnknownEvidence          = "LOAD_REUSE_RECEIPT"
	UnknownNextAuthorization = "PROVIDE_EXPLICIT_REUSE_AUTHORIZATION"
	UnknownNextEvidence      = "PUBLISH_ONE_EXACT_SUCCESSFUL_TEST_RECEIPT"
)

var compilerManifestPaths = []string{
	"cmd/gooo/generate_part01.go",
	"cmd/gooo/generate_pipeline_part03.go",
	"internal/generator/normalize_part01.go",
	"internal/generator/types_part01.go",
	"internal/meta/generation/semantic_retention.go",
	"internal/meta/publictestreuse/policy.go",
	"internal/meta/publictestreuse/receipt.go",
	"scripts/self-improvement-public-test-reuse/main.go",
	"scripts/self-improvement-public-test-reuse/model.go",
	"scripts/self-improvement-public-test-reuse/run.go",
	"scripts/self-improvement-public-test-reuse/verify.go",
}

type Binding struct {
	PolicySourceDigest      string `json:"policy_source_digest"`
	PolicySemanticDigest    string `json:"policy_semantic_digest"`
	PolicyEvaluatorDigest   string `json:"policy_evaluator_digest"`
	CanonicalSourceDigest   string `json:"canonical_source_digest"`
	CanonicalSemanticDigest string `json:"canonical_semantic_digest"`
	GeneratedOutputDigest   string `json:"generated_output_digest"`
	GeneratedSemanticDigest string `json:"generated_semantic_digest"`
	GeneratedManifestDigest string `json:"generated_manifest_digest"`
	CompilerDigest          string `json:"compiler_digest"`
	ReleasedToolDigest      string `json:"released_tool_digest"`
	ToolchainDigest         string `json:"toolchain_digest"`
	ToolchainVersion        string `json:"toolchain_version"`
	TestCommand             string `json:"test_command"`
	TestCommandDigest       string `json:"test_command_digest"`
	TestContractDigest      string `json:"test_contract_digest"`
}

type OriginalExecution struct {
	InvocationID    string `json:"invocation_id"`
	Operation       string `json:"operation"`
	ResultDigest    string `json:"result_digest"`
	Successful      bool   `json:"successful"`
	ExitCode        int    `json:"exit_code"`
	BuildExecutions int    `json:"build_executions"`
	TestExecutions  int    `json:"test_executions"`
	BuildMS         int64  `json:"build_ms"`
	TestMS          int64  `json:"test_ms"`
	WallMS          int64  `json:"wall_ms"`
	PeakRSSKib      int64  `json:"peak_rss_kib"`
}

type Receipt struct {
	Schema              string            `json:"schema"`
	ReceiptID           string            `json:"receipt_id"`
	Operation           string            `json:"operation"`
	Decision            string            `json:"decision"`
	Reason              string            `json:"reason"`
	Binding             Binding           `json:"binding"`
	Original            OriginalExecution `json:"original"`
	RepositoryWrites    int               `json:"repository_writes"`
	LocalTestExecutions int               `json:"local_test_executions"`
}

type UnknownState struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type Report struct {
	Schema                string        `json:"schema"`
	Operation             string        `json:"operation"`
	Decision              string        `json:"decision"`
	Reason                string        `json:"reason"`
	Unknown               *UnknownState `json:"unknown"`
	CaseID                string        `json:"case_id"`
	Binding               Binding       `json:"binding"`
	ReceiptDigest         string        `json:"receipt_digest"`
	OriginalReceiptID     string        `json:"original_receipt_id"`
	OriginalInvocationID  string        `json:"original_invocation_id"`
	BuildExecutions       int           `json:"build_executions"`
	TestExecutions        int           `json:"test_executions"`
	ReusedTestExecutions  int           `json:"reused_test_executions"`
	ReceiptHits           int           `json:"receipt_hits"`
	ReceiptMisses         int           `json:"receipt_misses"`
	BuildMS               int64         `json:"build_ms"`
	TestMS                int64         `json:"test_ms"`
	WallMS                int64         `json:"wall_ms"`
	PeakRSSKib            int64         `json:"peak_rss_kib"`
	GeneratedProgramBytes int64         `json:"generated_program_bytes"`
	GeneratedProgramLines int           `json:"generated_program_lines"`
	TestContractBytes     int64         `json:"test_contract_bytes"`
	TestContractLines     int           `json:"test_contract_lines"`
	GeneratedBytesEqual   bool          `json:"generated_bytes_equal"`
	SemanticEqual         bool          `json:"semantic_equal"`
	RepositoryWrites      int           `json:"repository_writes"`
	LocalTestExecutions   int           `json:"local_test_executions"`
}

func CompilerDigest(readFile func(string) ([]byte, error)) (string, error) {
	return generation.SemanticRetentionManifestDigest(readFile, compilerManifestPaths)
}

func ReleasedToolDigest(compilerDigest string) string {
	return cache.HashBytes([]byte("gooo-public-test-reuse-released-tool/v1\x00" + compilerDigest)).String()
}

func ToolchainDigest() string { return generation.SemanticRetentionToolchainDigest() }

func CurrentToolchainVersion() string { return runtime.Version() }

func TestCommandDigest(command, contractDigest string) string {
	return cache.HashBytes([]byte(command + "\x00" + contractDigest)).String()
}

func ReceiptContentDigest(receipt Receipt) (string, error) {
	receipt.ReceiptID = ""
	digest, err := cache.DigestOf(receipt)
	if err != nil {
		return "", fmt.Errorf("test reuse receipt content digest: %w", err)
	}
	return digest.String(), nil
}

func ValidateReceipt(receipt Receipt) error {
	if receipt.Schema != ReceiptSchema || receipt.Operation != OriginalOperation || receipt.Decision != DecisionClosed ||
		receipt.Reason != ReceiptReason || !cache.Digest(receipt.ReceiptID).Known() || receipt.RepositoryWrites != 0 || receipt.LocalTestExecutions != 0 {
		return errors.New("public test reuse receipt identity is invalid")
	}
	for _, value := range []string{
		receipt.Binding.PolicySourceDigest, receipt.Binding.PolicySemanticDigest, receipt.Binding.PolicyEvaluatorDigest,
		receipt.Binding.CanonicalSourceDigest, receipt.Binding.CanonicalSemanticDigest, receipt.Binding.GeneratedOutputDigest,
		receipt.Binding.GeneratedSemanticDigest, receipt.Binding.GeneratedManifestDigest, receipt.Binding.CompilerDigest,
		receipt.Binding.ReleasedToolDigest, receipt.Binding.ToolchainDigest, receipt.Binding.TestCommandDigest,
		receipt.Binding.TestContractDigest, receipt.Original.ResultDigest, receipt.Original.InvocationID,
	} {
		if !cache.Digest(value).Known() {
			return errors.New("public test reuse receipt contains unknown digest evidence")
		}
	}
	if receipt.Binding.ToolchainVersion == "" || receipt.Binding.TestCommand == "" ||
		receipt.Binding.TestCommandDigest != TestCommandDigest(receipt.Binding.TestCommand, receipt.Binding.TestContractDigest) {
		return errors.New("public test reuse receipt command binding is invalid")
	}
	original := receipt.Original
	if original.Operation != OriginalOperation || !original.Successful || original.ExitCode != 0 ||
		original.BuildExecutions != 1 || original.TestExecutions != 1 || original.BuildMS <= 0 || original.TestMS <= 0 ||
		original.WallMS <= 0 || original.PeakRSSKib <= 0 {
		return errors.New("public test reuse receipt does not contain one successful observed test")
	}
	contentDigest, err := ReceiptContentDigest(receipt)
	if err != nil || contentDigest != receipt.ReceiptID {
		return errors.New("public test reuse receipt is not content-addressed")
	}
	return nil
}

func VerifyReceipt(receipt Receipt, expected Binding) error {
	if err := ValidateReceipt(receipt); err != nil {
		return err
	}
	if receipt.Binding != expected {
		return errors.New("public test reuse receipt does not bind the exact current inputs")
	}
	return nil
}

func MarshalReceipt(receipt Receipt) ([]byte, error) {
	if err := ValidateReceipt(receipt); err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func DecodeReceipt(data []byte) (Receipt, error) {
	var receipt Receipt
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&receipt); err != nil {
		return Receipt{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return Receipt{}, errors.New("public test reuse receipt contains multiple JSON values")
	} else if err != io.EOF {
		return Receipt{}, fmt.Errorf("decode public test reuse receipt trailer: %w", err)
	}
	if err := ValidateReceipt(receipt); err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}

func WriteReceipt(filename string, receipt Receipt) error {
	data, err := MarshalReceipt(receipt)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o444)
	if err != nil {
		return fmt.Errorf("create immutable test reuse receipt: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("write immutable test reuse receipt: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync immutable test reuse receipt: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close immutable test reuse receipt: %w", err)
	}
	return nil
}

func ReadReceipt(filename string) (Receipt, error) {
	info, err := os.Lstat(filename)
	if err != nil {
		return Receipt{}, err
	}
	if !info.Mode().IsRegular() {
		return Receipt{}, errors.New("public test reuse receipt is not a regular file")
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return Receipt{}, err
	}
	return DecodeReceipt(data)
}

func EqualBinding(left, right Binding) bool { return left == right }

func Unknown(reason, step, next string, blockedBy ...string) *UnknownState {
	return &UnknownState{Stage: UnknownStage, Step: step, Reason: reason, UnknownClass: UnknownClass, NextOperation: next, BlockedBy: append([]string(nil), blockedBy...)}
}

func SameUnknown(left, right *UnknownState) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.Stage == right.Stage && left.Step == right.Step && left.Reason == right.Reason && left.UnknownClass == right.UnknownClass && left.NextOperation == right.NextOperation && strings.Join(left.BlockedBy, "\x00") == strings.Join(right.BlockedBy, "\x00")
}
