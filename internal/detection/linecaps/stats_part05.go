package linecaps

import (
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/duplicates"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

// EvaluateLineMetricIndicators connects raw measurements to policy and its
// actionable meta operations.
func EvaluateLineMetricIndicators(report LineMetricsReport, policy sourcepolicy.Policy) (sourcepolicy.Report, error) {
	topology := metricTopologyDirectories(report)
	observations := make([]sourcepolicy.Observation, 0, len(report.Files)+len(topology)*6+5)
	for _, file := range report.Files {
		dimension := sourcepolicy.Dimension("")
		switch file.Language {
		case FileLanguageGo:
			dimension = sourcepolicy.DimensionGoFileLines
		case FileLanguageGooo:
			dimension = sourcepolicy.DimensionGoooFileLines
		default:
			continue
		}
		observations = append(observations, metricObservation(file.Path, dimension, file.Lines))
	}
	producer := metricTopologyProducer(report)
	for _, directory := range topology {
		metrics := directoryMetricObservations(directory,
			isWorkflowDiscoveryRoot(report, directory))
		for index := range metrics {
			metrics[index].Producer = producer
		}
		observations = append(observations, metrics...)
	}
	analysis, err := Analyze(report.Root, nil, Limits{MaxFileLines: policy.MaxFileLines, MaxFunctionLines: policy.MaxFunctionLines})
	if err != nil {
		return sourcepolicy.Report{}, err
	}
	for _, finding := range analysis.Findings {
		if observation, ok := analysisObservation(finding); ok {
			observations = append(observations, observation)
		}
	}
	duplicateObservations, err := duplicates.Analyze(report.Root)
	if err != nil {
		return sourcepolicy.Report{}, err
	}
	observations = append(observations, duplicateObservations...)
	total := report.Total()
	observations = append(observations, metricObservation(".", sourcepolicy.DimensionGoFiles, total.GoFiles), metricObservation(".", sourcepolicy.DimensionGoooFiles, total.GoooFiles), metricObservation(".", sourcepolicy.DimensionGoLines, total.GoLines), metricObservation(".", sourcepolicy.DimensionGoooLines, total.GoooLines), rootREADMEObservation(report.Files))
	return sourcepolicy.Evaluate(policy, observations)
}

func isWorkflowDiscoveryRoot(report LineMetricsReport, directory DirectoryMetric) bool {
	if directory.Path != ".github/workflows" || directory.DirectFiles == 0 ||
		directory.DirectFolders != 0 {
		return false
	}
	files := 0
	for _, file := range report.Files {
		if filepath.ToSlash(filepath.Dir(file.Path)) != directory.Path {
			continue
		}
		extension := strings.ToLower(filepath.Ext(file.Path))
		if extension != ".yml" && extension != ".yaml" {
			return false
		}
		files++
	}
	return files == directory.DirectFiles
}

func metricObservation(subject string, dimension sourcepolicy.Dimension, value int) sourcepolicy.Observation {
	return sourcepolicy.Observation{Subject: subject, Dimension: dimension, Value: value, Producer: "linecaps.AnalyzeLineMetrics"}
}

func analysisObservation(finding Finding) (sourcepolicy.Observation, bool) {
	dimension := sourcepolicy.Dimension("")
	switch finding.Rule {
	case RuleFunctionLines:
		dimension = sourcepolicy.DimensionFunctionLines
	case RuleRefactorReturn:
		dimension = sourcepolicy.DimensionRefactorReturn
	case RuleRefactorAssign:
		dimension = sourcepolicy.DimensionRefactorAssign
	default:
		return sourcepolicy.Observation{}, false
	}
	subject := (sourcepolicy.SourceSubject{Path: finding.Path, Line: finding.StartLine, Name: finding.Name}).String()
	return sourcepolicy.Observation{Subject: subject, Dimension: dimension, Value: finding.Actual, Detail: finding.Detail, Producer: "linecaps.Analyze"}, true
}
