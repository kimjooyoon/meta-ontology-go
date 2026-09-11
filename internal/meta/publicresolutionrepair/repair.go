package publicresolutionrepair

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const (
	ProposalSchema      = "gooo/public-semantic-resolution-repair-proposal/v1"
	AuthorizationSchema = "gooo/public-semantic-resolution-repair-authorization/v1"
	OverlaySchema       = "gooo/public-semantic-resolution-repair-graph-overlay/v1"
	OverlayAppliedTo    = "CALLER_OWNED_IMMUTABLE_GRAPH_OVERLAY"
	ProposalReason      = "OBSERVED_AFFECTED_TEST_AND_COMPONENT"
	AuthorizationMethod = "EXPLICIT_HUMAN_AUTHORIZATION"
	RejectionMethod     = "EXPLICIT_HUMAN_REJECTION"
)

type AuthorizationArtifact struct {
	Schema          string `json:"schema"`
	ProposalDigest  string `json:"proposal_digest"`
	Decision        string `json:"decision"`
	Method          string `json:"method"`
	HumanAuthorized bool   `json:"human_authorized"`
	Reason          string `json:"reason"`
	Digest          string `json:"digest"`
}

type GraphOverlay struct {
	Schema              string `json:"schema"`
	OverlayID           string `json:"overlay_id"`
	BaseGraphDigest     string `json:"base_graph_digest"`
	OverlayGraphDigest  string `json:"overlay_graph_digest"`
	ProposalDigest      string `json:"proposal_digest"`
	AuthorizationDigest string `json:"authorization_digest"`
	AppliedTo           string `json:"applied_to"`
	BaseEdgeCount       int    `json:"base_edge_count"`
	AddedEdgeCount      int    `json:"added_edge_count"`
	OverlayEdgeCount    int    `json:"overlay_edge_count"`
	ContinuityEdgeCount int    `json:"continuity_edge_count"`
	AddedEdges          []Edge `json:"added_edges"`
	RepositoryWrites    int    `json:"repository_writes"`
}

func SynthesizeProposal(policy Policy, counterexample Counterexample) (Proposal, error) {
	if !counterexample.Valid || counterexample.CaseID != OriginalCounterexampleCaseID || counterexample.Decision != DecisionRefuted || counterexample.ChangedComponent == "" || counterexample.OmittedTarget == "" {
		return Proposal{}, errors.New("cannot synthesize repair proposal from an invalid counterexample")
	}
	for _, edge := range policy.CanonicalEdges {
		if edge.From == counterexample.ChangedComponent && edge.To == counterexample.OmittedTarget {
			proposal := Proposal{From: edge.From, To: edge.To, ProofMode: ProofRegression, Trigger: counterexample.CaseID, Reason: ProposalReason, Method: "DETERMINISTIC_MISSING_EDGE"}
			proposal.Digest = digestWithoutField(proposal, "digest")
			return proposal, nil
		}
	}
	return Proposal{}, errors.New("counterexample does not identify a canonical missing dependency edge")
}

func NewAuthorization(proposal Proposal, decision string) AuthorizationArtifact {
	artifact := AuthorizationArtifact{Schema: AuthorizationSchema, ProposalDigest: proposal.Digest, Decision: decision, Method: RejectionMethod, HumanAuthorized: false, Reason: "EXPLICIT_HUMAN_REJECTION"}
	if decision == AuthorizationAuthorized {
		artifact.Method = AuthorizationMethod
		artifact.HumanAuthorized = true
		artifact.Reason = "EXPLICIT_HUMAN_AUTHORIZATION"
	}
	artifact.Digest = digestWithoutField(artifact, "digest")
	return artifact
}

func ValidateAuthorization(authorization AuthorizationArtifact, proposal Proposal) error {
	if authorization.Schema != AuthorizationSchema || authorization.ProposalDigest != proposal.Digest || authorization.Method == "" || authorization.Reason == "" || authorization.Digest == "" || authorization.Digest != digestWithoutField(authorization, "digest") {
		return errors.New("semantic repair authorization artifact is not content-addressed or proposal-bound")
	}
	if authorization.Decision == AuthorizationAuthorized && (!authorization.HumanAuthorized || authorization.Method != AuthorizationMethod) {
		return errors.New("authorized repair is missing explicit human authorization")
	}
	if authorization.Decision == AuthorizationRejected && (authorization.HumanAuthorized || authorization.Method != RejectionMethod) {
		return errors.New("rejected repair authorization is contradictory")
	}
	if authorization.Decision != AuthorizationAuthorized && authorization.Decision != AuthorizationRejected {
		return errors.New("semantic repair authorization decision is unknown")
	}
	return nil
}

func BuildOverlay(policy Policy, counterexample Counterexample, proposal Proposal, authorization AuthorizationArtifact) (GraphOverlay, error) {
	if err := ValidateAuthorization(authorization, proposal); err != nil {
		return GraphOverlay{}, err
	}
	if authorization.Decision != AuthorizationAuthorized || proposal.ProofMode != ProofRegression || proposal.From != counterexample.ChangedComponent || proposal.To != counterexample.OmittedTarget {
		return GraphOverlay{}, errors.New("repair overlay is not authorized by the canonical regression transition")
	}
	base := DeclaredEdges(policy, counterexample, nil)
	if len(base) != policy.GraphEdgeCountBefore {
		return GraphOverlay{}, fmt.Errorf("repair overlay base edge count=%d want=%d", len(base), policy.GraphEdgeCountBefore)
	}
	added := []Edge{{From: proposal.From, To: proposal.To}}
	combined := append(append([]Edge(nil), base...), added...)
	combined = uniqueEdges(combined)
	if len(combined) != policy.GraphEdgeCountAfter {
		return GraphOverlay{}, fmt.Errorf("repair overlay edge count=%d want=%d", len(combined), policy.GraphEdgeCountAfter)
	}
	overlay := GraphOverlay{Schema: OverlaySchema, BaseGraphDigest: graphDigest(base), OverlayGraphDigest: graphDigest(combined), ProposalDigest: proposal.Digest, AuthorizationDigest: authorization.Digest, AppliedTo: OverlayAppliedTo, BaseEdgeCount: len(base), AddedEdgeCount: len(added), OverlayEdgeCount: len(combined), ContinuityEdgeCount: len(combined), AddedEdges: added, RepositoryWrites: 0}
	overlay.OverlayID = digestWithoutField(overlay, "overlay_id")
	return overlay, nil
}

func ValidateOverlay(overlay GraphOverlay, policy Policy, proposal Proposal, authorization AuthorizationArtifact) error {
	if overlay.Schema != OverlaySchema || overlay.OverlayID == "" || overlay.AppliedTo != OverlayAppliedTo || overlay.RepositoryWrites != 0 || overlay.ProposalDigest != proposal.Digest || overlay.AuthorizationDigest != authorization.Digest || overlay.BaseEdgeCount != policy.GraphEdgeCountBefore || overlay.OverlayEdgeCount != policy.GraphEdgeCountAfter || overlay.AddedEdgeCount != 1 || overlay.ContinuityEdgeCount != policy.ContinuityEdgeCount || overlay.OverlayID != digestWithoutField(overlay, "overlay_id") {
		return errors.New("semantic repair graph overlay is invalid or repository-bound")
	}
	if len(overlay.AddedEdges) != 1 || overlay.AddedEdges[0].From != proposal.From || overlay.AddedEdges[0].To != proposal.To {
		return errors.New("semantic repair graph overlay does not bind the proposal")
	}
	return nil
}

func digestWithoutField(value any, field string) string {
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	var object map[string]any
	if err := json.Unmarshal(data, &object); err != nil {
		return ""
	}
	delete(object, field)
	data, err = json.Marshal(object)
	if err != nil {
		return ""
	}
	return cache.HashBytes(data).String()
}

func DeclaredEdges(policy Policy, counterexample Counterexample, overlay *GraphOverlay) []Edge {
	edges := append([]Edge(nil), policy.CanonicalEdges...)
	if counterexample.OmittedTarget != "" {
		filtered := make([]Edge, 0, len(edges))
		for _, edge := range edges {
			if edge.From == counterexample.ChangedComponent && edge.To == counterexample.OmittedTarget {
				continue
			}
			filtered = append(filtered, edge)
		}
		edges = filtered
	}
	if overlay != nil {
		edges = uniqueEdges(append(edges, overlay.AddedEdges...))
	}
	return edges
}

func uniqueEdges(edges []Edge) []Edge {
	seen := map[string]bool{}
	result := make([]Edge, 0, len(edges))
	for _, edge := range edges {
		key := edge.From + ">" + edge.To
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, edge)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].From+">"+result[i].To < result[j].From+">"+result[j].To })
	return result
}
