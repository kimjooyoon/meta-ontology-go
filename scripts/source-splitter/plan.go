package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcesplit"
)

type splitPlan = sourcesplit.Plan
type splitPart = sourcesplit.Part
type declarationOrder = sourcesplit.DeclarationOrder

var errSplitBlocked = sourcesplit.ErrBlocked

func planSource(root, subject string, limit int) (splitPlan, error) {
	return sourcesplit.PlanSource(root, subject, limit)
}

func validateTopology(report metricevidence.Report, plan splitPlan) error {
	return sourcesplit.ValidateTopology(report, plan)
}

func physicalLines(data []byte) int { return sourcesplit.PhysicalLines(data) }
