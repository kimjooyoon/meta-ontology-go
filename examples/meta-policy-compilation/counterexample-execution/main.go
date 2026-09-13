package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/policycompilation"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, output, diagnostics io.Writer) error {
	flags := flag.NewFlagSet("counterexample-execution", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	policyPath := flags.String("policy", "", "original Gooo policy source")
	inputPath := flags.String("input", "", "counterexample and immutable case pairs")
	operationPath := flags.String("operation", "", "Gooo revision operation contract")
	profilePackage := flags.String("profile-package", "metapolicycompilation", "expected source package")
	profileNamespace := flags.String("profile-namespace", "metapolicycompilation", "expected source namespace")
	timeout := flags.Duration("timeout", 2*time.Minute, "native observation time budget")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *policyPath == "" || *inputPath == "" || *operationPath == "" || *timeout <= 0 {
		return errors.New("provide -policy, -input and -operation with a positive timeout and no positional arguments")
	}
	source, err := os.ReadFile(*policyPath)
	if err != nil {
		return err
	}
	input, err := os.ReadFile(*inputPath)
	if err != nil {
		return err
	}
	operation, err := os.ReadFile(*operationPath)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	report, observationError := policycompilation.ObserveGoooPolicyCounterexample(ctx, operation,
		*policyPath, source, *profilePackage, *profileNamespace, input)
	if report.Schema == "" {
		return observationError
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return errors.Join(observationError, encoder.Encode(report))
}
