package publicpartialreuse

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

const (
	ReceiptSchema    = "gooo/public-partial-test-reuse-receipt/v1"
	ReceiptOperation = Operation + ".baseline"
	ReceiptReason    = "SUCCESSFUL_PARTITION_TEST_EXECUTION"
	ReplayOperation  = Operation + ".selective"
	ReuseReason      = "AUTHORIZED_UNAFFECTED_PARTITION_RECEIPT_REUSE"
	UnknownStage     = "PARTIAL_TEST_REUSE"
	UnknownClass     = "INCOMPLETE_OR_AMBIGUOUS_EVIDENCE"
	UnknownNext      = "REPAIR_CANONICAL_GRAPH_OR_RECEIPT"
	RefutedReason    = "FAIL_CLOSED_PARTIAL_REUSE_CONTRADICTION"
)

var compilerManifestPaths = []string{
	"cmd/gooo/generate_part01.go",
	"cmd/gooo/generate_pipeline_part03.go",
	"internal/bidir/lower_document_part01.go",
	"internal/semantic/graph_part03.go",
	"internal/meta/publicpartialreuse/digest.go",
	"internal/meta/publicpartialreuse/evaluator.go",
	"internal/meta/publicpartialreuse/policy.go",
	"internal/meta/publicpartialreuse/receipt.go",
	"scripts/self-improvement-public-partial-reuse/main.go",
	"scripts/self-improvement-public-partial-reuse/model.go",
	"scripts/self-improvement-public-partial-reuse/run.go",
	"scripts/self-improvement-public-partial-reuse/verify.go",
}

type Binding struct {
	PolicySourceDigest              string `json:"policy_source_digest"`
	PolicySemanticDigest            string `json:"policy_semantic_digest"`
	PolicyEvaluatorDigest           string `json:"policy_evaluator_digest"`
	CanonicalSourceDigest           string `json:"canonical_source_digest"`
	CanonicalSemanticSubgraphDigest string `json:"canonical_semantic_subgraph_digest"`
	GeneratedArtifactDigest         string `json:"generated_artifact_digest"`
	GeneratedSemanticDigest         string `json:"generated_semantic_digest"`
	GeneratedManifestDigest         string `json:"generated_manifest_digest"`
	CompilerDigest                  string `json:"compiler_digest"`
	ReleasedToolDigest              string `json:"released_tool_digest"`
	ToolchainDigest                 string `json:"toolchain_digest"`
	ToolchainVersion                string `json:"toolchain_version"`
	TestCommand                     string `json:"test_command"`
	TestCommandDigest               string `json:"test_command_digest"`
	TestContractDigest              string `json:"test_contract_digest"`
	DependencyGraphDigest           string `json:"dependency_graph_digest"`
	OrchestrationReportDigest       string `json:"orchestration_report_digest"`
	OrchestrationOperation          string `json:"orchestration_operation"`
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

type Provenance struct {
	Operation string `json:"operation"`
	CaseID    string `json:"case_id"`
	Stage     string `json:"stage"`
}

type Receipt struct {
	Schema              string            `json:"schema"`
	ReceiptID           string            `json:"receipt_id"`
	Partition           string            `json:"partition"`
	Decision            string            `json:"decision"`
	Reason              string            `json:"reason"`
	Binding             Binding           `json:"binding"`
	Original            OriginalExecution `json:"original"`
	Provenance          Provenance        `json:"provenance"`
	RepositoryWrites    int               `json:"repository_writes"`
	LocalTestExecutions int               `json:"local_test_executions"`
}

func CompilerDigest(root string) (string, error) {
	var builder strings.Builder
	for _, relative := range compilerManifestPaths {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			return "", fmt.Errorf("read compiler identity %s: %w", relative, err)
		}
		builder.WriteString(relative)
		builder.WriteByte(0)
		builder.Write(data)
		builder.WriteByte(0)
	}
	return cache.HashBytes([]byte(builder.String())).String(), nil
}

func ReleasedToolDigest(compiler string) string {
	return cache.HashBytes([]byte("gooo-public-partial-reuse-released-tool/v1\x00" + compiler)).String()
}

func ToolchainDigest() string { return generation.SemanticRetentionToolchainDigest() }

func TestCommand(partition Partition) string {
	return "go test -tags partial_reuse_example -run ^" + partition.TestName + "$ -count=1 ."
}

func TestCommandDigest(command, contract string) string {
	return cache.HashBytes([]byte(command + "\x00" + contract)).String()
}

func BuildBinding(policy Policy, partition Partition, program, manifest, testContract []byte, compiler, orchestrationDigest, orchestrationOperation string) (Binding, error) {
	sourceDigest, err := partitionSourceDigest(policy, partition)
	if err != nil {
		return Binding{}, err
	}
	semanticDigest, err := partitionSemanticDigest(policy, partition)
	if err != nil {
		return Binding{}, err
	}
	generatedDigest, err := generatedPartitionDigest(program, partition)
	if err != nil {
		return Binding{}, err
	}
	wholeManifestDigest, err := manifestDigest(manifest)
	if err != nil {
		return Binding{}, err
	}
	contractDigest := cache.HashBytes(testContract).String()
	command := TestCommand(partition)
	return Binding{
		PolicySourceDigest:    cache.HashBytes([]byte(policy.SkeletonDigest())).String(),
		PolicySemanticDigest:  cache.HashBytes([]byte(policy.SkeletonDigest() + "\x00semantic")).String(),
		PolicyEvaluatorDigest: policy.EvaluatorDigest,
		CanonicalSourceDigest: sourceDigest, CanonicalSemanticSubgraphDigest: semanticDigest,
		GeneratedArtifactDigest: generatedDigest, GeneratedSemanticDigest: semanticDigest,
		GeneratedManifestDigest: hashJSON(struct{ Whole, Partition string }{wholeManifestDigest, semanticDigest}),
		CompilerDigest:          compiler, ReleasedToolDigest: ReleasedToolDigest(compiler), ToolchainDigest: ToolchainDigest(), ToolchainVersion: runtime.Version(),
		TestCommand: command, TestCommandDigest: TestCommandDigest(command, contractDigest), TestContractDigest: contractDigest,
		DependencyGraphDigest: policy.DependencyGraphDigest(), OrchestrationReportDigest: orchestrationDigest, OrchestrationOperation: orchestrationOperation,
	}, nil
}

func ReceiptContentDigest(receipt Receipt) (string, error) {
	receipt.ReceiptID = ""
	data, err := json.Marshal(receipt)
	if err != nil {
		return "", err
	}
	return cache.HashBytes(data).String(), nil
}

func ValidateReceipt(receipt Receipt) error {
	if receipt.Schema != ReceiptSchema || receipt.Partition == "" || receipt.Decision != DecisionClosed || receipt.Reason != ReceiptReason ||
		receipt.Original.Operation != ReceiptOperation || !receipt.Original.Successful || receipt.Original.ExitCode != 0 ||
		receipt.Original.BuildExecutions != 1 || receipt.Original.TestExecutions != 1 || receipt.Original.BuildMS <= 0 || receipt.Original.TestMS <= 0 ||
		receipt.Original.WallMS <= 0 || receipt.Original.PeakRSSKib <= 0 || receipt.RepositoryWrites != 0 || receipt.LocalTestExecutions != 0 ||
		receipt.Provenance.Operation != Operation || receipt.Provenance.Stage != "v14-authorized" {
		return errors.New("partial reuse receipt execution identity is invalid")
	}
	for _, value := range []string{
		receipt.ReceiptID, receipt.Binding.PolicySourceDigest, receipt.Binding.PolicySemanticDigest, receipt.Binding.PolicyEvaluatorDigest,
		receipt.Binding.CanonicalSourceDigest, receipt.Binding.CanonicalSemanticSubgraphDigest, receipt.Binding.GeneratedArtifactDigest,
		receipt.Binding.GeneratedSemanticDigest, receipt.Binding.GeneratedManifestDigest, receipt.Binding.CompilerDigest, receipt.Binding.ReleasedToolDigest,
		receipt.Binding.ToolchainDigest, receipt.Binding.TestCommandDigest, receipt.Binding.TestContractDigest, receipt.Binding.DependencyGraphDigest,
		receipt.Binding.OrchestrationReportDigest, receipt.Original.ResultDigest, receipt.Original.InvocationID,
	} {
		if !cache.Digest(value).Known() {
			return errors.New("partial reuse receipt has unknown digest evidence")
		}
	}
	if receipt.Binding.ToolchainVersion != "go1.27.0" || receipt.Binding.TestCommand == "" || receipt.Binding.OrchestrationOperation != "gooo.self-improvement.public-orchestration" ||
		receipt.Binding.TestCommandDigest != TestCommandDigest(receipt.Binding.TestCommand, receipt.Binding.TestContractDigest) {
		return errors.New("partial reuse receipt command or toolchain binding is invalid")
	}
	digest, err := ReceiptContentDigest(receipt)
	if err != nil || digest != receipt.ReceiptID {
		return errors.New("partial reuse receipt is not content-addressed")
	}
	return nil
}

func VerifyReceipt(receipt Receipt, expected Binding, partition string) error {
	if err := ValidateReceipt(receipt); err != nil {
		return err
	}
	if receipt.Partition != partition || receipt.Binding != expected {
		return errors.New("partial reuse receipt does not bind the exact partition inputs")
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

func WriteReceipt(filename string, receipt Receipt) error {
	data, err := MarshalReceipt(receipt)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o444)
	if err != nil {
		return fmt.Errorf("create immutable partial reuse receipt: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func ReadReceipt(filename string) (Receipt, error) {
	info, err := os.Lstat(filename)
	if err != nil || !info.Mode().IsRegular() {
		return Receipt{}, errors.New("partial reuse receipt is unavailable or not regular")
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return Receipt{}, err
	}
	var receipt Receipt
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&receipt); err != nil {
		return Receipt{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return Receipt{}, errors.New("partial reuse receipt contains trailing JSON")
	}
	if err := ValidateReceipt(receipt); err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}
