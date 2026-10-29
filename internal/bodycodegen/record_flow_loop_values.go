package bodycodegen

import "fmt"

func (c *recordFlowBuilder) sequenceValue(root uint16, element string) flowValue {
	value := flowScalar(root)
	for i, kind := range c.sequenceTypes {
		if kind.ElementTypeID == element {
			value.sequence = uint16(i + 1)
			return value
		}
	}
	if len(c.sequenceTypes) == recordTypeLimit+3 {
		c.err = fmt.Errorf("FLOW_SEQUENCE_TYPE_BOUND")
		return value
	}
	c.sequenceTypes = append(c.sequenceTypes, RecordFlowSequenceType{ElementTypeID: element})
	value.sequence = uint16(len(c.sequenceTypes))
	return value
}

func (c *recordFlowBuilder) sequenceElement(sequence flowValue, index uint16, span RecordFlowSpan) flowValue {
	return c.sequenceElementAt(sequence, index, span, -1, true)
}

func (c *recordFlowBuilder) sequenceElementAt(sequence flowValue, index uint16, span RecordFlowSpan,
	known int64, iteration bool) flowValue {
	root := c.add("element", [2]uint16{c.sequenceShape(sequence.scalar), index}, span, "", 0, "symbolic_any")
	if _, ok := c.sequenceOrigin(sequence.scalar); ok {
		projection := sequenceProjection{builder: c, index: index, known: known, span: span,
			iteration: iteration, remaining: recordFlowNodeLimit}
		value := projection.element(sequence.scalar, c.sequenceTypes[sequence.sequence-1].ElementTypeID)
		return c.projectSequenceFields(value, root, span)
	}
	element := c.sequenceTypes[sequence.sequence-1].ElementTypeID
	for i, record := range c.body.records {
		if record.ID == element {
			return c.symbolicElementFields(i, root, span)
		}
	}
	return flowScalar(root)
}

func (c *recordFlowBuilder) symbolicElementFields(record int, root uint16, span RecordFlowSpan) flowValue {
	value := flowValue{record: record, scalar: root}
	for i, field := range c.body.records[record].Fields {
		if c.err != nil {
			break
		}
		child := c.add("element_field", [2]uint16{root}, span, field.ID, 0, "symbolic_field")
		value.fields[i] = child
		if field.Cardinality != "" {
			c.sequenceValue(child, field.TypeID)
			continue
		}
		for j, nested := range c.body.records {
			if nested.ID == field.TypeID {
				c.symbolicElementFields(j, child, span)
				break
			}
		}
	}
	c.retainNominalValue(value)
	return value
}

func (c *recordFlowBuilder) openLoopCarry(value flowValue, span RecordFlowSpan, local uint16, field string) flowValue {
	if c.rejectConstructedSequenceCarry(value.scalar) {
		return flowScalar(0)
	}
	if value.record >= 0 {
		for i, declaration := range c.body.records[value.record].Fields {
			child := flowScalar(value.fields[i])
			if nested, ok := c.nominalValue(child.scalar); ok {
				child = nested
			}
			value.fields[i] = c.openLoopCarry(child, span, local, declaration.ID).scalar
		}
	}
	if value.record < 0 || c.compactRecords {
		seed := value.scalar
		value.scalar = c.add("loop_carry", [2]uint16{seed}, span, field, local, "previous_or_seed")
		if value.scalar != 0 {
			c.loopData.carries[value.scalar-1] = RecordFlowLoopCarry{Loop: c.loopData.active, Header: value.scalar, Seed: seed}
		}
	}
	if value.record >= 0 {
		c.retainNominalValue(value)
	}
	return value
}

func (c *recordFlowBuilder) closeLoopCarry(seed, header, yielded flowValue, span RecordFlowSpan,
	local uint16, field string, live bool) flowValue {
	if c.rejectConstructedSequenceCarry(seed.scalar, header.scalar, yielded.scalar) {
		return flowScalar(0)
	}
	value := seed
	if value.record >= 0 {
		for i, declaration := range c.body.records[value.record].Fields {
			a, b, z := flowScalar(seed.fields[i]), flowScalar(header.fields[i]), flowScalar(yielded.fields[i])
			if nested, ok := c.nominalValue(a.scalar); ok {
				a = nested
				b, _ = c.nominalValue(b.scalar)
				z, _ = c.nominalValue(z.scalar)
			}
			value.fields[i] = c.closeLoopCarry(a, b, z, span, local, declaration.ID, live).scalar
		}
	}
	if (value.record < 0 || c.compactRecords) && header.scalar != 0 && c.err == nil {
		carry := &c.loopData.carries[header.scalar-1]
		if live {
			carry.Yield = yielded.scalar
			if yielded.scalar != header.scalar {
				value.scalar = c.add("loop_exit", [2]uint16{seed.scalar, yielded.scalar}, span, field, local, "zero_or_more")
				if value.scalar != 0 {
					c.nodes[value.scalar-1].Condition = c.loopData.loops[carry.Loop-1].Entry
				}
			}
		}
		carry.Exit = value.scalar
	}
	if value.record >= 0 {
		c.retainNominalValue(value)
	}
	return value
}
