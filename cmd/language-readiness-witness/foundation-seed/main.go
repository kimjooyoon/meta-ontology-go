package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	root, input, expectedHead, output string
	check                             bool
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", "", "repository root")
	flag.StringVar(&cfg.input, "input", "", "exhausted predecessor resolution")
	flag.StringVar(&cfg.expectedHead, "expected-head", "", "exact current head SHA")
	flag.StringVar(&cfg.output, "output", "", "foundation receipt outside repository")
	flag.BoolVar(&cfg.check, "check", false, "require exact foundation authorization")
	flag.Parse()
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
