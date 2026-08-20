package linecaps

import (
	"fmt"
	"strings"
)

// Text returns a stable line-oriented report.
func (r LineMetricsReport) Text() string {
	var output strings.Builder
	sum := r.Total()
	fmt.Fprintf(&output, "line metrics: files=%d dirs=%d go_lines=%d gooo_lines=%d\n", sum.RecursiveFiles, sum.RecursiveFolders, sum.GoLines, sum.GoooLines)
	fmt.Fprintf(&output, "language totals: go_files=%d gooo_files=%d go_lines=%d gooo_lines=%d\n", sum.GoFiles, sum.GoooFiles, sum.GoLines, sum.GoooLines)
	fmt.Fprintf(&output, "meta indicators: total=%d blocking=%d failed=%d\n", len(r.Meta.Indicators), r.Meta.BlockingCount(), len(r.Meta.Failed()))
	writeLanguageFileSection := func(label FileLanguage, title string) {
		files := orderedFileMetrics(r.Files)
		count := 0
		lines := 0
		for _, file := range files {
			if file.Language != label {
				continue
			}
			count++
			lines += file.Lines
		}
		if count == 0 {
			return
		}
		fmt.Fprintf(&output, "%s files: count=%d lines=%d\n", title, count, lines)
		for _, file := range files {
			if file.Language != label {
				continue
			}
			fmt.Fprintf(&output, "  %s\tlines=%d\n", file.Path, file.Lines)
		}
	}
	writeLanguageFileSection(FileLanguageGo, "go")
	writeLanguageFileSection(FileLanguageGooo, "gooo")
	for _, directory := range orderedDirectoryMetrics(r.Directories) {
		fmt.Fprintf(&output, "%s: direct_folders=%d direct_files=%d folders=%d files=%d go_files=%d gooo_files=%d go_lines=%d gooo_lines=%d\n",
			directory.Path, directory.DirectFolders, directory.DirectFiles, directory.RecursiveFolders, directory.RecursiveFiles, directory.GoFiles, directory.GoooFiles, directory.GoLines, directory.GoooLines,
		)
	}
	return output.String()
}

// Summary returns bounded human output while JSON retains complete evidence.
func (r LineMetricsReport) Summary() string {
	var output strings.Builder
	total, failed := r.Total(), r.Meta.Failed()
	fmt.Fprintf(&output, "source metrics: commit=%s files=%d dirs=%d go_files=%d gooo_files=%d go_lines=%d gooo_lines=%d indicators=%d failed=%d\n", r.CommitSHA, total.RecursiveFiles, total.RecursiveFolders, total.GoFiles, total.GoooFiles, total.GoLines, total.GoooLines, len(r.Meta.Indicators), len(failed))
	for _, indicator := range failed {
		fmt.Fprintf(&output, "  %s\t%s\tvalue=%d limit=%d operation=%s proof=%s\n", indicator.MetricID, indicator.Subject, indicator.Value, indicator.Limit, indicator.MetaOperation, indicator.ProofChoice)
	}
	return output.String()
}

// Total returns aggregate metrics at the repository root.
func (r LineMetricsReport) Total() DirectoryMetric {
	if len(r.Directories) == 0 {
		return DirectoryMetric{}
	}
	for _, directory := range r.Directories {
		if directory.Path == "." {
			return directory
		}
	}
	return r.Directories[0]
}
