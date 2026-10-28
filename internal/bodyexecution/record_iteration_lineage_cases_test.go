package bodyexecution

import "testing"

// These eight inputs are authored contract probes, separate from native capture.
func TestIndexedGoooLineageAuthoredBoundaries(t *testing.T) {
	columns := []IndexedBridgeLineageColumns{}
	wants := []indexedTraceWant{}
	add := func(c IndexedBridgeLineageColumns, w indexedTraceWant) {
		w.Source, w.Case, w.Use = "authored:source", 0, 0
		columns, wants = append(columns, c), append(wants, w)
	}
	copy := authoredIndexedColumns(3, 2, 1)
	copy.History.Parents[2], copy.History.Modes[2] = 1, "field_copy"
	add(copy, indexedTraceWant{Low: 2, Roots: 1, Links: 1, Reachable: 2, Parents: 1,
		Status: "SCOPED_INDEXED_FIELD_LOCAL_ORIGINS", Complete: true})
	add(authoredIndexedColumns(34, 32, 32), indexedTraceWant{High: 1, Roots: 1, Links: 1, Reachable: 1,
		Status: "SCOPED_INDEXED_FIELD_LOCAL_ORIGINS", Complete: true})
	for _, status := range []string{"UNRESOLVED", "CONSTRUCTED"} {
		c := authoredIndexedColumns(1, 0, 0)
		c.Elements.Statuses[0], c.Elements.Indices[0] = status, -1
		if status == "CONSTRUCTED" {
			c.Elements.Indices[0] = 0
		}
		c.Elements.Ports[0], c.Elements.FieldIDs[0], c.Elements.OwnerIDs[0] = "", "", ""
		add(c, indexedTraceWant{Links: 1, Reachable: 1, Unresolved: 1, Status: "PARTIAL_LINKS", Complete: true})
	}
	dropped := authoredIndexedColumns(1, 0, 0)
	dropped.Elements.Refs[0], dropped.Elements.ReadDropped = -2, 1
	add(dropped, indexedTraceWant{Links: 1, Reachable: 1, Unresolved: 1, Status: "PARTIAL_RECORDING"})
	missing := authoredIndexedColumns(1, 0, 0)
	missing.Elements.ReadIDs[0] = 8
	add(missing, indexedTraceWant{Status: "INVALID_INPUT", Complete: true})
	typed := authoredIndexedColumns(1, 0, 0)
	typed.Elements.RecordIDs[0] = "fixture://other"
	add(typed, indexedTraceWant{Status: "INVALID_INPUT", Complete: true})
	duplicate := authoredIndexedColumns(1, 0, 0)
	duplicate.Elements.ReadIDs = append(duplicate.Elements.ReadIDs, 7)
	duplicate.Elements.Statuses = append(duplicate.Elements.Statuses, "RESOLVED")
	duplicate.Elements.Indices = append(duplicate.Elements.Indices, 3)
	duplicate.Elements.RecordIDs = append(duplicate.Elements.RecordIDs, "fixture://line")
	duplicate.Elements.Activities = append(duplicate.Elements.Activities, "fixture://build")
	duplicate.Elements.Ports = append(duplicate.Elements.Ports, "input")
	duplicate.Elements.FieldIDs = append(duplicate.Elements.FieldIDs, "fixture://basket/items")
	duplicate.Elements.OwnerIDs = append(duplicate.Elements.OwnerIDs, "fixture://basket")
	add(duplicate, indexedTraceWant{Status: "INVALID_INPUT", Complete: true})
	checkIndexedGoooPrograms(t, columns, wants)
	t.Log("8 authored probes:copy parent,write32 selected in upper word,unknown andconstructed origins,read loss,missing reference,nominal mismatch,duplicate readID;both interpreter/native;no extra captured-input denominator")
}

func authoredIndexedColumns(count, selected, leaf int) IndexedBridgeLineageColumns {
	c := IndexedBridgeLineageColumns{History: BridgeLineageColumns{FieldLineageColumns: FieldLineageColumns{
		SourceSHA256: "authored:source", JournalSHA256: "authored:bridge", SelectedUse: 0,
		RootAssignment: selected, RootClock: count, WriteCount: count, UseCount: 1, RecordingComplete: true,
		Bits: []int64{}, CompletionOrder: []int{}, Completions: []int{}, Parents: []int{}, UseStarts: []int{}, UseEnds: []int{},
		WriteCompleted: []bool{}, WriteDrops: []int{}, Modes: []string{}, Kinds: []string{}, InputKinds: []string{},
		Activities: []string{}, Ports: []string{}, Records: []string{}, Fields: []string{},
		UseAssignments: []int{selected}, UseClocks: []int{count}}, FieldUseCount: 1,
		OtherStarts: []int{}, OtherEnds: []int{}, NodeNames: []string{}, SourceOrdinals: []int{}, TypeIDs: []string{}},
		Elements: BridgeElementColumns{Refs: []int{}, ReadIDs: []int{7}, Statuses: []string{"RESOLVED"},
			Indices: []int{3}, RecordIDs: []string{"fixture://line"}, Activities: []string{"fixture://build"},
			Ports: []string{"input"}, FieldIDs: []string{"fixture://basket/items"}, OwnerIDs: []string{"fixture://basket"},
			ReadJournalSHA256: "authored:read", FieldSourcesSHA256: "authored:fields"}}
	for i := range count {
		h := &c.History
		h.Bits, h.CompletionOrder, h.Completions = append(h.Bits, int64(1)<<i), append(h.CompletionOrder, i), append(h.Completions, i)
		h.Parents, h.UseStarts, h.UseEnds = append(h.Parents, -1), append(h.UseStarts, 0), append(h.UseEnds, 0)
		h.WriteCompleted, h.WriteDrops = append(h.WriteCompleted, true), append(h.WriteDrops, 0)
		h.Modes, h.Kinds, h.InputKinds = append(h.Modes, "literal"), append(h.Kinds, "record_assignment"), append(h.InputKinds, "")
		h.Activities, h.Ports = append(h.Activities, "fixture://build"), append(h.Ports, "")
		h.Records, h.Fields = append(h.Records, "fixture://line"), append(h.Fields, "fixture://line/value")
		h.OtherStarts, h.OtherEnds = append(h.OtherStarts, 1), append(h.OtherEnds, 1)
		h.NodeNames, h.SourceOrdinals, h.TypeIDs = append(h.NodeNames, "item"), append(h.SourceOrdinals, 0), append(h.TypeIDs, "urn:gooo:type:integer")
		c.Elements.Refs = append(c.Elements.Refs, -1)
	}
	c.History.Modes[leaf], c.Elements.Refs[leaf] = "shared_rhs", 7
	return c
}
