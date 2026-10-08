package bodyexecution

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
)

type nativeArithmeticRows struct {
	Schema  string                     `json:"schema"`
	Outputs [][]json.RawMessage        `json:"outputs"`
	Faults  [][]*nativeArithmeticFault `json:"faults"`
	Blocked [][][]string               `json:"blocked"`
	Calls   []json.RawMessage          `json:"calls"`
}

func (graph compositionGraph) nativeArithmeticTraces(output []byte, suite CompositionCases,
	sites []CompositionFaultSite) ([]CompositionTrace, *CompositionOutcomes, error) {
	var observed nativeArithmeticRows
	if err := decode(output, &observed, 32<<20); err != nil {
		return nil, nil, err
	}
	count := len(suite.Cases)
	if observed.Schema != "gooo/arithmetic-outcome-observation/v1" || len(observed.Outputs) != count ||
		len(observed.Faults) != count || len(observed.Blocked) != count || observed.Calls == nil {
		return nil, nil, fmt.Errorf("compiled arithmetic outcome protocol or row count differs")
	}
	calls, err := graph.arithmeticCallRows(observed, count)
	if err != nil {
		return nil, nil, err
	}
	traces := make([]CompositionTrace, count)
	outcomes := &CompositionOutcomes{}
	for c, row := range observed.Outputs {
		if len(row) != graph.count || len(observed.Faults[c]) != graph.count || len(observed.Blocked[c]) != graph.count {
			return nil, nil, fmt.Errorf("compiled arithmetic outcome activity count differs")
		}
		trace := CompositionTrace{CaseIndex: c, Deliveries: make([]CompositionDelivery, 0, graph.count)}
		if calls != nil {
			trace.CalledInputsObserved, trace.Calls = true, calls[c]
		}
		for i := range graph.count {
			entry, err := graph.arithmeticDelivery(observed, trace, suite.Cases[c], sites, i)
			if err != nil {
				return nil, nil, err
			}
			countCompositionOutcome(outcomes, entry)
			trace.Deliveries = append(trace.Deliveries, entry)
		}
		traces[c] = trace
	}
	return traces, outcomes, nil
}

func (graph compositionGraph) arithmeticCallRows(observed nativeArithmeticRows, count int) ([][]CompositionCallInput, error) {
	if len(graph.plan.Preparations) == 0 {
		if len(observed.Calls) != 0 {
			return nil, fmt.Errorf("arithmetic protocol has undeclared observed calls")
		}
		return nil, nil
	}
	raw, err := json.Marshal(map[string]any{"schema": "gooo/called-input-observation/v1",
		"outputs": observed.Outputs, "calls": observed.Calls})
	if err != nil {
		return nil, err
	}
	_, calls, err := graph.nativeCallRows(raw, count)
	return calls, err
}

func (graph compositionGraph) arithmeticDelivery(observed nativeArithmeticRows, trace CompositionTrace,
	test CompositionCase, sites []CompositionFaultSite, index int) (CompositionDelivery, error) {
	node, c := graph.nodes[index], trace.CaseIndex
	entry := CompositionDelivery{ActivityID: node.ID, BlockedBy: graph.blockedInputs(node, trace)}
	if !slices.Equal(entry.BlockedBy, observed.Blocked[c][index]) {
		return entry, fmt.Errorf("compiled blocked dependencies differ from actual producer outcomes")
	}
	wireFault, actual := observed.Faults[c][index], observed.Outputs[c][index]
	if wireFault != nil || len(entry.BlockedBy) != 0 {
		if !bytes.Equal(bytes.TrimSpace(actual), []byte("null")) || wireFault != nil && len(entry.BlockedBy) != 0 {
			return entry, fmt.Errorf("faulted or blocked activity claimed a value or another outcome")
		}
		if wireFault != nil {
			fault, err := bindArithmeticFault(wireFault, sites, node.ID)
			if err != nil {
				return entry, err
			}
			entry.Fault = fault
		}
	} else {
		value, err := graph.canonicalValue(actual, node.OutputType)
		if err != nil {
			return entry, err
		}
		entry.Actual, entry.ActualFields = value, graph.recordFieldValues(node.OutputType, value)
	}
	if err := graph.arithmeticDeliveryInputs(&entry, node, test, trace); err != nil {
		return entry, err
	}
	if expected, present := test.Expected[node.Name]; present {
		entry.Expected, _ = graph.canonicalValue(expected, node.OutputType)
		match := entry.Fault == nil && len(entry.BlockedBy) == 0 && bytes.Equal(entry.Expected, entry.Actual)
		entry.Passed = &match
	}
	return entry, nil
}

func bindArithmeticFault(wire *nativeArithmeticFault, sites []CompositionFaultSite, root string) (*CompositionFault, error) {
	if wire.Site == nil || wire.Left == nil || wire.Right == nil || *wire.Site < 0 || *wire.Site >= len(sites) || *wire.Right != 0 {
		return nil, fmt.Errorf("compiled arithmetic fault lacks an exact site or zero divisor")
	}
	site := sites[*wire.Site]
	if site.RootActivityID != root || site.Operator != "/" && site.Operator != "%" {
		return nil, fmt.Errorf("compiled arithmetic fault has a different operation or root activity")
	}
	return &CompositionFault{Kind: "ZERO_DIVISOR", Site: site, Left: *wire.Left, Right: *wire.Right}, nil
}

func (graph compositionGraph) blockedInputs(node CompositionActivity, trace CompositionTrace) []string {
	var blocked []string
	for _, slot := range node.inputSlots() {
		if slot.From >= 0 {
			producer := trace.Deliveries[slot.From]
			if (producer.Fault != nil || len(producer.BlockedBy) != 0) && !slices.Contains(blocked, producer.ActivityID) {
				blocked = append(blocked, producer.ActivityID)
			}
		}
	}
	return blocked
}

func (graph compositionGraph) arithmeticDeliveryInputs(entry *CompositionDelivery, node CompositionActivity,
	test CompositionCase, trace CompositionTrace) error {
	if len(node.Inputs) > 0 {
		entry.Inputs = make([]CompositionPortDelivery, len(node.Inputs))
	}
	for p, slot := range node.inputSlots() {
		input, producer := test.Inputs[node.inputKey(slot.Port)], ""
		if slot.From >= 0 {
			input, producer = trace.Deliveries[slot.From].Actual, graph.nodes[slot.From].ID
		}
		var fields []CompositionRecordField
		if len(input) != 0 {
			value, err := graph.canonicalValue(input, slot.Type)
			if err != nil {
				return err
			}
			input, fields = value, graph.recordFieldValues(slot.Type, value)
		}
		if len(node.Inputs) == 0 {
			entry.Input, entry.ProducerID, entry.InputFields = input, producer, fields
		} else {
			entry.Inputs[p] = CompositionPortDelivery{Port: slot.Port, EntityID: slot.EntityID, ProducerID: producer, Value: input, Fields: fields}
		}
	}
	return nil
}

func countCompositionOutcome(counts *CompositionOutcomes, entry CompositionDelivery) {
	if entry.Passed == nil {
		return
	}
	switch {
	case entry.Fault != nil:
		counts.Faulted++
	case len(entry.BlockedBy) != 0:
		counts.Blocked++
	case *entry.Passed:
		counts.Matched++
	default:
		counts.Mismatched++
	}
}
