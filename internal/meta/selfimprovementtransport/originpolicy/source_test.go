package originpolicy

import (
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestOriginProjectionFresh1610(t *testing.T) {
	projection, err := bodycodegen.Generate("source.gooo", source, "BindProducerRun")
	if err != nil {
		t.Fatal(err)
	}
	checked, err := os.ReadFile("origin.generated.go")
	if err != nil || string(checked) != projection.Source {
		t.Fatalf("checked projection must match Gooo source: %v", err)
	}
	changed := strings.Replace(string(source), `workflow != run`, `subject != run`, 1)
	revised, err := bodycodegen.Generate("source.gooo", []byte(changed), "BindProducerRun")
	if err != nil || revised.Source == projection.Source ||
		strings.Contains(revised.Source, "workflow != run") ||
		revised.Report.SourceDigest == projection.Report.SourceDigest {
		t.Fatalf("source relation revision was not projected: %v", err)
	}
}

func TestOriginPolicyPartialInputsFresh1610(t *testing.T) {
	cases := []struct{ event, subject, workflow, run, reason string }{
		{"", "s", "w", "r", "PRODUCER_EVENT_UNKNOWN"},
		{"push", "s", "w", "", "PRODUCER_RUN_ORIGIN_UNKNOWN"},
		{"workflow_run", "s", "", "r", "PRODUCER_WORKFLOW_ORIGIN_UNKNOWN"},
		{"pull_request", "", "w", "r", "PRODUCER_SUBJECT_ORIGIN_UNKNOWN"},
		{"schedule", "s", "w", "r", "PRODUCER_EVENT_UNSUPPORTED"},
	}
	for _, row := range cases {
		if got := BindProducerRun(row.event, row.subject, row.workflow, row.run); got != row.reason {
			t.Errorf("%+v: got %s", row, got)
		}
	}
}
