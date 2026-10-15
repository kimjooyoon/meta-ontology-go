package main

import (
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type semanticContextIndex struct {
	nodes     map[string]semanticContextNode
	neighbors map[string][]string
}

func newSemanticContextIndex(dump graphDump, declarations []semantic.Node) semanticContextIndex {
	index := semanticContextIndex{nodes: map[string]semanticContextNode{}, neighbors: map[string][]string{}}
	for _, node := range dump.Nodes {
		index.nodes[node.ID] = semanticContextNode{graphNode: node}
	}
	for _, declaration := range declarations {
		node := index.nodes[string(declaration.ID)]
		node.ValueProgram = declaration.ValueProgram
		node.InputSequence, node.Assembly = declaration.InputSequence, declaration.Assembly
		index.nodes[node.ID] = node
	}
	for _, relation := range dump.Relations {
		index.link(relation.Subject, relation.Object)
	}
	for _, binding := range dump.RuntimeBindings {
		index.link(binding.ProducerActivity, binding.ConsumerActivity)
		index.link(binding.ProducerActivity, binding.Entity)
		index.link(binding.ConsumerActivity, binding.Entity)
	}
	for id, neighbors := range index.neighbors {
		slices.Sort(neighbors)
		index.neighbors[id] = slices.Compact(neighbors)
	}
	return index
}

func (index semanticContextIndex) link(left, right string) {
	index.neighbors[left] = append(index.neighbors[left], right)
	index.neighbors[right] = append(index.neighbors[right], left)
}

func (index semanticContextIndex) walk(root string, limits semanticContextLimits) ([]string, map[string]int) {
	order, depths := []string{root}, map[string]int{root: 0}
	for cursor := 0; cursor < len(order); cursor++ {
		id := order[cursor]
		if depths[id] == limits.Depth {
			continue
		}
		for _, neighbor := range index.neighbors[id] {
			if _, found := depths[neighbor]; found || len(order) == limits.MaxNodes {
				continue
			}
			depths[neighbor] = depths[id] + 1
			order = append(order, neighbor)
		}
	}
	return order, depths
}

func (index semanticContextIndex) project(dump graphDump, root string, limits semanticContextLimits,
	order []string, depths map[string]int) semanticContextReport {
	report := semanticContextReport{
		Schema: "gooo/semantic-context/v1", SourceDigest: dump.SourceDigest, SemanticDigest: dump.IR.SemanticDigest,
		Root: root, FactLayer: "deterministic", Limits: limits,
		Nodes: []semanticContextNode{}, Relations: []graphRelation{},
		RuntimeBindings: []graphRuntimeBinding{}, Reasons: []string{},
	}
	selected := map[string]bool{}
	for _, id := range order {
		selected[id] = true
		node := index.nodes[id]
		node.Depth = depths[id]
		report.Nodes = append(report.Nodes, node)
	}
	for _, relation := range dump.Relations {
		if selected[relation.Subject] && selected[relation.Object] {
			report.Relations = append(report.Relations, relation)
		}
	}
	for _, binding := range dump.RuntimeBindings {
		if selected[binding.ProducerActivity] && selected[binding.ConsumerActivity] && selected[binding.Entity] {
			report.RuntimeBindings = append(report.RuntimeBindings, binding)
		}
	}
	index.addFrontier(&report, selected, depths)
	return report
}

func (index semanticContextIndex) addFrontier(report *semanticContextReport,
	selected map[string]bool, depths map[string]int) {
	frontier := map[string]int{}
	for id := range selected {
		for _, neighbor := range index.neighbors[id] {
			if selected[neighbor] {
				continue
			}
			depth := depths[id] + 1
			if previous, found := frontier[neighbor]; !found || depth < previous {
				frontier[neighbor] = depth
			}
		}
	}
	reasons := map[string]bool{}
	for id, depth := range frontier {
		if report.NextRoot == "" || id < report.NextRoot {
			report.NextRoot = id
		}
		if _, admitted := depths[id]; admitted {
			reasons["max_bytes"] = true
		} else if depth > report.Limits.Depth {
			reasons["depth"] = true
		} else {
			reasons["max_nodes"] = true
		}
	}
	for reason := range reasons {
		report.Reasons = append(report.Reasons, reason)
	}
	slices.Sort(report.Reasons)
	report.OmittedAdjacentNodes = len(frontier)
	report.Truncated = len(frontier) > 0
}
