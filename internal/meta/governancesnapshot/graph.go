package governancesnapshot

import "fmt"

func graphEvidence(graph RawGraph, contract Contract) (GraphEvidence, string) {
	evidence := GraphEvidence{Schema: graph.SchemaVersion, ProgramDigest: graph.SourceDigest, GraphHash: graph.GraphHash,
		NodeCount: len(graph.Nodes), RelationCount: len(graph.Relations), ActivityCount: 0, BindingCount: 0}
	if graph.SchemaVersion != GraphSchema || graph.SourceDigest == "" || graph.GraphHash == "" {
		return evidence, "GOOO_GRAPH_SCHEMA_OR_DIGEST_INVALID"
	}
	for _, spec := range contract.Cells {
		activity, ok := graphNode(graph.Nodes, "Activity", spec.Activity)
		if !ok {
			return evidence, "GOOO_ACTIVITY_MISSING:" + spec.Activity
		}
		input, inputOK := graphNodeID(graph.Nodes, spec.InputID)
		output, outputOK := graphNodeID(graph.Nodes, spec.OutputID)
		if !inputOK || !outputOK {
			return evidence, "GOOO_ENTITY_MISSING:" + spec.ID
		}
		used := relationCount(graph.Relations, activity.ID, "used", input)
		generated := relationCount(graph.Relations, output, "wasGeneratedBy", activity.ID)
		if used != 1 || generated != 1 {
			return evidence, fmt.Sprintf("GOOO_EDGE_CARDINALITY:%s:%d/%d", spec.ID, used, generated)
		}
		evidence.Activities = append(evidence.Activities, activity.Name)
		evidence.Bindings = append(evidence.Bindings, GraphBinding{CellID: spec.ID, ActivityID: activity.ID,
			InputID: input, OutputID: output, UsedEdgeCount: used, GeneratedCount: generated})
	}
	evidence.ActivityCount = len(evidence.Activities)
	evidence.BindingCount = len(evidence.Bindings)
	return evidence, ""
}

func graphNode(nodes []GraphNode, kind, name string) (GraphNode, bool) {
	for _, node := range nodes {
		if node.Kind == kind && node.Name == name {
			return node, true
		}
	}
	return GraphNode{}, false
}

func graphNodeID(nodes []GraphNode, id string) (string, bool) {
	for _, node := range nodes {
		if node.ID == id {
			return node.ID, true
		}
	}
	return "", false
}

func relationCount(relations []GraphRelation, subject, predicate, object string) int {
	count := 0
	for _, relation := range relations {
		if relation.Subject == subject && relation.Predicate == predicate && relation.Object == object {
			count++
		}
	}
	return count
}

func appendGraphRefutation(cells []CellObservation, contract Contract, reason string) []CellObservation {
	for index, cell := range cells {
		if cell.ID == "HUMAN_DRIFT_REPORT" {
			cells[index] = refutedCell(cell, "gooo-graph:"+reason, reason, "actual graph node and edge bindings")
			cells[index].EvidenceDigest = digestJSON(cells[index])
			return cells
		}
	}
	return cells
}
