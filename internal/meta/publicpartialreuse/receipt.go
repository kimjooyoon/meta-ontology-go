package publicpartialreuse

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
	"github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"
)

const (
	ReceiptSchema    = "gooo/public-partial-test-reuse-receipt/v3"
	ReceiptOperation = Operation + ".baseline"
	ReceiptReason    = "SUCCESSFUL_PARTITION_TEST_EXECUTION"
	ReplayOperation  = Operation + ".selective"
	ReuseReason      = "AUTHORIZED_UNAFFECTED_PARTITION_RECEIPT_REUSE"
	UnknownStage     = "PARTIAL_TEST_REUSE"
	UnknownClass     = "INCOMPLETE_OR_AMBIGUOUS_EVIDENCE"
	UnknownNext      = "REPAIR_CANONICAL_GRAPH_OR_RECEIPT"
	RefutedReason    = "FAIL_CLOSED_PARTIAL_REUSE_CONTRADICTION"
)

// ErrUnobservedToolInputs keeps older records from authorizing current reuse.
var ErrUnobservedToolInputs = errors.New("historical partition receipt has no selected Go observations")

// ErrInvalidReceipt identifies invalid content, not unavailable input.
var ErrUnobservedPartitionInputs = errors.New("historical partition receipt has no independent package input scope")

var ErrInvalidReceipt = errors.New("partial reuse receipt content is invalid")

var compilerManifestPaths = []string{
	"cmd/gooo/generate_part01.go",
	"cmd/gooo/generate_pipeline_part03.go",
	"internal/bidir/lower_document_part01.go",
	"internal/semantic/graph_part03.go",
	"internal/meta/publicpartialreuse/assessment.go",
	"internal/meta/publicpartialreuse/observation.go",
	"internal/meta/publicpartialreuse/policy.gooo",
	"internal/meta/publicpartialreuse/selection.go",
	"internal/meta/publicpartialreuse/digest.go",
	"internal/meta/publicpartialreuse/evaluator.go",
	"internal/meta/publicpartialreuse/policy.go",
	"internal/meta/publicpartialreuse/receipt.go",
	"internal/meta/publicpartialreuse/package_origin.go",
	"internal/toolchaincli/model.go",
	"internal/toolchaincli/go_inputs.go",
	"internal/toolchaincli/package_inputs.go",
	"internal/toolchaincli/go_command.go",
	"internal/toolchaincli/resource_darwin.go",
	"internal/toolchaincli/resource_linux.go",
	"internal/toolchaincli/resource_other.go",
	"scripts/self-improvement-public-partial-reuse/main.go",
	"scripts/self-improvement-public-partial-reuse/model.go",
	"scripts/self-improvement-public-partial-reuse/origins.go",
	"scripts/self-improvement-public-partial-reuse/run.go",
	"scripts/self-improvement-public-partial-reuse/verify.go",
}

type Binding struct {
	PartitionProfile                string                             `json:"partition_input_profile"`
	SharedProgramDigest             string                             `json:"shared_program_digest"`
	ModuleDigest                    string                             `json:"module_digest"`
	ToolInputs                      toolchaincli.GeneratedTestGoInputs `json:"execution_tool_inputs"`
	PolicySourceDigest              string                             `json:"policy_source_digest"`
	PolicySemanticDigest            string                             `json:"policy_semantic_digest"`
	PolicyEvaluatorDigest           string                             `json:"policy_evaluator_digest"`
	CanonicalSourceDigest           string                             `json:"canonical_source_digest"`
	CanonicalSemanticSubgraphDigest string                             `json:"canonical_semantic_subgraph_digest"`
	GeneratedArtifactDigest         string                             `json:"generated_artifact_digest"`
	GeneratedSemanticDigest         string                             `json:"generated_semantic_digest"`
	GeneratedManifestDigest         string                             `json:"generated_manifest_digest"`
	CompilerDigest                  string                             `json:"compiler_digest"`
	ReleasedToolDigest              string                             `json:"released_tool_digest"`
	ToolchainDigest                 string                             `json:"toolchain_digest"`
	ToolchainVersion                string                             `json:"toolchain_version"`
	TestCommand                     string                             `json:"test_command"`
	TestCommandDigest               string                             `json:"test_command_digest"`
	TestContractDigest              string                             `json:"test_contract_digest"`
	DependencyGraphDigest           string                             `json:"dependency_graph_digest"`
	OrchestrationReportDigest       string                             `json:"orchestration_report_digest"`
	OrchestrationOperation          string                             `json:"orchestration_operation"`
}

type OriginalExecution struct {
	SemanticDigest  string `json:"generation_semantic_digest"`
	SourceDigest    string `json:"generation_source_digest"`
	ProgramDigest   string `json:"generation_program_digest"`
	ManifestDigest  string `json:"generation_manifest_digest"`
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

func TestCommand(partition Partition) string {
	return "go " + strings.Join(TestCommandArgs(partition), " ")
}

func TestCommandArgs(partition Partition) []string {
	return []string{"test", "-json", "-tags", "partial_reuse_example", "-run", "^" + regexp.QuoteMeta(partition.TestName) + "$", "-count=1", "."}
}

// ObserveTestOutput counts executed tests, never a requested or declared count.
func ObserveTestOutput(output []byte, expected []string) ([]string, error) {
	states := make(map[string]int, len(expected))
	for _, name := range expected {
		if name == "" || strings.Contains(name, "/") {
			return nil, errors.New("partial reuse expected test identity is invalid")
		}
		if _, exists := states[name]; exists {
			return nil, errors.New("partial reuse expected test identity is duplicated")
		}
		states[name] = 0
	}
	if len(states) == 0 {
		return nil, errors.New("partial reuse expected test cohort is empty")
	}
	decoder := json.NewDecoder(bytes.NewReader(output))
	var passed []string
	var packageName string
	started, completed := false, false
	for {
		var event struct{ Action, Package, Test string }
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("partial reuse test JSON: %w", err)
		}
		if event.Package == "" || (packageName != "" && packageName != event.Package) {
			return nil, errors.New("partial reuse test package identity is missing or ambiguous")
		}
		packageName = event.Package
		if event.Action == "fail" || event.Action == "skip" || (completed && event.Action != "output") {
			return nil, errors.New("partial reuse test execution is not a complete success")
		}
		if event.Test == "" {
			switch event.Action {
			case "start":
				if started {
					return nil, errors.New("partial reuse test package started twice")
				}
				started = true
			case "pass":
				if !started {
					return nil, errors.New("partial reuse test package did not start")
				}
				completed = true
			case "output":
			default:
				return nil, errors.New("partial reuse test package action is unknown")
			}
			continue
		}
		root, _, subtest := strings.Cut(event.Test, "/")
		state, exists := states[root]
		if !exists || !started || completed {
			return nil, errors.New("partial reuse observed an unexpected test or event order")
		}
		switch event.Action {
		case "run":
			if !subtest {
				if state != 0 {
					return nil, errors.New("partial reuse test ran more than once")
				}
				states[root] = 1
			}
		case "pass":
			if !subtest {
				if state != 1 {
					return nil, errors.New("partial reuse test passed without one run")
				}
				states[root] = 2
				passed = append(passed, root)
			}
		case "output", "pause", "cont":
		default:
			return nil, errors.New("partial reuse test action is unknown")
		}
	}
	if !completed || len(passed) != len(expected) {
		return nil, errors.New("partial reuse successful test cohort is incomplete")
	}
	return passed, nil
}

func TestCommandDigest(command, contract string) string {
	return cache.HashBytes([]byte(command + "\x00" + contract)).String()
}

func BuildBinding(policy Policy, partition Partition, program, manifest, testContract, module []byte, compiler, orchestrationDigest, orchestrationOperation string, tool toolchaincli.GeneratedTestGoInputs) (Binding, error) {
	if err := tool.Validate(); err != nil {
		return Binding{}, err
	}
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
	if len(module) == 0 {
		return Binding{}, errors.New("execution module input is missing")
	}
	if err := validateGenerationOrigin(policy, program, manifest); err != nil {
		return Binding{}, err
	}
	sharedDigest, err := sharedProgramDigest(policy, program)
	if err != nil {
		return Binding{}, err
	}
	contractDigest := cache.HashBytes(testContract).String()
	command := TestCommand(partition)
	return Binding{
		PartitionProfile: PartitionInputProfile, SharedProgramDigest: sharedDigest, ModuleDigest: cache.HashBytes(module).String(),
		PolicySourceDigest:    cache.HashBytes([]byte(policy.SkeletonDigest())).String(),
		PolicySemanticDigest:  cache.HashBytes([]byte(policy.SkeletonDigest() + "\x00semantic")).String(),
		PolicyEvaluatorDigest: policy.EvaluatorDigest,
		CanonicalSourceDigest: sourceDigest, CanonicalSemanticSubgraphDigest: semanticDigest,
		GeneratedArtifactDigest: generatedDigest, GeneratedSemanticDigest: semanticDigest,
		GeneratedManifestDigest: hashJSON(struct{ Profile, Source, Semantic, Generated, Shared string }{PartitionInputProfile, sourceDigest, semanticDigest, generatedDigest, sharedDigest}),
		CompilerDigest:          compiler, ReleasedToolDigest: ReleasedToolDigest(compiler), ToolchainDigest: tool.Digest(), ToolchainVersion: tool.Version, ToolInputs: tool,
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
	if receipt.Schema == "gooo/public-partial-test-reuse-receipt/v1" {
		return ErrUnobservedToolInputs
	}
	if receipt.Schema == "gooo/public-partial-test-reuse-receipt/v2" {
		return ErrUnobservedPartitionInputs
	}
	if receipt.Binding.PartitionProfile != PartitionInputProfile {
		return errors.New("partition input profile is missing")
	}
	if receipt.Binding.ToolInputs.Validate() != nil ||
		receipt.Binding.ToolchainDigest != receipt.Binding.ToolInputs.Digest() ||
		receipt.Binding.ToolchainVersion != receipt.Binding.ToolInputs.Version {
		return errors.New("partial reuse selected command binding is invalid")
	}
	if receipt.Schema != ReceiptSchema || receipt.Partition == "" || receipt.Decision != DecisionClosed || receipt.Reason != ReceiptReason ||
		receipt.Original.Operation != ReceiptOperation || !receipt.Original.Successful || receipt.Original.ExitCode != 0 ||
		receipt.Original.BuildExecutions != 1 || receipt.Original.TestExecutions != 1 || receipt.Original.BuildMS <= 0 || receipt.Original.TestMS <= 0 ||
		receipt.Original.WallMS <= 0 || receipt.Original.PeakRSSKib <= 0 || receipt.RepositoryWrites != 0 || receipt.LocalTestExecutions != 0 ||
		receipt.Provenance.Operation != Operation || receipt.Provenance.Stage != "v14-authorized" {
		return errors.New("partial reuse receipt execution identity is invalid")
	}
	for _, value := range []string{
		receipt.Binding.SharedProgramDigest, receipt.Binding.ModuleDigest, receipt.Original.SourceDigest, receipt.Original.SemanticDigest, receipt.Original.ProgramDigest, receipt.Original.ManifestDigest,
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
	if receipt.Binding.ToolchainVersion != "go1.27.3" || receipt.Binding.TestCommand == "" || receipt.Binding.OrchestrationOperation != "gooo.self-improvement.public-orchestration" ||
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
		if errors.Is(err, ErrUnobservedToolInputs) || errors.Is(err, ErrUnobservedPartitionInputs) {
			return err
		}
		return fmt.Errorf("%w: %w", ErrInvalidReceipt, err)
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
	file, err := os.Open(filename)
	if err != nil {
		return Receipt{}, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, (1<<20)+1))
	if err != nil {
		return Receipt{}, err
	}
	return DecodeReceipt(data)
}

func DecodeReceipt(data []byte) (Receipt, error) {
	if err := completeness.CheckJSON(data, 1<<20); err != nil {
		return Receipt{}, fmt.Errorf("%w: %w", ErrInvalidReceipt, err)
	}
	var receipt Receipt
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&receipt); err != nil {
		return Receipt{}, fmt.Errorf("%w: %w", ErrInvalidReceipt, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return Receipt{}, fmt.Errorf("%w: trailing JSON", ErrInvalidReceipt)
	}
	if err := ValidateReceipt(receipt); err != nil {
		if errors.Is(err, ErrUnobservedToolInputs) || errors.Is(err, ErrUnobservedPartitionInputs) {
			return Receipt{}, err
		}
		return Receipt{}, fmt.Errorf("%w: %w", ErrInvalidReceipt, err)
	}
	return receipt, nil
}
