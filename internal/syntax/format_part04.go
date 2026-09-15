package syntax

import (
	"fmt"
	"strings"
)

func formatActivity(output *strings.Builder, activity *ActivityDecl) error {
	if activity == nil {
		return fmt.Errorf("nil activity declaration")
	}
	if err := validateIdentifier(activity.Name, "activity name"); err != nil {
		return err
	}
	parameters, err := activityParameters(activity)
	if err != nil {
		return err
	}
	result, err := activityResult(activity)
	if err != nil {
		return err
	}
	output.WriteString("activity ")
	output.WriteString(activity.Name)
	output.WriteByte('(')
	for index, parameter := range parameters {
		if index > 0 {
			output.WriteString(", ")
		}
		output.WriteString(parameter.Name)
	}
	output.WriteString(") -> ")
	output.WriteString(result)
	return formatActivityValueProgram(output, activity)
}
func activityParameters(activity *ActivityDecl) ([]NameRef, error) {
	if activity.Inputs != nil && activity.Parameters != nil && !sameNames(activity.Inputs, activity.Parameters) {
		return nil, fmt.Errorf("activity %s has conflicting input aliases", activity.Name)
	}
	parameters := activity.Parameters
	if parameters == nil {
		parameters = activity.Inputs
	}
	for _, parameter := range parameters {
		if err := validateIdentifier(parameter.Name, "activity parameter"); err != nil {
			return nil, err
		}
	}
	return parameters, nil
}
func activityResult(activity *ActivityDecl) (string, error) {
	result := activity.Result.Name
	if result != "" && activity.Output != "" && result != activity.Output {
		return "", fmt.Errorf("activity %s has conflicting result aliases", activity.Name)
	}
	if result == "" {
		result = activity.Output
	}
	if err := validateIdentifier(result, "activity result"); err != nil {
		return "", err
	}
	return result, nil
}
func sameNames(left, right []NameRef) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].Name != right[index].Name {
			return false
		}
	}
	return true
}
