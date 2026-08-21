package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
	programverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/verify"
)

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("metric-meta-program", flag.ContinueOnError)
	flags.SetOutput(stderr)
	strategyPath := flags.String("strategy", "", "metric strategy plan JSON")
	verificationPath := flags.String("strategy-verification", "", "metric strategy verification JSON")
	outputDirectory := flags.String("out", "", "artifact output directory")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 || *strategyPath == "" || *verificationPath == "" || *outputDirectory == "" {
		fmt.Fprintln(stderr, "usage: metric-meta-program --strategy <plan.json> --strategy-verification <verification.json> --out <directory>")
		return 2
	}
	strategy, err := readBounded(*strategyPath)
	if err != nil {
		fmt.Fprintf(stderr, "metric-meta-program: read strategy: %v\n", err)
		return 1
	}
	strategyVerification, err := readBounded(*verificationPath)
	if err != nil {
		fmt.Fprintf(stderr, "metric-meta-program: read strategy verification: %v\n", err)
		return 1
	}
	program, source, err := metricprogram.Compile(strategy, strategyVerification)
	if err != nil {
		fmt.Fprintf(stderr, "metric-meta-program: compile: %v\n", err)
		return 1
	}
	programPayload, err := encodeJSON(program)
	if err != nil {
		fmt.Fprintf(stderr, "metric-meta-program: encode program: %v\n", err)
		return 1
	}
	report, err := programverify.Verify(strategy, strategyVerification, programPayload, source)
	if err != nil {
		fmt.Fprintf(stderr, "metric-meta-program: verify: %v\n", err)
		return 1
	}
	reportPayload, err := encodeJSON(report)
	if err != nil {
		fmt.Fprintf(stderr, "metric-meta-program: encode verification: %v\n", err)
		return 1
	}
	if err := writeArtifacts(*outputDirectory, source, programPayload, reportPayload); err != nil {
		fmt.Fprintf(stderr, "metric-meta-program: write artifacts: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "verified: %s bindings=%d operations=%d steps=%d\n", program.Digest, report.BindingCount, report.OperationCount, report.StepCount)
	return 0
}

func encodeJSON(value any) ([]byte, error) {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}
