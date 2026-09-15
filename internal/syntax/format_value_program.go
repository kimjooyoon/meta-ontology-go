package syntax

import "strings"

func formatActivityValueProgram(output *strings.Builder, activity *ActivityDecl) error {
	if activity.ValueProgramPresent || activity.ValueProgram != "" {
		output.WriteString(" computes ")
		output.WriteString(quoteString(activity.ValueProgram))
	}
	return nil
}
