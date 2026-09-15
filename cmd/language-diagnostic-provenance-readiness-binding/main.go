package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	root, head, concept, readiness, provenance, output, check string
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", ".", "logical repository root")
	flag.StringVar(&cfg.head, "expected-head", "", "exact commit SHA")
	flag.StringVar(&cfg.concept, "concept", "", "language concept artifact")
	flag.StringVar(&cfg.readiness, "readiness", "", "language readiness artifact")
	flag.StringVar(&cfg.provenance, "provenance", "", "diagnostic provenance artifact")
	flag.StringVar(&cfg.output, "output", "", "binding artifact output")
	flag.StringVar(&cfg.check, "check", "", "existing binding artifact to replay")
	flag.Parse()
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
