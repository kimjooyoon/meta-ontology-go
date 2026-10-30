package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
)

func planSource(root, subject string, limit int) (splitPlan, error) {
	return generation.NewGoSplitPlanner().PlanSource(root, subject, limit)
}

func prepareSource(root, subject string, report metricevidence.Report) (splitPlan, error) {
	return generation.NewGoSplitPlanner().Prepare(root, subject, report)
}
