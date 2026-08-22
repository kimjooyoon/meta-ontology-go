package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", "", "repository root")
	flag.StringVar(&cfg.input, "input", "", "predecessor collector input")
	flag.StringVar(&cfg.predecessorReceipt, "predecessor-receipt", "", "selected predecessor receipt")
	flag.StringVar(&cfg.report, "report", "", "exclusive semantic receipt output")
	flag.StringVar(&cfg.expectedDigest, "expected-digest", "", "expected semantic report digest")
	flag.BoolVar(&cfg.check, "check", false, "reject fail-closed semantic state")
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
