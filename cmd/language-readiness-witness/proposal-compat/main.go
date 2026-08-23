package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	root, input, expectedSHA, legacy, receipt string
	check                                     bool
}

func main() {
	var cfg config
	flag.StringVar(&cfg.root, "root", "", "repository root")
	flag.StringVar(&cfg.input, "input", "", "v2 proposal promotion receipt")
	flag.StringVar(&cfg.expectedSHA, "expected-sha", "", "exact current subject sha")
	flag.StringVar(&cfg.legacy, "legacy", "", "v1 guard-compatible output")
	flag.StringVar(&cfg.receipt, "receipt", "", "compatibility receipt output")
	flag.BoolVar(&cfg.check, "check", false, "compare existing outputs")
	flag.Parse()
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
