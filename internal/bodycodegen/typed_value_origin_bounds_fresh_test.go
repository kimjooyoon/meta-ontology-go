package bodycodegen

import (
	"fmt"
	"strings"
	"testing"
)

func TestTypedOriginBounds13193(t *testing.T) {
	source := "package boundorigins\nnamespace boundorigins\nentity Integer id \"urn:gooo:type:integer\"\nactivity Expand(Integer)->Integer id \"urn:typed13193:expand\" computes `let value=input\n"
	for i := range 180 {
		source += fmt.Sprintf("value=value+%d\n", 1601+i)
	}
	source += "return value`\n"
	r, err := InspectTypedValueOrigins("bounded13193.gooo", []byte(source), "Expand")
	if err != nil || r.Status != "UNRESOLVED" || r.Reason != "FLOW_NODE_BOUND" || len(r.Flow.Nodes) != 0 || len(r.Outputs) != 0 {
		t.Fatal("512-node bound publishes no prefix", err, r)
	}
	small := strings.Replace(source[:strings.Index(source, "value=value+")], "Expand", "Port", 1) + "return value`\n"
	r, err = InspectTypedValueOrigins("port13193.gooo", []byte(small), "Port")
	if err != nil || r.Status != "RESOLVED" || len(r.Inputs) != 1 || r.Flow.Nodes[r.Inputs[0].Node-1].Span.ActivityID != r.ActivityID {
		t.Fatal("root input identity", err, r)
	}
	r.Flow.Nodes = make([]RecordFlowNode, 61)
	if _, err := TypedValueOriginInputs(r, "bounds/reader:Inputs", "", false); err == nil {
		t.Fatal("61-node Gooo export silently accepted")
	}
	t.Log("fresh180-write graph stops at512 nodes without a resolved prefix; input activity ID retained;61-node adapter rejected")
}
