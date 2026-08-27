package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/experimentportfolio"
)

type sourceOptions struct {
	source string
	output string
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	var options sourceOptions
	flags := flag.NewFlagSet("experiment-portfolio-causal-source", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&options.source, "source", "", "candidate Gooo source")
	flags.StringVar(&options.output, "output", "", "source observation output")
	if err := flags.Parse(args); err != nil || options.source == "" || options.output == "" {
		fmt.Fprintln(stderr, "usage: experiment-portfolio-causal-source -source <file.gooo> -output <json>")
		return 2
	}
	source, err := os.ReadFile(options.source)
	if err != nil {
		fmt.Fprintf(stderr, "experiment-portfolio-causal-source: source: %v\n", err)
		return 2
	}
	observation, err := experimentportfolio.ObserveSource(options.source, source)
	if err != nil {
		fmt.Fprintf(stderr, "experiment-portfolio-causal-source: %v\n", err)
		return 1
	}
	if err := experimentportfolio.WriteSourceObservation(options.output, observation); err != nil {
		fmt.Fprintf(stderr, "experiment-portfolio-causal-source: output: %v\n", err)
		return 2
	}
	fmt.Fprintf(stdout, "source observation: %s %s\n", observation.SemanticValue, observation.SourceDigest)
	return 0
}
