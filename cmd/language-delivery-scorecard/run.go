package main

import (
	"flag"
	"fmt"
	"io"
)

type options struct {
	head, contract, manifest, user, conformance, lsp, release, execution, profile, readiness, out string
}

func run(args []string, stdout, stderr io.Writer) int {
	var value options
	flags := flag.NewFlagSet("language-delivery-scorecard", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&value.head, "expected-head", "", "exact subject commit")
	flags.StringVar(&value.contract, "contract", "", "delivery contract JSON")
	flags.StringVar(&value.manifest, "manifest", "", "source artifact manifest JSON")
	flags.StringVar(&value.user, "user-journey", "", "user journey receipt")
	flags.StringVar(&value.conformance, "conformance", "", "toolchain conformance receipt")
	flags.StringVar(&value.lsp, "lsp", "", "toolchain LSP receipt")
	flags.StringVar(&value.release, "release", "", "cross-platform release receipt")
	flags.StringVar(&value.execution, "execution", "", "language source execution receipt")
	flags.StringVar(&value.profile, "profile", "", "language profile receipt")
	flags.StringVar(&value.readiness, "readiness", "", "internal readiness artifact")
	flags.StringVar(&value.out, "out", "", "output report JSON")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if value.head == "" || value.contract == "" || value.manifest == "" || value.out == "" {
		fmt.Fprintln(stderr, "language-delivery-scorecard: required input missing")
		return 2
	}
	return evaluate(value, stdout, stderr)
}
