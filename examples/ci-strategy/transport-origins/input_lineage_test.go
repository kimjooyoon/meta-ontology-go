package transportorigins

import (
	"fmt"
	"strings"
	"testing"
)

func TestSavedDecisionOriginsIncludeBothSourceInputs(t *testing.T) {
	for _, apply := range []bool{false, true} {
		t.Run(fmt.Sprint(apply), func(t *testing.T) {
			_, raw, wire := runSavedReport(t, apply, 201, true, false)
			requests := wire.requests
			view, err := FromSavedReport(raw)
			if err != nil || view.F_status != "RECORDED_REPORT_ORIGINS" {
				t.Fatal(err, view)
			}
			base := "<urn:gooo:transport-report:" + contentDigest(raw)
			scope, before := base+":scope-record>", base+":before-record>"
			selected, next := base+":decision-record>", base+":next-record>"
			text := strings.Join(view.F_lines, "\n")
			for _, required := range []string{
				scope + " <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/ns/prov#Entity> .",
				scope + " <gooo://transport-origins/recordedIn> " + base + "> .",
				scope + " <gooo://transport-origins/recordPath> \"/scope\" .",
				before + " <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/ns/prov#Entity> .",
				before + " <gooo://transport-origins/recordedIn> " + base + "> .",
				before + " <gooo://transport-origins/recordPath> \"/before\" .",
				selected + " <http://www.w3.org/ns/prov#wasDerivedFrom> " + scope + " .",
				selected + " <http://www.w3.org/ns/prov#wasDerivedFrom> " + before + " .",
			} {
				if strings.Count(text, required) != 1 {
					t.Fatal("missing or repeated input origin", required, text)
				}
			}
			readbackScope := next + " <http://www.w3.org/ns/prov#wasDerivedFrom> " + scope + " ."
			if strings.Contains(text, readbackScope) != apply {
				t.Fatal("scope dependency must follow the recorded readback selection", text)
			}
			if wire.requests != requests || strings.Contains(text, "#Activity>") || strings.Contains(text, "#wasGeneratedBy>") {
				t.Fatal("content input origins changed execution semantics", text)
			}
		})
	}
}
