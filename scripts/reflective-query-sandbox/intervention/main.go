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
	"github.com/kimjooyoon/meta-ontology-go/internal/query"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const sourcePath = "examples/reflective-query-sandbox/main.gooo"

type probe struct {
	RawDigest      string `json:"raw_digest"`
	SemanticDigest string `json:"semantic_digest"`
	GraphDigest    string `json:"graph_digest"`
	Decision       string `json:"decision"`
	Resolution     string `json:"resolution"`
	Reason         string `json:"reason"`
	ClaimID        string `json:"claim_id"`
	ClaimState     string `json:"claim_state"`
}

type interventionEvidence struct {
	RawDigestChanged      bool   `json:"raw_digest_changed"`
	SemanticDigestChanged bool   `json:"semantic_digest_changed"`
	GraphDigestChanged    bool   `json:"graph_digest_changed"`
	QueryDecisionBefore   string `json:"query_decision_before"`
	QueryDecisionAfter    string `json:"query_decision_after"`
	QueryResolutionBefore string `json:"query_resolution_before"`
	QueryResolutionAfter  string `json:"query_resolution_after"`
	ClaimID               string `json:"claim_id"`
	ClaimStateBefore      string `json:"claim_state_before"`
	ClaimStateAfter       string `json:"claim_state_after"`
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
	semanticRaw := strings.Replace(string(raw), "activity ReflectMetrics(QuerySubject, MetricReadRelation", "activity ReflectMetrics(QuerySubject, ClaimPriorOpen", 1)
	if semanticRaw == string(raw) {
		fail("semantic intervention did not change a declared relation")
	}
	semanticProbe, err := probeSource(semanticRaw)
	if err != nil {
		fail("probe semantic intervention: %v", err)
	}
	nonsemanticProbe, err := probeSource(string(raw) + "\n// nonsemantic intervention\n")
	if err != nil {
		fail("probe nonsemantic intervention: %v", err)
	}
	result := report{
		Schema:                  schema,
		Base:                    base,
		SemanticIntervention:    compare(base, semanticProbe),
		NonsemanticIntervention: compare(base, nonsemanticProbe),
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fail("encode evidence: %v", err)
	}
	if err := os.WriteFile(*output, append(data, '\n'), 0o644); err != nil {
		fail("write evidence: %v", err)
	}
	fmt.Printf("intervention evidence: semantic query %s -> %s; nonsemantic semantic-digest-preserved=%t\n", base.Decision, semanticProbe.Decision, base.SemanticDigest == nonsemanticProbe.SemanticDigest)
}

const schema = "gooo/reflective-query-sandbox-intervention/v1"

func probeSource(raw string) (probe, error) {
	file, diagnostics := syntax.ParseFile(sourcePath, raw)
	if diagnostics.HasErrors() {
		return probe{}, diagnostics.Error()
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return probe{}, err
	}
	graph, err := query.FromSemanticIR(ir)
	if err != nil {
		return probe{}, err
	}
	var activity, target semantic.ID
	for _, node := range ir.Graph.Nodes() {
		if node.Kind == semantic.Activity && node.ValueProgram == "reflect.query:metrics" {
			activity = node.ID
		}
		if node.Kind == semantic.Entity && strings.Contains(node.ID.String(), "/metric/relation") {
			target = node.ID
		}
	}
	if activity == "" || target == "" {
		return probe{}, errors.New("metrics activity or metric relation is missing")
	}
	result, err := graph.ExactMatch(query.NewExactQuery(query.ID(activity.String()), query.Used, query.ID(target.String())))
	decision, resolution, reason := "UNKNOWN", "LOWER_RESOLUTION", "RELATION_NOT_OBSERVED"
	if err != nil {
		if errors.Is(err, query.ErrUnknownEndpoint) {
			reason = "UNKNOWN_TARGET"
		} else {
			decision, reason = "REFUTED", "QUERY_API_REJECTED"
		}
	} else if len(result.All()) == 1 {
		decision, resolution, reason = "PASS", "EXACT", "EXACT_RELATION_MATCH"
	}
	claimID, claimState := "", ""
	for _, node := range ir.Graph.Nodes() {
		marker := "/claim/"
		index := strings.Index(node.ID.String(), marker)
		if node.Kind != semantic.Entity || index < 0 {
			continue
		}
		parts := strings.Split(node.ID.String()[index+len(marker):], "/")
		if len(parts) == 6 && parts[4] == "reflect.metrics" {
			claimID, claimState = strings.ToLower(parts[0])+"."+parts[1], "OPEN"
			if decision == "PASS" {
				claimState = "DISCHARGED"
			}
		}
	}
	sum := sha256.Sum256([]byte(raw))
	return probe{RawDigest: hex.EncodeToString(sum[:]), SemanticDigest: ir.StableHash(), GraphDigest: graph.StableHash(), Decision: decision, Resolution: resolution, Reason: reason, ClaimID: claimID, ClaimState: claimState}, nil
}

func compare(base, changed probe) interventionEvidence {
	return interventionEvidence{RawDigestChanged: base.RawDigest != changed.RawDigest, SemanticDigestChanged: base.SemanticDigest != changed.SemanticDigest, GraphDigestChanged: base.GraphDigest != changed.GraphDigest, QueryDecisionBefore: base.Decision, QueryDecisionAfter: changed.Decision, QueryResolutionBefore: base.Resolution, QueryResolutionAfter: changed.Resolution, ClaimID: changed.ClaimID, ClaimStateBefore: base.ClaimState, ClaimStateAfter: changed.ClaimState}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
