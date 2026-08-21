package main

import (
	"flag"
	"fmt"
	"io"

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
	if err := verifyRepositorySHA(cfg.root, cfg.sha); err != nil {
		return err
	}
	report, err := metricevidence.Load(cfg.metrics, cfg.sha)
	if err != nil {
		return err
	}
	indicators := report.GoSplitIndicators()
	if cfg.check {
		return checkSplitPlans(cfg, report, indicators)
	}
	if !metricevidence.Contains(indicators, cfg.subject) {
		return fmt.Errorf("subject %q is not an actionable split indicator", cfg.subject)
	}
	plan, err := planSource(cfg.root, cfg.subject, report.Meta.Policy.MaxFileLines)
	if err != nil {
		return err
	}
	if err := validateTopology(report, plan); err != nil {
		return err
	}
	if err := applySplit(plan); err != nil {
		return err
	}
	fmt.Printf("source-splitter: subject=%s outputs=%d write=true\n", cfg.subject, len(plan.Parts))
	return nil
}

func parseConfig(args []string) (config, error) {
	var cfg config
	flags := flag.NewFlagSet("source-splitter", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&cfg.root, "root", ".", "repository root")
	flags.StringVar(&cfg.metrics, "metrics", "", "exact-SHA source metrics JSON")
	flags.StringVar(&cfg.sha, "sha", "", "expected repository SHA")
	flags.StringVar(&cfg.subject, "subject", "", "single metric subject")
	flags.BoolVar(&cfg.check, "check", false, "plan every actionable split without writing")
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
