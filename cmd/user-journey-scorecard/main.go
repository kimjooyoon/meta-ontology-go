package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	root, executable, head, contract, upstream, profile, output, check string
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", "", "absolute repository root")
	flag.StringVar(&cfg.executable, "executable", "", "profiled gooo executable")
	flag.StringVar(&cfg.head, "expected-head", "", "exact subject SHA")
	flag.StringVar(&cfg.contract, "contract", "", "versioned scorecard contract")
	flag.StringVar(&cfg.upstream, "upstream", "", "exact toolchain CLI receipt")
	flag.StringVar(&cfg.profile, "profile", "", "runner resource observations")
	flag.StringVar(&cfg.output, "output", "", "scorecard output")
	flag.StringVar(&cfg.check, "check", "", "existing scorecard to replay")
	flag.Parse()
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
