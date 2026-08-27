package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	canonicalSourcePath    = "examples/reflective-query-sandbox/main.gooo"
	proposalManifestSchema = "gooo/reflective-query-sandbox/proposal-receipt/v1"
	proposalArtifactSchema = "gooo/reflective-query-sandbox/refinement-proposal/v1"
)

type coordinate struct {
	Satisfied int `json:"satisfied"`
	Total     int `json:"total"`
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

type sourceSnapshot struct {
	Path           string `json:"path"`
	SourceDigest   string `json:"source_digest"`
	SemanticDigest string `json:"semantic_digest"`
}

type observedAttempt struct {
	ID               string   `json:"id"`
	Operation        string   `json:"operation"`
	Decision         string   `json:"decision"`
	Resolution       string   `json:"resolution"`
	Reason           string   `json:"reason"`
	Stage            string   `json:"stage"`
	Step             string   `json:"step"`
	ProofChoice      string   `json:"proof_choice"`
	Target           string   `json:"target"`
	EvidenceClaimIDs []string `json:"evidence_claim_ids"`
}

type observedClaim struct {
	Sequence        int    `json:"sequence"`
	ClaimID         string `json:"claim_id"`
	EvidenceAttempt string `json:"evidence_attempt"`
	From            string `json:"from"`
	To              string `json:"to"`
	Reason          string `json:"reason"`
}

type observation struct {
	Schema            string            `json:"schema"`
	ProvisionalDigest string            `json:"provisional_digest"`
	Source            sourceSnapshot    `json:"source"`
	Attempts          []observedAttempt `json:"attempts"`
	Claims            []observedClaim   `json:"claims"`
}

type sourceModel struct {
	ProposalEntity string
	Policy         string
	Authority      string
	MetaOperation  string
	MetaProgram    string
}

type proposalResult struct {
	CaseID                 string     `json:"case_id"`
	Outcome                string     `json:"outcome"`
	ProposalEmitted        bool       `json:"proposal_emitted"`
	ProposalCoordinate     coordinate `json:"proposal_coordinate"`
	RejectionCoordinate    coordinate `json:"rejection_coordinate"`
	ClaimID                string     `json:"claim_id"`
	ClaimTransition        string     `json:"claim_transition"`
	ProposalTransition     string     `json:"proposal_transition"`
	Path                   string     `json:"path"`
	Bytes                  int        `json:"bytes"`
	BytesDigest            string     `json:"bytes_digest"`
	ProducerSemanticDigest string     `json:"producer_semantic_digest"`
	ConsumerSemanticDigest string     `json:"consumer_semantic_digest"`
}

type rejection struct {
	Name       string `json:"name"`
	Decision   string `json:"decision"`
	Resolution string `json:"resolution"`
	Stage      string `json:"stage"`
	Step       string `json:"step"`
	Reason     string `json:"reason"`
}

type directionalBoundary struct {
	ForbiddenImportsObserved int        `json:"forbidden_imports_observed"`
	MaximumAllowed           int        `json:"maximum_allowed"`
	Conformance              coordinate `json:"conformance"`
	EvidenceDigest           string     `json:"evidence_digest"`
}

type verification struct {
	Schema                    string              `json:"schema"`
	Decision                  string              `json:"decision"`
	Resolution                string              `json:"resolution"`
	Reason                    string              `json:"reason"`
	SourcePath                string              `json:"source_path"`
	SourceRawDigest           string              `json:"source_raw_digest"`
	SourceSemanticDigest      string              `json:"source_semantic_digest"`
	QueryReceiptDigest        string              `json:"query_receipt_digest"`
	Authority                 string              `json:"authority"`
	RepositoryWrites          int                 `json:"repository_writes"`
	MutationAuthority         bool                `json:"mutation_authority"`
	Cases                     []proposalResult    `json:"cases"`
	Emitted                   coordinate          `json:"emitted"`
	Rejected                  coordinate          `json:"rejected"`
	GeneratedArtifactCount    coordinate          `json:"generated_artifact_count"`
	GeneratedReconsumption    coordinate          `json:"generated_reconsumption"`
	OpenClaimTransition       coordinate          `json:"open_claim_transition"`
	Counterexamples           []rejection         `json:"counterexamples"`
	CounterexampleDenominator int                 `json:"counterexample_denominator"`
	DirectionalImportBoundary directionalBoundary `json:"directional_import_boundary"`
	PortableReceiptDigest     string              `json:"portable_receipt_digest"`
}

type validationFault struct {
	Stage  string
	Step   string
	Reason string
}

func (e validationFault) Error() string { return e.Stage + "/" + e.Step + "/" + e.Reason }

func main() {
	input := flag.String("input", "", "producer observation directory")
	sourcePath := flag.String("source", "", "canonical Gooo source")
	output := flag.String("output", "", "independent proposal verification receipt")
	importsPath := flag.String("producer-imports-evidence", "", "consumer dependency evidence")
	maximum := flag.Int("producer-imports-maximum", 0, "maximum forbidden producer dependency count")
	flag.Parse()
	if *input == "" || *sourcePath != canonicalSourcePath || *output == "" || *importsPath == "" {
		fail("usage: proposal-consumer -input DIR -source examples/reflective-query-sandbox/main.gooo -output FILE -producer-imports-evidence FILE -producer-imports-maximum N")
	}
	verification, err := verify(*input, *sourcePath, *importsPath, *maximum)
	if err != nil {
		fail("proposal verification: %v", err)
	}
	data, err := json.MarshalIndent(verification, "", "  ")
	if err != nil {
		fail("encode proposal verification: %v", err)
	}
	if err := os.WriteFile(*output, append(data, '\n'), 0o644); err != nil {
		fail("write proposal verification: %v", err)
	}
	fmt.Printf("proposal consumer: decision=%s cases=%d emitted=%d/%d rejected=%d/%d reconsumption=%d/%d open=%d/%d counterexamples=%d/%d\n", verification.Decision, len(verification.Cases), verification.Emitted.Satisfied, verification.Emitted.Total, verification.Rejected.Satisfied, verification.Rejected.Total, verification.GeneratedReconsumption.Satisfied, verification.GeneratedReconsumption.Total, verification.OpenClaimTransition.Satisfied, verification.OpenClaimTransition.Total, len(verification.Counterexamples), verification.CounterexampleDenominator)
}

func verify(input, sourcePath, importsPath string, maximum int) (verification, error) {
	observationData, err := os.ReadFile(filepath.Join(input, "observation.json"))
	if err != nil {
		return verification{}, err
	}
	var value observation
	if err := json.Unmarshal(observationData, &value); err != nil {
		return verification{}, err
	}
	manifestData, err := os.ReadFile(filepath.Join(input, "proposal-receipt.json"))
	if err != nil {
		return verification{}, err
	}
	var manifest proposalManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return verification{}, err
	}
	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return verification{}, err
	}
	model, sourceSemantic, err := deriveSourceModel(sourcePath, sourceData)
	if err != nil {
		return verification{}, err
	}
	if value.Schema != "gooo/reflective-query-sandbox-observation/v4" || manifest.Schema != proposalManifestSchema {
		return verification{}, errors.New("proposal input schema mismatch")
	}
	if manifest.SourcePath != sourcePath || manifest.SourceRawDigest != semantic.StableHash(sourceData) || manifest.SourceSemanticDigest != sourceSemantic || manifest.QueryReceiptDigest != value.ProvisionalDigest {
		return verification{}, errors.New("proposal receipt source or query binding mismatch")
	}
	if manifest.Authority != "NONE" || manifest.RepositoryWrites != 0 || manifest.MutationAuthority {
		return verification{}, errors.New("proposal receipt exceeds non-authoritative boundary")
	}
	if sealProposalManifest(manifest) != manifest.PortableReceiptDigest {
		return verification{}, errors.New("portable proposal receipt digest mismatch")
	}
	if len(manifest.Cases) != 3 || manifest.ProposalCount != 3 {
		return verification{}, errors.New("proposal fixed case denominator is not three")
	}
	expected, err := expectedCases(value, model, sourceSemantic)
	if err != nil {
		return verification{}, err
	}
	results := make([]proposalResult, 0, len(expected))
	emitted, rejected, open := coordinate{}, coordinate{}, coordinate{}
	for index, want := range expected {
		if err := validateCaseBinding(manifest.Cases[index], want); err != nil {
			return verification{}, fmt.Errorf("case %s: %w", want.CaseID, err)
		}
		item := manifest.Cases[index]
		artifactPath, err := safeArtifactPath(input, item.Path, item.CaseID)
		if err != nil {
			return verification{}, err
		}
		raw, err := os.ReadFile(artifactPath)
		if err != nil {
			return verification{}, err
		}
		artifactDigest, metadata, err := parseProposalArtifact(artifactPath, raw)
		if err != nil {
			return verification{}, fmt.Errorf("case %s artifact: %w", item.CaseID, err)
		}
		if artifactDigest != item.SemanticDigest || semantic.StableHash(raw) != item.BytesDigest || len(raw) != item.Bytes {
			return verification{}, fmt.Errorf("case %s artifact digest or byte binding mismatch", item.CaseID)
		}
		if err := validateArtifactMetadata(metadata, item); err != nil {
			return verification{}, fmt.Errorf("case %s artifact metadata: %w", item.CaseID, err)
		}
		result := proposalResult{CaseID: item.CaseID, Outcome: item.Outcome, ProposalEmitted: item.ProposalEmitted, ClaimID: item.ClaimID, ClaimTransition: item.ObservedClaimFrom + "->" + item.ObservedClaimTo, ProposalTransition: item.ProposalTransitionFrom + "->" + item.ProposalTransitionTo, Path: item.Path, Bytes: item.Bytes, BytesDigest: item.BytesDigest, ProducerSemanticDigest: item.SemanticDigest, ConsumerSemanticDigest: artifactDigest}
		result.ProposalCoordinate = item.ProposalCoordinate
		result.RejectionCoordinate = item.RejectionCoordinate
		results = append(results, result)
		emitted.Total++
		if item.ProposalEmitted {
			emitted.Satisfied++
		}
		rejected.Total++
		if item.Outcome == "REJECTED" {
			rejected.Satisfied++
		}
		if item.CaseID == "unknown-observation" {
			open.Total++
			if item.ObservedClaimFrom == "OPEN" && item.ObservedClaimTo == "OPEN" && item.ProposalTransitionFrom == "OPEN" && item.ProposalTransitionTo == "OPEN" {
				open.Satisfied++
			}
		}
	}
	if emitted != manifest.Emitted || rejected != manifest.Rejected {
		return verification{}, errors.New("proposal emission coordinates differ from producer receipt")
	}
	if open != (coordinate{Satisfied: 1, Total: 1}) {
		return verification{}, errors.New("unknown claim was not preserved OPEN->OPEN")
	}
	counterexamples := buildCounterexamples(manifest.Cases, expected)
	for _, item := range counterexamples {
		if item.Decision != "REFUTED" || item.Resolution != "EXACT" || item.Stage == "" || item.Step == "" || item.Reason == "" {
			return verification{}, errors.New("counterexample did not fail closed at exact coordinates")
		}
	}
	importsData, err := os.ReadFile(importsPath)
	if err != nil {
		return verification{}, err
	}
	observedImports := 0
	for _, line := range strings.Split(string(importsData), "\n") {
		if strings.Contains(line, "/scripts/reflective-query-sandbox/producer") {
			observedImports++
		}
	}
	boundary := directionalBoundary{ForbiddenImportsObserved: observedImports, MaximumAllowed: maximum, Conformance: coordinate{Total: 1}, EvidenceDigest: semantic.StableHash(importsData)}
	if observedImports <= maximum {
		boundary.Conformance.Satisfied = 1
	} else {
		return verification{}, errors.New("directional producer import boundary exceeded")
	}
	return verification{
		Schema: "gooo/reflective-query-sandbox/proposal-verification/v1", Decision: "PROPOSAL_ONLY", Resolution: "EXACT_RECONSTRUCTION", Reason: "PROPOSAL_EMITTED_WITHOUT_PROMOTION",
		SourcePath: sourcePath, SourceRawDigest: semantic.StableHash(sourceData), SourceSemanticDigest: sourceSemantic, QueryReceiptDigest: value.ProvisionalDigest,
		Authority: "NONE", RepositoryWrites: 0, MutationAuthority: false, Cases: results, Emitted: emitted, Rejected: rejected,
		GeneratedArtifactCount: coordinate{Satisfied: len(results), Total: len(expected)}, GeneratedReconsumption: coordinate{Satisfied: 1, Total: 1}, OpenClaimTransition: open,
		Counterexamples: counterexamples, CounterexampleDenominator: 5, DirectionalImportBoundary: boundary, PortableReceiptDigest: manifest.PortableReceiptDigest,
	}, nil
}

func deriveSourceModel(path string, data []byte) (sourceModel, string, error) {
	file, diagnostics := syntax.ParseFile(path, string(data))
	if diagnostics.HasErrors() {
		return sourceModel{}, "", diagnostics.Error()
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return sourceModel{}, "", err
	}
	model := sourceModel{}
	for _, node := range ir.Graph.Nodes() {
		id := node.ID.String()
		switch {
		case node.Kind == semantic.Entity && strings.Contains(id, "/proposal/refinement"):
			model.ProposalEntity = id
		case node.Kind == semantic.Entity && strings.Contains(id, "/proposal/policy/"):
			model.Policy = id
		case node.Kind == semantic.Entity && strings.Contains(id, "/proposal/authority/"):
			model.Authority = id
		case node.Kind == semantic.Entity && strings.Contains(id, "/proposal/meta-operation/"):
			model.MetaOperation = id
		case node.Kind == semantic.Activity && node.ValueProgram == "reflect.meta:proposal:refine-unknown":
			model.MetaProgram = node.ValueProgram
		}
	}
	if model.ProposalEntity == "" || model.Policy == "" || model.Authority == "" || model.MetaOperation == "" || model.MetaProgram == "" || tail(model.Authority) != "NONE" {
		return sourceModel{}, "", errors.New("source proposal declarations are incomplete")
	}
	return model, ir.StableHash(), nil
}

func expectedCases(value observation, model sourceModel, sourceSemantic string) ([]proposalCase, error) {
	exact, ok := findAttempt(value.Attempts, func(item observedAttempt) bool {
		return item.Operation == "query" && item.Decision == "PASS" && item.Resolution == "EXACT" && len(item.EvidenceClaimIDs) > 0
	})
	if !ok {
		return nil, errors.New("exact observation is missing")
	}
	unknown, ok := findAttempt(value.Attempts, func(item observedAttempt) bool { return item.ID == "unknown.target" })
	if !ok {
		return nil, errors.New("unknown observation is missing")
	}
	mutation, ok := findAttempt(value.Attempts, func(item observedAttempt) bool { return item.ID == "mutation.attempt" })
	if !ok {
		return nil, errors.New("mutation observation is missing")
	}
	unknownOutcome, unknownReason := "REJECTED", "UNKNOWN_POLICY_REJECTS_REFINEMENT"
	if tail(model.Policy) == "unknown-preserve-open" {
		unknownOutcome, unknownReason = "EMITTED", "UNKNOWN_REFINE_PROPOSAL_EMITTED"
	}
	makeCase := func(id, kind, outcome, reason string, emitted bool, attempt observedAttempt, claimID, requested string) proposalCase {
		claim := latestClaim(value.Claims, claimID)
		return proposalCase{CaseID: id, ObservationKind: kind, Outcome: outcome, OutcomeReason: reason, ProposalEmitted: emitted, Path: "proposals/" + safeToken(id) + ".gooo", SourceRawDigest: value.Source.SourceDigest, SourceSemanticDigest: sourceSemantic, QueryReceiptDigest: value.ProvisionalDigest, ClaimID: claimID, TargetSemanticAddress: attempt.Target, UnknownStage: attempt.Stage, UnknownStep: attempt.Step, UnknownReason: attempt.Reason, RequestedRefinement: requested, ProofChoice: attempt.ProofChoice, MetaOperation: model.MetaProgram, Authority: "NONE", ObservedClaimFrom: claim.From, ObservedClaimTo: claim.To, ObservedClaimReason: claim.Reason}
	}
	exactClaimID := exact.EvidenceClaimIDs[0]
	unknownClaimID := firstClaimID(unknown, "guardrail.unknown-closed")
	mutationClaimID := firstClaimID(mutation, "outcome.immutable-id-patch-rejected")
	result := []proposalCase{
		makeCase("exact-observation", "EXACT", "NOT_EMITTED", "EXACT_OBSERVATION_NEEDS_NO_REFINEMENT", false, exact, exactClaimID, "none-exact-observation"),
		makeCase("unknown-observation", "UNKNOWN", unknownOutcome, unknownReason, unknownOutcome == "EMITTED", unknown, unknownClaimID, "refine-unknown-target-resolution"),
		makeCase("mutation-request", "MUTATION_REQUEST", "REJECTED", "MUTATION_REQUEST_IS_NOT_A_REFINEMENT", false, mutation, mutationClaimID, "mutation-not-authorized"),
	}
	for index := range result {
		result[index].ProposalCoordinate.Total = 1
		result[index].RejectionCoordinate.Total = 1
		if result[index].ProposalEmitted {
			result[index].ProposalCoordinate.Satisfied = 1
		}
		if result[index].Outcome == "REJECTED" {
			result[index].RejectionCoordinate.Satisfied = 1
		}
	}
	result[1].ProposalTransitionFrom, result[1].ProposalTransitionTo, result[1].ProposalTransitionReason = "OPEN", "OPEN", "UNKNOWN_PRESERVED_OPEN"
	return result, nil
}

func validateCaseBinding(got, want proposalCase) error {
	checks := []struct{ actual, expected, stage, step, reason string }{
		{got.CaseID, want.CaseID, "PROPOSAL", "bind-case-identity", "CASE_ID_MISMATCH"},
		{got.SourceRawDigest, want.SourceRawDigest, "PROPOSAL", "bind-source-raw-digest", "SOURCE_RAW_DIGEST_MISMATCH"},
		{got.SourceSemanticDigest, want.SourceSemanticDigest, "PROPOSAL", "bind-source-semantic-digest", "SOURCE_SEMANTIC_DIGEST_MISMATCH"},
		{got.QueryReceiptDigest, want.QueryReceiptDigest, "PROPOSAL", "bind-query-receipt-digest", "QUERY_RECEIPT_DIGEST_MISMATCH"},
		{got.TargetSemanticAddress, want.TargetSemanticAddress, "PROPOSAL", "bind-target-address", "TARGET_SEMANTIC_ADDRESS_MISMATCH"},
		{got.UnknownStage, want.UnknownStage, "PROPOSAL", "bind-unknown-causality", "UNKNOWN_CAUSALITY_MISMATCH"},
		{got.UnknownStep, want.UnknownStep, "PROPOSAL", "bind-unknown-causality", "UNKNOWN_CAUSALITY_MISMATCH"},
		{got.UnknownReason, want.UnknownReason, "PROPOSAL", "bind-unknown-causality", "UNKNOWN_CAUSALITY_MISMATCH"},
		{got.Authority, "NONE", "PROPOSAL", "validate-authority", "PROPOSAL_AUTHORITY_ESCALATED"},
	}
	for _, check := range checks {
		if check.actual != check.expected {
			return validationFault{Stage: check.stage, Step: check.step, Reason: check.reason}
		}
	}
	if got.Outcome != want.Outcome || got.OutcomeReason != want.OutcomeReason || got.ProposalEmitted != want.ProposalEmitted || got.ClaimID != want.ClaimID || got.RequestedRefinement != want.RequestedRefinement || got.ProofChoice != want.ProofChoice || got.MetaOperation != want.MetaOperation || got.ObservedClaimFrom != want.ObservedClaimFrom || got.ObservedClaimTo != want.ObservedClaimTo || got.ObservedClaimReason != want.ObservedClaimReason || got.ProposalTransitionFrom != want.ProposalTransitionFrom || got.ProposalTransitionTo != want.ProposalTransitionTo || got.ProposalTransitionReason != want.ProposalTransitionReason {
		return validationFault{Stage: "PROPOSAL", Step: "bind-source-observation", Reason: "PROPOSAL_OBSERVATION_BINDING_MISMATCH"}
	}
	if got.ProposalCoordinate != want.ProposalCoordinate || got.RejectionCoordinate != want.RejectionCoordinate {
		return validationFault{Stage: "PROPOSAL", Step: "bind-case-coordinate", Reason: "PROPOSAL_COORDINATE_MISMATCH"}
	}
	return nil
}

func validateArtifactMetadata(metadata map[string]string, item proposalCase) error {
	expected := map[string]string{
		"reflect.meta:proposal-artifact": "reflect.meta:proposal-artifact", "schema": proposalArtifactSchema, "case": item.CaseID,
		"observation": item.ObservationKind, "proposal.outcome": item.Outcome, "proposal.reason": item.OutcomeReason,
		"proposal.emitted": fmt.Sprintf("%t", item.ProposalEmitted), "claim": item.ClaimID, "target": item.TargetSemanticAddress,
		"source.path": canonicalSourcePath, "source.raw": item.SourceRawDigest, "source.semantic": item.SourceSemanticDigest,
		"query.receipt": item.QueryReceiptDigest, "unknown.stage": item.UnknownStage, "unknown.step": item.UnknownStep, "unknown.reason": item.UnknownReason,
		"requested.refinement": item.RequestedRefinement, "proof.choice": item.ProofChoice, "meta.operation": item.MetaOperation, "authority": "NONE",
		"policy": "unknown-preserve-open", "claim.from": item.ObservedClaimFrom, "claim.to": item.ObservedClaimTo, "claim.reason": item.ObservedClaimReason,
		"proposal.from": item.ProposalTransitionFrom, "proposal.to": item.ProposalTransitionTo, "proposal.transition.reason": item.ProposalTransitionReason,
	}
	for key, value := range expected {
		if metadata[key] != value {
			return fmt.Errorf("field %s mismatch", key)
		}
	}
	if len(metadata) != len(expected) {
		return errors.New("artifact metadata has undeclared fields")
	}
	return nil
}

func parseProposalArtifact(path string, raw []byte) (string, map[string]string, error) {
	file, diagnostics := syntax.ParseFile(path, string(raw))
	if diagnostics.HasErrors() {
		return "", nil, diagnostics.Error()
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return "", nil, err
	}
	metadata := map[string]string{}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind == semantic.Entity && strings.HasPrefix(node.Name, "SourceRawDigest_") {
			metadata["source.raw"] = strings.TrimPrefix(node.Name, "SourceRawDigest_")
		}
		if node.Kind != semantic.Activity || !strings.HasPrefix(node.ValueProgram, "reflect.meta:proposal-artifact;") {
			continue
		}
		for index, part := range strings.Split(node.ValueProgram, ";") {
			if index == 0 {
				metadata[part] = part
				continue
			}
			pair := strings.SplitN(part, "=", 2)
			if len(pair) != 2 || pair[0] == "" {
				return "", nil, errors.New("malformed proposal metadata")
			}
			if _, exists := metadata[pair[0]]; exists {
				return "", nil, errors.New("duplicate proposal metadata")
			}
			metadata[pair[0]] = pair[1]
		}
	}
	if len(metadata) == 0 {
		return "", nil, errors.New("proposal activity is missing")
	}
	return ir.StableHash(), metadata, nil
}

func buildCounterexamples(got []proposalCase, expected []proposalCase) []rejection {
	mutations := []struct {
		name string
		edit func(*proposalCase)
	}{
		{"target-address-tamper", func(item *proposalCase) { item.TargetSemanticAddress += "/tampered" }},
		{"unknown-reason-substitution", func(item *proposalCase) { item.UnknownReason = "UNKNOWN_TARGET_REPLACED" }},
		{"authority-escalation", func(item *proposalCase) { item.Authority = "REPOSITORY_WRITE" }},
		{"query-receipt-digest-mismatch", func(item *proposalCase) { item.QueryReceiptDigest = strings.Repeat("0", 64) }},
		{"source-semantic-digest-mismatch", func(item *proposalCase) { item.SourceSemanticDigest = strings.Repeat("f", 64) }},
	}
	result := make([]rejection, 0, len(mutations))
	for index, mutation := range mutations {
		caseIndex := 1
		if index == 0 {
			caseIndex = 1
		}
		candidate := got[caseIndex]
		mutation.edit(&candidate)
		err := validateCaseBinding(candidate, expected[caseIndex])
		var fault validationFault
		if !errors.As(err, &fault) {
			result = append(result, rejection{Name: mutation.name, Decision: "UNKNOWN", Resolution: "LOWER_RESOLUTION", Stage: "PROPOSAL", Step: "counterexample", Reason: "COUNTEREXAMPLE_NOT_REJECTED"})
			continue
		}
		result = append(result, rejection{Name: mutation.name, Decision: "REFUTED", Resolution: "EXACT", Stage: fault.Stage, Step: fault.Step, Reason: fault.Reason})
	}
	return result
}

func safeArtifactPath(input, relative, caseID string) (string, error) {
	if relative != "proposals/"+safeToken(caseID)+".gooo" || filepath.IsAbs(relative) || filepath.Clean(relative) != relative || strings.HasPrefix(relative, "../") {
		return "", errors.New("proposal artifact path is not canonical")
	}
	return filepath.Join(input, filepath.FromSlash(relative)), nil
}

func findAttempt(attempts []observedAttempt, predicate func(observedAttempt) bool) (observedAttempt, bool) {
	for _, item := range attempts {
		if predicate(item) {
			return item, true
		}
	}
	return observedAttempt{}, false
}
func firstClaimID(attempt observedAttempt, fallback string) string {
	if len(attempt.EvidenceClaimIDs) > 0 {
		return attempt.EvidenceClaimIDs[0]
	}
	return fallback
}
func latestClaim(claims []observedClaim, id string) observedClaim {
	var result observedClaim
	for _, item := range claims {
		if item.ClaimID == id && item.Sequence > result.Sequence {
			result = item
		}
	}
	return result
}
func safeToken(value string) string {
	var b strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
	}
	return b.String()
}
func tail(value string) string {
	parts := strings.Split(strings.TrimSuffix(value, "/"), "/")
	return parts[len(parts)-1]
}
func sealProposalManifest(manifest proposalManifest) string {
	manifest.PortableReceiptDigest = ""
	raw, err := json.Marshal(manifest)
	if err != nil {
		panic(err)
	}
	return semantic.StableHash(raw)
}
func fail(format string, args ...any) { fmt.Fprintf(os.Stderr, format+"\n", args...); os.Exit(1) }
