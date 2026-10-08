package bodycodegen

import (
	"encoding/json"
	"fmt"
	"go/constant"
	"go/token"
	"strconv"
	"strings"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
)

func recordGraphContext(p recordAssemblyPlan, flow *RecordValueFlow) *RecordOrdinalContext {
	r := &RecordOrdinalContext{Schema: "gooo/record-value-graph-context/v3", Status: "ENCODED",
		FeatureVersion: jointdecision.RecordGraphSharedFeatureVersion,
		Scope:          "ordered source value graph, canonical literals, input types and complete intent; source provenance retained; inputs, expected outputs and predictions excluded"}
	raw, _ := json.Marshal(flow)
	r.ValueFlowSHA256 = digest(raw)
	if flow == nil || flow.Status != "RESOLVED" {
		r.Status, r.Reason = "DECLINED_TO_DETERMINISTIC", "SOURCE_VALUE_FLOW_UNRESOLVED"
		if flow != nil {
			r.Reason += ":" + flow.Reason
		}
		return r
	}
	if len(p.choices) != 3 || len(flow.Choices) != 3 {
		r.Status, r.Reason = "DECLINED_TO_DETERMINISTIC", "THREE_FIELD_CHOICES_REQUIRED"
		return r
	}
	input, err := recordGraphInput(p, flow)
	if err == nil {
		r.Text, err = jointdecision.EncodeRecordGraphThree(input)
	}
	if err != nil {
		r.Status, r.Reason = "DECLINED_TO_DETERMINISTIC", "SOURCE_VALUE_GRAPH_UNREPRESENTABLE:"+err.Error()
		return r
	}
	for _, choice := range input.Choices {
		raw, _ := json.Marshal(choice)
		r.Parts = append(r.Parts, string(raw))
	}
	r.SHA256 = digest([]byte(r.Text))
	return r
}

func recordGraphInput(p recordAssemblyPlan, flow *RecordValueFlow) (jointdecision.RecordGraphInput, error) {
	var input jointdecision.RecordGraphInput
	views, err := recordGraphLiteralViews(p)
	if err != nil {
		return input, err
	}
	for _, node := range flow.Nodes {
		n := jointdecision.RecordGraphNode{Kind: node.Kind, Operator: node.Operator, FieldID: node.FieldID,
			Parents: node.Parents, Guard: node.Guard, Condition: node.Condition}
		switch node.Kind {
		case "input":
			n.Input, n.InputType = node.Local, recordGraphInputType(p.body, node)
		case "choice":
			n.Operator = ""
		case "literal":
			n.Operator, n.Literal, err = recordGraphLiteral(node, views)
			if err != nil {
				return input, err
			}
		}
		input.Nodes = append(input.Nodes, n)
	}
	for i, choice := range p.choices {
		input.Choices[i] = jointdecision.RecordGraphChoice{
			Field: choice.Field, First: choice.First, Second: choice.Second, Intent: choice.Intent,
			FieldID: choice.FieldID, Roots: [2]uint16{flow.Choices[i].First, flow.Choices[i].Second}}
	}
	return input, nil
}

func recordGraphInputType(body preparedBody, node RecordFlowNode) string {
	if node.FieldID != "" {
		for _, record := range body.allRecords {
			for _, field := range record.Fields {
				if field.ID == node.FieldID {
					if field.Presence == "optional" {
						return ""
					}
					return recordGoType(field.TypeID)
				}
			}
		}
		return ""
	}
	if node.Local == 0 || int(node.Local) > len(body.parameters) {
		return ""
	}
	return body.parameters[node.Local-1].Type
}

func recordGraphLiteralViews(p recordAssemblyPlan) (map[string]string, error) {
	views := map[string]string{"/computes": p.body.body, "/baseline": p.body.body}
	var closure strings.Builder
	closure.WriteString(p.body.body)
	for _, choice := range p.choices {
		views["/alternative:"+choice.ID] = choice.Second
		closure.WriteString("\n_ = (" + choice.Second + ")")
	}
	functions, _, _, err := p.body.resolvePureCalls(closure.String())
	for _, function := range functions {
		views[function.identity.ActivityID+"/computes"] = function.body
	}
	return views, err
}

func recordGraphLiteral(node RecordFlowNode, views map[string]string) (string, string, error) {
	if node.Operator == "true" || node.Operator == "false" {
		return "BOOL", node.Operator, nil
	}
	if node.Operator == "0" {
		return "INT", "0", nil
	}
	body := views[node.Span.ActivityID+"/"+node.Span.View]
	if node.Span.Start < 0 || node.Span.Start >= node.Span.End || node.Span.End > len(body) {
		return "", "", fmt.Errorf("literal source span is unresolved")
	}
	text := body[node.Span.Start:node.Span.End]
	if node.Operator == "STRING" {
		value, err := strconv.Unquote(text)
		return "STRING", strconv.Quote(value), err
	}
	if node.Operator == "INT" {
		value := constant.MakeFromLiteral(text, token.INT, 0)
		if value.Kind() == constant.Int {
			return "INT", value.ExactString(), nil
		}
	}
	return "", "", fmt.Errorf("source literal is outside the graph model profile")
}
