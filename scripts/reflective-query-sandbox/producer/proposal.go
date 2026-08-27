package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	sandbox "github.com/kimjooyoon/meta-ontology-go/internal/meta/reflectivequerysandbox"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	proposalManifestSchema = "gooo/reflective-query-sandbox/proposal-receipt/v1"
	proposalArtifactSchema = "gooo/reflective-query-sandbox/refinement-proposal/v1"
	proposalProducerName   = "reflective-query-sandbox.proposal-producer"
)

type proposalSourceModel struct {
	ProposalEntity semantic.ID
	Policy         semantic.ID
	Authority      semantic.ID
	MetaOperation  semantic.ID
	MetaProgram    string
}

type proposalCase struct {
	CaseID                   string     `json:"case_id"`
	ObservationKind          string     `json:"observation_kind"`
	Outcome                  string     `json:"outcome"`
	OutcomeReason            string     `json:"outcome_reason"`
	ProposalEmitted          bool       `json:"proposal_emitted"`
	ProposalCoordinate       coordinate `json:"proposal_coordinate"`
	RejectionCoordinate      coordinate `json:"rejection_coordinate"`
	Path                     string     `json:"path"`
	Bytes                    int        `json:"bytes"`
	BytesDigest              string     `json:"bytes_digest"`
	SemanticDigest           string     `json:"semantic_digest"`
	SourceRawDigest          string     `json:"source_raw_digest"`
	SourceSemanticDigest     string     `json:"source_semantic_digest"`
	QueryReceiptDigest       string     `json:"query_receipt_digest"`
	ClaimID                  string     `json:"claim_id"`
	TargetSemanticAddress    string     `json:"target_semantic_address"`
	UnknownStage             string     `json:"unknown_stage"`
	UnknownStep              string     `json:"unknown_step"`
	UnknownReason            string     `json:"unknown_reason"`
	RequestedRefinement      string     `json:"requested_refinement"`
	ProofChoice              string     `json:"proof_choice"`
	MetaOperation            string     `json:"meta_operation"`
	Authority                string     `json:"authority"`
	ObservedClaimFrom        string     `json:"observed_claim_from"`
	ObservedClaimTo          string     `json:"observed_claim_to"`
	ObservedClaimReason      string     `json:"observed_claim_reason"`
	ProposalTransitionFrom   string     `json:"proposal_transition_from"`
	ProposalTransitionTo     string     `json:"proposal_transition_to"`
	ProposalTransitionReason string     `json:"proposal_transition_reason"`
}

type proposalManifest struct {
	Schema                string         `json:"schema"`
	SourcePath            string         `json:"source_path"`
	SourceRawDigest       string         `json:"source_raw_digest"`
	SourceSemanticDigest  string         `json:"source_semantic_digest"`
	QueryReceiptDigest    string         `json:"query_receipt_digest"`
	Authority             string         `json:"authority"`
	RepositoryWrites      int            `json:"repository_writes"`
	MutationAuthority     bool           `json:"mutation_authority"`
	Producer              string         `json:"producer"`
	ProposalCount         int            `json:"proposal_count"`
	Emitted               coordinate     `json:"emitted"`
	Rejected              coordinate     `json:"rejected"`
	Cases                 []proposalCase `json:"cases"`
	PortableReceiptDigest string         `json:"portable_receipt_digest"`
}

type coordinate struct {
	Satisfied int `json:"satisfied"`
	Total     int `json:"total"`
}

func writeProposalArtifacts(sourcePath, proposalDir, manifestPath string, observation sandbox.Observation) error {
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read proposal source: %w", err)
	}
	model, err := deriveProposalSourceModel(sourcePath, data)
	if err != nil {
		return err
	}
	cases, err := proposalCases(observation, model)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(proposalDir, 0o755); err != nil {
		return fmt.Errorf("create proposal directory: %w", err)
	}
	for index := range cases {
		cases[index].ProposalCoordinate.Total = 1
		cases[index].RejectionCoordinate.Total = 1
		if cases[index].ProposalEmitted {
			cases[index].ProposalCoordinate.Satisfied = 1
		}
		if cases[index].Outcome == "REJECTED" {
			cases[index].RejectionCoordinate.Satisfied = 1
		}
		raw := renderProposal(cases[index], model)
		relative := filepath.ToSlash(filepath.Join("proposals", caseFileName(cases[index].CaseID)))
		path := filepath.Join(proposalDir, caseFileName(cases[index].CaseID))
		if err := os.WriteFile(path, raw, 0o644); err != nil {
			return fmt.Errorf("write proposal %s: %w", cases[index].CaseID, err)
		}
		semanticDigest, err := proposalSemanticDigest(path, raw)
		if err != nil {
			return fmt.Errorf("lower proposal %s: %w", cases[index].CaseID, err)
		}
		cases[index].Path = relative
		cases[index].Bytes = len(raw)
		cases[index].BytesDigest = semantic.StableHash(raw)
		cases[index].SemanticDigest = semanticDigest
	}
	manifest := proposalManifest{
		Schema: proposalManifestSchema, SourcePath: sourcePath,
		SourceRawDigest: observation.Source.SourceDigest, SourceSemanticDigest: observation.Source.SemanticDigest,
		QueryReceiptDigest: observation.ProvisionalDigest, Authority: "NONE", RepositoryWrites: 0,
		MutationAuthority: false, Producer: proposalProducerName, ProposalCount: len(cases), Cases: cases,
	}
	for _, item := range cases {
		if item.ProposalEmitted {
			manifest.Emitted.Satisfied++
		}
		manifest.Emitted.Total++
		if item.Outcome == "REJECTED" {
			manifest.Rejected.Satisfied++
		}
		manifest.Rejected.Total++
	}
	manifest.PortableReceiptDigest = sealProposalManifest(manifest)
	encoded, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("encode proposal receipt: %w", err)
	}
	if err := os.WriteFile(manifestPath, append(encoded, '\n'), 0o644); err != nil {
		return fmt.Errorf("write proposal receipt: %w", err)
	}
	return nil
}

func deriveProposalSourceModel(path string, data []byte) (proposalSourceModel, error) {
	file, diagnostics := syntax.ParseFile(path, string(data))
	if diagnostics.HasErrors() {
		return proposalSourceModel{}, diagnostics.Error()
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return proposalSourceModel{}, fmt.Errorf("lower proposal policy source: %w", err)
	}
	model := proposalSourceModel{}
	for _, node := range ir.Graph.Nodes() {
		id := node.ID.String()
		switch {
		case node.Kind == semantic.Entity && strings.Contains(id, "/proposal/refinement"):
			model.ProposalEntity = node.ID
		case node.Kind == semantic.Entity && strings.Contains(id, "/proposal/policy/"):
			model.Policy = node.ID
		case node.Kind == semantic.Entity && strings.Contains(id, "/proposal/authority/"):
			model.Authority = node.ID
		case node.Kind == semantic.Entity && strings.Contains(id, "/proposal/meta-operation/"):
			model.MetaOperation = node.ID
		case node.Kind == semantic.Activity && node.ValueProgram == "reflect.meta:proposal:refine-unknown":
			model.MetaProgram = node.ValueProgram
		}
	}
	if model.ProposalEntity == "" || model.Policy == "" || model.Authority == "" || model.MetaOperation == "" || model.MetaProgram == "" {
		return proposalSourceModel{}, fmt.Errorf("source proposal policy is incomplete")
	}
	if tail(model.Authority) != "NONE" {
		return proposalSourceModel{}, fmt.Errorf("proposal authority is not NONE")
	}
	return model, nil
}

func proposalCases(observation sandbox.Observation, model proposalSourceModel) ([]proposalCase, error) {
	exact, ok := findProposalAttempt(observation.Attempts, func(attempt sandbox.Attempt) bool {
		return attempt.Operation == "query" && attempt.Decision == "PASS" && attempt.Resolution == "EXACT" && len(attempt.EvidenceClaimIDs) > 0
	})
	if !ok {
		return nil, fmt.Errorf("exact observation attempt is missing")
	}
	unknown, ok := findProposalAttempt(observation.Attempts, func(attempt sandbox.Attempt) bool { return attempt.ID == "unknown.target" })
	if !ok {
		return nil, fmt.Errorf("unknown observation attempt is missing")
	}
	mutation, ok := findProposalAttempt(observation.Attempts, func(attempt sandbox.Attempt) bool { return attempt.ID == "mutation.attempt" })
	if !ok {
		return nil, fmt.Errorf("mutation observation attempt is missing")
	}
	policy := tail(model.Policy)
	unknownOutcome, unknownReason := "REJECTED", "UNKNOWN_POLICY_REJECTS_REFINEMENT"
	if policy == "unknown-preserve-open" {
		unknownOutcome, unknownReason = "EMITTED", "UNKNOWN_REFINE_PROPOSAL_EMITTED"
	}
	exactClaim := observedClaim(observation.Claims, exact.EvidenceClaimIDs[0])
	unknownClaim := observedClaim(observation.Claims, firstClaimID(unknown, "guardrail.unknown-closed"))
	mutationClaim := observedClaim(observation.Claims, firstClaimID(mutation, "outcome.immutable-id-patch-rejected"))
	base := func(caseID, kind, outcome, reason string, emitted bool, attempt sandbox.Attempt, claim sandbox.ClaimTransition, requested string) proposalCase {
		return proposalCase{
			CaseID: caseID, ObservationKind: kind, Outcome: outcome, OutcomeReason: reason, ProposalEmitted: emitted,
			SourceRawDigest: observation.Source.SourceDigest, SourceSemanticDigest: observation.Source.SemanticDigest,
			QueryReceiptDigest: observation.ProvisionalDigest, ClaimID: claim.ClaimID, TargetSemanticAddress: attempt.Target,
			UnknownStage: attempt.Stage, UnknownStep: attempt.Step, UnknownReason: attempt.Reason,
			RequestedRefinement: requested, ProofChoice: attempt.ProofChoice, MetaOperation: model.MetaProgram,
			Authority: "NONE", ObservedClaimFrom: claim.From, ObservedClaimTo: claim.To, ObservedClaimReason: claim.Reason,
		}
	}
	cases := []proposalCase{
		base("exact-observation", "EXACT", "NOT_EMITTED", "EXACT_OBSERVATION_NEEDS_NO_REFINEMENT", false, exact, exactClaim, "none-exact-observation"),
		base("unknown-observation", "UNKNOWN", unknownOutcome, unknownReason, unknownOutcome == "EMITTED", unknown, unknownClaim, "refine-unknown-target-resolution"),
		base("mutation-request", "MUTATION_REQUEST", "REJECTED", "MUTATION_REQUEST_IS_NOT_A_REFINEMENT", false, mutation, mutationClaim, "mutation-not-authorized"),
	}
	cases[1].ProposalTransitionFrom, cases[1].ProposalTransitionTo, cases[1].ProposalTransitionReason = "OPEN", "OPEN", "UNKNOWN_PRESERVED_OPEN"
	return cases, nil
}

func renderProposal(item proposalCase, model proposalSourceModel) []byte {
	value := strings.Join([]string{
		"reflect.meta:proposal-artifact", "schema=" + proposalArtifactSchema, "case=" + item.CaseID,
		"observation=" + item.ObservationKind, "proposal.outcome=" + item.Outcome,
		"proposal.reason=" + item.OutcomeReason, fmt.Sprintf("proposal.emitted=%t", item.ProposalEmitted),
		"claim=" + item.ClaimID, "target=" + item.TargetSemanticAddress, "source.path=" + sandbox.SourcePath,
		"source.raw=" + item.SourceRawDigest, "source.semantic=" + item.SourceSemanticDigest,
		"query.receipt=" + item.QueryReceiptDigest, "unknown.stage=" + item.UnknownStage,
		"unknown.step=" + item.UnknownStep, "unknown.reason=" + item.UnknownReason,
		"requested.refinement=" + item.RequestedRefinement, "proof.choice=" + item.ProofChoice,
		"meta.operation=" + item.MetaOperation, "authority=" + item.Authority,
		"policy=" + tail(model.Policy), "claim.from=" + item.ObservedClaimFrom,
		"claim.to=" + item.ObservedClaimTo, "claim.reason=" + item.ObservedClaimReason,
		"proposal.from=" + item.ProposalTransitionFrom, "proposal.to=" + item.ProposalTransitionTo,
		"proposal.transition.reason=" + item.ProposalTransitionReason,
	}, ";")
	return []byte(fmt.Sprintf(`package reflectivequeryproposal
namespace reflectivequeryproposal

entity SourceRawDigest_%s id "gooo://reflective-query-sandbox/proposal-artifact/source-raw"
entity RefinementProposal id "gooo://reflective-query-sandbox/proposal-artifact/case/%s/outcome/%s"
activity RecordProposal(RefinementProposal) -> RefinementProposal computes "%s"
`, safeToken(item.SourceRawDigest), safeToken(item.CaseID), strings.ToLower(safeToken(item.Outcome)), valueWithoutRawDigest(value)))
}

func valueWithoutRawDigest(value string) string {
	parts := strings.Split(value, ";")
	filtered := parts[:0]
	for _, part := range parts {
		if !strings.HasPrefix(part, "source.raw=") {
			filtered = append(filtered, part)
		}
	}
	return strings.Join(filtered, ";")
}

func proposalSemanticDigest(path string, raw []byte) (string, error) {
	file, diagnostics := syntax.ParseFile(path, string(raw))
	if diagnostics.HasErrors() {
		return "", diagnostics.Error()
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return "", err
	}
	return ir.StableHash(), nil
}

func sealProposalManifest(manifest proposalManifest) string {
	manifest.PortableReceiptDigest = ""
	raw, err := json.Marshal(manifest)
	if err != nil {
		panic(err)
	}
	return semantic.StableHash(raw)
}

func findProposalAttempt(attempts []sandbox.Attempt, predicate func(sandbox.Attempt) bool) (sandbox.Attempt, bool) {
	for _, attempt := range attempts {
		if predicate(attempt) {
			return attempt, true
		}
	}
	return sandbox.Attempt{}, false
}

func observedClaim(claims []sandbox.ClaimTransition, id string) sandbox.ClaimTransition {
	var result sandbox.ClaimTransition
	for _, claim := range claims {
		if claim.ClaimID == id && claim.Sequence > result.Sequence {
			result = claim
		}
	}
	return result
}

func firstClaimID(attempt sandbox.Attempt, fallback string) string {
	if len(attempt.EvidenceClaimIDs) > 0 {
		return attempt.EvidenceClaimIDs[0]
	}
	return fallback
}

func caseFileName(caseID string) string { return safeToken(caseID) + ".gooo" }

func safeToken(value string) string {
	var builder strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			builder.WriteRune(r)
		} else {
			builder.WriteByte('-')
		}
	}
	return builder.String()
}

func tail(id semantic.ID) string {
	parts := strings.Split(strings.TrimSuffix(id.String(), "/"), "/")
	return parts[len(parts)-1]
}
