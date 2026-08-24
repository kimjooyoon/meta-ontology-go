package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", "", "repository root")
	flag.StringVar(&cfg.input, "input", "", "source authority evidence input")
	flag.StringVar(&cfg.output, "output", "", "exclusive shadow receipt output")
	flag.StringVar(&cfg.expectedSHA, "expected-sha", "", "exact expected subject SHA")
	flag.BoolVar(&cfg.check, "check", false, "reject a non-satisfied shadow observation")
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
