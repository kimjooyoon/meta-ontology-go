package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct { root, corpus, concept, expectedHead, output, check string }

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", "", "repository root")
	flag.StringVar(&cfg.corpus, "corpus", "", "versioned LSP corpus")
	flag.StringVar(&cfg.concept, "concept", "", "language concept artifact")
	flag.StringVar(&cfg.expectedHead, "expected-head", "", "exact commit SHA")
	flag.StringVar(&cfg.output, "output", "", "new report outside repository")
	flag.StringVar(&cfg.check, "check", "", "existing report outside repository")
	flag.Parse()
	if err := run(cfg, os.Stdout); err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
}
