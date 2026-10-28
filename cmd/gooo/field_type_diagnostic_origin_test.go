package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestSemanticFieldTypeDiagnosticRetainsTheExactReference(t *testing.T) {
	for _, spelling := range []string{"Missing", `"urn:missing:type"`} {
		t.Run(spelling, func(t *testing.T) {
			source := "package origin\nnamespace origin\n// 한글 앞부분\n" +
				"entity Data id \"urn:type-origin:data\" fields {\n" +
				"field label id \"urn:type-origin:label\" type string required one\n" +
				"field broken id \"urn:type-origin:broken\" type " + spelling + " required one\n}\n"
			var first bytes.Buffer
			for attempt := range 2 {
				var output, diagnostic bytes.Buffer
				code := runCheck([]string{"--semantic", "--json", "오류.gooo"},
					fixtureReader{source: source}, EntityFieldsCLIParser{}, &output, &diagnostic)
				var report jsonReport
				if err := json.Unmarshal(output.Bytes(), &report); err != nil {
					t.Fatal(err)
				}
				if code != exitFailure || report.Status != "error" || len(report.Diagnostics) != 1 || diagnostic.Len() != 0 {
					t.Fatal("unknown field type must remain a semantic failure", code, report, diagnostic.String())
				}
				entry := report.Diagnostics[0]
				start := strings.Index(source, "type "+spelling) + len("type ")
				if entry.Code != "semantic.lowering" || entry.Span.File != "오류.gooo" ||
					entry.Span.Start.Offset != start || entry.Span.End.Offset != start+len(spelling) ||
					!strings.Contains(entry.Message, "urn:type-origin:broken") {
					t.Fatal("failure must identify the actual field and its original type reference", entry)
				}
				if attempt == 0 {
					first.Write(output.Bytes())
				} else if !bytes.Equal(first.Bytes(), output.Bytes()) {
					t.Fatal("repeating semantic inspection changed the diagnostic")
				}
			}
		})
	}
}
