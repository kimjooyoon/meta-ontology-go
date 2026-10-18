package bodycodegen

import (
	"fmt"
	"strings"
	"testing"
)

func TestPureActivityDefinitionLimitCountsEntry(t *testing.T) {
	for _, helpers := range []int{31, 32} {
		t.Run(fmt.Sprint(helpers), func(t *testing.T) {
			var source strings.Builder
			source.WriteString("package limits\nnamespace limits\nentity Integer id \"urn:gooo:type:integer\"\n")
			calls := make([]string, helpers)
			for i := range helpers {
				fmt.Fprintf(&source, "activity F%d(Integer) -> Integer id \"urn:limit:f%d\" computes `return input`\n", i, i)
				calls[i] = fmt.Sprintf("F%d(input)", i)
			}
			fmt.Fprintf(&source, "activity Run(Integer) -> Integer id \"urn:limit:run\" computes `return %s`\n", strings.Join(calls, "+"))
			generated, err := Generate("limits.gooo", []byte(source.String()), "Run")
			if helpers == 31 {
				if err != nil || generated.Report.CallClosure == nil || len(generated.Report.CallClosure.Activities) != 31 ||
					CurrentProfileLimits().ActivitiesPerCallClosure != 32 {
					t.Fatal("entry plus31 helpers must fit the reported32-definition bound", err, generated.Report.CallClosure)
				}
			} else if err == nil || !strings.Contains(err.Error(), "32 activities") || generated.Source != "" {
				t.Fatal("entry plus32 helpers escaped the definition bound", err)
			}
		})
	}
}
