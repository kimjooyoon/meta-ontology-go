package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

// Rendering consumes the existing result only; no selection, model or execution
// is added. JSON remains the default and keeps its original wire shape.
func writeBodyConstructResult(out io.Writer, result bodyConstructOutput, format string) error {
	if format == "" || format == "json" {
		return json.NewEncoder(out).Encode(result)
	}
	if format != "text" && format != "markdown" {
		return fmt.Errorf("body-construct format must be json, text or markdown")
	}
	var text strings.Builder
	if format == "markdown" {
		fmt.Fprint(&text, "# Gooo construction result\n\n| Observation | Result |\n| --- | --- |\n")
	} else {
		fmt.Fprintln(&text, "Gooo construction result")
	}
	for _, row := range constructionReportRows(result) {
		if format == "markdown" {
			fmt.Fprintf(&text, "| %s | %s |\n", row[0], outcomeMarkdownCell(row[1]))
		} else {
			fmt.Fprintf(&text, "%s: %s\n", row[0], strconv.Quote(row[1]))
		}
	}
	fmt.Fprint(&text, "\nShares count the named supplied checks or expected activity outputs, not all possible inputs.\n")
	fmt.Fprint(&text, "Input overlap counts caller root tuples only; local-example and model-training exposure are unknown.\n")
	fmt.Fprint(&text, "Candidate uniqueness and behavior on other inputs are not assessed by this report.\n")
	fmt.Fprint(&text, "Use --format json for all attempts, values and evidence. Saved --out files remain full JSON and source.\n")
	_, err := io.WriteString(out, text.String())
	return err
}

func constructionReportRows(result bodyConstructOutput) [][2]string {
	c, e := result.Construction, result.Evaluation
	mode := "saved construction replay"
	if result.GeneratedNow {
		mode = "new construction"
	}
	rows := [][2]string{{"Mode", mode}, {"Entry activity", reportedValue(c.Initial.Plan.EntryActivity)},
		{"Construction stage", reportedValue(c.Stage)},
		{"Construction decision", reportedValue(c.Decision)}, {"Stop reason", reportedValue(c.StopReason)},
		{"Attempted programs / budget", fmt.Sprintf("%d / %d", len(c.Attempts), c.ProgramBudget)},
		{"Source-bounded candidate combinations", reportedValue(c.CandidateSpace)}}
	rows = append(rows, constructionSelectionRows(c)...)
	rows = append(rows, [2]string{"Evaluation stage", reportedValue(e.Runtime.Stage)},
		[2]string{"Evaluation expected outputs", constructionRuntimeShare(e.Runtime)})
	if e.InputSeparation.Scope == "" {
		rows = append(rows, [2]string{"Evaluation input overlap", "not measured"})
	} else {
		s := e.InputSeparation
		rows = append(rows, [2]string{"Evaluation input overlap", fmt.Sprintf("%d unique tuples: %d consumed during caller construction, %d other; %d duplicate rows", s.UniqueInputs, s.ConstructionInputs, s.OtherInputs, s.DuplicateRows)})
	}
	if !result.GeneratedNow {
		rows = append(rows, [2]string{"Saved history replayed", strconv.FormatBool(e.ConstructionReplayed)},
			[2]string{"New model calls during replay/evaluation", strconv.Itoa(e.NewModelCalls)})
	}
	rows = append(rows, constructionFailureRows(result)...)
	return rows
}

func constructionSelectionRows(c bodyexecution.JointConstruction) [][2]string {
	rows := [][2]string{}
	rejected := 0
	for _, attempt := range c.Attempts {
		if attempt.Rejection != nil {
			rejected++
		}
	}
	rows = append(rows, [2]string{"Rejected before caller execution", strconv.Itoa(rejected)})
	if c.SelectedAttempt < 0 || c.SelectedAttempt >= len(c.Attempts) {
		return append(rows, [2]string{"Selected attempt", "none recorded"})
	}
	a := c.Attempts[c.SelectedAttempt]
	rows = append(rows, [2]string{"Selected attempt", strconv.Itoa(c.SelectedAttempt + 1)},
		[2]string{"Selected local checks", constructionShare(a.LocalPassed, a.LocalTotal)},
		[2]string{"Caller expected outputs used for selection", constructionRuntimeShare(a.Runtime)})
	if len(c.Attempts) > 1 {
		rows = append(rows, [2]string{"First attempt caller expected outputs", constructionRuntimeShare(c.Attempts[0].Runtime)})
	}
	return rows
}

func constructionShare(passed, total int) string {
	if total <= 0 {
		return "not measured (no scored checks)"
	}
	if passed < 0 || passed > total {
		return "inconsistent recorded counts; inspect JSON"
	}
	return fmt.Sprintf("%d/%d (%.2f%%)", passed, total, 100*float64(passed)/float64(total))
}

func reportedValue(value string) string {
	if value == "" {
		return "not recorded"
	}
	return value
}
