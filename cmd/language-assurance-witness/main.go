package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	root, subjectSHA, input, output, expectCandidate string
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", "", "repository root")
	flag.StringVar(&cfg.subjectSHA, "subject-sha", "", "exact evaluated commit sha")
	flag.StringVar(&cfg.input, "input", "", "transaction input")
	flag.StringVar(&cfg.output, "output", "", "report path outside the repository")
	flag.StringVar(&cfg.expectCandidate, "expect-candidate", "", "required candidate decision")
	flag.Parse()
	if err := run(cfg, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
