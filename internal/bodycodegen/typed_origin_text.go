package bodycodegen

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

// WriteTypedOriginText projects the existing source graph without evaluating it.
func WriteTypedOriginText(w io.Writer, report TypedValueOrigins, source []byte) error {
	if digest(source) != report.SourceSHA256 {
		return fmt.Errorf("origin display source differs")
	}
	if _, err := fmt.Fprintf(w, "Activity: %s\nSource: %s\nSource digest: %s\nStatus: %s\nScope: possible source dependencies\n", report.ActivityID, report.Filename, report.SourceSHA256, report.Status); err != nil {
		return err
	}
	if report.Status != "RESOLVED" {
		_, err := fmt.Fprintf(w, "Reason: %s\n", report.Reason)
		return err
	}
	if report.Flow == nil || len(report.Flow.Nodes) > recordFlowNodeLimit || len(report.Sources) != len(report.Flow.Nodes) {
		return fmt.Errorf("origin display graph extent differs")
	}
	for _, output := range report.Outputs {
		name := output.FieldID
		if name == "" {
			name = "return"
		}
		if _, err := fmt.Fprintf(w, "%s <- #%d\n", name, output.Root); err != nil {
			return err
		}
	}
	for i, node := range report.Flow.Nodes {
		if err := writeOriginTextNode(w, report, node, report.Sources[i], source); err != nil {
			return err
		}
	}
	return nil
}

func writeOriginTextNode(w io.Writer, report TypedValueOrigins, node RecordFlowNode, location BodyExpressionSource, source []byte) error {
	if _, err := fmt.Fprintf(w, "\n#%d %s [%s]", node.ID, OriginRole(node.Kind), node.Kind); err != nil {
		return err
	}
	for _, input := range report.Inputs {
		if input.Node == node.ID {
			if _, err := fmt.Fprintf(w, " %s", input.Port); err != nil {
				return err
			}
		}
	}
	if node.FieldID != "" {
		if _, err := fmt.Fprintf(w, " field=%s", node.FieldID); err != nil {
			return err
		}
	}
	if err := writeOriginTextFacts(w, node); err != nil {
		return err
	}
	if err := writeOriginTextLocation(w, location, source); err != nil {
		return err
	}
	for _, parent := range node.Parents {
		if parent != 0 {
			if _, err := fmt.Fprintf(w, "  data <- #%d\n", parent); err != nil {
				return err
			}
		}
	}
	if node.Guard != 0 {
		if _, err := fmt.Fprintf(w, "  control <- #%d\n", node.Guard); err != nil {
			return err
		}
	}
	if node.Condition != 0 {
		if _, err := fmt.Fprintf(w, "  condition <- #%d\n", node.Condition); err != nil {
			return err
		}
	}
	return nil
}

func writeOriginTextFacts(w io.Writer, node RecordFlowNode) error {
	if node.Operator != "" {
		label := "operator"
		if node.Kind == "guard" && (node.Operator == "true" || node.Operator == "false") {
			label = "when"
		}
		if _, err := fmt.Fprintf(w, " %s=%s", label, strconv.Quote(node.Operator)); err != nil {
			return err
		}
	}
	if node.KnownTruth != "" {
		_, err := fmt.Fprintf(w, " constant-condition=%s", strconv.Quote(node.KnownTruth))
		return err
	}
	return nil
}

func writeOriginTextLocation(w io.Writer, location BodyExpressionSource, source []byte) error {
	if location.Status != "MAPPED" {
		_, err := fmt.Fprintf(w, "\n  source location: %s\n", location.Reason)
		return err
	}
	if location.Start < 0 || location.End <= location.Start || location.End > len(source) {
		return fmt.Errorf("origin display location differs")
	}
	line := bytes.Count(source[:location.Start], []byte("\n")) + 1
	begin := bytes.LastIndexByte(source[:location.Start], '\n') + 1
	column := utf8.RuneCount(source[begin:location.Start]) + 1
	expression := strings.ReplaceAll(location.Expression, "\n", " ")
	count := 0
	for i := range expression {
		if count == 160 {
			expression = expression[:i] + "…"
			break
		}
		count++
	}
	_, err := fmt.Fprintf(w, "\n  at %d:%d %s\n", line, column, strconv.Quote(expression))
	return err
}
