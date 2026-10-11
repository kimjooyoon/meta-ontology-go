package toolchainrelease

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func validateHelperFillPasses(history *bodycodegen.LocalCallHistory, helperSHA string, caller bool) error {
	phases := []string{"fill-training", "fill-training", "fill-probe", "fill-selected-training", "fill-selected-holdout"}
	inputs := [][]int64{{2, 0}, {-2, -4}, {-5, -1}, {-2, -4}, {-3}}
	if caller {
		phases, inputs = []string{"fill-caller-training", "fill-caller-holdout"}, [][]int64{{-2, -4}, {-3}}
	}
	if history == nil || history.Schema != "gooo/local-call-inputs/v1" || history.Status != "RECORDED" ||
		len(history.Evaluations) != len(phases) {
		return fmt.Errorf("helper fill evaluator history is missing or partial")
	}
	for i, pass := range history.Evaluations {
		completed, candidate, count := true, "works", len(inputs[i])
		if !caller && i == 0 {
			completed, candidate, count = false, "fails", 1
		}
		if pass == nil || pass.Schema != history.Schema || pass.Status != "RECORDED" ||
			pass.ActivityID != "fillhistory://activity/build" || pass.GeneratedSHA256 == "" || pass.SuiteSHA256 == "" ||
			pass.Phase != phases[i] || pass.CandidateID != candidate || pass.EvaluationCompleted != completed ||
			pass.CasesStarted != len(inputs[i]) || pass.CasesCompleted != count || len(pass.Calls) != len(inputs[i]) ||
			completed && pass.EvaluationError != "" || !completed && !strings.Contains(pass.EvaluationError, "zero divisor") {
			return fmt.Errorf("helper fill evaluator pass %d changed its failure or completion", i)
		}
		for j, call := range pass.Calls {
			if call.CaseIndex != j || call.ActivityID != "fillhistory://activity/seed" || call.ProgramSHA256 != helperSHA ||
				!samePackageSourceValue(call.Inputs, []byte(fmt.Sprintf("[%d]", inputs[i][j]))) {
				return fmt.Errorf("helper fill evaluator pass %d changed helper input %d", i, j)
			}
		}
	}
	return nil
}
