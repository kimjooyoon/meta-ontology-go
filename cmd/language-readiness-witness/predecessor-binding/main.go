package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	root        string
	expectedSHA string
	output      string
	check       bool
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", "", "repository root")
	flag.StringVar(&cfg.expectedSHA, "expected-sha", "", "exact observed head SHA")
	flag.StringVar(&cfg.output, "output", "", "exclusive output path outside the repository")
	flag.BoolVar(&cfg.check, "check", false, "fail unless the binding evidence is fully resolved")
	flag.Parse()
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
