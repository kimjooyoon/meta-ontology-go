package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", "", "repository root")
	flag.StringVar(&cfg.coverage, "coverage", "", "operation coverage report")
	flag.StringVar(&cfg.provenance, "provenance", "", "self-improvement provenance")
	flag.StringVar(&cfg.output, "output", "", "resolution input outside the repository")
	flag.StringVar(&cfg.ciConclusion, "ci-conclusion", "success", "CI conclusion")
	flag.Parse()
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
