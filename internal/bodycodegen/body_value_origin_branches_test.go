package bodycodegen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOriginBranchAndCollectionLinks8123(t *testing.T) {
	cases := []struct {
		name, body, condition string
		want, absent          []string
	}{
		{"branch", "let result = 0; if input >= 59 { result = input + 61 } else { result = input - 67 }; if result > 991 { return result }; return input", "result > 991", []string{"input + 61", "input - 67", "input >= 59"}, nil},
		{"shadow", "let 합계 = input + 71; if input > 73 { let 합계 = input + 79; if 합계 > 83 { return 합계 } }; if 합계 > 89 { return 합계 }; return 0", "합계 > 89", []string{"input + 71"}, []string{"input + 79"}},
		{"collection", "let 합계 = 0; for 항목 in input.amounts { if 항목 > 97 { 합계 = 합계 + 항목 } }; if 합계 > 101 { return 합계 }; return 0", "합계 > 101", []string{"input.amounts", "항목 > 97", "합계 + 항목"}, nil},
	}
	var records []map[string]any
	for _, c := range cases {
		flow := InspectBodyValueOrigins(c.body, 1)
		at := strings.Index(c.body, c.condition)
		link := flow.conditionSlice(FillOriginRange{at, at + len(c.condition)}, flow.BodySHA256)
		if link == nil || flow.Status == "UNMAPPED" {
			t.Fatalf("%s missing condition lineage: %+v", c.name, flow)
		}
		seen := map[string]bool{}
		for _, node := range link.Nodes {
			if node.Decoded.Start < 0 || node.Decoded.End > len(c.body) || node.Decoded.Start > node.Decoded.End {
				t.Fatalf("%s original range: %+v", c.name, node)
			}
			seen[c.body[node.Decoded.Start:node.Decoded.End]] = true
		}
		for _, expression := range c.want {
			if !seen[expression] {
				t.Fatalf("%s missing %q", c.name, expression)
			}
		}
		for _, expression := range c.absent {
			if seen[expression] {
				t.Fatalf("%s leaked %q", c.name, expression)
			}
		}
		if c.name == "collection" && flow.Status != "PARTIAL" {
			t.Fatal("collection static scope lost partial status")
		}
		records = append(records, map[string]any{"name": c.name, "body": c.body, "flow": flow, "condition": link})
	}
	if directory := os.Getenv("GOOO_CARDINALITY_OBSERVATION"); directory != "" {
		raw, err := json.MarshalIndent(records, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, "origins8123.json"), raw, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("8123: three new bodies retain branch alternatives, lexical shadow boundaries, original Korean byte ranges and collection input dependencies; native0/model0")
}
