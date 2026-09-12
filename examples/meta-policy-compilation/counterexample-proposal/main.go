package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/policycompilation"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, output, diagnostics io.Writer) error {
	flags := flag.NewFlagSet("counterexample-proposal", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	policyPath := flags.String("policy", "", "original Gooo policy source")
	inputPath := flags.String("counterexample", "", "source-bound counterexample JSON")
	profilePackage := flags.String("profile-package", "metapolicycompilation", "expected source package")
	profileNamespace := flags.String("profile-namespace", "metapolicycompilation", "expected source namespace")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *policyPath == "" || *inputPath == "" {
		return errors.New("provide -policy and -counterexample without positional arguments")
	}
	raw, err := os.ReadFile(*inputPath)
	if err != nil {
		return fmt.Errorf("read counterexample: %w", err)
	}
	input, err := policycompilation.DecodePolicyRevisionCounterexample(raw)
	if err != nil {
		return fmt.Errorf("decode counterexample: %w", err)
	}
	source, err := os.ReadFile(*policyPath)
	if err != nil {
		return fmt.Errorf("read policy: %w", err)
	}
	report, proposalError := policycompilation.ProposePolicyRevisionFromCounterexample(
		*policyPath, source, *profilePackage, *profileNamespace, input,
	)
	if report.Schema == "" {
		return proposalError
	}
	report.CounterexampleArtifactDigest = policycompilation.DigestBytes(raw)
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return errors.Join(proposalError, encoder.Encode(report))
}
