package bodyexecution

import (
	"encoding/json"
	"fmt"
)

type NativeOriginQuery struct {
	RootActivityID string `json:"root_activity_id,omitempty"`
	ActivityID     string `json:"activity_id,omitempty"`
	FieldID        string `json:"field_id,omitempty"`
	Kind           string `json:"kind"`
	Offset         int    `json:"offset"`
	Limit          int    `json:"limit"`
}

// NativeOriginInput names recorded claims used by a new Gooo invocation.
// Reading a saved record does not re-attest the producer's execution.
type NativeOriginInput struct {
	Schema                 string                    `json:"schema"`
	FieldSourcesSHA256     string                    `json:"field_sources_sha256,omitempty"`
	FieldFileSourcesSHA256 string                    `json:"field_file_sources_sha256,omitempty"`
	ReceiptSHA256          string                    `json:"receipt_sha256"`
	SelectedSourceSHA256   string                    `json:"selected_source_sha256"`
	RuntimeSuiteSHA256     string                    `json:"runtime_suite_sha256"`
	Query                  NativeOriginQuery         `json:"query"`
	SourceRows             *int                      `json:"source_rows,omitempty"`
	TotalRows              int                       `json:"total_rows"`
	ReturnedRows           int                       `json:"returned_rows"`
	NextOffset             *int                      `json:"next_offset,omitempty"`
	Truncated              bool                      `json:"truncated"`
	Observation            *CompositionReadSummary   `json:"recorded_observation"`
	FieldObservation       *CompositionFieldSummary  `json:"recorded_field_observation,omitempty"`
	BridgeObservation      *CompositionBridgeSummary `json:"recorded_field_local_bridge,omitempty"`
	LocalObservation       *CompositionLocalSummary  `json:"recorded_local_observation,omitempty"`
	ProducerExecutions     int                       `json:"producer_program_executions"`
	ProducerModelCalls     int                       `json:"producer_model_calls"`
	Scope                  string                    `json:"scope"`
	InputShape             string                    `json:"input_shape,omitempty"`
	ConsumerRows           int                       `json:"consumer_rows,omitempty"`
}

func ValidateNativeOriginQuery(query NativeOriginQuery) error {
	switch query.Kind {
	case "reads", "constructions", "elements", "links", "faults", "local-writes", "local-uses", "local-elements", "local-fill-fragments", "local-input-locations", "field-writes", "field-uses", "field-input-locations":
	default:
		return fmt.Errorf("origin kind must be reads, constructions, elements, links, faults, local-writes, local-uses, local-elements, local-fill-fragments, local-input-locations, field-writes, field-uses or field-input-locations")
	}
	maximum := 65536
	if query.Kind == "local-fill-fragments" || query.Kind == "local-input-locations" || query.Kind == "field-input-locations" {
		maximum = 1<<31 - 1
	}
	if query.Offset < 0 || query.Offset > maximum || query.Limit < 1 || query.Limit > 128 {
		return fmt.Errorf("origin offset must be 0..%d and limit 1..128", maximum)
	}
	return validateOriginSubjectSelection(query)
}

type nativeOriginPage struct {
	query             NativeOriginQuery
	fileBindings      *localFileBindingIndices
	fieldFileBindings *localFileBindingIndices
	sourceRows        int
	total             int
	rows              []json.RawMessage
	err               error
}

func (page *nativeOriginPage) add(row any) {
	index := page.total
	page.total++
	if page.err != nil || index < page.query.Offset || len(page.rows) >= page.query.Limit {
		return
	}
	raw, err := json.Marshal(row)
	if err != nil {
		page.err = err
		return
	}
	page.rows = append(page.rows, raw)
}

// SelectRecordedNativeOrigins walks retained observations without materializing
// every fact or running a body, candidate, native artifact or model.
func SelectRecordedNativeOrigins(runtime CompositionRuntime, query NativeOriginQuery) ([]json.RawMessage, *NativeOriginInput, error) {
	return selectRecordedNativeOrigins(runtime, query, false)
}

// SelectRecordedNativeOriginPage also represents an empty table at offset zero.
// A column consumer can distinguish zero retained fragments from an invalid
// page offset while the existing row-input API retains its nonempty contract.
func SelectRecordedNativeOriginPage(runtime CompositionRuntime, query NativeOriginQuery) ([]json.RawMessage, *NativeOriginInput, error) {
	return selectRecordedNativeOrigins(runtime, query, true)
}

func selectRecordedNativeOrigins(runtime CompositionRuntime, query NativeOriginQuery, emptyTable bool) ([]json.RawMessage, *NativeOriginInput, error) {
	if err := ValidateNativeOriginQuery(query); err != nil {
		return nil, nil, err
	}
	if err := validateRecordedNativeOrigins(runtime); err != nil {
		return nil, nil, err
	}
	if query.Kind == "faults" {
		if err := validateRecordedFaultOrigins(runtime); err != nil {
			return nil, nil, err
		}
	}
	if query.Kind != "reads" && runtime.ReadObservation.Schema != "gooo/native-read-summary/v2" {
		return nil, nil, fmt.Errorf("saved record has no construction observations; choose reads or a newer record")
	}
	if isLocalOriginKind(query.Kind) && runtime.LocalObservation == nil {
		return nil, nil, fmt.Errorf("saved record has no scalar local observations")
	}
	if isFieldOriginKind(query.Kind) && runtime.FieldObservation == nil {
		return nil, nil, fmt.Errorf("saved record has no per-field execution history")
	}
	page := nativeOriginPage{query: query, rows: []json.RawMessage{}}
	for _, trace := range runtime.Traces {
		page.visit(runtime, trace)
	}
	if page.err != nil {
		return nil, nil, page.err
	}
	if len(page.rows) == 0 && !(emptyTable && query.Offset == 0 && page.total == 0) {
		return nil, nil, fmt.Errorf("origin selection contains no rows: kind=%s offset=%d total=%d", query.Kind, query.Offset, page.total)
	}
	observed := *runtime.ReadObservation
	report := &NativeOriginInput{Schema: "gooo/native-origin-input/v1", SelectedSourceSHA256: runtime.SelectedSourceSHA256, RuntimeSuiteSHA256: runtime.RuntimeSuiteSHA256, Query: query, TotalRows: page.total, ReturnedRows: len(page.rows), Truncated: query.Offset+len(page.rows) < page.total, Observation: &observed,
		Scope: "recorded claims read without re-execution; the new consumer evaluates these rows only; construction read windows describe reads during evaluation, not complete scalar dependencies; inspect original observation drops and the selected page"}
	if query.hasSubjectSelection() {
		report.SourceRows = &page.sourceRows
	}
	if runtime.FieldSources != nil {
		report.FieldSourcesSHA256 = runtime.FieldSources.SHA256
	}
	if runtime.FieldFileSources != nil {
		report.FieldFileSourcesSHA256 = runtime.FieldFileSources.SHA256
	}
	if query.Kind == "field-input-locations" {
		report.Scope = "recorded original-file field anchors; authored field targets and values are separate static annotations referenced by dynamic events; construction fragments identify template/candidate/grammar origins; missing history stays UNMAPPED; mapping coverage does not establish causal completeness or producer correctness"
	}
	if query.Kind == "faults" {
		report.Scope = "recorded arithmetic faults and retained operation identities; selected_local_rhs anchors cover only the RHS in the selected Gooo carrier; other source mappings remain UNMAPPED; no producer execution or new correctness denominator"
	}
	if runtime.LocalObservation != nil {
		local := *runtime.LocalObservation
		report.LocalObservation = &local
	}
	if runtime.FieldObservation != nil {
		copy := *runtime.FieldObservation
		report.FieldObservation = &copy
	}
	if runtime.BridgeObservation != nil {
		copy := *runtime.BridgeObservation
		report.BridgeObservation = &copy
		if copy.Schema == "gooo/native-field-local-bridge-summary/v2" &&
			(query.Kind == "field-writes" || query.Kind == "field-uses") {
			report.Schema = "gooo/native-origin-input/v2"
		}
	}
	if report.Truncated {
		next := query.Offset + len(page.rows)
		report.NextOffset = &next
	}
	return page.rows, report, nil
}

func isLocalOriginKind(kind string) bool {
	return kind == "local-writes" || kind == "local-uses" || kind == "local-elements" || kind == "local-fill-fragments" || kind == "local-input-locations"
}

func (page *nativeOriginPage) visit(runtime CompositionRuntime, trace CompositionTrace) {
	if page.query.Kind == "faults" {
		page.visitFaults(runtime, trace)
		return
	}
	if page.query.Kind == "field-input-locations" {
		page.visitFieldFileLocations(runtime, trace)
		return
	}
	if isFieldOriginKind(page.query.Kind) {
		page.visitFields(runtime, trace)
		return
	}
	if isLocalOriginKind(page.query.Kind) {
		page.visitLocals(runtime, trace)
		return
	}
	if page.query.Kind == "reads" || page.query.Kind == "links" {
		for i, event := range trace.Reads {
			if page.query.Kind == "reads" {
				page.add(compositionReadFact(runtime, trace.CaseIndex, event))
			} else if o := event.Construction; o != nil {
				page.add(CompositionConstructedReadFact{CaseIndex: trace.CaseIndex, ReadIndex: i, InstanceIndex: o.Instance, ElementIndex: o.Index})
			}
		}
		return
	}
	for instance, created := range trace.Constructions {
		site := runtime.ConstructionSites[created.SiteIndex]
		if page.query.Kind == "constructions" {
			page.add(compositionConstructionInstanceFact(trace.CaseIndex, instance, created, site))
			continue
		}
		for i, element := range created.Elements {
			page.add(compositionConstructionFact(trace.CaseIndex, instance, i, element, site))
		}
	}
}
