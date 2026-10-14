package bodyexecution

import "fmt"

func validateFieldBinding(binding CompositionFieldBinding) error {
	if binding.ExpressionKind != "" {
		if binding.ExpressionKind != "return" || binding.ExpressionOrdinal < 0 || binding.VariableOrdinal < 0 || binding.InputKind != "" || binding.InputPort != "" || binding.Declaration == nil || binding.Name != fmt.Sprintf("<return:%d>", binding.ExpressionOrdinal) {
			return fmt.Errorf("field return-expression binding differs")
		}
	} else if binding.ExpressionOrdinal != 0 {
		return fmt.Errorf("field local/input claims an expression ordinal")
	}
	if binding.VariableOrdinal < 0 {
		if binding.InputPort == "" {
			return fmt.Errorf("field input binding has no port")
		}
		switch binding.InputKind {
		case "root_input":
			if binding.ActivityID != binding.RootActivityID || binding.InputFrom != "" {
				return fmt.Errorf("field root input identity differs")
			}
		case "activity_input":
			if binding.ActivityID != binding.RootActivityID || binding.InputFrom == "" {
				return fmt.Errorf("field activity input identity differs")
			}
		case "helper_parameter":
			if binding.InputFrom != "" {
				return fmt.Errorf("field helper parameter has a graph producer")
			}
		default:
			return fmt.Errorf("field input kind differs")
		}
	} else if binding.InputKind != "" || binding.InputPort != "" || binding.InputFrom != "" {
		return fmt.Errorf("field local binding claims an input port")
	}
	fields := map[string]bool{}
	for _, field := range binding.Record.Fields {
		if field.ID == "" || field.GoName == "" || fields[field.ID] || field.Presence != "" || field.Cardinality != "" {
			return fmt.Errorf("field record shape or scalar profile differs")
		}
		fields[field.ID] = true
		switch field.TypeID {
		case "urn:gooo:type:integer", "urn:gooo:type:boolean", "urn:gooo:type:string":
		default:
			return fmt.Errorf("field record type is outside required scalar profile")
		}
	}
	return nil
}
