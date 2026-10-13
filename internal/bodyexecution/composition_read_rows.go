package bodyexecution

import (
	"encoding/json"
	"fmt"
)

type nativeReadEvent struct {
	Case   *int `json:"case_index"`
	Site   *int `json:"site_index"`
	Index  *int `json:"index"`
	Length *int `json:"length"`
	Origin *struct {
		Input *int `json:"input_collection"`
		Index *int `json:"element_index"`
	} `json:"origin"`
	Reason string `json:"reason"`
}

func readInputLengths(suite CompositionCases, inputs []CompositionReadInput) ([][]int, error) {
	lengths := make([][]int, len(suite.Cases))
	for c, test := range suite.Cases {
		lengths[c] = make([]int, len(inputs))
		for i, input := range inputs {
			fields, err := recordObject(test.Inputs[input.InputKey])
			if err != nil {
				return nil, err
			}
			if _, present := fields[input.FieldName]; !present && input.Presence == "optional" {
				continue
			}
			var elements []json.RawMessage
			if err := json.Unmarshal(fields[input.FieldName], &elements); err != nil || elements == nil {
				return nil, fmt.Errorf("read input requires a decoded collection")
			}
			lengths[c][i] = len(elements)
		}
	}
	return lengths, nil
}

func (graph compositionGraph) nativeReadFile(raw []byte, suite CompositionCases, sites []CompositionReadSite, inputs []CompositionReadInput) (
	[][]CompositionReadObservation, []int, []int, *CompositionReadSummary, error) {
	var file struct {
		Schema              string            `json:"schema"`
		Reads               []nativeReadEvent `json:"reads"`
		Dropped             []int             `json:"dropped"`
		RegistrationDropped []int             `json:"registration_dropped"`
	}
	fail := func(err error) ([][]CompositionReadObservation, []int, []int, *CompositionReadSummary, error) {
		return nil, nil, nil, nil, err
	}
	if err := decode(raw, &file, 16<<20); err != nil {
		return fail(err)
	}
	cases := len(suite.Cases)
	if cases < 1 || cases > 128 || file.Schema != "gooo/native-read-observation/v1" || file.Reads == nil || len(file.Reads) > readObservationLimit || len(file.Dropped) != cases || len(file.RegistrationDropped) != cases {
		return fail(fmt.Errorf("collection read protocol or bounds differ"))
	}
	lengths, err := readInputLengths(suite, inputs)
	if err != nil {
		return fail(err)
	}
	rows := make([][]CompositionReadObservation, cases)
	summary := &CompositionReadSummary{Schema: "gooo/native-read-summary/v1", Status: "COMPLETE", RecordedReads: len(file.Reads), SHA256: digest(raw),
		Scope: "successful scalar collection index and range reads; input element identity is preserved through slice views, record copies, binds and helper calls; constructed collections remain unresolved; scalar arithmetic is not tracked as element identity"}
	lastCase, lastRoot := -1, -1
	for _, event := range file.Reads {
		if event.Case == nil || event.Site == nil || event.Index == nil || event.Length == nil || *event.Case < 0 || *event.Case >= cases || *event.Site < 0 || *event.Site >= len(sites) || *event.Index < 0 || *event.Length < 1 || *event.Length > 256 || *event.Index >= *event.Length {
			return fail(fmt.Errorf("collection read requires a valid case, site and successful index"))
		}
		c, s := *event.Case, *event.Site
		root := graph.index(sites[s].RootActivityID)
		if root < 0 || c < lastCase || c == lastCase && root < lastRoot {
			return fail(fmt.Errorf("collection read evaluation order differs"))
		}
		observed := CompositionReadObservation{SiteIndex: s, Index: *event.Index, Length: *event.Length, Reason: event.Reason}
		if event.Origin != nil {
			o := event.Origin
			if o.Input == nil || o.Index == nil || *o.Input < 0 || *o.Input >= len(inputs) || *o.Index < 0 || *o.Index >= lengths[c][*o.Input] || event.Reason != "" {
				return fail(fmt.Errorf("collection read input origin is incomplete or out of range"))
			}
			owner := graph.index(inputs[*o.Input].ActivityID)
			if owner < 0 || owner > root || inputs[*o.Input].ElementType != sites[s].ElementType {
				return fail(fmt.Errorf("collection read input origin differs from decoded type or delivery order"))
			}
			observed.Origin = &CompositionReadOrigin{*o.Input, *o.Index}
			summary.ResolvedReads++
		} else {
			if event.Reason != "NO_REGISTERED_INPUT_ELEMENT" {
				return fail(fmt.Errorf("unresolved collection read requires an explicit reason"))
			}
			summary.UnresolvedReads++
		}
		rows[c] = append(rows[c], observed)
		lastCase, lastRoot = c, root
	}
	for c, dropped := range file.Dropped {
		if dropped < 0 || dropped > int(^uint(0)>>1)-summary.DroppedReads || dropped > 0 && (len(file.Reads) != readObservationLimit || c < lastCase) {
			return fail(fmt.Errorf("collection read claimed impossible dropped events"))
		}
		summary.DroppedReads += dropped
		total := 0
		for _, n := range lengths[c] {
			total += n
		}
		if file.RegistrationDropped[c] != max(0, total-readObservationLimit) {
			return fail(fmt.Errorf("collection input registration bound differs"))
		}
		summary.RegistrationDropped += file.RegistrationDropped[c]
	}
	if summary.UnresolvedReads+summary.DroppedReads+summary.RegistrationDropped > 0 {
		summary.Status = "PARTIAL"
	}
	return rows, file.Dropped, file.RegistrationDropped, summary, nil
}
