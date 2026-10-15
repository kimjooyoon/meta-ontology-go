package bodycodegen

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

func recordedSourceControlReport(t *testing.T) TypedValueOrigins {
	t.Helper()
	f, err := os.Open("testdata/source-control-report.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	var report TypedValueOrigins
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	return report
}

func TestSourceControlColumnsRetainIRFacts(t *testing.T) {
	report := recordedSourceControlReport(t)
	raw, err := TypedSourceControlInputs(report, "source/control:Inspect")
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Inputs []map[string]SourceControlColumns `json:"inputs"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil || len(envelope.Inputs) != 1 {
		t.Fatal(err)
	}
	c := envelope.Inputs[0]["source/control:Inspect"]
	if c.Schema != "gooo/source-control-columns/v1" || c.Scope != "retained_source_graph" || c.Count != 25 ||
		c.SourceSHA256 != report.SourceSHA256 || c.ActivityID != report.ActivityID || !strings.HasPrefix(c.GraphSHA256, "sha256:") {
		t.Fatal("source graph identity", c)
	}
	for i, node := range report.Flow.Nodes {
		if c.IDs[i] != node.ID || c.Bits[i] != int64(1)<<i || c.ParentA[i] != node.Parents[0] || c.ParentB[i] != node.Parents[1] ||
			c.Guards[i] != node.Guard || c.Conditions[i] != node.Condition || c.Kinds[i] != node.Kind || c.Operators[i] != node.Operator ||
			c.KnownTruth[i] != node.KnownTruth || c.NodeActivities[i] != node.Span.ActivityID || c.FieldIDs[i] != node.FieldID {
			t.Fatal("source node changed", i, node)
		}
	}
	if c.KnownTruth[4] != "false" || c.Operators[5] != "true" || c.Operators[6] != "false" || c.NodeActivities[5] != "urn:guardmetadata:adjust" {
		t.Fatal("static condition/polarity provenance missing", c)
	}
	prior := c.GraphSHA256
	report.Flow.Nodes[5].Operator = "false"
	changed, err := TypedSourceControlInputs(report, "source/control:Inspect")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(changed, &envelope); err != nil || envelope.Inputs[0]["source/control:Inspect"].GraphSHA256 == prior {
		t.Fatal("changed raw IR inherited the old graph identity", err)
	}
}

func TestSourceControlExportKeepsLegacyWire(t *testing.T) {
	report := recordedSourceControlReport(t)
	raw, err := TypedValueOriginInputs(report, "guard/reader:Inspect", "", true)
	want, readErr := os.ReadFile("testdata/source-control-legacy-inputs.json")
	if err != nil || readErr != nil || !bytes.Equal(append(raw, '\n'), want) {
		t.Fatal("legacy input export changed", err, readErr)
	}
}

func TestSourceControlExportRejectsInvalidBounds(t *testing.T) {
	for _, kind := range []string{"missing", "unresolved", "oversized", "forward", "target"} {
		r := recordedSourceControlReport(t)
		target := "source/control:Inspect"
		switch kind {
		case "missing":
			r.Flow = nil
		case "unresolved":
			r.Flow.Status = "UNRESOLVED"
		case "oversized":
			r.Flow.Nodes = make([]RecordFlowNode, 61)
		case "forward":
			r.Flow.Nodes[0].Parents[0] = 1
		case "target":
			target = "not-a-consumer"
		}
		if _, err := TypedSourceControlInputs(r, target); err == nil {
			t.Fatal("invalid projection accepted", kind)
		}
	}
}
