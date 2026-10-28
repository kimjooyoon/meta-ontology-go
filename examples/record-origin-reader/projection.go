package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

// Rows and pages only change transport layout. They retain original node IDs,
// all four dependency kinds and the exact symbolic record payload dictionaries.
type node struct {
	NodeID     uint16 `json:"node_id"`
	A          uint16 `json:"a"`
	B          uint16 `json:"b"`
	Guard      uint16 `json:"guard"`
	Condition  uint16 `json:"condition"`
	Version    uint16 `json:"version"`
	Value      uint16 `json:"value"`
	Bit        int64  `json:"bit"`
	Kind       string `json:"kind"`
	FieldID    string `json:"field_id"`
	Port       string `json:"port"`
	CalleeID   string `json:"callee_id"`
	ActivityID string `json:"activity_id"`
}

type nodePage struct {
	First int    `json:"first"`
	Nodes []node `json:"nodes"`
}
type field struct {
	FieldID     string `json:"field_id"`
	TypeID      string `json:"type_id"`
	Cardinality string `json:"cardinality"`
}
type layout struct {
	TypeID  string  `json:"type_id"`
	Members []field `json:"members"`
}
type value struct {
	Layout uint16   `json:"layout"`
	Roots  []uint16 `json:"roots"`
}
type valuePage struct {
	First  int     `json:"first"`
	Values []value `json:"values"`
}

// Parent zero is an explicit root-loop value in the Gooo input contract.
type loop struct {
	Number         uint16 `json:"number"`
	Parent         uint16 `json:"parent"`
	Collection     uint16 `json:"collection"`
	SequenceType   uint16 `json:"sequence_type"`
	Index          uint16 `json:"index"`
	Element        uint16 `json:"element"`
	Entry          uint16 `json:"entry"`
	Done           uint16 `json:"done"`
	Continues      bool   `json:"continues"`
	IndexKind      string `json:"index_kind"`
	SourceActivity string `json:"source_activity"`
	SourceView     string `json:"source_view"`
	SourceStart    int    `json:"source_start"`
	SourceEnd      int    `json:"source_end"`
}

type carryPage struct {
	First   int                               `json:"first"`
	Carries []bodycodegen.RecordFlowLoopCarry `json:"carries"`
}

type originGraph struct {
	Schema        string      `json:"schema"`
	SourceSHA256  string      `json:"source_sha256"`
	FlowSHA256    string      `json:"flow_sha256"`
	ReportSHA256  string      `json:"report_sha256"`
	ActivityID    string      `json:"activity_id"`
	FieldID       string      `json:"field_id"`
	Root          uint16      `json:"root"`
	Count         int         `json:"count"`
	ValueCount    int         `json:"value_count"`
	Control       bool        `json:"control"`
	History       bool        `json:"history"`
	Pages         []nodePage  `json:"pages"`
	Layouts       []layout    `json:"layouts"`
	ValuePages    []valuePage `json:"value_pages"`
	CarryPages    []carryPage `json:"carry_pages"`
	CarryCount    int         `json:"carry_count"`
	Loops         []loop      `json:"loops"`
	SequenceTypes []string    `json:"sequence_types"`
}

func project(r bodycodegen.TypedValueOrigins, fieldID string, control, history bool) ([]byte, error) {
	if r.Status != "RESOLVED" || r.Flow == nil ||
		(r.Schema != "gooo/typed-value-origins/v2" && r.Schema != "gooo/typed-value-origins/v3") {
		return nil, fmt.Errorf("source origins unresolved or outside v2/v3: %s", r.Reason)
	}
	var root uint16
	for _, out := range r.Outputs {
		if out.FieldID == fieldID {
			root = out.Root
		}
	}
	if root == 0 {
		return nil, fmt.Errorf("choose a declared output field ID from body-origins outputs")
	}
	flow, err := json.Marshal(r.Flow)
	if err != nil {
		return nil, err
	}
	report, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	rows := projectNodes(r)
	g := originGraph{
		Schema: "gooo/record-origin-graph/v2", SourceSHA256: r.SourceSHA256,
		FlowSHA256: sha(flow), ReportSHA256: sha(report), ActivityID: r.ActivityID,
		FieldID: fieldID, Root: root, Count: len(rows), ValueCount: len(r.Flow.RecordValues),
		Control: control, History: history,
		CarryCount:    len(r.Flow.LoopCarries),
		CarryPages:    []carryPage{},
		Loops:         []loop{},
		SequenceTypes: []string{},
	}
	for _, sequence := range r.Flow.SequenceTypes {
		g.SequenceTypes = append(g.SequenceTypes, sequence.ElementTypeID)
	}
	for _, l := range r.Flow.Loops {
		g.Loops = append(g.Loops, loop{l.ID, l.Parent, l.Collection, l.SequenceType,
			l.Index, l.Element, l.Entry, l.Done, l.Continues, l.IndexKind,
			l.Span.ActivityID, l.Span.View, l.Span.Start, l.Span.End})
	}
	for first := 0; first < len(r.Flow.LoopCarries); first += 256 {
		g.CarryPages = append(g.CarryPages, carryPage{first + 1, r.Flow.LoopCarries[first:min(first+256, len(r.Flow.LoopCarries))]})
	}
	for first := 0; first < len(rows); first += 256 {
		g.Pages = append(g.Pages, nodePage{first + 1, rows[first:min(first+256, len(rows))]})
	}
	projectTables(&g, r.Flow)
	return json.Marshal(g)
}

func projectNodes(r bodycodegen.TypedValueOrigins) []node {
	values := make([]uint16, len(r.Flow.Nodes)+1)
	ports := make([]string, len(r.Flow.Nodes)+1)
	for _, v := range r.Flow.RecordVersions {
		values[v.Node] = v.Value
	}
	for _, in := range r.Inputs {
		ports[in.Node] = in.Port
	}
	rows := make([]node, len(r.Flow.Nodes))
	for i, n := range r.Flow.Nodes {
		rows[i] = node{
			NodeID: n.ID, A: n.Parents[0], B: n.Parents[1], Guard: n.Guard,
			Condition: n.Condition, Version: n.RecordVersion, Value: values[n.ID],
			Bit: int64(1) << ((n.ID - 1) % 32), Kind: n.Kind, FieldID: n.FieldID,
			Port: ports[n.ID], CalleeID: n.CalleeID, ActivityID: n.Span.ActivityID,
		}
	}
	return rows
}

func projectTables(g *originGraph, f *bodycodegen.RecordValueFlow) {
	for _, l := range f.RecordLayouts {
		members := make([]field, len(l.Fields))
		for i, m := range l.Fields {
			members[i] = field{m.ID, m.TypeID, m.Cardinality}
		}
		g.Layouts = append(g.Layouts, layout{l.ID, members})
	}
	values := make([]value, len(f.RecordValues))
	for i, v := range f.RecordValues {
		values[i] = value{v.Layout, v.Fields}
	}
	for first := 0; first < len(values); first += 256 {
		g.ValuePages = append(g.ValuePages, valuePage{first + 1, values[first:min(first+256, len(values))]})
	}
}

func sha(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}
