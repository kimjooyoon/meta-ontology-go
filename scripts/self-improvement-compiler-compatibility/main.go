package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/compatibilitypolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/compilercompatibility"
)

func main() {
	mode := flag.String("mode", "", "identity, authorize, receipt, certify, run, or verify")
	contract := flag.String("contract", "", "canonical compatibility .gooo policy")
	input := flag.String("input", "", "canonical input .gooo")
	root := flag.String("output-root", "", "generate output root")
	role := flag.String("role", "", "predecessor or successor")
	candidate := flag.String("candidate-stable-id", "", "canonical candidate stable ID")
	compilerDigest := flag.String("compiler-digest", "", "compiler implementation digest")
	testResult := flag.String("test-result", "", "independent test contract result")
	authorizationPath := flag.String("authorization", "", "caller-owned authorization JSON")
	predecessorPath := flag.String("predecessor", "", "predecessor execution receipt")
	successorPath := flag.String("successor", "", "successor execution receipt")
	certificatePath := flag.String("certificate", "", "compatibility certificate")
	consumptionPath := flag.String("consumption-report", "", "actual gooo generate compatibility report")
	publicationRoot := flag.String("publication-root", "", "caller-owned publication root")
	output := flag.String("output", "", "caller-owned output JSON")
	flag.Parse()

	var err error
	switch *mode {
	case "identity":
		err = runIdentity(*contract, *output)
	case "authorize":
		err = runAuthorize(*candidate, *input, *compilerDigest, *output)
	case "receipt":
		err = runReceipt(*contract, *input, *root, *role, *candidate, *compilerDigest, *testResult, *authorizationPath, *output)
	case "certify":
		err = runCertify(*contract, *predecessorPath, *successorPath, *authorizationPath, *output)
	case "run":
		err = runConformance(*contract, *predecessorPath, *successorPath, *authorizationPath, *certificatePath, *consumptionPath, *publicationRoot, *output)
	case "verify":
		err = verifyConformance(*contract, *certificatePath, *consumptionPath, *publicationRoot, *output)
	default:
		err = errors.New("self-improvement-compiler-compatibility requires a supported mode")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type identityReport struct {
	Schema                 string `json:"schema"`
	CompilerImplementation string `json:"compiler_implementation_digest"`
	GoToolchain            string `json:"go_toolchain_digest"`
	Policy                 string `json:"policy_digest"`
	PolicyEvaluator        string `json:"policy_evaluator_digest"`
	TestContract           string `json:"test_contract_digest"`
}

func runIdentity(contractPath, outputPath string) error {
	policy, err := loadPolicy(contractPath)
	if err != nil {
		return err
	}
	compilerDigest, err := compilercompatibility.CompilerImplementationDigest(os.ReadFile)
	if err != nil {
		return fmt.Errorf("compiler implementation identity: %w", err)
	}
	return writeJSON(outputPath, identityReport{Schema: "gooo/compiler-successor-identity/v1", CompilerImplementation: compilerDigest,
		GoToolchain: compilercompatibility.CurrentToolchainDigest(), Policy: policy.SourceDigest, PolicyEvaluator: policy.EvaluatorDigest,
		TestContract: compilercompatibility.TestContractDigest()})
}

func runAuthorize(candidateID, inputPath, successorDigest, outputPath string) error {
	if candidateID == "" || inputPath == "" || successorDigest == "" {
		return errors.New("authorize requires candidate-stable-id, input, compiler-digest, and output")
	}
	input, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}
	authorization, err := compilercompatibility.BuildAuthorization(candidateID, cache.HashBytes(input).String(), successorDigest)
	if err != nil {
		return err
	}
	return writeJSON(outputPath, authorization)
}

func loadPolicy(path string) (compatibilitypolicy.Policy, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return compatibilitypolicy.Policy{}, fmt.Errorf("read compatibility policy: %w", err)
	}
	policy, err := compatibilitypolicy.Load(path, source)
	if err != nil {
		return compatibilitypolicy.Policy{}, err
	}
	return policy, nil
}
