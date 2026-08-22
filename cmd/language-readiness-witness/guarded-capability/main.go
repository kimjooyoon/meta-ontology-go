package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	root, currentHead, foundationArchive, output, check string
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", "", "repository root")
	flag.StringVar(&cfg.currentHead, "current-head", "", "exact current commit sha")
	flag.StringVar(&cfg.foundationArchive, "foundation-archive", "", "external foundation archive")
	flag.StringVar(&cfg.output, "output", "", "capability receipt outside the repository")
	flag.StringVar(&cfg.check, "check", "", "existing capability receipt outside the repository")
	flag.Parse()
	if err := run(cfg, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
