package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func constructionInputHistoryRows(r bodyexecution.CompositionRuntime) [][2]string {
	if r.ConstructionInputHistorySHA256 == "" {
		return [][2]string{{"Constructed input history", "not recorded"}}
	}
	s := r.InputSeparation
	if s.Scope == "" || s.UniqueInputs < 0 || s.DuplicateRows < 0 || s.OverlappingInputs < 0 ||
		s.DisjointInputs < 0 || s.UnknownInputs < 0 || s.DisjointCasesPassed < 0 || s.DisjointCasesPassed > s.DisjointInputs ||
		s.UniqueInputs != s.OverlappingInputs+s.DisjointInputs+s.UnknownInputs {
		return [][2]string{{"Constructed input history", "inconsistent recorded counts; inspect JSON"}}
	}
	return [][2]string{
		{"Constructed input history", reportedValue(s.Status) + ": " + reportedValue(s.Reason)},
		{"Actual constructed inputs", fmt.Sprintf("%d unique root tuples: %d overlapping, %d disjoint, %d unknown; %d duplicate rows",
			s.UniqueInputs, s.OverlappingInputs, s.DisjointInputs, s.UnknownInputs, s.DuplicateRows)},
		{"Disjoint cases matching every expectation", constructionShare(s.DisjointCasesPassed, s.DisjointInputs)},
	}
}
