package bodycodegen

import (
	"encoding/json"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
)

func recordPlanModelContext(p recordAssemblyPlan, version string) *RecordOrdinalContext {
	if version == jointdecision.RecordOriginSharedFeatureVersion {
		return recordOriginContext(p.choices, recordValueFlow(p))
	}
	return recordModelContext(p.choices, version)
}

func recordOriginContext(choices []RecordValueChoice, flow *RecordValueFlow) *RecordOrdinalContext {
	r := &RecordOrdinalContext{Schema: "gooo/record-field-origin-context/v2", Status: "ENCODED",
		FeatureVersion: jointdecision.RecordOriginSharedFeatureVersion,
		Scope:          "complete expressions and intent with bounded unique typed source ancestors; conditional guards retained; input values and finite cases excluded"}
	rawFlow, _ := json.Marshal(flow)
	r.ValueFlowSHA256 = digest(rawFlow)
	if flow == nil || flow.Status != "RESOLVED" {
		r.Status, r.Reason = "DECLINED_TO_DETERMINISTIC", "SOURCE_VALUE_FLOW_UNRESOLVED"
		if flow != nil {
			r.Reason += ":" + flow.Reason
		}
		return r
	}
	input := recordOriginParts(choices, flow, r)
	if len(choices) != len(input) {
		r.Status, r.Reason = "DECLINED_TO_DETERMINISTIC", "THREE_FIELD_CHOICES_REQUIRED"
		return r
	}
	text, err := jointdecision.EncodeRecordOriginThree(input)
	if err != nil {
		r.Status, r.Reason = "DECLINED_TO_DETERMINISTIC", "COMPLETE_FIELD_ORIGIN_CONTEXT_EXCEEDS_MODEL_BOUND"
		return r
	}
	r.Text, r.SHA256 = text, digest([]byte(text))
	return r
}

func recordOriginParts(choices []RecordValueChoice, flow *RecordValueFlow, r *RecordOrdinalContext) [3]jointdecision.RecordOriginChoice {
	var input [3]jointdecision.RecordOriginChoice
	for i, choice := range choices {
		part := jointdecision.RecordOriginChoice{
			Field: choice.Field, First: choice.First, Second: choice.Second, Intent: choice.Intent}
		part.Origins[0] = recordOriginCounts(flow, flow.Choices[i].First, choice.FieldID)
		part.Origins[1] = recordOriginCounts(flow, flow.Choices[i].Second, choice.FieldID)
		raw, _ := json.Marshal(part)
		r.Parts = append(r.Parts, string(raw))
		if i < len(input) {
			input[i] = part
		}
	}
	return input
}

func recordOriginCounts(flow *RecordValueFlow, root uint16, fieldID string) [16]uint16 {
	var counts [16]uint16
	var seen [recordFlowNodeLimit]bool
	var stack [recordFlowNodeLimit]uint16
	count := 0
	push := func(id uint16) {
		if id != 0 && !seen[id-1] {
			seen[id-1] = true
			stack[count] = id
			count++
		}
	}
	push(root)
	for count != 0 {
		count--
		node := flow.Nodes[stack[count]-1]
		counts[recordOriginSlot(node, fieldID)]++
		for _, id := range [4]uint16{node.Parents[0], node.Parents[1], node.Guard, node.Condition} {
			push(id)
		}
	}
	return counts
}

func recordOriginSlot(node RecordFlowNode, fieldID string) int {
	role := 1
	if node.FieldID == "" {
		role = 2
	} else if node.FieldID == fieldID {
		role = 0
	}
	switch node.Kind {
	case "input":
		return role
	case "literal":
		return 3
	case "expression":
		return 4
	case "copy":
		return 5 + role
	case "write":
		return 8 + role
	case "choice":
		return 11 + min(role, 1)
	case "join":
		return 13
	case "guard", "guard_join":
		return 14
	default:
		return 15
	}
}
