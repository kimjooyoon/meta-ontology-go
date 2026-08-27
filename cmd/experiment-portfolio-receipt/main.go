package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/experimentportfolio"
)

type receiptOptions struct {
	candidate string
	subject   string
	source    string
	output    string
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	var options receiptOptions
	flags := flag.NewFlagSet("experiment-portfolio-receipt", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&options.candidate, "candidate", "", "portfolio candidate id")
	flags.StringVar(&options.subject, "subject-sha", "", "exact experiment subject SHA")
	flags.StringVar(&options.source, "source", "", "candidate Gooo source")
	flags.StringVar(&options.output, "output", "", "receipt output")
	if err := flags.Parse(args); err != nil || options.candidate == "" || options.subject == "" || options.source == "" || options.output == "" {
		fmt.Fprintln(stderr, "usage: experiment-portfolio-receipt -candidate <id> -subject-sha <sha> -source <file.gooo> -output <json>")
		return 2
	}
	source, err := os.ReadFile(options.source)
	if err != nil {
		fmt.Fprintf(stderr, "experiment-portfolio-receipt: source: %v\n", err)
		return 2
	}
	receipt, err := experimentportfolio.ProduceReceipt(options.subject, options.source, source, options.candidate)
	if err != nil {
		fmt.Fprintf(stderr, "experiment-portfolio-receipt: %v\n", err)
		return 1
	}
	if err := experimentportfolio.WriteReceipt(options.output, receipt); err != nil {
		fmt.Fprintf(stderr, "experiment-portfolio-receipt: output: %v\n", err)
		return 2
	}
	fmt.Fprintf(stdout, "portfolio receipt: %s %s\n", receipt.CandidateID, receipt.Digest)
	return 0
}
