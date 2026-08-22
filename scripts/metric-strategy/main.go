package main

import (
	"flag"
	"fmt"
	"os"
)

type options struct {
	mode, metrics, intervention, interventionVerification string
	root, repository, subjectSHA, plan, replayPlan        string
	strategyOptions
}

type strategyOptions struct {
	strategyVerification, predecessorSHA, selectedProposal, githubAPI, output string
}

func main() {
	var value options
	flag.StringVar(&value.mode, "mode", "generate", "generate, verify, or measure a metric strategy")
	flag.StringVar(&value.metrics, "metrics", "", "source metric report")
	flag.StringVar(&value.intervention, "intervention", "", "metric intervention ledger")
	flag.StringVar(&value.interventionVerification, "intervention-verification", "", "metric intervention verification")
	flag.StringVar(&value.root, "root", ".", "repository root for language concept binding")
	flag.StringVar(&value.repository, "repository", "", "repository identity")
	flag.StringVar(&value.subjectSHA, "subject-sha", "", "exact subject commit")
	flag.StringVar(&value.plan, "plan", "", "metric strategy plan for verification")
	flag.StringVar(&value.replayPlan, "replay-plan", "", "independently replayed metric strategy plan")
	flag.StringVar(&value.strategyVerification, "strategy-verification", "", "metric strategy verification receipt")
	flag.StringVar(&value.predecessorSHA, "predecessor-sha", "", "merged predecessor commit")
	flag.StringVar(&value.selectedProposal, "selected-proposal", "", "selected predecessor proposal contract")
	flag.StringVar(&value.githubAPI, "github-api", os.Getenv("GITHUB_API_URL"), "GitHub API root")
	flag.StringVar(&value.output, "output", "", "output JSON path")
	flag.Parse()
	if err := run(value); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
