package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	root           string
	input          string
	report         string
	expectedDigest string
	check          bool
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", "", "repository root")
	flag.StringVar(&cfg.input, "input", "", "resolution input JSON")
	flag.StringVar(&cfg.report, "report", "", "receipt path outside the repository")
	flag.StringVar(&cfg.expectedDigest, "expected-digest", "", "expected report digest")
	flag.BoolVar(&cfg.check, "check", false, "reject fail-closed resolution")
	flag.Parse()
	failClosed, err := run(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if failClosed {
		os.Exit(1)
	}
}
