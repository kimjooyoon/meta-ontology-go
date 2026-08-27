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
	RawDigest                    string `json:"raw_digest"`
	SemanticDigest               string `json:"semantic_digest"`
	GraphDigest                  string `json:"graph_digest"`
	MutationField                string `json:"mutation_field"`
	MutationPayload              string `json:"mutation_payload"`
	MutationDecision             string `json:"mutation_decision"`
	MutationResolution           string `json:"mutation_resolution"`
	MutationReason               string `json:"mutation_reason"`
	MutationAPIOutcome           string `json:"mutation_api_outcome"`
	DetachedGraphPatchCapability string `json:"detached_graph_patch_capability"`
	OverallAuthority             string `json:"overall_authority"`
	GraphDigestBefore            string `json:"graph_digest_before"`
	OriginalGraphDigestAfter     string `json:"original_graph_digest_after"`
	ReturnedGraphDigest          string `json:"returned_graph_digest,omitempty"`
	ClaimID                      string `json:"claim_id"`
	ClaimState                   string `json:"claim_state"`
	ProposalPolicy               string `json:"proposal_policy"`
	ProposalSemanticDigest       string `json:"proposal_semantic_digest"`
	ProposalOutcome              string `json:"proposal_outcome"`
	ProposalTransition           string `json:"proposal_transition"`
}

type interventionEvidence struct {
	RawDigestChanged                   bool   `json:"raw_digest_changed"`
	RawDigestBefore                    string `json:"raw_digest_before"`
	RawDigestAfter                     string `json:"raw_digest_after"`
	SemanticDigestChanged              bool   `json:"semantic_digest_changed"`
	SemanticDigestBefore               string `json:"semantic_digest_before"`
	SemanticDigestAfter                string `json:"semantic_digest_after"`
	GraphDigestChanged                 bool   `json:"graph_digest_changed"`
	MutationFieldBefore                string `json:"mutation_field_before"`
	MutationFieldAfter                 string `json:"mutation_field_after"`
	MutationPayloadBefore              string `json:"mutation_payload_before"`
	MutationPayloadAfter               string `json:"mutation_payload_after"`
	MutationDecisionBefore             string `json:"mutation_decision_before"`
	MutationDecisionAfter              string `json:"mutation_decision_after"`
	MutationResolutionBefore           string `json:"mutation_resolution_before"`
	MutationResolutionAfter            string `json:"mutation_resolution_after"`
	MutationAPIOutcomeBefore           string `json:"mutation_api_outcome_before"`
	MutationAPIOutcomeAfter            string `json:"mutation_api_outcome_after"`
	DetachedGraphPatchCapabilityBefore string `json:"detached_graph_patch_capability_before"`
	DetachedGraphPatchCapabilityAfter  string `json:"detached_graph_patch_capability_after"`
	OverallAuthorityBefore             string `json:"overall_authority_before"`
	OverallAuthorityAfter              string `json:"overall_authority_after"`
	GraphDigestBefore                  string `json:"graph_digest_before"`
	OriginalGraphDigestAfter           string `json:"original_graph_digest_after"`
	ReturnedGraphDigestBefore          string `json:"returned_graph_digest_before,omitempty"`
	ReturnedGraphDigestAfter           string `json:"returned_graph_digest_after,omitempty"`
	ClaimID                            string `json:"claim_id"`
	ClaimStateBefore                   string `json:"claim_state_before"`
	ClaimStateAfter                    string `json:"claim_state_after"`
	ProposalPolicyBefore               string `json:"proposal_policy_before"`
	ProposalPolicyAfter                string `json:"proposal_policy_after"`
	ProposalSemanticDigestChanged      bool   `json:"proposal_semantic_digest_changed"`
	ProposalSemanticDigestBefore       string `json:"proposal_semantic_digest_before"`
	ProposalSemanticDigestAfter        string `json:"proposal_semantic_digest_after"`
	ProposalOutcomeBefore              string `json:"proposal_outcome_before"`
	ProposalOutcomeAfter               string `json:"proposal_outcome_after"`
	ProposalTransitionBefore           string `json:"proposal_transition_before"`
	ProposalTransitionAfter            string `json:"proposal_transition_after"`
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
	if *source != sourcePath {
		fail("source path is not canonical: got %q want %q", *source, sourcePath)
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
	semanticRaw = strings.Replace(semanticRaw, "gooo://reflective-query-sandbox/proposal/policy/unknown-preserve-open", "gooo://reflective-query-sandbox/proposal/policy/unknown-reject-refinement", 1)
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
	result := report{Schema: "gooo/reflective-query-sandbox-intervention/v3", Base: base, SemanticIntervention: compare(base, semanticProbe), NonsemanticIntervention: compare(base, nonsemanticProbe)}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fail("encode evidence: %v", err)
	}
	if err := os.WriteFile(*output, append(data, '\n'), 0o644); err != nil {
		fail("write evidence: %v", err)
	}
	fmt.Printf("intervention evidence: immutable-id patch %s/%s -> %s/%s; detached capability=%s; overall authority=%s; nonsemantic semantic-digest-preserved=%t\n", base.MutationDecision, base.MutationAPIOutcome, semanticProbe.MutationDecision, semanticProbe.MutationAPIOutcome, semanticProbe.DetachedGraphPatchCapability, semanticProbe.OverallAuthority, base.SemanticDigest == nonsemanticProbe.SemanticDigest)
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
	capability := "NOT_OBSERVED"
	if callErr != nil {
		var conflict semantic.GraphPatchConflict
		if errors.As(callErr, &conflict) && conflict.Code == semantic.PatchImmutableField && conflict.Detail == field && originalAfter == before {
			decision, resolution, reason, outcome = "DENIED", "EXACT_REJECTION", "IMMUTABLE_ID_PATCH_REJECTED", "REJECTED"
		} else {
			decision, resolution, reason, outcome = "UNKNOWN", "LOWER_RESOLUTION", "MUTATION_API_ERROR", "ERROR"
		}
	} else {
		returned = patched.StableHash()
		decision, resolution, reason, outcome, capability = "REFUTED", "EXACT", "DETACHED_GRAPH_PATCH_ACCEPTED", "ACCEPTED", "OBSERVED"
	}
	state := "OPEN"
	if decision == "DENIED" && resolution == "EXACT_REJECTION" {
		state = "DISCHARGED"
	}
	if decision == "REFUTED" {
		state = "REFUTED"
	}
	claimID := "outcome.immutable-id-patch-rejected"
	sum := sha256.Sum256([]byte(raw))
	policy := ""
	metaProgram := ""
	for _, sourceNode := range ir.Graph.Nodes() {
		id := sourceNode.ID.String()
		if sourceNode.Kind == semantic.Entity && strings.Contains(id, "/proposal/policy/") {
			policy = tail(sourceNode.ID)
		}
		if sourceNode.Kind == semantic.Activity && sourceNode.ValueProgram == "reflect.meta:proposal:refine-unknown" {
			metaProgram = sourceNode.ValueProgram
		}
	}
	if policy == "" || metaProgram == "" {
		return probe{}, errors.New("proposal policy declaration is missing")
	}
	proposalOutcome := "REJECTED"
	proposalTransition := ""
	if policy == "unknown-preserve-open" {
		proposalOutcome, proposalTransition = "EMITTED", "OPEN->OPEN"
	}
	proposalRaw := renderInterventionProposal(semantic.StableHash([]byte(raw)), ir.StableHash(), policy, proposalOutcome, proposalTransition, metaProgram)
	proposalFile, proposalDiagnostics := syntax.ParseFile("intervention-proposal.gooo", string(proposalRaw))
	if proposalDiagnostics.HasErrors() {
		return probe{}, proposalDiagnostics.Error()
	}
	proposalIR, err := bidir.Lower(proposalFile)
	if err != nil {
		return probe{}, err
	}
	return probe{RawDigest: hex.EncodeToString(sum[:]), SemanticDigest: ir.StableHash(), GraphDigest: before, MutationField: field, MutationPayload: payload, MutationDecision: decision, MutationResolution: resolution, MutationReason: reason, MutationAPIOutcome: outcome, DetachedGraphPatchCapability: capability, OverallAuthority: "UNKNOWN", GraphDigestBefore: before, OriginalGraphDigestAfter: originalAfter, ReturnedGraphDigest: returned, ClaimID: claimID, ClaimState: state, ProposalPolicy: policy, ProposalSemanticDigest: proposalIR.StableHash(), ProposalOutcome: proposalOutcome, ProposalTransition: proposalTransition}, nil
}

func tail(id semantic.ID) string {
	parts := strings.Split(strings.TrimSuffix(id.String(), "/"), "/")
	return parts[len(parts)-1]
}
func renderInterventionProposal(sourceRawDigest, sourceSemanticDigest, policy, outcome, transition, metaProgram string) []byte {
	value := strings.Join([]string{
		"reflect.meta:proposal-artifact", "schema=gooo/reflective-query-sandbox/refinement-proposal/v1", "case=unknown-observation",
		"observation=UNKNOWN", "proposal.outcome=" + outcome, "proposal.reason=INTERVENTION_POLICY",
		"proposal.emitted=" + fmt.Sprintf("%t", outcome == "EMITTED"), "claim=guardrail.unknown-closed", "target=unknown.target",
		"source.path=examples/reflective-query-sandbox/main.gooo", "source.semantic=" + sourceSemanticDigest,
		"query.receipt=intervention-query-receipt", "unknown.stage=UNKNOWN", "unknown.step=resolve-unknown-subject", "unknown.reason=UNKNOWN_TARGET",
		"requested.refinement=refine-unknown-target-resolution", "proof.choice=REGRESSION", "meta.operation=" + metaProgram, "authority=NONE",
		"policy=" + policy, "claim.from=OPEN", "claim.to=OPEN", "claim.reason=UNKNOWN_PRESERVED", "proposal.from=" + transition,
		"proposal.to=" + transition, "proposal.transition.reason=UNKNOWN_PRESERVED_OPEN",
	}, ";")
	return []byte(fmt.Sprintf("package reflectivequeryproposal\nnamespace reflectivequeryproposal\n\nentity SourceRawDigest_%s id \"gooo://reflective-query-sandbox/proposal-artifact/source-raw\"\nentity RefinementProposal id \"gooo://reflective-query-sandbox/proposal-artifact/intervention\"\nactivity RecordProposal(RefinementProposal) -> RefinementProposal computes \"%s\"\n", sourceRawDigest, value))
}
func compare(base, changed probe) interventionEvidence {
	return interventionEvidence{RawDigestChanged: base.RawDigest != changed.RawDigest, RawDigestBefore: base.RawDigest, RawDigestAfter: changed.RawDigest, SemanticDigestChanged: base.SemanticDigest != changed.SemanticDigest, SemanticDigestBefore: base.SemanticDigest, SemanticDigestAfter: changed.SemanticDigest, GraphDigestChanged: base.GraphDigest != changed.GraphDigest, MutationFieldBefore: base.MutationField, MutationFieldAfter: changed.MutationField, MutationPayloadBefore: base.MutationPayload, MutationPayloadAfter: changed.MutationPayload, MutationDecisionBefore: base.MutationDecision, MutationDecisionAfter: changed.MutationDecision, MutationResolutionBefore: base.MutationResolution, MutationResolutionAfter: changed.MutationResolution, MutationAPIOutcomeBefore: base.MutationAPIOutcome, MutationAPIOutcomeAfter: changed.MutationAPIOutcome, DetachedGraphPatchCapabilityBefore: base.DetachedGraphPatchCapability, DetachedGraphPatchCapabilityAfter: changed.DetachedGraphPatchCapability, OverallAuthorityBefore: base.OverallAuthority, OverallAuthorityAfter: changed.OverallAuthority, GraphDigestBefore: changed.GraphDigestBefore, OriginalGraphDigestAfter: changed.OriginalGraphDigestAfter, ReturnedGraphDigestBefore: base.ReturnedGraphDigest, ReturnedGraphDigestAfter: changed.ReturnedGraphDigest, ClaimID: changed.ClaimID, ClaimStateBefore: base.ClaimState, ClaimStateAfter: changed.ClaimState, ProposalPolicyBefore: base.ProposalPolicy, ProposalPolicyAfter: changed.ProposalPolicy, ProposalSemanticDigestChanged: base.ProposalSemanticDigest != changed.ProposalSemanticDigest, ProposalSemanticDigestBefore: base.ProposalSemanticDigest, ProposalSemanticDigestAfter: changed.ProposalSemanticDigest, ProposalOutcomeBefore: base.ProposalOutcome, ProposalOutcomeAfter: changed.ProposalOutcome, ProposalTransitionBefore: base.ProposalTransition, ProposalTransitionAfter: changed.ProposalTransition}
}
func fail(format string, args ...any) { fmt.Fprintf(os.Stderr, format+"\n", args...); os.Exit(1) }
