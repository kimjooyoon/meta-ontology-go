package bodycodegen

import (
	"encoding/json"
	"fmt"
	"strings"
)

// InputDeclarationColumns transports one row per original signature position.
// Record field nodes sharing one input retain one signature row.
type InputDeclarationColumns struct {
	Schema       string   `json:"schema"`
	Scope        string   `json:"scope"`
	SourceSHA256 string   `json:"source_sha256"`
	ActivityID   string   `json:"activity_id"`
	Count        int      `json:"count"`
	Indices      []int    `json:"indices"`
	Ports        []string `json:"ports"`
	TypeNames    []string `json:"type_names"`
	TypeIDs      []string `json:"type_ids"`
	Starts       []int    `json:"starts"`
	Ends         []int    `json:"ends"`
}

func TypedInputDeclarationInputs(r TypedValueOrigins, target string) ([]byte, error) {
	pkg, activity, ok := strings.Cut(target, ":")
	if !ok || pkg == "" || activity == "" || strings.ContainsAny(target, " \t\r\n") || strings.Contains(activity, ":") {
		return nil, fmt.Errorf("consumer input needs package/path:Activity")
	}
	rows, count, err := inputDeclarationRows(r)
	if err != nil {
		return nil, err
	}
	c := InputDeclarationColumns{Schema: "gooo/input-declaration-columns/v1", Scope: "activity_signature",
		SourceSHA256: r.SourceSHA256, ActivityID: r.ActivityID, Count: count,
		Indices: make([]int, count), Ports: make([]string, count), TypeNames: make([]string, count),
		TypeIDs: make([]string, count), Starts: make([]int, count), Ends: make([]int, count)}
	for i := range count {
		d := rows[i]
		if d == nil {
			return nil, fmt.Errorf("input declaration positions are not contiguous")
		}
		c.Indices[i], c.Ports[i], c.TypeNames[i], c.TypeIDs[i] = i, d.Port, d.TypeName, d.TypeID
		c.Starts[i], c.Ends[i] = d.Start, d.End
	}
	output := struct {
		Schema string                               `json:"schema"`
		Inputs []map[string]InputDeclarationColumns `json:"inputs"`
	}{Schema: "gooo/body-composition-inputs/v1", Inputs: []map[string]InputDeclarationColumns{{target: c}}}
	raw, err := json.Marshal(output)
	if err == nil && len(raw) > 32<<10 {
		err = fmt.Errorf("input declarations exceed 32 KiB")
	}
	return raw, err
}

func inputDeclarationRows(r TypedValueOrigins) ([16]*TypedInputDeclaration, int, error) {
	var rows [16]*TypedInputDeclaration
	count := 0
	if r.Status != "RESOLVED" || r.SourceSHA256 == "" || r.ActivityID == "" || r.Flow == nil ||
		r.Flow.Status != "RESOLVED" || r.Flow.ActivityID != r.ActivityID || len(r.Inputs) == 0 {
		return rows, count, fmt.Errorf("input declarations require resolved source inputs")
	}
	for _, input := range r.Inputs {
		d := input.Declaration
		if d == nil || d.Status != "MAPPED" || d.View != "activity_signature" || d.InputIndex < 0 || d.InputIndex >= len(rows) ||
			d.SourceSHA256 != r.SourceSHA256 || d.ActivityID != r.ActivityID || d.Port != input.Port ||
			d.Start < 0 || d.End <= d.Start || d.TypeName == "" || d.Expression != d.TypeName ||
			input.Node == 0 || int(input.Node) > len(r.Flow.Nodes) || r.Flow.Nodes[input.Node-1].Kind != "input" {
			return rows, count, fmt.Errorf("input declaration source binding unavailable or inconsistent")
		}
		node := r.Flow.Nodes[input.Node-1]
		if node.ID != input.Node || int(node.Local) != d.InputIndex+1 || node.FieldID != input.FieldID ||
			node.Span.ActivityID != r.ActivityID {
			return rows, count, fmt.Errorf("input declaration differs from its source graph node")
		}
		if previous := rows[d.InputIndex]; previous != nil && *previous != *d {
			return rows, count, fmt.Errorf("record fields disagree about their signature position")
		}
		if rows[d.InputIndex] == nil {
			count++
		}
		rows[d.InputIndex] = d
	}
	return rows, count, nil
}
