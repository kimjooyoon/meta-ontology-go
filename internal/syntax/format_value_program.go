package syntax

import (
	"fmt"
	"strings"
)

func formatActivityValueProgram(output *strings.Builder, activity *ActivityDecl) error {
	if activity.ValueProgramPresent || activity.ValueProgram != "" {
		output.WriteString(" computes ")
		output.WriteString(quoteString(activity.ValueProgram))
	}
	if activity.Assembly != nil && (!activity.ValueProgramPresent && activity.ValueProgram == "") {
		return fmt.Errorf("assembling requires an activity computes body")
	}
	return formatAssembly(output, activity.Assembly)
}
