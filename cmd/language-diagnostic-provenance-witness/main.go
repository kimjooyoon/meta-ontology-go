package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	root, head, concept, registry, output, check string
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", ".", "logical repository root")
	flag.StringVar(&cfg.head, "expected-head", "", "exact commit SHA")
	flag.StringVar(&cfg.concept, "concept", "", "language concept artifact")
	flag.StringVar(&cfg.registry, "registry", "", "fixed diagnostic provenance registry")
	flag.StringVar(&cfg.output, "output", "", "diagnostic provenance artifact output")
	flag.StringVar(&cfg.check, "check", "", "existing artifact to replay")
	flag.Parse()
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
