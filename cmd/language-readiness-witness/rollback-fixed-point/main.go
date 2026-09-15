package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	root, guard, transformation, expectedSHA, output string
	check                                            bool
}

func main() {
	var cfg config
	flag.StringVar(&cfg.root, "root", "", "repository root")
	flag.StringVar(&cfg.guard, "guard", "", "guarded promotion report")
	flag.StringVar(&cfg.transformation, "transformation", "", "transformation ledger")
	flag.StringVar(&cfg.expectedSHA, "expected-sha", "", "exact current subject sha")
	flag.StringVar(&cfg.output, "output", "", "recovery report output")
	flag.BoolVar(&cfg.check, "check", false, "compare existing output")
	flag.Parse()
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
