package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	root   string
	output string
	check  string
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", "", "repository root")
	flag.StringVar(&cfg.output, "output", "", "artifact path outside the repository")
	flag.StringVar(&cfg.check, "check", "", "existing artifact to consume outside the repository")
	flag.Parse()
	if err := run(cfg, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
