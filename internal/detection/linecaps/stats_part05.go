package linecaps

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

// EvaluateLineMetricIndicators connects raw measurements to policy and its
// actionable meta operations.
func EvaluateLineMetricIndicators(report LineMetricsReport, policy sourcepolicy.Policy) (sourcepolicy.Report, error) {
	observations := make([]sourcepolicy.Observation, 0, len(report.Files)+len(report.Directories)*6+4)
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
	for _, directory := range report.Directories {
		observations = append(observations, directoryMetricObservations(directory)...)
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
	total := report.Total()
	observations = append(observations,
		metricObservation(".", sourcepolicy.DimensionGoFiles, total.GoFiles),
		metricObservation(".", sourcepolicy.DimensionGoooFiles, total.GoooFiles),
		metricObservation(".", sourcepolicy.DimensionGoLines, total.GoLines),
		metricObservation(".", sourcepolicy.DimensionGoooLines, total.GoooLines),
	)
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
	subject := fmt.Sprintf("%s:%d:%s", finding.Path, finding.StartLine, finding.Name)
	return sourcepolicy.Observation{Subject: subject, Dimension: dimension, Value: finding.Actual, Detail: finding.Detail, Producer: "linecaps.Analyze"}, true
}
