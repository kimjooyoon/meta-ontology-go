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

func storageDirectoryBinding(directory directoryMetric, index indicatorIndex, readmeValue int) (metaBinding, error) {
	kinds := 0
	if directory.DirectFiles > 0 {
		kinds++
	}
	if directory.DirectFolders > 0 {
		kinds++
	}
	expected := []expectedMetric{{"gooo.metric.layout.direct-entries.v1", directory.DirectFiles + directory.DirectFolders}, {"gooo.metric.layout.direct-files.v1", directory.DirectFiles}, {"gooo.metric.layout.direct-folders.v1", directory.DirectFolders}, {"gooo.metric.layout.entry-kinds.v1", kinds}, {"gooo.metric.layout.recursive-files.v1", directory.RecursiveFiles}, {"gooo.metric.layout.recursive-folders.v1", directory.RecursiveFolders}}
	if directory.Path == "." {
		expected = append(expected, expectedMetric{rootREADMEMetric, readmeValue})
		expected = append(expected, rootSummaryMetrics(directory)...)
	}
	rows := make([]sourceIndicator, 0, len(expected))
	for _, item := range expected {
		row, err := lookupIndicator(index, directory.SubjectKind, directory.Path, item.id, item.value)
		if err != nil {
			return metaBinding{}, err
		}
		if item.id == rootREADMEMetric && row.Detail != "ontology="+rootREADMEOntology {
			return metaBinding{}, fmt.Errorf("root README indicator lost ontology binding")
		}
		if err := validateRootSummaryIndicator(row); err != nil {
			return metaBinding{}, err
		}
		rows = append(rows, row)
	}
	if err := validateDirectoryApplicability(directory, rows); err != nil {
		return metaBinding{}, err
	}
	return sourceBinding(rows, "FOUNDATION"), nil
}
