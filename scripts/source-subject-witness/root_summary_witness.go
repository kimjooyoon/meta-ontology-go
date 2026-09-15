package main

import "fmt"

const rootSummarySpace = "LOGICAL_SOURCE_SUMMARY"

func compileRootSummaryWitness(directories []directoryMetric, index indicatorIndex) (subjectWitness, error) {
	roots := make([]directoryMetric, 0, 1)
	for _, directory := range directories {
		if directory.Path == "." || directory.SubjectKind == "PROJECT_ROOT" {
			roots = append(roots, directory)
		}
	}
	if len(roots) != 1 || roots[0].Path != "." || roots[0].SubjectKind != "PROJECT_ROOT" {
		return subjectWitness{}, fmt.Errorf("logical source root is not an exact project root")
	}
	root := roots[0]
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
