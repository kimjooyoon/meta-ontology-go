package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	assurance, head, syntax, semantics, binding string
	useCases, toolchain, release, output        string
	unavailableBoundary                        string
	unavailableAssurance                       bool
}

func main() {
	var cfg config
	flag.StringVar(&cfg.assurance, "assurance", "", "merged official assurance report")
	flag.StringVar(&cfg.head, "head", "", "current exact subject SHA")
	flag.StringVar(&cfg.syntax, "syntax", "", "syntax receipt")
	flag.StringVar(&cfg.semantics, "semantics", "", "semantic receipt")
	flag.StringVar(&cfg.binding, "binding", "", "semantic binding receipt")
	flag.StringVar(&cfg.useCases, "use-cases", "", "executable use-case receipt")
	flag.StringVar(&cfg.toolchain, "toolchain", "", "toolchain conformance receipt")
	flag.StringVar(&cfg.release, "release", "", "cross-platform release receipt")
	flag.StringVar(&cfg.output, "output", "", "shadow report output")
	flag.StringVar(&cfg.unavailableBoundary, "unavailable-boundary", "", "boundary to omit")
	flag.BoolVar(&cfg.unavailableAssurance, "unavailable-assurance", false, "omit assurance")
	flag.Parse()
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
