package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	root, input, promotion, output, check, expectedSHA string
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", "", "repository root")
	flag.StringVar(&cfg.input, "input", "", "language concept artifact outside the repository")
	flag.StringVar(&cfg.promotion, "proposal-promotion", "", "verified proposal promotion outside the repository")
	flag.StringVar(&cfg.output, "output", "", "readiness artifact path outside the repository")
	flag.StringVar(&cfg.check, "check", "", "existing readiness artifact outside the repository")
	flag.StringVar(&cfg.expectedSHA, "expected-sha", "", "exact 40 character commit sha")
	flag.Parse()
	if err := run(cfg, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
