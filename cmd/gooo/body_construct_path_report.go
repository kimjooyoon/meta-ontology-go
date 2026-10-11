package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

// These rows read the original local history, separately from caller attempts.
// Saved replay retains this history; none of its counters are new model calls.
func constructionPathRows(c bodyexecution.JointConstruction) [][2]string {
	var rows [][2]string
	shown, omitted := 0, 0
	for _, step := range c.Initial.ConstructionSteps() {
		p := step.Generation.Report.BodyPaths
		if p == nil {
			continue
		}
		if shown == 8 {
			omitted++
			continue
		}
		shown++
		name := reportedValue(step.Generation.Report.Activity)
		rows = append(rows, [2]string{"Initial typed path body", name})
		rows = append(rows, constructionInitialPathRows(p)...)
		selected := constructionSelectedPath(c, step.Generation.Report.ActivityID)
		rows = append(rows, [2]string{"Selected caller-program body checks", constructionSelectedPathShare(selected)})
	}
	if omitted > 0 {
		rows = append(rows, [2]string{"Additional initial typed path bodies", fmt.Sprintf("%d; use --format json", omitted)})
	}
	return rows
}

func constructionInitialPathRows(p *bodycodegen.BodyPathReceipt) [][2]string {
	if p.Search.Schema == "" {
		return [][2]string{{"Initial local path search", "not recorded for this path mode"}}
	}
	s := p.Search
	rows := [][2]string{{"Initial local path search", fmt.Sprintf("%d recorded candidates; %d recorded local model calls", len(s.Attempts), s.Selection.ModelCalls)}}
	if len(s.Attempts) == 0 {
		return append(rows, [2]string{"First local candidate", "none recorded"})
	}
	first := s.Attempts[0]
	rows = append(rows, [2]string{"First local candidate", fmt.Sprintf("mask %d; %s", first.Mask, reportedValue(first.Status))},
		[2]string{"First local candidate output checks", constructionShare(first.Passed, first.Total)},
		[2]string{"First local candidate recorded condition checks", constructionRecordedConditions(first.Conditions)})
	for _, condition := range first.Conditions {
		if !condition.Passed {
			rows = append(rows, [2]string{"First unmet local condition", constructionConditionDetail(condition)})
			break
		}
	}
	return rows
}

func constructionSelectedPath(c bodyexecution.JointConstruction, activityID string) *bodycodegen.PathCandidate {
	if activityID == "" || c.SelectedAttempt < 0 || c.SelectedAttempt >= len(c.Attempts) {
		return nil
	}
	var selected *bodycodegen.PathCandidate
	for i := range c.Attempts[c.SelectedAttempt].PathCandidates {
		candidate := &c.Attempts[c.SelectedAttempt].PathCandidates[i]
		if candidate.ActivityID == activityID {
			if selected != nil {
				return nil // Ambiguous history must not choose an arbitrary row.
			}
			selected = candidate
		}
	}
	return selected
}

func constructionSelectedPathShare(candidate *bodycodegen.PathCandidate) string {
	if candidate == nil {
		return "not uniquely recorded; inspect JSON"
	}
	conditions := "not recorded"
	if c := candidate.Conditions; c != nil {
		conditions = fmt.Sprintf("%s; %d not reached", constructionShare(c.Passed, c.Declared), c.NotReached)
	}
	return fmt.Sprintf("mask %d; outputs %s; conditions %s", candidate.Mask, constructionShare(candidate.Passed, candidate.Total), conditions)
}

func constructionRecordedConditions(conditions []pathplan.ConditionResult) string {
	if len(conditions) == 0 {
		return "not recorded"
	}
	passed, notReached := 0, 0
	for _, c := range conditions {
		if c.Passed {
			if !c.Observation.Reached {
				return "inconsistent recorded condition; inspect JSON"
			}
			passed++
		}
		if !c.Observation.Reached {
			notReached++
		}
	}
	return fmt.Sprintf("%s; %d not reached", constructionShare(passed, len(conditions)), notReached)
}

func constructionConditionDetail(c pathplan.ConditionResult) string {
	actual := "not reached"
	if c.Observation.Reached {
		actual = strconv.FormatBool(c.Observation.Value)
	}
	choice, _ := json.Marshal(c.Case.ChoiceID)
	return fmt.Sprintf("choice %s; input %d; expected %t; observed %s", constructionValue(choice), c.Case.Input, c.Case.Expected, actual)
}
