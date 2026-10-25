package main

import (
	"encoding/json"
	"strings"
	"testing"

	completion "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcecompletion"
)

func pendingSplitReceipt(t *testing.T) splitReport {
	t.Helper()
	raw := `{"schema":"gooo.logical-source-split.v1","source_sha":"abc",` +
		`"decision":"PROGRESS","resolution":"LOGICAL_ONLY","reason":"SUCCESSOR_PROJECTION_REQUIRED",` +
		`"next_operation":"project-generated-paths","policy_source_sha256":"` + completion.SourceDigest() + `",` +
		`"plan_source_sha256":"sha256:` + strings.Repeat("a", 64) + `","exact":true,"unhandled":["pending.go"],` +
		`"subjects":[{"changed_files":["d.go","d_split02.go"],"created_files":["d_split02.go"]}],` +
		`"coordinates":{"input_subjects":2,"unhandled_subjects":1,"selected_subjects":1,"applied_subjects":1,` +
		`"changed_paths":2,"created_paths":1,"unknowns":0,"deferred_topology_subjects":1}}`
	var report splitReport
	if err := json.Unmarshal([]byte(raw), &report); err != nil {
		t.Fatal(err)
	}
	return report
}

func TestWriteSetKeepsLogicalPathsExactWhileProjectionIsPending(t *testing.T) {
	density, extraction, _ := exactReports()
	split := pendingSplitReceipt(t)
	observed := []string{"a.go", "c.go", "d.go", "d_split02.go"}
	report := reduce("abc", density, extraction, split, observed, []string{"d_split02.go"})
	if report.Decision != "PASS" || !report.Exact || !equalPaths(report.Expected, observed) {
		t.Fatal("logical write-set verification confused pending construction with unknown writes", report.Reason)
	}
	if split.Decision != "PROGRESS" || split.Resolution != "LOGICAL_ONLY" {
		t.Fatal("path verification changed the unfinished source outcome")
	}
}
