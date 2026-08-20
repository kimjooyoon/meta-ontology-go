package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/exactsha"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
)

type config struct {
	root    string
	metrics string
	sha     string
	subject string
	check   bool
}

func run(args []string) error {
	cfg, err := parseConfig(args)
	if err != nil {
		return err
	}
	if err := exactsha.Verify(cfg.root, cfg.sha); err != nil {
		return err
	}
	report, err := metricevidence.Load(cfg.metrics, cfg.sha)
	if err != nil {
		return err
	}
	indicators := report.GoSplitIndicators()
	if cfg.check {
		return checkPlans(cfg, report, indicators)
	}
	if !metricevidence.Contains(indicators, cfg.subject) {
		return fmt.Errorf("subject %q is not an actionable split indicator", cfg.subject)
	}
	plan, err := planRepack(cfg.root, cfg.subject, report.Meta.Policy.MaxFileLines)
	if err != nil {
		return err
	}
	if err := applyRepack(plan); err != nil {
		return err
	}
	fmt.Printf("source-repacker: subject=%s destination=%s write=true\n", cfg.subject, plan.Edits[1].Subject)
	return nil
}

func parseConfig(args []string) (config, error) {
	var cfg config
	flags := flag.NewFlagSet("source-repacker", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&cfg.root, "root", ".", "repository root")
	flags.StringVar(&cfg.metrics, "metrics", "", "exact-SHA source metrics JSON")
	flags.StringVar(&cfg.sha, "sha", "", "expected repository SHA")
	flags.StringVar(&cfg.subject, "subject", "", "single metric subject")
	flags.BoolVar(&cfg.check, "check", false, "plan actionable repacks without writing")
	if err := flags.Parse(args); err != nil {
		return cfg, err
	}
	if flags.NArg() != 0 || cfg.metrics == "" || cfg.sha == "" {
		return cfg, fmt.Errorf("metrics and sha are required and positional arguments are forbidden")
	}
	if !cfg.check && cfg.subject == "" {
		return cfg, fmt.Errorf("subject is required in write mode")
	}
	return cfg, nil
}
