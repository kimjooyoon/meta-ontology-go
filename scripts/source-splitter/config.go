package main

import (
	"flag"
	"fmt"
	"io"
)

type config struct {
	root         string
	metrics      string
	sha          string
	subject      string
	plan         string
	output       string
	check        bool
	evidenceJSON bool
}

func parseConfig(args []string) (config, error) {
	var cfg config
	flags := flag.NewFlagSet("source-splitter", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&cfg.root, "root", ".", "repository root")
	flags.StringVar(&cfg.metrics, "metrics", "", "exact-SHA source metrics JSON")
	flags.StringVar(&cfg.sha, "sha", "", "expected repository SHA")
	flags.StringVar(&cfg.subject, "subject", "", "single metric subject")
	flags.StringVar(&cfg.plan, "plan", "", "logical split plan for batch write mode")
	flags.StringVar(&cfg.output, "output", "", "batch split receipt JSON")
	flags.BoolVar(&cfg.check, "check", false, "plan every actionable split without writing")
	flags.BoolVar(&cfg.evidenceJSON, "evidence-json", false, "emit raw write evidence as JSON")
	if err := flags.Parse(args); err != nil {
		return cfg, err
	}
	if flags.NArg() != 0 || cfg.metrics == "" || cfg.sha == "" {
		return cfg, fmt.Errorf("metrics and sha are required and positional arguments are forbidden")
	}
	if cfg.plan != "" || cfg.output != "" {
		if cfg.plan == "" || cfg.output == "" || cfg.subject != "" || cfg.check || cfg.evidenceJSON {
			return cfg, fmt.Errorf("plan and output require exclusive batch write mode")
		}
		return cfg, nil
	}
	if !cfg.check && cfg.subject == "" {
		return cfg, fmt.Errorf("subject is required in write mode")
	}
	if cfg.check && cfg.evidenceJSON {
		return cfg, fmt.Errorf("evidence-json requires write mode")
	}
	return cfg, nil
}
