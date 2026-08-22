package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	root           string
	actionability  string
	observations   string
	report         string
	check          bool
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", "", "repository root")
	flag.StringVar(&cfg.actionability, "actionability", "", "meta-actionability report JSON")
	flag.StringVar(&cfg.observations, "observations", "", "artifact observation JSON")
	flag.StringVar(&cfg.report, "report", "", "coverage report path outside the repository")
	flag.BoolVar(&cfg.check, "check", false, "reject fail-closed coverage")
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
