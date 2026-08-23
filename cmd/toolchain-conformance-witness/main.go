package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	root, head, concept, corpus, syntax, semantic, query, interop string
	diagnostic, runtime, cli, formatFix, useCases, output, check string
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", "", "absolute logical repository root")
	flag.StringVar(&cfg.head, "expected-head", "", "exact commit SHA")
	flag.StringVar(&cfg.concept, "concept", "", "language concept artifact")
	flag.StringVar(&cfg.corpus, "corpus", "", "fixed conformance corpus")
	flag.StringVar(&cfg.syntax, "syntax", "", "syntax receipt")
	flag.StringVar(&cfg.semantic, "semantic", "", "semantic receipt")
	flag.StringVar(&cfg.query, "query", "", "query receipt")
	flag.StringVar(&cfg.interop, "interop", "", "Go interoperation receipt")
	flag.StringVar(&cfg.diagnostic, "diagnostic", "", "diagnostic receipt")
	flag.StringVar(&cfg.runtime, "runtime", "", "package runtime receipt")
	flag.StringVar(&cfg.cli, "cli", "", "CLI receipt")
	flag.StringVar(&cfg.formatFix, "format-fix", "", "format/fix receipt")
	flag.StringVar(&cfg.useCases, "use-cases", "", "executable use cases receipt")
	flag.StringVar(&cfg.output, "output", "", "conformance report output")
	flag.StringVar(&cfg.check, "check", "", "existing conformance report")
	flag.Parse()
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
