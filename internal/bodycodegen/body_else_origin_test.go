package bodycodegen

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

func TestElsePolicyRetainsItsGeneratedSourceOrigin(t *testing.T) {
	raw, err := os.ReadFile("../meta/bodysurface/generation-origin.json")
	if err != nil {
		t.Fatal(err)
	}
	var origin map[string]string
	if err := json.Unmarshal(raw, &origin); err != nil {
		t.Fatal(err)
	}
	for file, wanted := range map[string]string{
		"../meta/bodysurface/else.gooo": origin["source_sha256"],
		"else_policy_generated.go":      origin["generated_sha256"],
	} {
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(source)
		if hex.EncodeToString(digest[:]) != wanted {
			t.Fatal("named source facts lost their generated implementation", file)
		}
	}
}

func TestElseLayoutKeepsTheOriginalGuardCoordinates(t *testing.T) {
	source, err := os.ReadFile("../../examples/else-layout/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	report, err := InspectTypedValueOrigins("main.gooo", source, "Adjust")
	if err != nil || report.Status != "RESOLVED" || report.Flow == nil {
		t.Fatal("multiline else source inspection", err, report.Reason)
	}
	if len(report.Sources) != len(report.Flow.Nodes) {
		t.Fatal("source rows are not aligned with the graph")
	}
	guards := 0
	for i, node := range report.Flow.Nodes {
		if node.Kind != "guard" {
			continue
		}
		row := report.Sources[i]
		if row.Status != "MAPPED" || row.SourceSHA256 != report.SourceSHA256 ||
			row.Start < 0 || row.End < row.Start || row.End > len(source) ||
			row.Expression != string(source[row.Start:row.End]) ||
			(row.Expression != "input.enabled" && row.Expression != "조정값 > 41") {
			t.Fatal("guard lost its authored condition", node.ID, row)
		}
		guards++
	}
	if guards != 4 {
		t.Fatal("original two branch conditions lost a source guard", guards)
	}
}
