package bodyexecution

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func (graph compositionGraph) nativeConditionRows(output []byte, cases int,
	sites []CompositionConditionSite) ([]byte, [][]CompositionConditionObservation, error) {
	if len(sites) == 0 {
		return output, nil, nil
	}
	var observed struct {
		Schema     string          `json:"schema"`
		Result     json.RawMessage `json:"result"`
		Conditions []struct {
			CaseIndex *int  `json:"case_index"`
			SiteIndex *int  `json:"site_index"`
			Value     *bool `json:"value"`
		} `json:"conditions"`
	}
	if err := decode(output, &observed, 32<<20); err != nil {
		return nil, nil, err
	}
	if observed.Schema != conditionObservationSchema || len(observed.Result) == 0 ||
		bytes.Equal(bytes.TrimSpace(observed.Result), []byte("null")) || observed.Conditions == nil ||
		len(observed.Conditions) > conditionObservationLimit {
		return nil, nil, fmt.Errorf("compiled native if observation protocol or bound differs")
	}
	rows := make([][]CompositionConditionObservation, cases)
	lastCase, lastRoot := -1, -1
	for _, event := range observed.Conditions {
		if event.CaseIndex == nil || event.SiteIndex == nil || event.Value == nil ||
			*event.CaseIndex < 0 || *event.CaseIndex >= cases || *event.SiteIndex < 0 || *event.SiteIndex >= len(sites) {
			return nil, nil, fmt.Errorf("compiled native if observation lacks a case, site or truth value")
		}
		c, s := *event.CaseIndex, *event.SiteIndex
		root := graph.index(sites[s].RootActivityID)
		if c < lastCase || root < 0 || c == lastCase && root < lastRoot {
			return nil, nil, fmt.Errorf("compiled native if observation root or evaluation order differs")
		}
		rows[c] = append(rows[c], CompositionConditionObservation{SiteIndex: s, Value: *event.Value})
		lastCase, lastRoot = c, root
	}
	return observed.Result, rows, nil
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
