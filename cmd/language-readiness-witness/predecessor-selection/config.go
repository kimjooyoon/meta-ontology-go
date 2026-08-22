package main

import "flag"

type config struct {
	root, repository, currentHead, predecessor, branch, workflow string
	baseline, reference, bindingBaseline, receipt string
}

func parseConfig() config {
	value := config{}
	flag.StringVar(&value.root, "root", "", "repository root")
	flag.StringVar(&value.repository, "repository", "", "owner/repository")
	flag.StringVar(&value.currentHead, "current-head", "", "current exact head SHA")
	flag.StringVar(&value.predecessor, "predecessor-sha", "", "optional exact predecessor SHA")
	flag.StringVar(&value.branch, "branch", "", "canonical predecessor branch")
	flag.StringVar(&value.workflow, "workflow", "", "canonical predecessor workflow")
	flag.StringVar(&value.baseline, "baseline", "", "selected readiness artifact output")
	flag.StringVar(&value.reference, "reference", "", "selected baseline reference output")
	flag.StringVar(&value.bindingBaseline, "binding-baseline", "", "selected binding report output")
	flag.StringVar(&value.receipt, "receipt", "", "selection receipt output")
	flag.Parse()
	return value
}
