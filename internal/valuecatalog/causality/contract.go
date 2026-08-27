package causality

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

var claimAxes = [...]string{
	"catalog-singleton",
	"identity-versioned",
	"signature-typed",
	"operand-typed",
	"effect-explicit",
	"determinism-explicit",
	"failure-set-closed",
	"authority-zero",
	"invocation-ir-bound",
}

type edgeSpec struct {
	from int
	to   int
	kind string
}

var edgeContract = [...]edgeSpec{
	{from: 0, to: 1, kind: "CATALOG_IDENTIFIES"},
	{from: 1, to: 2, kind: "IDENTITY_TYPES_SIGNATURE"},
	{from: 2, to: 3, kind: "SIGNATURE_TYPES_OPERAND"},
	{from: 1, to: 4, kind: "IDENTITY_CLASSIFIES_EFFECT"},
	{from: 4, to: 5, kind: "EFFECT_CONSTRAINS_DETERMINISM"},
	{from: 4, to: 6, kind: "EFFECT_CLOSES_FAILURE_SET"},
	{from: 0, to: 7, kind: "CATALOG_DENIES_AUTHORITY"},
	{from: 3, to: 8, kind: "OPERAND_BINDS_INVOCATION_IR"},
	{from: 5, to: 8, kind: "DETERMINISM_BINDS_INVOCATION_IR"},
	{from: 6, to: 8, kind: "FAILURE_SET_BINDS_INVOCATION_IR"},
	{from: 7, to: 8, kind: "AUTHORITY_BINDS_INVOCATION_IR"},
}

func buildGraph(claimIDs []string) (GraphContract, error) {
	if len(claimIDs) != ClaimTotal {
		return GraphContract{}, fmt.Errorf("claim total: got %d want %d", len(claimIDs), ClaimTotal)
	}
	graph := GraphContract{
		Schema:                     GraphSchema,
		Authority:                  "DECLARED_EXPERIMENTAL_CONTRACT",
		Completeness:               "CLOSED_WORLD_OS9_ONLY",
		SemanticCorrectnessClaimed: false,
		NodeTotal:                  ClaimTotal,
		EdgeTotal:                  EdgeTotal,
		Nodes:                      make([]GraphNode, 0, ClaimTotal),
		Edges:                      make([]GraphEdge, 0, EdgeTotal),
	}
	for index, claimID := range claimIDs {
		if claimID == "" || !strings.HasSuffix(claimID, claimAxes[index]+".v1") {
			return GraphContract{}, fmt.Errorf("claim %d does not bind axis %q: %q", index+1, claimAxes[index], claimID)
		}
		graph.Nodes = append(graph.Nodes, GraphNode{
			Ordinal: index + 1,
			Axis:    claimAxes[index],
			ClaimID: claimID,
		})
	}
	for index, edge := range edgeContract {
		graph.Edges = append(graph.Edges, GraphEdge{
			EdgeID:      fmt.Sprintf("E%02d", index+1),
			FromClaimID: claimIDs[edge.from],
			ToClaimID:   claimIDs[edge.to],
			Kind:        edge.kind,
		})
	}
	digest, err := graphDigest(graph)
	if err != nil {
		return GraphContract{}, err
	}
	graph.Digest = digest
	return graph, nil
}

func graphDigest(graph GraphContract) (string, error) {
	graph.Digest = ""
	return digestJSON(graph)
}

func digestJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
