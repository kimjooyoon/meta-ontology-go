package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	root, head, concept, corpus, output, check string
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", ".", "logical repository root")
	flag.StringVar(&cfg.head, "expected-head", "", "exact commit SHA")
	flag.StringVar(&cfg.concept, "concept", "", "language concept artifact")
	flag.StringVar(&cfg.corpus, "corpus", "", "fixed package runtime corpus")
	flag.StringVar(&cfg.output, "output", "", "package runtime report output")
	flag.StringVar(&cfg.check, "check", "", "existing package runtime report")
	flag.Parse()
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
