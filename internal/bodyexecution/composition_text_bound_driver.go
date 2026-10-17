package bodyexecution

import (
	"fmt"
	"strings"
)

func compositionNativePreamble(graph compositionGraph) string {
	for _, node := range graph.nodes[:graph.count] {
		if record, ok := graph.recordType(node.OutputType); ok {
			for _, field := range record.Fields {
				if field.TextBound.IsSet() {
					return "package main\nimport(\"encoding/json\";\"os\";\"unicode/utf8\")\nfunc main(){\n"
				}
			}
		}
	}
	return "package main\nimport(\"encoding/json\";\"os\")\nfunc main(){\n"
}

// JSON serialization replaces malformed Go strings. Validate source-bounded
// Text before that irreversible conversion; size admission remains in the
// source-bound record receiver and its unchanged aggregate output budget.
func writeNativeTextUTF8Checks(source *strings.Builder, graph compositionGraph, node CompositionActivity, value string) {
	record, ok := graph.recordType(node.OutputType)
	if !ok {
		return
	}
	for _, field := range record.Fields {
		if !field.TextBound.IsSet() {
			continue
		}
		expression := value + "." + field.GoName
		if field.Presence == "optional" {
			fmt.Fprintf(source, "if %s!=nil{\n", expression)
			expression = "*" + expression
		}
		if field.Cardinality == "many" {
			fmt.Fprintf(source, "for _,goooBoundText:=range %s{\n", expression)
			expression = "goooBoundText"
		}
		message := fmt.Sprintf("record field %s: Text requires valid UTF-8 before JSON encoding\n", field.ID)
		fmt.Fprintf(source, "if !utf8.ValidString(%s){_,_=os.Stderr.WriteString(%q);os.Exit(4)}\n", expression, message)
		if field.Cardinality == "many" {
			source.WriteString("}\n")
		}
		if field.Presence == "optional" {
			source.WriteString("}\n")
		}
	}
}
