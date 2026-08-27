package nonmonotonicrefutation

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	sourceContractSchema = "gooo/meta-nonmonotonic-refutation-source/v3"
	noEvidenceTarget     = "none"
)

type sourceModel struct {
	Contract       Contract
	SemanticDigest string
}

type sourceBinding struct {
	RawDigest      string `json:"raw_digest"`
	SemanticDigest string `json:"semantic_digest"`
	PolicyID       string `json:"policy_id"`
	PolicyDigest   string `json:"policy_digest"`
}

// Produce parses and lowers the actual .gooo source before constructing the
// source-derived producer receipt. The repository argument is deliberately a
// scoped net-status observation, not evidence of zero transient writes.
func Produce(sourcePath string, source []byte, netRepositoryStatusUnchanged bool) (ProducerReport, error) {
	model, err := reconstructSource(sourcePath, source)
	if err != nil {
		return ProducerReport{}, err
	}
	rawDigest := DigestBytes(source)
	report := ProducerReport{
		Schema:               ProducerSchema,
		Contract:             model.Contract,
		SourcePath:           sourcePath,
		SourceDigest:         rawDigest,
		SourceSemanticDigest: model.SemanticDigest,
		SourceBindingDigest: DigestJSON(sourceBinding{
			RawDigest: rawDigest, SemanticDigest: model.SemanticDigest,
			PolicyID: model.Contract.Policy.ID, PolicyDigest: model.Contract.Policy.PolicyDigest,
		}),
		SourceModelDigest: DigestJSON(model.Contract),
		Producer:          ProducerID,
		Consumer:          ConsumerID,
		MetaOperation:     MetaOperation,
		ProofChoice:       ProofRegression,
		Effects: Effects{
			NetRepositoryStatusUnchanged: netRepositoryStatusUnchanged,
			RepositoryWriteObservation:   repositoryStatusObservation(netRepositoryStatusUnchanged),
			MutationAuthorityResolution:  "UNKNOWN",
			PromotionOperationsObserved:  0,
		},
		NotClaimed: []string{
			"external domain truth: all observations are HISTORICAL_FIXTURE material",
			"current evidence beyond the accepted append-only ledger",
			"authority to infer transient writes or mutate/promote the repository",
		},
	}
	report.ReceiptDigest = DigestJSON(report)
	return report, nil
}

func repositoryStatusObservation(unchanged bool) string {
	if unchanged {
		return "NONE_OBSERVED_IN_NET_STATUS"
	}
	return "NET_STATUS_CHANGED"
}

func reconstructSource(sourcePath string, source []byte) (sourceModel, error) {
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if diagnostics.HasErrors() {
		return sourceModel{}, fmt.Errorf("parse %s: diagnostics contain errors", sourcePath)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return sourceModel{}, fmt.Errorf("lower %s: %w", sourcePath, err)
	}
	model := sourceModel{SemanticDigest: "sha256:" + semantic.StableHash([]byte(ir.SemanticCanonical()))}
	entities := make(map[string]string)
	for _, declaration := range file.Declarations {
		entity, ok := declaration.(*syntax.EntityDecl)
		if !ok {
			continue
		}
		if entity.ID == "" {
			return sourceModel{}, fmt.Errorf("entity %q has no stable ID", entity.Name)
		}
		entities[entity.Name] = entity.ID
		if strings.HasPrefix(entity.ID, "gooo://nonmonotonic-refutation/claim/") {
			model.Contract.Claims = append(model.Contract.Claims, Claim{ID: entity.ID})
		}
	}
	if len(model.Contract.Claims) != 3 {
		return sourceModel{}, fmt.Errorf("source claims = %d, want 3", len(model.Contract.Claims))
	}

	var policySeen bool
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || !activity.ValueProgramPresent || !strings.HasPrefix(activity.ValueProgram, "meta.revision-policy:v1;") {
			continue
		}
		if policySeen || len(activity.Inputs) != 1 || activity.Output == "" {
			return sourceModel{}, fmt.Errorf("revision policy activity %q is not unique and bound", activity.Name)
		}
		policySeen = true
		policyID := entities[activity.Inputs[0].Name]
		outputID := entities[activity.Output]
		if !strings.HasPrefix(policyID, "gooo://nonmonotonic-refutation/policy/") ||
			!strings.HasPrefix(outputID, "gooo://nonmonotonic-refutation/policy-binding/") {
			return sourceModel{}, fmt.Errorf("revision policy activity %q has invalid endpoints", activity.Name)
		}
		fields, err := policyFields(activity.ValueProgram)
		if err != nil {
			return sourceModel{}, fmt.Errorf("activity %q: %w", activity.Name, err)
		}
		if fields["policy_id"] != policyID {
			return sourceModel{}, fmt.Errorf("revision policy activity %q policy ID mismatch", activity.Name)
		}
		model.Contract.Policy = RevisionPolicy{
			ID:                          fields["policy_id"],
			CorrectionRelation:          fields["correction_relation"],
			CorrectionTarget:            fields["correction_target"],
			UnknownAction:               fields["unknown_action"],
			InsufficientAction:          fields["insufficient_action"],
			OrdinarySupportAfterRefuted: fields["ordinary_support_after_refuted"],
			FoundationRule:              fields["foundation_rule"],
			CoherenceRule:               fields["coherence_rule"],
			RegressionRule:              fields["regression_rule"],
			FixtureClass:                fields["fixture_class"],
			PolicyDigest:                fields["policy_digest"],
		}
		if err := validatePolicy(model.Contract.Policy); err != nil {
			return sourceModel{}, fmt.Errorf("activity %q: %w", activity.Name, err)
		}
	}
	if !policySeen {
		return sourceModel{}, fmt.Errorf("source has no revision policy object")
	}

	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || !activity.ValueProgramPresent || !strings.HasPrefix(activity.ValueProgram, "meta.observe:v3;") {
			continue
		}
		if len(activity.Inputs) != 5 || activity.Output == "" {
			return sourceModel{}, fmt.Errorf("observation activity %q has incomplete endpoints", activity.Name)
		}
		observationID, ok := entities[activity.Output]
		if !ok || !strings.HasPrefix(observationID, "gooo://nonmonotonic-refutation/observation/") {
			return sourceModel{}, fmt.Errorf("observation activity %q has an unbound output", activity.Name)
		}
		fields, err := observationFields(activity.ValueProgram)
		if err != nil {
			return sourceModel{}, fmt.Errorf("activity %q: %w", activity.Name, err)
		}
		claimID, ok := claimIDForKey(model.Contract.Claims, fields["claim"])
		if !ok || entities[activity.Inputs[0].Name] != claimID {
			return sourceModel{}, fmt.Errorf("activity %q claim binding is invalid", activity.Name)
		}
		subjectID := entities[activity.Inputs[2].Name]
		inputID := entities[activity.Inputs[3].Name]
		if entities[activity.Inputs[1].Name] != "gooo://nonmonotonic-refutation/predicate/"+fields["predicate"] ||
			subjectID != "gooo://nonmonotonic-refutation/subject/"+fields["subject"] ||
			inputID != "gooo://nonmonotonic-refutation/input/"+fields["input"] ||
			entities[activity.Inputs[4].Name] != "gooo://nonmonotonic-refutation/value/one" || fields["expected"] != "1" {
			return sourceModel{}, fmt.Errorf("activity %q subject/input/proposition endpoint mismatch", activity.Name)
		}
		observation := Observation{
			ID: observationID, Activity: activity.Name, ClaimID: claimID, Sequence: len(model.Contract.Observations) + 1,
			Proposition: fields["proposition"], Subject: fields["subject"], Input: fields["input"],
			Predicate: fields["predicate"], ExpectedValue: fields["expected"], ObservedValue: fields["observed"],
			ObservedMaterial: fields["observed_material"], ProviderClass: fields["provider_class"], Provenance: fields["provenance"],
			ObservationQuality: fields["observation_quality"],
			RevisionRelation:   fields["revision_relation"], SupersedesEvidenceDigest: fields["supersedes_evidence_digest"], SupersedesClaimID: fields["supersedes_claim_id"],
			PolicyID: fields["policy_id"], PolicyDigest: fields["policy_digest"], Producer: fields["producer"],
			Consumer: fields["consumer"], MetaOperation: fields["meta_operation"], ProofChoice: fields["proof_choice"],
			Coordinate:    Coordinate{Stage: fields["stage"], Step: fields["step"]},
			TargetAddress: subjectID + "|" + inputID,
		}
		if err := validateObservation(observation, model.Contract.Policy); err != nil {
			return sourceModel{}, fmt.Errorf("activity %q: %w", activity.Name, err)
		}
		model.Contract.Observations = append(model.Contract.Observations, observation)
	}
	if len(model.Contract.Observations) != 8 {
		return sourceModel{}, fmt.Errorf("source observations = %d, want 8", len(model.Contract.Observations))
	}
	model.Contract.Schema = sourceContractSchema
	model.Contract.FixedCaseTotal = len(model.Contract.Claims)
	model.Contract.FixedClaimTotal = len(model.Contract.Claims)
	model.Contract.FixedObservationTotal = len(model.Contract.Observations)
	model.Contract.FixedLedgerRowTotal = len(model.Contract.Observations)
	if err := completeClaims(&model.Contract); err != nil {
		return sourceModel{}, err
	}
	for index := range model.Contract.Observations {
		model.Contract.Observations[index].EvidenceDigest = EvidenceDigest(model.Contract.Observations[index])
	}
	return model, nil
}

func policyFields(program string) (map[string]string, error) {
	fields, err := parseFields(program, "meta.revision-policy:v1", knownPolicyField)
	if err != nil {
		return nil, err
	}
	for _, key := range []string{"policy_id", "correction_relation", "correction_target", "unknown_action", "insufficient_action", "ordinary_support_after_refuted", "foundation_rule", "coherence_rule", "regression_rule", "fixture_class", "policy_digest"} {
		if fields[key] == "" {
			return nil, fmt.Errorf("missing policy field %q", key)
		}
	}
	return fields, nil
}

func observationFields(program string) (map[string]string, error) {
	fields, err := parseFields(program, "meta.observe:v3", knownObservationField)
	if err != nil {
		return nil, err
	}
	for _, key := range []string{"claim", "proposition", "subject", "input", "predicate", "expected", "observed_material", "observation_quality", "provider_class", "provenance", "revision_relation", "supersedes_evidence_digest", "supersedes_claim_id", "policy_id", "policy_digest", "producer", "consumer", "meta_operation", "proof_choice", "stage", "step"} {
		if fields[key] == "" {
			return nil, fmt.Errorf("missing source field %q", key)
		}
	}
	if _, ok := fields["observed"]; !ok {
		return nil, fmt.Errorf("missing source field %q", "observed")
	}
	return fields, nil
}

func parseFields(program, marker string, known func(string) bool) (map[string]string, error) {
	fields := make(map[string]string)
	for _, part := range strings.Split(program, ";") {
		if part == marker {
			continue
		}
		key, value, ok := strings.Cut(part, "=")
		if !ok || strings.TrimSpace(key) == "" || !known(key) || (strings.TrimSpace(value) == "" && key != "observed") {
			return nil, fmt.Errorf("malformed or unsupported source field %q", part)
		}
		if _, exists := fields[key]; exists {
			return nil, fmt.Errorf("duplicate source field %q", key)
		}
		fields[key] = value
	}
	return fields, nil
}

func knownPolicyField(key string) bool {
	for _, known := range []string{"policy_id", "correction_relation", "correction_target", "unknown_action", "insufficient_action", "ordinary_support_after_refuted", "foundation_rule", "coherence_rule", "regression_rule", "fixture_class", "policy_digest"} {
		if key == known {
			return true
		}
	}
	return false
}

func knownObservationField(key string) bool {
	for _, known := range []string{"claim", "proposition", "subject", "input", "predicate", "expected", "observed", "observed_material", "observation_quality", "provider_class", "provenance", "revision_relation", "supersedes_evidence_digest", "supersedes_claim_id", "policy_id", "policy_digest", "producer", "consumer", "meta_operation", "proof_choice", "stage", "step"} {
		if key == known {
			return true
		}
	}
	return false
}

func claimIDForKey(claims []Claim, key string) (string, bool) {
	want := "gooo://nonmonotonic-refutation/claim/" + key
	for _, claim := range claims {
		if claim.ID == want {
			return claim.ID, true
		}
	}
	return "", false
}

func validatePolicy(policy RevisionPolicy) error {
	if !strings.HasPrefix(policy.ID, "gooo://nonmonotonic-refutation/policy/") ||
		policy.CorrectionRelation != RevisionSupersedes || policy.CorrectionTarget != PolicyCorrectionTargetEvidence ||
		policy.UnknownAction != PolicyUnknownRetain || policy.InsufficientAction != PolicyInsufficientRetain ||
		policy.OrdinarySupportAfterRefuted != PolicyOrdinarySupportRetain ||
		policy.FoundationRule != PolicyFoundationFirstClaimEvent || policy.CoherenceRule != PolicyCoherenceLaterClaimOpening ||
		policy.RegressionRule != PolicyRegressionTargetedHistory || policy.FixtureClass != ProviderHistoricalFixture {
		return fmt.Errorf("revision policy values are not the bounded v4 policy")
	}
	if !validDigest(policy.PolicyDigest) {
		return fmt.Errorf("policy digest is not canonical sha256 hex")
	}
	candidate := policy
	candidate.PolicyDigest = ""
	if DigestJSON(candidate) != policy.PolicyDigest {
		return fmt.Errorf("policy digest does not match canonical policy")
	}
	return nil
}

func validateObservation(observation Observation, policy RevisionPolicy) error {
	if observation.Proposition == "" || observation.Subject == "" || observation.Input == "" || observation.Predicate == "" || observation.ExpectedValue == "" || observation.ObservedMaterial == "" {
		return fmt.Errorf("proposition, observable subject/input, and evidence recipe are required")
	}
	if observation.ObservationQuality != "SUFFICIENT" && observation.ObservationQuality != "UNRESOLVED" {
		return fmt.Errorf("observation quality is not bounded")
	}
	if observation.ProviderClass != policy.FixtureClass || observation.ProviderClass != ProviderHistoricalFixture {
		return fmt.Errorf("evidence must be explicitly classified as historical fixture")
	}
	if observation.PolicyID != policy.ID || observation.PolicyDigest != policy.PolicyDigest {
		return fmt.Errorf("observation policy binding mismatch")
	}
	if observation.Producer != ProducerID || observation.Consumer != ConsumerID || observation.MetaOperation != MetaOperation {
		return fmt.Errorf("producer/consumer/meta-operation provenance mismatch")
	}
	if observation.ProofChoice != ProofFoundation && observation.ProofChoice != ProofCoherence && observation.ProofChoice != ProofRegression {
		return fmt.Errorf("proof choice is not bounded")
	}
	if observation.RevisionRelation != RevisionNone && observation.RevisionRelation != RevisionSupersedes {
		return fmt.Errorf("revision relation is not bounded")
	}
	if observation.RevisionRelation == RevisionNone && observation.SupersedesEvidenceDigest != noEvidenceTarget {
		return fmt.Errorf("non-correction observation must target none")
	}
	if observation.RevisionRelation == RevisionNone && observation.SupersedesClaimID != noEvidenceTarget {
		return fmt.Errorf("non-correction observation must target no claim")
	}
	if observation.RevisionRelation == RevisionSupersedes && observation.SupersedesEvidenceDigest != noEvidenceTarget && !validDigest(observation.SupersedesEvidenceDigest) {
		return fmt.Errorf("correction target must be a canonical evidence digest or explicit none")
	}
	if observation.RevisionRelation == RevisionSupersedes && observation.SupersedesClaimID != noEvidenceTarget && !strings.HasPrefix(observation.SupersedesClaimID, "gooo://nonmonotonic-refutation/claim/") {
		return fmt.Errorf("correction target claim must be canonical claim ID or explicit none")
	}
	if observation.RevisionRelation == RevisionNone && observation.SupersedesEvidenceDigest == "" {
		return fmt.Errorf("supersession target is required even when none")
	}
	if observation.Provenance == "" || observation.Coordinate.Stage == "" || observation.Coordinate.Step == "" || observation.TargetAddress == "" {
		return fmt.Errorf("fixture provenance, target, and coordinate are required")
	}
	return nil
}

func validDigest(value string) bool {
	if len(value) != len("sha256:")+64 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

func EvidenceDigest(observation Observation) string {
	return DigestJSON(EvidenceMaterial{
		ClaimID: observation.ClaimID, Proposition: observation.Proposition, TargetAddress: observation.TargetAddress,
		ObservedMaterial: observation.ObservedMaterial, ObservedValue: observation.ObservedValue,
		ObservationQuality: observation.ObservationQuality, ProviderClass: observation.ProviderClass, Sequence: observation.Sequence,
		SupersededEvidenceDigest: observation.SupersedesEvidenceDigest, SupersededClaimID: observation.SupersedesClaimID,
	})
}

func completeClaims(contract *Contract) error {
	for index := range contract.Claims {
		claim := &contract.Claims[index]
		for _, observation := range contract.Observations {
			if observation.ClaimID != claim.ID {
				continue
			}
			if claim.Proposition == "" {
				claim.Proposition = observation.Proposition
				claim.Subject = observation.Subject
				claim.Input = observation.Input
				claim.Predicate = observation.Predicate
				claim.ExpectedValue = observation.ExpectedValue
			}
			if claim.Proposition != observation.Proposition || claim.Subject != observation.Subject || claim.Input != observation.Input || claim.Predicate != observation.Predicate || claim.ExpectedValue != observation.ExpectedValue {
				return fmt.Errorf("claim %q changes proposition or subject/input", claim.ID)
			}
		}
		if claim.Proposition == "" {
			return fmt.Errorf("claim %q has no observation", claim.ID)
		}
	}
	return nil
}
