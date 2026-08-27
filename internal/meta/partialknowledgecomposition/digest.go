package partialknowledgecomposition

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) string {
	raw, _ := json.Marshal(value)
	return digestBytes(raw)
}

func receiptDigest(receipt Receipt) string {
	receipt.Digest = ""
	return digestValue(receipt)
}

func transitionDigest(transition ClaimTransition) string {
	transition.Digest = ""
	return digestValue(transition)
}

func rawEvidenceReceiptDigest(receipt RawEvidenceReceipt) string {
	receipt.Digest = ""
	return digestValue(receipt)
}

func snapshotDigest(snapshot Snapshot) string {
	return digestValue(struct {
		Tracked   []string `json:"tracked"`
		Untracked []string `json:"untracked"`
		Status    []string `json:"status"`
	}{snapshot.Tracked, snapshot.Untracked, snapshot.Status})
}

func evidenceDigest(evidence Evidence) string {
	evidence.EvidenceDigest = ""
	return digestValue(evidence)
}

func workspaceEvidenceDigest(workspace WorkspaceObservation) string {
	workspace.EvidenceDigest = ""
	return digestValue(workspace)
}

func capabilityEvidenceDigest(capability CapabilityObservation) string {
	capability.EvidenceDigest = ""
	return digestValue(capability)
}

func upstreamClaimEvidenceDigest(claim UpstreamClaim) string {
	return digestValue(struct {
		ClaimID                 string `json:"claim_id"`
		Proposition             string `json:"proposition"`
		PropositionDigest       string `json:"proposition_digest"`
		Predicate               string `json:"predicate"`
		State                   string `json:"state"`
		Resolution              string `json:"resolution"`
		Stage                   string `json:"stage"`
		Step                    string `json:"step"`
		Reason                  string `json:"reason"`
		RawSourceDigest         string `json:"raw_source_digest"`
		SemanticDigest          string `json:"semantic_digest"`
		WorkspaceSnapshotDigest string `json:"workspace_snapshot_digest"`
		TargetOperation         string `json:"target_operation"`
		TargetOutput            string `json:"target_output"`
	}{claim.ClaimID, claim.Proposition, claim.PropositionDigest, claim.Predicate, claim.State, claim.Resolution, claim.Stage, claim.Step, claim.Reason, claim.RawSourceDigest, claim.SemanticDigest, claim.WorkspaceSnapshotDigest, claim.TargetOperation, claim.TargetOutput})
}

func semanticProjectionDigest(receipt Receipt) string {
	type semanticCase struct {
		ID                string `json:"id"`
		Result            Value  `json:"result"`
		Predicate         string `json:"predicate"`
		Proposition       string `json:"proposition"`
		PropositionDigest string `json:"proposition_digest"`
		TargetAddress     string `json:"target_address"`
		TargetOperation   string `json:"target_operation"`
		TargetOutput      string `json:"target_output"`
		Decision          string `json:"decision"`
		Resolution        string `json:"resolution"`
		Stage             string `json:"stage"`
		Step              string `json:"step"`
		Reason            string `json:"reason"`
		TopSuccess        bool   `json:"top_success"`
	}
	type semanticClaim struct {
		ClaimID           string `json:"claim_id"`
		From              string `json:"from"`
		To                string `json:"to"`
		Predicate         string `json:"predicate"`
		Proposition       string `json:"proposition"`
		PropositionDigest string `json:"proposition_digest"`
		TargetAddress     string `json:"target_address"`
		TargetOperation   string `json:"target_operation"`
		TargetOutput      string `json:"target_output"`
		Stage             string `json:"stage"`
		Step              string `json:"step"`
		Reason            string `json:"reason"`
	}
	cases := make([]semanticCase, 0, len(receipt.Cases))
	for _, current := range receipt.Cases {
		cases = append(cases, semanticCase{current.ID, current.Result, current.Predicate, current.Proposition, current.PropositionDigest, current.TargetAddress, current.TargetOperation, current.TargetOutput, current.Decision, current.Resolution, current.Stage, current.Step, current.Reason, current.TopSuccess})
	}
	claims := make([]semanticClaim, 0, len(receipt.Claims))
	for _, current := range receipt.Claims {
		claims = append(claims, semanticClaim{current.ClaimID, current.From, current.To, current.Predicate, current.Proposition, current.PropositionDigest, current.TargetAddress, current.TargetOperation, current.TargetOutput, current.Stage, current.Step, current.Reason})
	}
	return digestValue(struct {
		SemanticIRDigest string          `json:"semantic_ir_digest"`
		Cases            []semanticCase  `json:"cases"`
		Claims           []semanticClaim `json:"claims"`
		Summary          Summary         `json:"summary"`
	}{receipt.SemanticIRDigest, cases, claims, receipt.Summary})
}
