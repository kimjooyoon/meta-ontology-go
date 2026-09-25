package main

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/workgraph"

func workgraphEvaluate(contract workgraph.Contract, observation workgraph.Observation) (workgraph.Report, error) {
	return workgraph.Evaluate(contract, observation)
}
