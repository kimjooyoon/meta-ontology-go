package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	root, head, registry, concept, output, check string
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", "", "repository root")
	flag.StringVar(&cfg.head, "head", "", "exact 40 character commit sha")
	flag.StringVar(&cfg.registry, "registry", "", "versioned use case registry")
	flag.StringVar(&cfg.concept, "concept-artifact", "", "language concept artifact outside the repository")
	flag.StringVar(&cfg.output, "output", "", "receipt path outside the repository")
	flag.StringVar(&cfg.check, "check", "", "existing receipt outside the repository")
	flag.Parse()
	if err := run(cfg, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
