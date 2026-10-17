package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	root, executable, head, concept, corpus, output, check, compare string
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", "", "absolute logical repository root")
	flag.StringVar(&cfg.executable, "executable", "", "exact gooo executable outside the repository")
	flag.StringVar(&cfg.head, "expected-head", "", "exact commit SHA")
	flag.StringVar(&cfg.concept, "concept", "", "language concept artifact")
	flag.StringVar(&cfg.corpus, "corpus", "", "fixed toolchain format/fix corpus")
	flag.StringVar(&cfg.output, "output", "", "toolchain format/fix report output")
	flag.StringVar(&cfg.check, "check", "", "existing toolchain format/fix report")
	flag.StringVar(&cfg.compare, "compare", "", "compare this saved report with --check; no native execution")
	flag.Parse()
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
