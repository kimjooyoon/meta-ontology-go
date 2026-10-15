package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func constructionCriteriaQuestionRows(c bodyexecution.JointConstruction) [][2]string {
	var rows [][2]string
	for _, step := range c.Initial.ConstructionSteps() {
		r := step.Generation.Report.RecordAssembly
		if r == nil || r.CriteriaQuestions == nil {
			continue
		}
		o := r.CriteriaQuestions
		rows = append(rows,
			[2]string{"Gooo criteria questions", reportedValue(step.Generation.Report.Activity) + ": " + reportedValue(o.Status)},
			[2]string{"Criteria input obligations", fmt.Sprintf("%d initial + %d acquired = %d effective", len(o.InitialInputs), len(o.AcquiredInputs), o.EffectiveInputs)},
			[2]string{"Criteria question work", fmt.Sprintf("%d candidate evaluations; %d predicate evaluations; %d reused outputs; %d reused predicate results", o.CandidateEvaluations, o.PredicateEvaluations, o.ReusedOutputs, o.ReusedCriteriaResults)})
		if o.Options.CriteriaStrategy != "" {
			rows = append(rows, [2]string{"Criteria question strategy", o.Options.CriteriaStrategy})
			for _, round := range o.Rounds {
				if round.AcquiredInput != "" {
					rows = append(rows, [2]string{"Acquired criteria input", reportedValue(round.AcquiredInput) + ": " + round.RecommendationReason})
				}
			}
		}
		if len(o.Rounds) == 0 {
			continue
		}
		last := o.Rounds[len(o.Rounds)-1]
		rows = append(rows, [2]string{"Alternatives after criteria questions", fmt.Sprintf("%d surviving; %d rejected; %d unavailable; %d unobserved", len(last.SurvivingMasks), len(last.RejectedMasks), len(last.UnavailableMasks), last.Unobserved)})
		if last.RecommendedIndex != nil && *last.RecommendedIndex >= 0 && *last.RecommendedIndex < len(last.Probes) && last.AcquiredInput == "" {
			rows = append(rows, [2]string{"Unanswered criteria input", reportedValue(last.Probes[*last.RecommendedIndex].Input)})
		}
	}
	return rows
}
