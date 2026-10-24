package linecaps

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/duplicates"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

// EvaluateLineMetricIndicators connects raw measurements to policy and its
// actionable meta operations.
func EvaluateLineMetricIndicators(report LineMetricsReport, policy sourcepolicy.Policy) (sourcepolicy.Report, error) {
	topology := metricTopologyDirectories(report)
	observations := make([]sourcepolicy.Observation, 0, len(report.Files)+len(topology)*6+5)
	goFiles := make(map[string]int)
	for index, file := range report.Files {
		dimension := sourcepolicy.Dimension("")
		switch file.Language {
		case FileLanguageGo:
			dimension = sourcepolicy.DimensionGoFileLines
			goFiles[file.Path] = index
		case FileLanguageGooo:
			dimension = sourcepolicy.DimensionGoooFileLines
		default:
			continue
		}
		observation := metricObservation(file.Path, dimension, file.Lines)
		if file.Language == FileLanguageGo {
			observation.ModuleRoot = file.ModuleRoot
			observation.ModuleBoundaryReason = file.ModuleBoundaryReason
		}
		observations = append(observations, observation)
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
			observations = append(observations,
				bindModuleObservation(observation, []string{finding.Path}, report.Files, goFiles))
		}
	}
	duplicateObservations, err := duplicates.Analyze(report.Root)
	if err != nil {
		return sourcepolicy.Report{}, err
	}
	for _, observation := range duplicateObservations {
		observations = append(observations,
			bindModuleObservation(observation, observation.SourcePaths, report.Files, goFiles))
	}
	total := report.Total()
	observations = append(observations, metricObservation(".", sourcepolicy.DimensionGoFiles, total.GoFiles), metricObservation(".", sourcepolicy.DimensionGoooFiles, total.GoooFiles), metricObservation(".", sourcepolicy.DimensionGoLines, total.GoLines), metricObservation(".", sourcepolicy.DimensionGoooLines, total.GoooLines), rootREADMEObservation(report.Files))
	return sourcepolicy.Evaluate(policy, observations)
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
