package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func constructionRecordObservationRows(c bodyexecution.JointConstruction) [][2]string {
	var rows [][2]string
	for _, step := range c.Initial.ConstructionSteps() {
		r := step.Generation.Report.RecordAssembly
		if r == nil || r.Observation == nil {
			continue
		}
		o := r.Observation
		rows = append(rows, [2]string{"Record source questions", reportedValue(step.Generation.Report.Activity) + ": " + reportedValue(o.Status)},
			[2]string{"Record local cases", fmt.Sprintf("%d initial + %d acquired = %d effective", len(o.InitialCases), o.EffectiveCases-len(o.InitialCases), o.EffectiveCases)},
			[2]string{"Record observation work", fmt.Sprintf("%d candidate evaluations; %d reused outputs; %d oracle inputs evaluated", o.CandidateEvaluations, o.ReusedOutputs, o.OracleEvaluations)})
		if len(o.Rounds) > 0 {
			last := o.Rounds[len(o.Rounds)-1]
			rows = append(rows, [2]string{"Record alternatives after questions", fmt.Sprintf("%d surviving; %d unobserved; %d unavailable", len(last.SurvivingMasks), last.Unobserved, len(last.UnavailableMasks))})
			if last.RecommendedIndex != nil && *last.RecommendedIndex >= 0 && *last.RecommendedIndex < len(last.Probes) && last.Observation == nil {
				rows = append(rows, [2]string{"Unanswered record input", reportedValue(last.Probes[*last.RecommendedIndex].Input)})
			}
		}
	}
	return rows
}
