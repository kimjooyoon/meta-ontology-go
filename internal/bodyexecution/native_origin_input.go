package bodyexecution

import (
	"encoding/json"
	"fmt"
)

type NativeOriginQuery struct {
	Kind   string `json:"kind"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

// NativeOriginInput names recorded claims used by a new Gooo invocation.
// Reading a saved record does not re-attest the producer's execution.
type NativeOriginInput struct {
	Schema               string                   `json:"schema"`
	ReceiptSHA256        string                   `json:"receipt_sha256"`
	SelectedSourceSHA256 string                   `json:"selected_source_sha256"`
	RuntimeSuiteSHA256   string                   `json:"runtime_suite_sha256"`
	Query                NativeOriginQuery        `json:"query"`
	TotalRows            int                      `json:"total_rows"`
	ReturnedRows         int                      `json:"returned_rows"`
	NextOffset           *int                     `json:"next_offset,omitempty"`
	Truncated            bool                     `json:"truncated"`
	Observation          *CompositionReadSummary  `json:"recorded_observation"`
	LocalObservation     *CompositionLocalSummary `json:"recorded_local_observation,omitempty"`
	ProducerExecutions   int                      `json:"producer_program_executions"`
	ProducerModelCalls   int                      `json:"producer_model_calls"`
	Scope                string                   `json:"scope"`
}

func ValidateNativeOriginQuery(query NativeOriginQuery) error {
	switch query.Kind {
	case "reads", "constructions", "elements", "links", "local-writes", "local-uses", "local-elements":
	default:
		return fmt.Errorf("origin kind must be reads, constructions, elements, links, local-writes, local-uses or local-elements")
	}
	if query.Offset < 0 || query.Offset > 65536 || query.Limit < 1 || query.Limit > 128 {
		return fmt.Errorf("origin offset must be 0..65536 and limit 1..128")
	}
	return nil
}

type nativeOriginPage struct {
	query NativeOriginQuery
	total int
	rows  []json.RawMessage
	err   error
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
	if err := ValidateNativeOriginQuery(query); err != nil {
		return nil, nil, err
	}
	if err := validateRecordedNativeOrigins(runtime); err != nil {
		return nil, nil, err
	}
	if query.Kind != "reads" && runtime.ReadObservation.Schema != "gooo/native-read-summary/v2" {
		return nil, nil, fmt.Errorf("saved record has no construction observations; choose reads or a newer record")
	}
	if isLocalOriginKind(query.Kind) && runtime.LocalObservation == nil {
		return nil, nil, fmt.Errorf("saved record has no scalar local observations")
	}
	page := nativeOriginPage{query: query, rows: []json.RawMessage{}}
	for _, trace := range runtime.Traces {
		page.visit(runtime, trace)
	}
	if page.err != nil {
		return nil, nil, page.err
	}
	if len(page.rows) == 0 {
		return nil, nil, fmt.Errorf("origin selection contains no rows: kind=%s offset=%d total=%d", query.Kind, query.Offset, page.total)
	}
	observed := *runtime.ReadObservation
	report := &NativeOriginInput{Schema: "gooo/native-origin-input/v1", SelectedSourceSHA256: runtime.SelectedSourceSHA256, RuntimeSuiteSHA256: runtime.RuntimeSuiteSHA256, Query: query, TotalRows: page.total, ReturnedRows: len(page.rows), Truncated: query.Offset+len(page.rows) < page.total, Observation: &observed,
		Scope: "recorded claims read without re-execution; the new consumer evaluates these rows only; construction read windows describe reads during evaluation, not complete scalar dependencies; inspect original observation drops and the selected page"}
	if runtime.LocalObservation != nil {
		local := *runtime.LocalObservation
		report.LocalObservation = &local
	}
	if report.Truncated {
		next := query.Offset + len(page.rows)
		report.NextOffset = &next
	}
	return page.rows, report, nil
}

func isLocalOriginKind(kind string) bool {
	return kind == "local-writes" || kind == "local-uses" || kind == "local-elements"
}

func (page *nativeOriginPage) visit(runtime CompositionRuntime, trace CompositionTrace) {
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
