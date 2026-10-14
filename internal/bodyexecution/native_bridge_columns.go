package bodyexecution

import (
	"encoding/json"
	"fmt"
)

// These 32 primitive columns retain the two read namespaces and true scalar
// source identities. Scalar locals have no invented record or field identity.
type BridgeLineageColumns struct {
	FieldLineageColumns
	OtherStarts    []int    `json:"other_starts"`
	OtherEnds      []int    `json:"other_ends"`
	NodeNames      []string `json:"node_names"`
	SourceOrdinals []int    `json:"source_ordinals"`
	TypeIDs        []string `json:"type_ids"`
	FieldUseCount  int      `json:"field_use_count"`
}

func SelectRecordedBridgeLineage(runtime CompositionRuntime, query NativeOriginQuery) (BridgeLineageColumns, *NativeOriginInput, error) {
	var c BridgeLineageColumns
	if query.Kind != "field-uses" || query.Limit != 1 {
		return c, nil, fmt.Errorf("field-local-columns selects exactly one field use")
	}
	rows, report, err := SelectRecordedNativeOrigins(runtime, query)
	if err != nil {
		return c, report, err
	}
	if runtime.BridgeObservation == nil {
		return c, report, fmt.Errorf("saved runtime has no native field/local bridge; no program is replayed to invent one")
	}
	var selected CompositionFieldFact
	if err := json.Unmarshal(rows[0], &selected); err != nil {
		return c, report, err
	}
	t := runtime.Traces[selected.CaseIndex]
	n, u := len(t.FieldWrites)+len(t.LocalWrites), len(t.FieldUses)+len(t.LocalUses)
	if n > 60 || u > 128 {
		return c, report, fmt.Errorf("field-local-columns needs an unsliced case with at most60 writes/128 uses; got%d/%d", n, u)
	}
	c = recordedBridgeColumns(runtime, t, selected.EventIndex)
	report.InputShape, report.ConsumerRows = "gooo/field-local-lineage-columns/v1", 1
	report.Scope = "one unsliced recorded field/local case; shared completion clocks and two evaluated read windows per write; receiving Gooo owns ancestry traversal and partial/unknown status; source declarations retain scalar-local identities; no producer execution or authentication"
	return c, report, nil
}

func newBridgeColumnArrays(n, u int) BridgeLineageColumns {
	c := BridgeLineageColumns{
		WriteCount: n, UseCount: u, Bits: make([]int64, n), CompletionOrder: make([]int, n),
		Completions: make([]int, n), Parents: make([]int, n), UseStarts: make([]int, n), UseEnds: make([]int, n),
		WriteCompleted: make([]bool, n), WriteDrops: make([]int, n), Modes: make([]string, n), Kinds: make([]string, n),
		InputKinds: make([]string, n), Activities: make([]string, n), Ports: make([]string, n), Records: make([]string, n),
		Fields: make([]string, n), UseAssignments: make([]int, u), UseClocks: make([]int, u),
		OtherStarts: make([]int, n), OtherEnds: make([]int, n), NodeNames: make([]string, n), SourceOrdinals: make([]int, n), TypeIDs: make([]string, n)}
	return c
}

func recordedBridgeColumns(r CompositionRuntime, t CompositionTrace, event int) BridgeLineageColumns {
	c := newBridgeColumnArrays(len(t.FieldWrites)+len(t.LocalWrites), len(t.FieldUses)+len(t.LocalUses))
	c.SourceSHA256, c.JournalSHA256 = r.SelectedSourceSHA256, r.BridgeObservation.SHA256
	c.CaseIndex, c.SelectedUse = t.CaseIndex, event
	c.RootAssignment, c.RootClock = t.FieldUses[event].Assignment, t.Bridge.FieldClocks[event]
	c.RecordingComplete, c.FieldUseCount = bridgeCaseRecordingComplete(t), len(t.FieldUses)
	for i, w := range t.FieldWrites {
		c.bridgeField(r, t, i, w)
	}
	for i, w := range t.LocalWrites {
		c.bridgeLocal(r, t, i, w)
	}
	for i := range c.CompletionOrder {
		c.CompletionOrder[i] = -1
	}
	for i, completion := range c.Completions {
		if c.WriteCompleted[i] {
			c.CompletionOrder[completion] = i
		}
	}
	for i, use := range t.FieldUses {
		c.UseAssignments[i], c.UseClocks[i] = use.Assignment, t.Bridge.FieldClocks[i]
	}
	for i, use := range t.LocalUses {
		assignment := use.Assignment
		if assignment >= 0 {
			assignment += len(t.FieldWrites)
		}
		c.UseAssignments[len(t.FieldUses)+i], c.UseClocks[len(t.FieldUses)+i] = assignment, t.Bridge.LocalClocks[i]
	}
	return c
}

func (c *BridgeLineageColumns) bridgeField(r CompositionRuntime, t CompositionTrace, i int, w CompositionFieldWrite) {
	c.appendWrite(r, i, w)
	b, site := t.Bridge.Fields[i], r.FieldSites[w.SiteIndex]
	binding := r.FieldBindings[site.Binding]
	c.Completions[i] = b.Completion
	c.OtherStarts[i], c.OtherEnds[i] = len(t.FieldUses)+b.OtherStart, len(t.FieldUses)+b.OtherEnd
	c.WriteDrops[i] += b.OtherDrops
	c.NodeNames[i], c.SourceOrdinals[i], c.TypeIDs[i] = binding.Name, binding.VariableOrdinal, binding.Record.Fields[site.Field].TypeID
	c.Modes[i] = bridgeLeafMode(c.Modes[i], site.Expression, w.UseEnd-w.UseStart+b.OtherEnd-b.OtherStart)
}

func (c *BridgeLineageColumns) bridgeLocal(r CompositionRuntime, t CompositionTrace, local int, w CompositionLocalWrite) {
	i, b, site := len(t.FieldWrites)+local, t.Bridge.Locals[local], r.LocalSites[w.SiteIndex]
	c.Bits[i], c.Completions[i], c.Parents[i] = int64(1)<<i, b.Completion, -1
	c.UseStarts[i], c.UseEnds[i] = len(t.FieldUses)+w.UseStart, len(t.FieldUses)+w.UseEnd
	c.OtherStarts[i], c.OtherEnds[i] = b.OtherStart, b.OtherEnd
	c.WriteCompleted[i], c.WriteDrops[i] = w.Completed, w.UsesDropped+b.OtherDrops
	c.Kinds[i], c.Activities[i] = "scalar_local", site.ActivityID
	c.NodeNames[i], c.SourceOrdinals[i] = site.Name, site.VariableOrdinal
	c.TypeIDs[i] = map[string]string{"int64": "urn:gooo:type:integer", "bool": "urn:gooo:type:boolean", "string": "urn:gooo:type:string"}[site.ScalarType]
	c.Modes[i] = bridgeLeafMode("scalar_rhs", site.Expression, w.UseEnd-w.UseStart+b.OtherEnd-b.OtherStart)
	if w.ReadEnd > w.ReadStart || w.ReadsDropped > 0 {
		c.Modes[i] = "collection_rhs"
	}
}

func bridgeCaseRecordingComplete(t CompositionTrace) bool {
	if t.Bridge == nil || t.FieldFrameDropped+t.FieldWriteDropped+t.FieldUseDropped+t.LocalFrameDropped+t.LocalWriteDropped+t.LocalUseDropped > 0 {
		return false
	}
	for _, w := range t.FieldWrites {
		if !w.Completed || w.UsesDropped > 0 {
			return false
		}
	}
	for _, w := range t.LocalWrites {
		if !w.Completed || w.ReadsDropped+w.UsesDropped > 0 {
			return false
		}
	}
	for _, windows := range [][]CompositionBridgeWindow{t.Bridge.Fields, t.Bridge.Locals} {
		for _, w := range windows {
			if w.OtherDrops > 0 {
				return false
			}
		}
	}
	for _, u := range t.FieldUses {
		if u.Assignment < 0 {
			return false
		}
	}
	for _, u := range t.LocalUses {
		if u.Assignment < 0 {
			return false
		}
	}
	return true
}
