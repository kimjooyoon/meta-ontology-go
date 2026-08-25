package main

import "fmt"

const rootSummarySpace = "LOGICAL_SOURCE_SUMMARY"

func compileRootSummaryWitness(root directoryMetric, index indicatorIndex) (subjectWitness, error) {
	if root.Path != "." || root.SubjectKind != "PROJECT_ROOT" {
		return subjectWitness{}, fmt.Errorf("logical source root is not an exact project root")
	}
	rows := make([]sourceIndicator, 0, rootSummaryCount)
	for _, item := range rootSummaryMetrics(root) {
		row, err := lookupIndicator(index, root.SubjectKind, root.Path, item.id, item.value)
		if err != nil {
			return subjectWitness{}, err
		}
		if err := validateRootSummaryIndicator(row); err != nil {
			return subjectWitness{}, err
		}
		rows = append(rows, row)
	}
	return directoryWitness(rootSummarySpace, root, sourceBinding(rows, "COHERENCE")), nil
}
