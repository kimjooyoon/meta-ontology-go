package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	root, expectedHead, concept, corpus, receipts string
	output, bundle, check                         string
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", "", "repository root")
	flag.StringVar(&cfg.expectedHead, "expected-head", "", "exact source commit")
	flag.StringVar(&cfg.concept, "concept", "", "language concept artifact")
	flag.StringVar(&cfg.corpus, "corpus", "", "fixed release corpus")
	flag.StringVar(&cfg.receipts, "receipts", "", "external platform receipt directory")
	flag.StringVar(&cfg.output, "output", "", "external report path")
	flag.StringVar(&cfg.bundle, "bundle", "", "external release bundle directory")
	flag.StringVar(&cfg.check, "check", "", "existing external report path")
	flag.Parse()
	if err := run(cfg, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
