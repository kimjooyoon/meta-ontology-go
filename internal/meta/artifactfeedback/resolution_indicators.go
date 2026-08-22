package artifactfeedback

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/semanticresolution"

type resolutionIndicatorSpec struct {
	metric, unit, operation, activity string
	class                            semanticresolution.IndicatorClass
	relation                         semanticresolution.Relation
	proof                            semanticresolution.ProofChoice
	target                           int
}

var resolutionIndicatorSpecs = []resolutionIndicatorSpec{
	{MetricResolutionRecovery, "basis_points", "select-coarse-recovery-operation", "SelectCoarseRecoveryOperation", semanticresolution.ClassOutcome, semanticresolution.RelationGreaterOrEqual, semanticresolution.ProofCoherence, 10000},
	{MetricConflictObservation, "basis_points", "observe-semantic-conflict", "ObserveSemanticConflict", semanticresolution.ClassDriver, semanticresolution.RelationGreaterOrEqual, semanticresolution.ProofFoundation, 10000},
	{MetricMonotoneDescent, "basis_points", "lower-semantic-resolution", "LowerSemanticResolution", semanticresolution.ClassDriver, semanticresolution.RelationGreaterOrEqual, semanticresolution.ProofCoherence, 10000},
	{MetricTransitionReplay, "basis_points", "replay-resolution-transition", "ReplayResolutionTransition", semanticresolution.ClassDriver, semanticresolution.RelationGreaterOrEqual, semanticresolution.ProofRegression, 10000},
	{MetricFalseFixedPoint, "decisions", "select-coarse-recovery-operation", "SelectCoarseRecoveryOperation", semanticresolution.ClassGuardrail, semanticresolution.RelationLessOrEqual, semanticresolution.ProofCoherence, 0},
	{MetricResolutionDescents, "descents", "preserve-resolution-descent-bound", "PreserveResolutionDescentBound", semanticresolution.ClassGuardrail, semanticresolution.RelationLessOrEqual, semanticresolution.ProofFoundation, semanticresolution.MaxResolutionDescents},
	{MetricResolutionWrites, "repository_writes", "preserve-read-only-resolution", "PreserveReadOnlyResolution", semanticresolution.ClassGuardrail, semanticresolution.RelationLessOrEqual, semanticresolution.ProofFoundation, 0},
}

func resolutionIndicators(report ResolutionReport, conflict bool) []ResolutionIndicator {
	valid, monotone := resolutionValidity(report, conflict)
	falseFixedPoint := 0
	if conflict && report.Decision == "FIXED_POINT" {
		falseFixedPoint = 1
	}
	values := []int{valid, 10000, monotone, valid, falseFixedPoint, report.Descents, report.RepositoryWrites}
	indicators := make([]ResolutionIndicator, 0, len(resolutionIndicatorSpecs))
	for index, spec := range resolutionIndicatorSpecs {
		satisfied := values[index] >= spec.target
		if spec.relation == semanticresolution.RelationLessOrEqual {
			satisfied = values[index] <= spec.target
		}
		indicators = append(indicators, ResolutionIndicator{
			MetricID: spec.metric, Class: spec.class, Target: spec.target,
			Unit: spec.unit, Relation: spec.relation, ProofChoice: spec.proof,
			Producer: "artifactfeedback.EvaluateWithResolution",
			Consumer: "self-improvement-cycle", MetaOperation: spec.operation,
			Activity: spec.activity, Value: values[index], Satisfied: satisfied,
		})
	}
	return indicators
}

func resolutionValidity(report ResolutionReport, conflict bool) (int, int) {
	if !conflict {
		valid := report.Decision == report.Feedback.Decision &&
			report.ToResolution == report.FromResolution &&
			report.Descents == report.PreviousDescents
		return basisPoints(valid), basisPoints(valid)
	}
	if report.Decision == DecisionLowerResolution {
		next, ok := semanticresolution.LowerSemanticResolution(report.FromResolution)
		valid := ok && report.ToResolution == next &&
			report.Descents == report.PreviousDescents+1 &&
			report.NextOperation == NextOperationReevaluateFeedback
		return basisPoints(valid), basisPoints(valid)
	}
	valid := report.Decision == "FAIL_CLOSED" && report.NextOperation == ""
	return basisPoints(valid), basisPoints(valid)
}

func basisPoints(value bool) int {
	if value {
		return 10000
	}
	return 0
}
