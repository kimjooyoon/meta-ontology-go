package bodycodegen

import "fmt"

func (c *recordFlowBuilder) nominalValue(root uint16) (flowValue, bool) {
	if root == 0 || c.nominal == nil {
		return flowValue{}, false
	}
	value := c.nominal[root-1]
	return value, value.scalar == root
}

func (c *recordFlowBuilder) retainNominalValue(value flowValue) {
	if value.scalar == 0 || c.err != nil {
		return
	}
	if c.nominal == nil {
		c.nominal = new([recordFlowNodeLimit]flowValue)
	}
	c.nominal[value.scalar-1] = value
}

func (c *recordFlowBuilder) inputRecordFields(record int, span RecordFlowSpan, local, parent uint16) flowValue {
	value := flowValue{record: record, scalar: parent}
	for i, field := range c.body.records[record].Fields {
		if c.err != nil {
			break
		}
		node := c.add("input", [2]uint16{parent}, span, field.ID, local, "")
		value.fields[i] = node
		if field.Cardinality != "" {
			c.sequenceValue(node, field.TypeID)
			continue
		}
		for j, child := range c.body.records {
			if child.ID == field.TypeID {
				c.inputRecordFields(j, span, local, node)
				break
			}
		}
	}
	c.retainNominalValue(value)
	return value
}

func (c *recordFlowBuilder) recordFieldRoot(value flowValue, span RecordFlowSpan, field string, local uint16, kind string) uint16 {
	operator := ""
	if kind == "record" {
		if value.record < 0 {
			return value.scalar
		}
		kind, operator = "expression", "record_field"
	}
	root := c.add(kind, [2]uint16{value.scalar}, span, field, local, operator)
	if value.record >= 0 {
		value.scalar = root
		c.retainNominalValue(value)
	}
	return root
}

func (c *recordFlowBuilder) nominalRoot(root uint16, span RecordFlowSpan) uint16 {
	value, ok := c.nominalValue(root)
	if !ok {
		return root
	}
	result := uint16(0)
	for i, field := range c.body.records[value.record].Fields {
		part := c.nominalRoot(value.fields[i], span)
		if result == 0 {
			result = part
		} else {
			result = c.add("expression", [2]uint16{result, part}, span, field.ID, 0, "record_fields")
		}
	}
	if c.compactRecords {
		result = c.add("expression", [2]uint16{result}, span, "", 0, "record_fields")
		c.linkRecordVersion(result, root)
	}
	return result
}

func (c *recordFlowBuilder) joinField(a, b uint16, span RecordFlowSpan, field string, local, guard uint16) uint16 {
	if a == b {
		return a
	}
	x, xok := c.nominalValue(a)
	y, yok := c.nominalValue(b)
	root := c.join(a, b, span, field, local, guard)
	if !xok && !yok {
		return root
	}
	if xok != yok || x.record != y.record {
		c.err = fmt.Errorf("FLOW_RECORD_JOIN_UNRESOLVED")
		return 0
	}
	x.scalar = root
	for i, field := range c.body.records[x.record].Fields {
		x.fields[i] = c.joinField(x.fields[i], y.fields[i], span, field.ID, local, guard)
	}
	c.retainNominalValue(x)
	return root
}
