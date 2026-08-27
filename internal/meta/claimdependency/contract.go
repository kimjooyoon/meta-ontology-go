package claimdependency

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const (
	ProducerID      = "gooo://meta/claim-dependency/producer/v1"
	ConsumerID      = "gooo://meta/claim-dependency/independent-judge/v1"
	MetaOperationID = "classify-claim-state-causality"
	ProofChoice     = "COHERENCE"
)

var claimContract = [...]Claim{
	{Ordinal: 1, Axis: "source-observed", ClaimID: "gooo.claim.dependency.source-observed.v1", Statement: "the Gooo source is the observed subject", Producer: ProducerID, Consumer: ConsumerID, MetaOperation: "observe-gooo-source", ProofChoice: "FOUNDATION", Coordinate: Coordinate{Stage: "READ", Step: "gooo-source", Reason: "SOURCE_READ"}},
	{Ordinal: 2, Axis: "producer-bound", ClaimID: "gooo.claim.dependency.producer-bound.v1", Statement: "the receipt identifies its deterministic producer", Producer: ProducerID, Consumer: ConsumerID, MetaOperation: "bind-producer", ProofChoice: "FOUNDATION", Coordinate: Coordinate{Stage: "BIND", Step: "producer", Reason: "PRODUCER_IDENTIFIED"}},
	{Ordinal: 3, Axis: "proof-choice-bound", ClaimID: "gooo.claim.dependency.proof-choice-bound.v1", Statement: "the state claim names a proof choice", Producer: ProducerID, Consumer: ConsumerID, MetaOperation: "choose-proof-route", ProofChoice: ProofChoice, Coordinate: Coordinate{Stage: "BIND", Step: "proof-choice", Reason: "PROOF_CHOICE_DECLARED"}},
	{Ordinal: 4, Axis: "consumer-bound", ClaimID: "gooo.claim.dependency.consumer-bound.v1", Statement: "the receipt names an independent decision consumer", Producer: ProducerID, Consumer: ConsumerID, MetaOperation: "bind-consumer", ProofChoice: "COHERENCE", Coordinate: Coordinate{Stage: "BIND", Step: "consumer", Reason: "CONSUMER_IDENTIFIED"}},
	{Ordinal: 5, Axis: "read-only-bound", ClaimID: "gooo.claim.dependency.read-only-bound.v1", Statement: "the experiment cannot mutate the repository", Producer: ProducerID, Consumer: ConsumerID, MetaOperation: "deny-repository-mutation", ProofChoice: "REGRESSION", Coordinate: Coordinate{Stage: "GUARD", Step: "authority", Reason: "READ_ONLY"}},
	{Ordinal: 6, Axis: "decision-replay-bound", ClaimID: "gooo.claim.dependency.decision-replay-bound.v1", Statement: "an independent judge can replay the state decision", Producer: ProducerID, Consumer: ConsumerID, MetaOperation: MetaOperationID, ProofChoice: "REGRESSION", Coordinate: Coordinate{Stage: "JUDGE", Step: "replay-decision", Reason: "INDEPENDENT_REPLAY"}},
}

type edgeSpec struct {
	from int
	to   int
	kind string
}

var edgeContract = [...]edgeSpec{
	{from: 0, to: 1, kind: "SOURCE_INFORMS_PRODUCER"},
	{from: 1, to: 2, kind: "PRODUCER_SELECTS_PROOF"},
	{from: 1, to: 3, kind: "PRODUCER_BINDS_CONSUMER"},
	{from: 1, to: 4, kind: "PRODUCER_DENIES_MUTATION"},
	{from: 2, to: 5, kind: "PROOF_SUPPORTS_DECISION"},
	{from: 3, to: 5, kind: "CONSUMER_ACCEPTS_RECEIPT"},
	{from: 4, to: 5, kind: "AUTHORITY_GUARDRAIL"},
	{from: 1, to: 5, kind: "PRODUCER_TRACEABLE_DECISION"},
}

func buildGraph() (Graph, error) {
	graph := Graph{
		Schema: GraphSchema, Authority: "DECLARED_EXPERIMENTAL_CONTRACT",
		Completeness: "CLOSED_WORLD_FIXED_SIX_CLAIMS", NodeTotal: ClaimTotal,
		EdgeTotal: EdgeTotal, Nodes: append([]Claim(nil), claimContract...),
		Edges: make([]Edge, 0, EdgeTotal),
	}
	for index, edge := range edgeContract {
		graph.Edges = append(graph.Edges, Edge{
			EdgeID: fmt.Sprintf("E%02d", index+1), FromClaimID: claimContract[edge.from].ClaimID,
			ToClaimID: claimContract[edge.to].ClaimID, Kind: edge.kind,
		})
	}
	digest, err := graphDigest(graph)
	if err != nil {
		return Graph{}, err
	}
	graph.Digest = digest
	return graph, nil
}

func graphDigest(graph Graph) (string, error) {
	graph.Digest = ""
	return digestJSON(graph)
}

func receiptDigest(receipt Receipt) (string, error) {
	receipt.Digest = ""
	return digestJSON(receipt)
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
