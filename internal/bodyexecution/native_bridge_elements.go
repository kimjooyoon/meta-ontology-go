package bodyexecution

import (
	"fmt"
	"slices"
)

// The existing 32-column history and new element columns keep separate record
// shapes. References name original case-local reads, never equal-valued items.
type IndexedBridgeLineageColumns struct {
	History  BridgeLineageColumns `json:"history"`
	Elements BridgeElementColumns `json:"elements"`
}

type BridgeElementColumns struct {
	Refs                []int    `json:"refs"`
	ReadIDs             []int    `json:"read_ids"`
	Statuses            []string `json:"statuses"`
	Indices             []int    `json:"indices"`
	RecordIDs           []string `json:"record_ids"`
	Activities          []string `json:"activities"`
	Ports               []string `json:"ports"`
	FieldIDs            []string `json:"field_ids"`
	OwnerIDs            []string `json:"owner_ids"`
	ReadDropped         int      `json:"read_dropped"`
	RegistrationDropped int      `json:"registration_dropped"`
	ReadJournalSHA256   string   `json:"read_journal_sha256"`
	FieldSourcesSHA256  string   `json:"field_sources_sha256"`
}

func SelectRecordedIndexedBridgeLineagePage(r CompositionRuntime, query NativeOriginQuery) ([]IndexedBridgeLineageColumns, *NativeOriginInput, error) {
	history, report, err := selectRecordedBridgeLineagePage(r, query, true)
	if err != nil {
		return nil, report, err
	}
	result := make([]IndexedBridgeLineageColumns, 0, len(history))
	for _, row := range history {
		elements, err := recordedBridgeElements(r, r.Traces[row.CaseIndex])
		if err != nil {
			return nil, report, err
		}
		result = append(result, IndexedBridgeLineageColumns{History: row, Elements: elements})
	}
	report.InputShape, report.ConsumerRows = "gooo/field-local-lineage-columns/v2", len(result)
	report.Scope = "unsliced recorded field/local cases and referenced original element-read metadata; Gooo owns traversal, input-root selection and partial status; element references preserve occurrence IDs and original positions; no producer execution, model call or ancestry calculation in transport"
	return result, report, nil
}

func recordedBridgeElements(r CompositionRuntime, t CompositionTrace) (BridgeElementColumns, error) {
	c := BridgeElementColumns{Refs: make([]int, len(t.FieldWrites)+len(t.LocalWrites)),
		ReadIDs: []int{}, Statuses: []string{}, Indices: []int{}, RecordIDs: []string{},
		Activities: []string{}, Ports: []string{}, FieldIDs: []string{}, OwnerIDs: []string{},
		ReadDropped: t.ReadsDropped, RegistrationDropped: t.ReadRegistrationDropped,
		ReadJournalSHA256: r.ReadObservation.SHA256}
	if r.FieldSources != nil {
		c.FieldSourcesSHA256 = r.FieldSources.SHA256
	}
	for i := range c.Refs {
		c.Refs[i] = -1
	}
	copy(c.Refs, t.Bridge.ElementReads)
	records := map[int]string{}
	for i, ref := range c.Refs[:len(t.FieldWrites)] {
		if ref < 0 {
			continue
		}
		site := r.FieldSites[t.FieldWrites[i].SiteIndex]
		record := r.FieldBindings[site.Binding].Record.ID
		if prior, found := records[ref]; found {
			if prior != record {
				return c, fmt.Errorf("one element read claims different nominal records")
			}
			continue
		}
		records[ref] = record
		c.ReadIDs = append(c.ReadIDs, ref)
	}
	slices.Sort(c.ReadIDs)
	for _, ref := range c.ReadIDs {
		c.appendElement(r, t, ref, records[ref])
	}
	return c, nil
}

func (c *BridgeElementColumns) appendElement(r CompositionRuntime, t CompositionTrace, ref int, record string) {
	read := t.Reads[ref]
	fact := compositionReadFact(r, t.CaseIndex, read)
	activity, port, field, owner := fact.ActivityID, "", "", ""
	if read.Origin != nil {
		input := r.ReadInputs[read.Origin.Input]
		activity, port, field, owner = input.ActivityID, input.Port, input.FieldID, input.EntityID
	}
	c.Statuses, c.Indices = append(c.Statuses, fact.Status), append(c.Indices, fact.OriginIndex)
	c.RecordIDs, c.Activities = append(c.RecordIDs, record), append(c.Activities, activity)
	c.Ports, c.FieldIDs, c.OwnerIDs = append(c.Ports, port), append(c.FieldIDs, field), append(c.OwnerIDs, owner)
}
