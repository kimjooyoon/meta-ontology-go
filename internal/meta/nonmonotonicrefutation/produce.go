package nonmonotonicrefutation

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const sourceContractSchema = "gooo/meta-nonmonotonic-refutation-source/v2"

type sourceModel struct {
	Contract       Contract
	SemanticDigest string
}

type sourceBinding struct {
	RawDigest      string `json:"raw_digest"`
	SemanticDigest string `json:"semantic_digest"`
}

// Produce parses and lowers the actual .gooo source before constructing the
// source-derived producer receipt. repositoryWrites comes from a caller's
// before/after repository snapshot; no mutation or promotion is authorized.
func Produce(sourcePath string, source []byte, repositoryWrites int) (ProducerReport, error) {
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
		SourceBindingDigest:  DigestJSON(sourceBinding{RawDigest: rawDigest, SemanticDigest: model.SemanticDigest}),
		SourceModelDigest:    DigestJSON(model.Contract),
		Producer:             ProducerID,
		Consumer:             ConsumerID,
		MetaOperation:        MetaOperation,
		ProofChoice:          ProofRegression,
		Effects:              Effects{RepositoryWrites: repositoryWrites, MutationAuthority: false, PromotionCount: 0},
		NotClaimed: []string{
			"truth of the domain proposition outside the parsed fixture",
			"probabilistic confidence or source credibility ranking",
			"automatic repository mutation or promotion",
		},
	}
	report.ReceiptDigest = DigestJSON(report)
	return report, nil
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
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || !activity.ValueProgramPresent || !strings.HasPrefix(activity.ValueProgram, "meta.observe:v2;") {
			continue
		}
		if len(activity.Inputs) < 5 || activity.Output == "" {
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
		if entities[activity.Inputs[1].Name] != "gooo://nonmonotonic-refutation/predicate/"+fields["predicate"] ||
			entities[activity.Inputs[2].Name] != "gooo://nonmonotonic-refutation/subject/"+fields["subject"] ||
			entities[activity.Inputs[3].Name] != "gooo://nonmonotonic-refutation/input/"+fields["input"] ||
			entities[activity.Inputs[4].Name] != "gooo://nonmonotonic-refutation/value/one" || fields["expected"] != "1" {
			return sourceModel{}, fmt.Errorf("activity %q subject/input/proposition endpoint mismatch", activity.Name)
		}
		observation := Observation{
			ID: observationID, Activity: activity.Name, ClaimID: claimID, Sequence: len(model.Contract.Observations) + 1,
			Proposition: fields["proposition"], Subject: fields["subject"], Input: fields["input"],
			Predicate: fields["predicate"], ExpectedValue: fields["expected"], ObservedValue: fields["observed"],
			Provenance: fields["provenance"], EvidenceDigest: fields["evidence_digest"], Producer: fields["producer"],
			Consumer: fields["consumer"], MetaOperation: fields["meta_operation"], ProofChoice: fields["proof_choice"],
			Coordinate: Coordinate{Stage: fields["stage"], Step: fields["step"]},
		}
		if err := validateObservation(observation); err != nil {
			return sourceModel{}, fmt.Errorf("activity %q: %w", activity.Name, err)
		}
		model.Contract.Observations = append(model.Contract.Observations, observation)
	}
	if len(model.Contract.Observations) != 6 {
		return sourceModel{}, fmt.Errorf("source observations = %d, want 6", len(model.Contract.Observations))
	}
	model.Contract.Schema = sourceContractSchema
	model.Contract.FixedCaseTotal = len(model.Contract.Claims)
	model.Contract.FixedClaimTotal = len(model.Contract.Claims)
	model.Contract.FixedObservationTotal = len(model.Contract.Observations)
	model.Contract.FixedLedgerRowTotal = len(model.Contract.Observations)
	if err := completeClaims(&model.Contract); err != nil {
		return sourceModel{}, err
	}
	return model, nil
}

func observationFields(program string) (map[string]string, error) {
	fields := make(map[string]string)
	for _, part := range strings.Split(program, ";") {
		if part == "meta.observe:v2" {
			continue
		}
		key, value, ok := strings.Cut(part, "=")
		if !ok || strings.TrimSpace(key) == "" || (strings.TrimSpace(value) == "" && key != "observed") {
			return nil, fmt.Errorf("malformed source field %q", part)
		}
		if !knownObservationField(key) {
			return nil, fmt.Errorf("unsupported source field %q", key)
		}
		if _, exists := fields[key]; exists {
			return nil, fmt.Errorf("duplicate source field %q", key)
		}
		fields[key] = value
	}
	for _, key := range []string{"claim", "proposition", "subject", "input", "predicate", "expected", "provenance", "evidence_digest", "producer", "consumer", "meta_operation", "proof_choice", "stage", "step"} {
		if fields[key] == "" {
			return nil, fmt.Errorf("missing source field %q", key)
		}
	}
	if _, ok := fields["observed"]; !ok {
		return nil, fmt.Errorf("missing source field %q", "observed")
	}
	return fields, nil
}

func knownObservationField(key string) bool {
	for _, known := range []string{"claim", "proposition", "subject", "input", "predicate", "expected", "observed", "provenance", "evidence_digest", "producer", "consumer", "meta_operation", "proof_choice", "stage", "step"} {
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

func validateObservation(observation Observation) error {
	if observation.Proposition == "" || observation.Subject == "" || observation.Input == "" || observation.Predicate == "" || observation.ExpectedValue == "" {
		return fmt.Errorf("proposition and observable subject/input are required")
	}
	if observation.Producer != ProducerID || observation.Consumer != ConsumerID || observation.MetaOperation != MetaOperation {
		return fmt.Errorf("producer/consumer/meta-operation provenance mismatch")
	}
	if observation.ProofChoice == "" || observation.Coordinate.Stage == "" || observation.Coordinate.Step == "" {
		return fmt.Errorf("proof choice and coordinate are required")
	}
	if observation.Provenance == "" || len(observation.EvidenceDigest) != len("sha256:")+64 || !strings.HasPrefix(observation.EvidenceDigest, "sha256:") {
		return fmt.Errorf("evidence provenance and digest are required")
	}
	return nil
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
