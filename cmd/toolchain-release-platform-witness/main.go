package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	root, outputDir, expectedHead                   string
	platformID, runner, expectedGOOS, expectedGOARCH string
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", "", "repository root")
	flag.StringVar(&cfg.outputDir, "output-dir", "", "external platform artifact directory")
	flag.StringVar(&cfg.expectedHead, "expected-head", "", "exact source commit")
	flag.StringVar(&cfg.platformID, "platform-id", "", "fixed platform identifier")
	flag.StringVar(&cfg.runner, "runner", "", "fixed GitHub runner label")
	flag.StringVar(&cfg.expectedGOOS, "expected-goos", "", "expected native GOOS")
	flag.StringVar(&cfg.expectedGOARCH, "expected-goarch", "", "expected native GOARCH")
	flag.Parse()
	if err := run(cfg, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
