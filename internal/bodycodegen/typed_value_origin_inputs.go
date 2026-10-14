package bodycodegen

import (
	"encoding/json"
	"fmt"
	"strings"
)

// TypedOriginColumns transports graph facts; the Gooo consumer owns traversal.
type TypedOriginColumns struct {
	SourceSHA256   string   `json:"source_sha256"`
	Root           uint16   `json:"root"`
	Count          int      `json:"count"`
	IncludeControl bool     `json:"include_control"`
	Bits           []int64  `json:"bits"`
	IDs            []uint16 `json:"ids"`
	ParentA        []uint16 `json:"parent_a"`
	ParentB        []uint16 `json:"parent_b"`
	Guards         []uint16 `json:"guards"`
	Conditions     []uint16 `json:"conditions"`
	Kinds          []string `json:"kinds"`
	FieldIDs       []string `json:"field_ids"`
	InputPorts     []string `json:"input_ports"`
}

// TypedValueOriginInputs emits one input row for a selected scalar/record field.
// The accompanying Gooo example uses a 60-bit set, so this adapter refuses larger
// graphs instead of dropping nodes. The inspection report itself allows 512.
func TypedValueOriginInputs(r TypedValueOrigins, target, field string, control bool) ([]byte, error) {
	pkg, activity, ok := strings.Cut(target, ":")
	if !ok || pkg == "" || activity == "" || strings.ContainsAny(target, " \t\r\n") || strings.Contains(activity, ":") {
		return nil, fmt.Errorf("consumer input needs package/path:Activity")
	}
	if r.Status != "RESOLVED" || r.Flow == nil || len(r.Flow.Nodes) < 1 || len(r.Flow.Nodes) > 60 {
		return nil, fmt.Errorf("Gooo ancestry input requires a resolved graph with 1..60 nodes")
	}
	root := uint16(0)
	for _, output := range r.Outputs {
		if output.FieldID == field {
			root = output.Root
		}
	}
	if root == 0 || int(root) > len(r.Flow.Nodes) {
		return nil, fmt.Errorf("select an output using its stable field ID; scalar outputs use an empty field ID")
	}
	c := TypedOriginColumns{SourceSHA256: r.SourceSHA256, Root: root, Count: len(r.Flow.Nodes), IncludeControl: control}
	ports := make(map[uint16]string, len(r.Inputs))
	for _, input := range r.Inputs {
		ports[input.Node] = input.Port
	}
	for i, node := range r.Flow.Nodes {
		if int(node.ID) != i+1 || node.Parents[0] >= node.ID || node.Parents[1] >= node.ID || node.Guard >= node.ID || node.Condition >= node.ID {
			return nil, fmt.Errorf("source graph is not topologically ordered")
		}
		c.Bits = append(c.Bits, int64(1)<<i)
		c.IDs = append(c.IDs, node.ID)
		c.ParentA = append(c.ParentA, node.Parents[0])
		c.ParentB = append(c.ParentB, node.Parents[1])
		c.Guards = append(c.Guards, node.Guard)
		c.Conditions = append(c.Conditions, node.Condition)
		c.Kinds = append(c.Kinds, node.Kind)
		c.FieldIDs = append(c.FieldIDs, node.FieldID)
		c.InputPorts = append(c.InputPorts, ports[node.ID])
	}
	output := struct {
		Schema string                          `json:"schema"`
		Inputs []map[string]TypedOriginColumns `json:"inputs"`
	}{Schema: "gooo/body-composition-inputs/v1", Inputs: []map[string]TypedOriginColumns{{target: c}}}
	raw, err := json.Marshal(output)
	if err == nil && len(raw) > 32<<10 {
		err = fmt.Errorf("Gooo ancestry input exceeds 32 KiB")
	}
	return raw, err
}
