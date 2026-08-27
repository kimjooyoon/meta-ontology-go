package semanticdeltareceiptconsumer

import (
	"bytes"
	"sort"
)

func textualDelta(before, after []byte) TextualDelta {
	changed := !bytes.Equal(before, after)
	decision := rawFixedPoint
	if changed {
		decision = rawChanged
	}
	changedBytes := len(after) - len(before)
	if changedBytes < 0 {
		changedBytes = -changedBytes
	}
	for index := 0; index < len(before) && index < len(after); index++ {
		if before[index] != after[index] {
			changedBytes++
		}
	}
	return TextualDelta{Changed: changed, Decision: decision, BeforeBytes: len(before), AfterBytes: len(after), ChangedBytes: changedBytes, BeforeDigest: digestBytes(before), AfterDigest: digestBytes(after)}
}

func structuralDelta(before, after projectedSource) StructuralDelta {
	result := StructuralDelta{Status: "KNOWN"}
	leftNodes, rightNodes := keyedNodes(before.nodes), keyedNodes(after.nodes)
	leftFacts, rightFacts := keyedFacts(before.facts), keyedFacts(after.facts)
	for key, value := range rightNodes {
		if _, ok := leftNodes[key]; !ok {
			result.AddedNodes = append(result.AddedNodes, value)
		}
	}
	for key, value := range leftNodes {
		if _, ok := rightNodes[key]; !ok {
			result.RemovedNodes = append(result.RemovedNodes, value)
		}
	}
	for key, value := range rightFacts {
		if _, ok := leftFacts[key]; !ok {
			result.AddedFacts = append(result.AddedFacts, value)
		}
	}
	for key, value := range leftFacts {
		if _, ok := rightFacts[key]; !ok {
			result.RemovedFacts = append(result.RemovedFacts, value)
		}
	}
	sort.Slice(result.AddedNodes, func(i, j int) bool { return result.AddedNodes[i].ID < result.AddedNodes[j].ID })
	sort.Slice(result.RemovedNodes, func(i, j int) bool { return result.RemovedNodes[i].ID < result.RemovedNodes[j].ID })
	sort.Slice(result.AddedFacts, func(i, j int) bool { return factLess(result.AddedFacts[i], result.AddedFacts[j]) })
	sort.Slice(result.RemovedFacts, func(i, j int) bool { return factLess(result.RemovedFacts[i], result.RemovedFacts[j]) })
	return result
}

func keyedNodes(values []Node) map[string]Node {
	result := make(map[string]Node, len(values))
	for _, value := range values {
		result[value.ID+"\x00"+value.Kind] = value
	}
	return result
}

func keyedFacts(values []Fact) map[string]Fact {
	result := make(map[string]Fact, len(values))
	for _, value := range values {
		result[value.Subject+"\x00"+value.Predicate+"\x00"+value.Object] = value
	}
	return result
}
