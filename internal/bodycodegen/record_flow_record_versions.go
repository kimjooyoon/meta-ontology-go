package bodycodegen

import "fmt"

// Layout fields keep their authored order and stable IDs. Version values are
// symbolic node references, not JSON values or a value-equality interning key.
type RecordFlowRecordField struct {
	ID     string `json:"id"`
	TypeID string `json:"type_id"`
}
type RecordFlowRecordLayout struct {
	ID     string                  `json:"id"`
	Fields []RecordFlowRecordField `json:"fields"`
}
type RecordFlowRecordValue struct {
	Layout uint16   `json:"layout"`
	Fields []uint16 `json:"fields"`
}
type RecordFlowRecordVersion struct {
	Node  uint16 `json:"node"`
	Value uint16 `json:"value"`
}
type recordFlowValueKey struct {
	layout uint16
	fields [RecordFieldLimit]uint16
}

func recordFlowNeedsVersionFeatures(flow *RecordValueFlow) bool {
	return flow != nil && (flow.Schema == recordVersionFlowSchema || len(flow.RecordVersions) != 0)
}

func (c *recordFlowBuilder) exportRecordVersions(flow *RecordValueFlow) {
	if !c.compactRecords || c.err != nil {
		return
	}
	staged := RecordValueFlow{}
	layouts := map[string]uint16{}
	values := map[recordFlowValueKey]uint16{}
	for i := 0; i < c.count; i++ {
		value, ok := c.nominalValue(uint16(i + 1))
		if !ok {
			continue
		}
		record := c.body.records[value.record]
		layout := layouts[record.ID]
		if layout == 0 {
			if len(layouts) == recordTypeLimit {
				c.err = fmt.Errorf("FLOW_RECORD_TYPE_BOUND")
				return
			}
			item := RecordFlowRecordLayout{ID: record.ID}
			for _, field := range record.Fields {
				item.Fields = append(item.Fields, RecordFlowRecordField{field.ID, field.TypeID})
			}
			staged.RecordLayouts = append(staged.RecordLayouts, item)
			layout = uint16(len(staged.RecordLayouts))
			layouts[record.ID] = layout
		}
		key := recordFlowValueKey{layout: layout, fields: value.fields}
		index := values[key]
		if index == 0 {
			for j := range record.Fields {
				if value.fields[j] == 0 || int(value.fields[j]) > c.count {
					c.err = fmt.Errorf("FLOW_RECORD_FIELD_UNRESOLVED")
					return
				}
			}
			staged.RecordValues = append(staged.RecordValues, RecordFlowRecordValue{Layout: layout, Fields: append([]uint16(nil), value.fields[:len(record.Fields)]...)})
			index = uint16(len(staged.RecordValues))
			values[key] = index
		}
		staged.RecordVersions = append(staged.RecordVersions, RecordFlowRecordVersion{Node: uint16(i + 1), Value: index})
	}
	flow.RecordLayouts, flow.RecordValues, flow.RecordVersions = staged.RecordLayouts, staged.RecordValues, staged.RecordVersions
}
