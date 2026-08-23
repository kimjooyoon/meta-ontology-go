package guardedcapability

import (
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedpromotion"
)

func TestEmbeddedFoundationReplaysBySection(t *testing.T) {
	report, err := foundationReport()
	if err != nil {
		t.Fatal(err)
	}
	replay := guardedpromotion.Build(report.Source)
	sections := []struct {
		name       string
		foundation any
		replayed   any
	}{
		{name: "header", foundation: []string{report.Schema, report.Decision, report.Reason, report.Resolution}, replayed: []string{replay.Schema, replay.Decision, replay.Reason, replay.Resolution}},
		{name: "summary", foundation: report.Summary, replayed: replay.Summary},
		{name: "coordinates", foundation: report.Coordinates, replayed: replay.Coordinates},
		{name: "indicators", foundation: report.Indicators, replayed: replay.Indicators},
		{name: "proofs", foundation: report.Proofs, replayed: replay.Proofs},
		{name: "digest", foundation: report.ReportDigest, replayed: replay.ReportDigest},
	}
	for _, section := range sections {
		if !reflect.DeepEqual(section.foundation, section.replayed) {
			t.Fatalf("foundation %s does not replay: foundation=%#v replayed=%#v", section.name, section.foundation, section.replayed)
		}
	}
}
