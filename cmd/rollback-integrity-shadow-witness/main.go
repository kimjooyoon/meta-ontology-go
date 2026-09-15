package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	assurance, output string
	unavailable       bool
}

func main() {
	var cfg config
	flag.StringVar(&cfg.assurance, "assurance", "", "official assurance report")
	flag.StringVar(&cfg.output, "output", "", "shadow report output")
	flag.BoolVar(&cfg.unavailable, "unavailable", false, "evaluate unavailable evidence")
	flag.Parse()
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
