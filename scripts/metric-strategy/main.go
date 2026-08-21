package main

import (
	"flag"
	"fmt"
	"os"
)

type options struct {
	mode, metrics, intervention, interventionVerification string
	repository, subjectSHA, plan, output                 string
}

func main() {
	var value options
	flag.StringVar(&value.mode, "mode", "generate", "generate or verify a metric strategy")
	flag.StringVar(&value.metrics, "metrics", "", "source metric report")
	flag.StringVar(&value.intervention, "intervention", "", "metric intervention ledger")
	flag.StringVar(&value.interventionVerification, "intervention-verification", "", "metric intervention verification")
	flag.StringVar(&value.repository, "repository", "", "repository identity")
	flag.StringVar(&value.subjectSHA, "subject-sha", "", "exact subject commit")
	flag.StringVar(&value.plan, "plan", "", "metric strategy plan for verification")
	flag.StringVar(&value.output, "output", "", "output JSON path")
	flag.Parse()
	if err := run(value); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

