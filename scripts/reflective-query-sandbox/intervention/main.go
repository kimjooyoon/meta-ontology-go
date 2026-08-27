package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const sourcePath = "examples/reflective-query-sandbox/main.gooo"

type probe struct {
	RawDigest                string `json:"raw_digest"`
	SemanticDigest           string `json:"semantic_digest"`
	GraphDigest              string `json:"graph_digest"`
	MutationField            string `json:"mutation_field"`
	MutationPayload          string `json:"mutation_payload"`
	MutationDecision         string `json:"mutation_decision"`
	MutationResolution       string `json:"mutation_resolution"`
	MutationReason           string `json:"mutation_reason"`
	MutationAPIOutcome       string `json:"mutation_api_outcome"`
	MutationAuthority        bool   `json:"mutation_authority"`
	GraphDigestBefore        string `json:"graph_digest_before"`
	OriginalGraphDigestAfter string `json:"original_graph_digest_after"`
	ReturnedGraphDigest      string `json:"returned_graph_digest,omitempty"`
	ClaimID                  string `json:"claim_id"`
	ClaimState               string `json:"claim_state"`
}

type interventionEvidence struct {
	RawDigestChanged          bool   `json:"raw_digest_changed"`
	SemanticDigestChanged     bool   `json:"semantic_digest_changed"`
	GraphDigestChanged        bool   `json:"graph_digest_changed"`
	MutationFieldBefore       string `json:"mutation_field_before"`
	MutationFieldAfter        string `json:"mutation_field_after"`
	MutationPayloadBefore     string `json:"mutation_payload_before"`
	MutationPayloadAfter      string `json:"mutation_payload_after"`
	MutationDecisionBefore    string `json:"mutation_decision_before"`
	MutationDecisionAfter     string `json:"mutation_decision_after"`
	MutationResolutionBefore  string `json:"mutation_resolution_before"`
	MutationResolutionAfter   string `json:"mutation_resolution_after"`
	MutationAPIOutcomeBefore  string `json:"mutation_api_outcome_before"`
	MutationAPIOutcomeAfter   string `json:"mutation_api_outcome_after"`
	MutationAuthorityBefore   bool   `json:"mutation_authority_before"`
	MutationAuthorityAfter    bool   `json:"mutation_authority_after"`
	GraphDigestBefore         string `json:"graph_digest_before"`
	OriginalGraphDigestAfter  string `json:"original_graph_digest_after"`
	ReturnedGraphDigestBefore string `json:"returned_graph_digest_before,omitempty"`
	ReturnedGraphDigestAfter  string `json:"returned_graph_digest_after,omitempty"`
	ClaimID                   string `json:"claim_id"`
	ClaimStateBefore          string `json:"claim_state_before"`
	ClaimStateAfter           string `json:"claim_state_after"`
}

type report struct {
	Schema                  string               `json:"schema"`
	Base                    probe                `json:"base"`
	SemanticIntervention    interventionEvidence `json:"semantic_intervention"`
	NonsemanticIntervention interventionEvidence `json:"nonsemantic_intervention"`
}

func main() {
	source := flag.String("source", "", "Gooo source")
	output := flag.String("output", "", "intervention evidence")
	flag.Parse()
	if *source == "" || *output == "" {
		fail("usage: intervention -source FILE -output FILE")
	}
	raw, err := os.ReadFile(*source)
	if err != nil {
		fail("read source: %v", err)
	}
	base, err := probeSource(string(raw))
	if err != nil {
		fail("probe base source: %v", err)
	}
	semanticRaw := strings.Replace(string(raw), "gooo://reflective-query-sandbox/mutation/field/id", "gooo://reflective-query-sandbox/mutation/field/name", 1)
	semanticRaw = strings.Replace(semanticRaw, "gooo://reflective-query-sandbox/mutation/payload/identity-preserving", "gooo://reflective-query-sandbox/mutation/payload/intervened-name", 1)
	if semanticRaw == string(raw) {
		fail("semantic intervention did not change mutation contract")
	}
	semanticProbe, err := probeSource(semanticRaw)
	if err != nil {
		fail("probe semantic intervention: %v", err)
	}
	nonsemanticProbe, err := probeSource(string(raw) + "\n// nonsemantic intervention\n")
	if err != nil {
		fail("probe nonsemantic intervention: %v", err)
	}
	result := report{Schema: "gooo/reflective-query-sandbox-intervention/v2", Base: base, SemanticIntervention: compare(base, semanticProbe), NonsemanticIntervention: compare(base, nonsemanticProbe)}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fail("encode evidence: %v", err)
	}
	if err := os.WriteFile(*output, append(data, '\n'), 0o644); err != nil {
		fail("write evidence: %v", err)
	}
	fmt.Printf("intervention evidence: mutation %s/%s -> %s/%s; semantic authority=%t; nonsemantic semantic-digest-preserved=%t\n", base.MutationDecision, base.MutationAPIOutcome, semanticProbe.MutationDecision, semanticProbe.MutationAPIOutcome, semanticProbe.MutationAuthority, base.SemanticDigest == nonsemanticProbe.SemanticDigest)
}

func probeSource(raw string) (probe, error) {
	file, diagnostics := syntax.ParseFile(sourcePath, raw)
	if diagnostics.HasErrors() {
		return probe{}, diagnostics.Error()
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return probe{}, err
	}
	var activity, target, fieldNode, payloadNode, intentNode, localityNode semantic.ID
	for _, node := range ir.Graph.Nodes() {
		id := node.ID.String()
		if node.Kind == semantic.Activity && node.ValueProgram == "reflect.attempt:mutation" {
			activity = node.ID
		}
		if node.Kind == semantic.Entity && strings.Contains(id, "/mutation/request") {
			target = node.ID
		}
		if node.Kind == semantic.Entity && strings.Contains(id, "/mutation/field/") {
			fieldNode = node.ID
		}
		if node.Kind == semantic.Entity && strings.Contains(id, "/mutation/payload/") {
			payloadNode = node.ID
		}
		if node.Kind == semantic.Entity && strings.Contains(id, "/mutation/intent/") {
			intentNode = node.ID
		}
		if node.Kind == semantic.Entity && strings.Contains(id, "/mutation/locality/") {
			localityNode = node.ID
		}
	}
	if activity == "" || target == "" || fieldNode == "" || payloadNode == "" || intentNode == "" || localityNode == "" {
		return probe{}, errors.New("mutation activity or source-declared contract is missing")
	}
	node, ok := ir.Graph.Node(target)
	if !ok {
		return probe{}, errors.New("mutation target is missing")
	}
	field, payload, intent, locality := tail(fieldNode), tail(payloadNode), tail(intentNode), tail(localityNode)
	before := ir.Graph.StableHash()
	fieldHash := ""
	if field != "id" {
		fieldHash, err = semantic.NodeFieldHash(node, field)
		if err != nil {
			return probe{}, err
		}
	}
	request := semantic.GraphPatchRequest{SchemaVersion: semantic.GraphPatchSchemaVersion, Operation: semantic.GraphPatchSetNodeField, ExpectedGraphHash: before, NodeID: node.ID, ExpectedNodeHash: node.StableHash(), Field: field, ExpectedFieldHash: fieldHash, ExpectedSourceDigest: ir.StableHash(), ExpectedIRDigest: ir.StableHash(), AllowedIntent: intent, Locality: locality}
	patched, callErr := ir.Graph.ApplyGraphPatch(semantic.GraphPatchBase{SourceDigest: ir.StableHash(), IRDigest: ir.StableHash()}, request, semantic.GraphPatchMutation{Name: payload})
	originalAfter := ir.Graph.StableHash()
	returned := ""
	decision, resolution, reason, outcome := "", "", "", ""
	authority := false
	if callErr != nil {
		var conflict semantic.GraphPatchConflict
		if errors.As(callErr, &conflict) && conflict.Code == semantic.PatchImmutableField && conflict.Detail == field && originalAfter == before {
			decision, resolution, reason, outcome = "DENIED", "EXACT_REJECTION", "IMMUTABLE_FIELD_REJECTED", "REJECTED"
		} else {
			decision, resolution, reason, outcome = "UNKNOWN", "LOWER_RESOLUTION", "MUTATION_API_ERROR", "ERROR"
		}
	} else {
		returned = patched.StableHash()
		decision, resolution, reason, outcome, authority = "REFUTED", "EXACT", "MUTATION_CAPABILITY_ACCEPTED", "ACCEPTED", true
	}
	state := "OPEN"
	if decision == "DENIED" && resolution == "EXACT_REJECTION" {
		state = "DISCHARGED"
	}
	if decision == "REFUTED" {
		state = "REFUTED"
	}
	claimID := "outcome.mutation-denied"
	sum := sha256.Sum256([]byte(raw))
	return probe{RawDigest: hex.EncodeToString(sum[:]), SemanticDigest: ir.StableHash(), GraphDigest: before, MutationField: field, MutationPayload: payload, MutationDecision: decision, MutationResolution: resolution, MutationReason: reason, MutationAPIOutcome: outcome, MutationAuthority: authority, GraphDigestBefore: before, OriginalGraphDigestAfter: originalAfter, ReturnedGraphDigest: returned, ClaimID: claimID, ClaimState: state}, nil
}

func tail(id semantic.ID) string {
	parts := strings.Split(strings.TrimSuffix(id.String(), "/"), "/")
	return parts[len(parts)-1]
}
func compare(base, changed probe) interventionEvidence {
	return interventionEvidence{RawDigestChanged: base.RawDigest != changed.RawDigest, SemanticDigestChanged: base.SemanticDigest != changed.SemanticDigest, GraphDigestChanged: base.GraphDigest != changed.GraphDigest, MutationFieldBefore: base.MutationField, MutationFieldAfter: changed.MutationField, MutationPayloadBefore: base.MutationPayload, MutationPayloadAfter: changed.MutationPayload, MutationDecisionBefore: base.MutationDecision, MutationDecisionAfter: changed.MutationDecision, MutationResolutionBefore: base.MutationResolution, MutationResolutionAfter: changed.MutationResolution, MutationAPIOutcomeBefore: base.MutationAPIOutcome, MutationAPIOutcomeAfter: changed.MutationAPIOutcome, MutationAuthorityBefore: base.MutationAuthority, MutationAuthorityAfter: changed.MutationAuthority, GraphDigestBefore: changed.GraphDigestBefore, OriginalGraphDigestAfter: changed.OriginalGraphDigestAfter, ReturnedGraphDigestBefore: base.ReturnedGraphDigest, ReturnedGraphDigestAfter: changed.ReturnedGraphDigest, ClaimID: changed.ClaimID, ClaimStateBefore: base.ClaimState, ClaimStateAfter: changed.ClaimState}
}
func fail(format string, args ...any) { fmt.Fprintf(os.Stderr, format+"\n", args...); os.Exit(1) }
