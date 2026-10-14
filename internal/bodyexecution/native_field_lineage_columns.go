package bodyexecution

import (
	"encoding/json"
	"fmt"
)

// FieldLineageColumns transports one complete case journal and one selected
// read. The receiving Gooo program owns traversal, root selection and status.
type FieldLineageColumns struct {
	SourceSHA256      string   `json:"source_sha256"`
	JournalSHA256     string   `json:"journal_sha256"`
	CaseIndex         int      `json:"case_index"`
	SelectedUse       int      `json:"selected_use"`
	RootAssignment    int      `json:"root_assignment"`
	RootClock         int      `json:"root_clock"`
	WriteCount        int      `json:"write_count"`
	UseCount          int      `json:"use_count"`
	RecordingComplete bool     `json:"recording_complete"`
	Bits              []int64  `json:"bits"`
	CompletionOrder   []int    `json:"completion_order"`
	Completions       []int    `json:"completions"`
	Parents           []int    `json:"parents"`
	UseStarts         []int    `json:"use_starts"`
	UseEnds           []int    `json:"use_ends"`
	WriteCompleted    []bool   `json:"write_completed"`
	WriteDrops        []int    `json:"write_drops"`
	Modes             []string `json:"modes"`
	Kinds             []string `json:"kinds"`
	InputKinds        []string `json:"input_kinds"`
	Activities        []string `json:"activities"`
	Ports             []string `json:"ports"`
	Records           []string `json:"record_ids"`
	Fields            []string `json:"field_ids"`
	UseAssignments    []int    `json:"use_assignments"`
	UseClocks         []int    `json:"use_clocks"`
}

// SelectRecordedFieldLineage exports facts without calculating an ancestry
// closure. Array order follows case-local assignment/use numbering; the extra
// completion order is an index of the producer's retained completion clock.
func SelectRecordedFieldLineage(runtime CompositionRuntime, query NativeOriginQuery) (FieldLineageColumns, *NativeOriginInput, error) {
	var c FieldLineageColumns
	if query.Kind != "field-uses" || query.Limit != 1 {
		return c, nil, fmt.Errorf("field-lineage-columns selects exactly one field use; choose field-uses and limit1")
	}
	rows, report, err := SelectRecordedNativeOrigins(runtime, query)
	if err != nil {
		return c, report, err
	}
	var selected CompositionFieldFact
	if err := json.Unmarshal(rows[0], &selected); err != nil {
		return c, report, err
	}
	trace := runtime.Traces[selected.CaseIndex]
	n, u := len(trace.FieldWrites), len(trace.FieldUses)
	if n > 60 || u > 128 {
		return c, report, fmt.Errorf("field-lineage-columns needs an unsliced case with at most60 writes/128 uses; got%d/%d", n, u)
	}
	read := trace.FieldUses[selected.EventIndex]
	c = newFieldLineageColumns(runtime, trace, selected.EventIndex, read)
	for i, write := range trace.FieldWrites {
		c.appendWrite(runtime, i, write)
	}
	for i, use := range trace.FieldUses {
		c.UseAssignments[i], c.UseClocks[i] = use.Assignment, use.CompletionClock
		if use.Assignment < 0 {
			c.RecordingComplete = false
		}
	}
	report.InputShape, report.ConsumerRows = "gooo/field-lineage-columns/v1", 1
	report.Scope = "one unsliced recorded case with at most60 field writes and128 field uses; columns preserve producer field identities, parent claims and completion clocks; receiving Gooo owns traversal and status; scalar-local/control dependencies and producer authentication are outside this journal profile"
	return c, report, nil
}

func newFieldLineageColumns(runtime CompositionRuntime, trace CompositionTrace, event int, read CompositionFieldUse) FieldLineageColumns {
	n, u := len(trace.FieldWrites), len(trace.FieldUses)
	c := FieldLineageColumns{SourceSHA256: runtime.SelectedSourceSHA256, JournalSHA256: runtime.FieldObservation.SHA256,
		CaseIndex: trace.CaseIndex, SelectedUse: event, RootAssignment: read.Assignment, RootClock: read.CompletionClock,
		WriteCount: n, UseCount: u, RecordingComplete: trace.FieldFrameDropped+trace.FieldWriteDropped+trace.FieldUseDropped == 0,
		Bits: make([]int64, n), CompletionOrder: make([]int, n), Completions: make([]int, n), Parents: make([]int, n),
		UseStarts: make([]int, n), UseEnds: make([]int, n), WriteCompleted: make([]bool, n), WriteDrops: make([]int, n),
		Modes: make([]string, n), Kinds: make([]string, n), InputKinds: make([]string, n), Activities: make([]string, n),
		Ports: make([]string, n), Records: make([]string, n), Fields: make([]string, n), UseAssignments: make([]int, u), UseClocks: make([]int, u)}
	for i := range c.CompletionOrder {
		c.CompletionOrder[i] = -1
	}
	return c
}

func (c *FieldLineageColumns) appendWrite(runtime CompositionRuntime, i int, write CompositionFieldWrite) {
	site := runtime.FieldSites[write.SiteIndex]
	binding := runtime.FieldBindings[site.Binding]
	c.Bits[i], c.Completions[i], c.Parents[i] = int64(1)<<i, write.Completion, write.ParentAssignment
	c.UseStarts[i], c.UseEnds[i], c.WriteCompleted[i], c.WriteDrops[i] = write.UseStart, write.UseEnd, write.Completed, write.UsesDropped
	c.Modes[i], c.Kinds[i], c.InputKinds[i] = site.DependencyMode, site.Kind, binding.InputKind
	c.Activities[i], c.Ports[i], c.Records[i], c.Fields[i] = binding.ActivityID, binding.InputPort, binding.Record.ID, binding.Record.Fields[site.Field].ID
	if write.Completed {
		c.CompletionOrder[write.Completion] = i
	} else {
		c.RecordingComplete = false
	}
}
