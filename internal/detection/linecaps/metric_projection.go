package linecaps

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

// AnalyzeProjectedLineMetrics binds logical source metrics to a separately
// materialized physical storage topology.
func AnalyzeProjectedLineMetrics(root, storageRoot string) (LineMetricsReport, error) {
	if storageRoot == "" {
		return LineMetricsReport{}, fmt.Errorf("line metrics storage root must not be empty")
	}
	sourcePath, err := filepath.Abs(root)
	if err != nil {
		return LineMetricsReport{}, err
	}
	storagePath, err := filepath.Abs(storageRoot)
	if err != nil {
		return LineMetricsReport{}, err
	}
	if sourcePath == storagePath {
		return AnalyzeLineMetrics(root)
	}
	report, err := AnalyzeLineMetrics(root)
	if err != nil {
		return LineMetricsReport{}, err
	}
	storage, err := AnalyzeLineMetrics(storageRoot)
	if err != nil {
		return LineMetricsReport{}, err
	}
	report.StorageRoot = filepath.ToSlash(storageRoot)
	report.StorageDirectories = storage.Directories
	report.Meta, err = EvaluateLineMetricIndicators(report, sourcepolicy.Default())
	if err != nil {
		return LineMetricsReport{}, err
	}
	return report, nil
}

func metricTopologyDirectories(report LineMetricsReport) []DirectoryMetric {
	if len(report.StorageDirectories) > 0 {
		return report.StorageDirectories
	}
	return report.Directories
}

func metricTopologyProducer(report LineMetricsReport) string {
	if len(report.StorageDirectories) > 0 {
		return "repository-projector.topology"
	}
	return "linecaps.AnalyzeLineMetrics"
}

func writeDirectoryMetrics(output *strings.Builder, plane string, directories []DirectoryMetric) {
	fmt.Fprintf(output, "%s directories:\n", plane)
	for _, directory := range orderedDirectoryMetrics(directories) {
		fmt.Fprintf(output, "%s: direct_folders=%d direct_files=%d folders=%d files=%d go_files=%d gooo_files=%d go_lines=%d gooo_lines=%d\n",
			directory.Path, directory.DirectFolders, directory.DirectFiles, directory.RecursiveFolders, directory.RecursiveFiles, directory.GoFiles, directory.GoooFiles, directory.GoLines, directory.GoooLines,
		)
	}
}
