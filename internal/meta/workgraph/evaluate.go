package workgraph

import "fmt"

func Evaluate(contract Contract, observation Observation) (Report, error) {
	if err := ValidateContract(contract); err != nil {
		return Report{}, err
	}
	if len(observation.HeadSHA) != 40 {
		return Report{}, fmt.Errorf("head SHA must contain 40 characters")
	}
	cells := buildCells(contract, observation)
	summary := summarize(cells)
	result, resolution, reason, next := decision(summary, cells)
	report := Report{
		Schema: ReportSchema, HeadSHA: observation.HeadSHA, Project: contract.Project,
		Decision: result, Resolution: resolution, Reason: reason, NextOperation: next,
		MutationAllowed: false, ContractDigest: DigestValue(contract), SourceDigest: observation.SourceDigest,
		GeneratedDigest: observation.GeneratedDigest, ReplayDigest: observation.ReplayDigest,
		Cells: cells, Resource: observation.Resource, Summary: summary,

		Claim: claimLifecycle(contract, observation, cells, summary, next)}
	report.Indicators = indicators(report)
	return report, nil
}
