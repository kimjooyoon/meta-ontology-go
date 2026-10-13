package bodyexecution

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type nativeConditionEvent struct {
	CaseIndex *int  `json:"case_index"`
	SiteIndex *int  `json:"site_index"`
	Value     *bool `json:"value"`
}

func (graph compositionGraph) nativeConditionRows(output []byte, cases int,
	sites []CompositionConditionSite) ([]byte, [][]CompositionConditionObservation, error) {
	if len(sites) == 0 {
		return output, nil, nil
	}
	var observed struct {
		Schema     string                 `json:"schema"`
		Result     json.RawMessage        `json:"result"`
		Conditions []nativeConditionEvent `json:"conditions"`
	}
	if err := decode(output, &observed, 32<<20); err != nil {
		return nil, nil, err
	}
	if observed.Schema != conditionObservationSchema || len(observed.Result) == 0 ||
		bytes.Equal(bytes.TrimSpace(observed.Result), []byte("null")) || observed.Conditions == nil ||
		len(observed.Conditions) > conditionObservationLimit {
		return nil, nil, fmt.Errorf("compiled native if observation protocol or bound differs")
	}
	rows, err := graph.nativeConditionEvents(observed.Conditions, cases, sites)
	return observed.Result, rows, err
}

func (graph compositionGraph) nativeConditionEvents(events []nativeConditionEvent, cases int,
	sites []CompositionConditionSite) ([][]CompositionConditionObservation, error) {
	if cases < 1 || cases > 128 || len(events) > conditionObservationLimit {
		return nil, fmt.Errorf("compiled native if observation exceeds its event or case bound")
	}
	rows := make([][]CompositionConditionObservation, cases)
	lastCase, lastRoot := -1, -1
	for _, event := range events {
		if event.CaseIndex == nil || event.SiteIndex == nil || event.Value == nil ||
			*event.CaseIndex < 0 || *event.CaseIndex >= cases || *event.SiteIndex < 0 || *event.SiteIndex >= len(sites) {
			return nil, fmt.Errorf("compiled native if observation lacks a case, site or truth value")
		}
		c, s := *event.CaseIndex, *event.SiteIndex
		root := graph.index(sites[s].RootActivityID)
		if c < lastCase || root < 0 || c == lastCase && root < lastRoot {
			return nil, fmt.Errorf("compiled native if observation root or evaluation order differs")
		}
		rows[c] = append(rows[c], CompositionConditionObservation{SiteIndex: s, Value: *event.Value})
		lastCase, lastRoot = c, root
	}
	return rows, nil
}

func (graph compositionGraph) nativeConditionFile(raw []byte, cases int, sites []CompositionConditionSite) (
	[][]CompositionConditionObservation, []int, *CompositionConditionSummary, error) {
	var observed struct {
		Schema     string                 `json:"schema"`
		Conditions []nativeConditionEvent `json:"conditions"`
		Dropped    []int                  `json:"dropped"`
	}
	if err := decode(raw, &observed, 8<<20); err != nil {
		return nil, nil, nil, err
	}
	if observed.Schema != "gooo/native-condition-observation/v2" || observed.Conditions == nil || len(observed.Dropped) != cases {
		return nil, nil, nil, fmt.Errorf("native if file protocol or dropped-case count differs")
	}
	rows, err := graph.nativeConditionEvents(observed.Conditions, cases, sites)
	if err != nil {
		return nil, nil, nil, err
	}
	summary := &CompositionConditionSummary{Schema: "gooo/native-condition-summary/v1", Status: "COMPLETE",
		RecordedEvents: len(observed.Conditions), SHA256: digest(raw)}
	last := -1
	if len(observed.Conditions) != 0 {
		last = *observed.Conditions[len(observed.Conditions)-1].CaseIndex
	}
	for c, dropped := range observed.Dropped {
		if dropped < 0 || dropped > 0 && (summary.RecordedEvents != conditionObservationLimit || c < last) ||
			dropped > int(^uint(0)>>1)-summary.DroppedEvents {
			return nil, nil, nil, fmt.Errorf("native if file claimed impossible dropped events")
		}
		summary.DroppedEvents += dropped
	}
	if summary.DroppedEvents > 0 {
		summary.Status = "PARTIAL"
	}
	return rows, observed.Dropped, summary, nil
}

func (graph compositionGraph) bindNativeConditions(traces []CompositionTrace,
	rows [][]CompositionConditionObservation, sites []CompositionConditionSite) error {
	if rows == nil {
		return nil
	}
	if len(traces) != len(rows) {
		return fmt.Errorf("native if observations differ from the actual case count")
	}
	for c := range traces {
		for _, event := range rows[c] {
			root := graph.index(sites[event.SiteIndex].RootActivityID)
			if len(traces[c].Deliveries[root].BlockedBy) != 0 {
				return fmt.Errorf("blocked activity claimed an evaluated if condition")
			}
		}
		traces[c].ConditionsObserved, traces[c].Conditions = true, rows[c]
	}
	return nil
}
