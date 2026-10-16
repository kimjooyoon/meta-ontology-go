package policyprojection

import (
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func sourceActivity(filename string, source []byte) (string, error) {
	file, diagnostics := bodycodegen.ParseBodyFile(filename, source)
	if file == nil || diagnostics.HasErrors() {
		return "", fmt.Errorf("policy source syntax: %v", diagnostics)
	}
	var selected *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || activity.ID != ActivityID {
			continue
		}
		if selected != nil {
			return "", fmt.Errorf("policy activity ID is duplicated")
		}
		selected = activity
	}
	if selected == nil || selected.Assembly == nil || len(selected.Inputs) != 2 ||
		len(selected.Assembly.Spec.Choices) != 0 || selected.Assembly.Spec.MaxAttempts != 1 ||
		len(selected.Assembly.Spec.ValueCases) != 4 || len(selected.Assembly.Spec.ValueHoldoutCases) != 0 {
		return "", fmt.Errorf("policy requires one fixed source activity and all four Boolean inputs")
	}
	var seen uint8
	for _, row := range selected.Assembly.Spec.ValueCases {
		var inputs []json.RawMessage
		if json.Unmarshal([]byte(row.Inputs), &inputs) != nil || len(inputs) != 2 {
			return "", fmt.Errorf("policy cases require two Boolean inputs")
		}
		mask := uint8(0)
		for index, input := range inputs {
			switch string(input) {
			case "true":
				mask |= 1 << index
			case "false":
			default:
				return "", fmt.Errorf("policy case input is not Boolean")
			}
		}
		if seen&(1<<mask) != 0 {
			return "", fmt.Errorf("policy case input is duplicated")
		}
		seen |= 1 << mask
	}
	if seen != 15 {
		return "", fmt.Errorf("policy Boolean input domain is incomplete")
	}
	return selected.Name, nil
}
