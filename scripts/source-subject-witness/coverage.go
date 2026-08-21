package main

import "fmt"

type expectedMetric struct {
	id    string
	value int
}

func fileBinding(file fileMetric, index indicatorIndex) (metaBinding, error) {
	if file.Language == "other" {
		return derivedBinding("DERIVED_OBSERVATION", "observe-other-file", "FOUNDATION"), nil
	}
	metric := "gooo.metric.source." + file.Language + "-file-lines.v1"
	row, err := lookupIndicator(index, "FILE", file.Path, metric, file.Lines)
	if err != nil {
		return metaBinding{}, err
	}
	want := "split-" + file.Language + "-declarations"
	if file.Language == "gooo" {
		want = "split-gooo-sections"
	}
	if row.MetaOperation != want {
		return metaBinding{}, fmt.Errorf("file %q operation %q, want %q", file.Path, row.MetaOperation, want)
	}
	return sourceBinding([]sourceIndicator{row}, "COHERENCE"), nil
}

func logicalDirectoryBinding() metaBinding {
	return derivedBinding("DERIVED_PROJECTION", "observe-logical-directory", "COHERENCE")
}

func storageDirectoryBinding(directory directoryMetric, index indicatorIndex) (metaBinding, error) {
	kinds := 0
	if directory.DirectFiles > 0 {
		kinds++
	}
	if directory.DirectFolders > 0 {
		kinds++
	}
	expected := []expectedMetric{
		{"gooo.metric.layout.direct-entries.v1", directory.DirectFiles + directory.DirectFolders},
		{"gooo.metric.layout.direct-files.v1", directory.DirectFiles},
		{"gooo.metric.layout.direct-folders.v1", directory.DirectFolders},
		{"gooo.metric.layout.entry-kinds.v1", kinds},
		{"gooo.metric.layout.recursive-files.v1", directory.RecursiveFiles},
		{"gooo.metric.layout.recursive-folders.v1", directory.RecursiveFolders},
	}
	rows := make([]sourceIndicator, 0, len(expected))
	for _, item := range expected {
		row, err := lookupIndicator(index, directory.SubjectKind, directory.Path, item.id, item.value)
		if err != nil {
			return metaBinding{}, err
		}
		rows = append(rows, row)
	}
	notApplicable := 0
	for _, row := range rows {
		if row.Applicability == "NOT_APPLICABLE" {
			notApplicable++
		}
	}
	if (directory.Path == "." && notApplicable != 2) || (directory.Path != "." && notApplicable != 0) {
		return metaBinding{}, fmt.Errorf("directory %q has %d topology exemptions", directory.Path, notApplicable)
	}
	return sourceBinding(rows, "FOUNDATION"), nil
}

func sourceBinding(rows []sourceIndicator, route string) metaBinding {
	return metaBinding{Kind: "SOURCE_INDICATORS", Operation: operationSet(rows), Route: route,
		IndicatorCount: len(rows), IndicatorDigest: digestValues(rows)}
}

func derivedBinding(kind, operation, route string) metaBinding {
	return metaBinding{Kind: kind, Operation: operation, Route: route, IndicatorDigest: digestValues([]sourceIndicator{})}
}
