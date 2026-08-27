package nonmonotonicrefutation

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const sourceContractSchema = "gooo/meta-nonmonotonic-refutation-source/v1"

type sourceModel struct {
	Contract       Contract
	SemanticDigest string
}

// Produce parses and lowers the actual .gooo source before constructing the
// producer wire model. repositoryWrites is supplied by the caller's before /
// after repository snapshot; it is never a declared constant in the receipt.
func Produce(sourcePath string, source []byte, repositoryWrites int) (ProducerReport, error) {
	model, err := reconstructSource(sourcePath, source)
	if err != nil {
		return ProducerReport{}, err
	}
	report := ProducerReport{
		Schema:               ProducerSchema,
		Contract:             model.Contract,
		SourcePath:           sourcePath,
		SourceDigest:         DigestBytes(source),
		SourceSemanticDigest: model.SemanticDigest,
		SourceModelDigest:    DigestJSON(model.Contract),
		Producer:             ProducerID,
		Consumer:             ConsumerID,
		MetaOperation:        MetaOperation,
		ProofChoice:          ProofRegression,
		Effects:              Effects{RepositoryWrites: repositoryWrites, MutationAuthority: false},
		NotClaimed: []string{
			"truth of the domain claim outside the parsed fixture",
			"probabilistic confidence or source credibility ranking",
			"automatic repository mutation or event-log replication",
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
		if !ok || !activity.ValueProgramPresent || !strings.HasPrefix(activity.ValueProgram, "meta.observe:v1;") {
			continue
		}
		if len(activity.Inputs) < 3 || activity.Output == "" {
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
		if entities[activity.Inputs[1].Name] != "gooo://nonmonotonic-refutation/predicate/"+fields["predicate"] || entities[activity.Inputs[2].Name] != "gooo://nonmonotonic-refutation/value/one" || fields["expected"] != "1" {
			return sourceModel{}, fmt.Errorf("activity %q has unresolved predicate or expected-value input", activity.Name)
		}
		sequence := len(model.Contract.Observations) + 1
		observation := Observation{
			ID: observationID, Activity: activity.Name, ClaimID: claimID, Sequence: sequence,
			Predicate: fields["predicate"], ExpectedValue: fields["expected"],
			ObservedValue: fields["observed"], Provenance: fields["provenance"],
			EvidenceDigest: fields["evidence_digest"], PriorState: fields["prior"],
			RevisionPolicy: fields["revision_policy"], Producer: fields["producer"],
			Consumer: fields["consumer"], MetaOperation: fields["meta_operation"],
			ProofChoice: fields["proof_choice"],
			Coordinate:  Coordinate{Stage: fields["stage"], Step: fields["step"]},
		}
		if err := validateObservation(model.Contract.Claims, observation); err != nil {
			return sourceModel{}, fmt.Errorf("activity %q: %w", activity.Name, err)
		}
		model.Contract.Observations = append(model.Contract.Observations, observation)
	}
	if len(model.Contract.Observations) != 6 {
		return sourceModel{}, fmt.Errorf("source observations = %d, want 6", len(model.Contract.Observations))
	}
	model.Contract.Schema = sourceContractSchema
	model.Contract.FixedClaimTotal = len(model.Contract.Claims)
	model.Contract.FixedObservationTotal = len(model.Contract.Observations)
	model.Contract.FixedTransitionTotal = len(model.Contract.Observations)
	if err := completeClaims(&model.Contract); err != nil {
		return sourceModel{}, err
	}
	return model, nil
}

func observationFields(program string) (map[string]string, error) {
	fields := make(map[string]string)
	for _, part := range strings.Split(program, ";") {
		if part == "meta.observe:v1" {
			continue
		}
		key, value, ok := strings.Cut(part, "=")
		if !ok || strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
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
	for _, key := range []string{"claim", "predicate", "expected", "observed", "provenance", "evidence_digest", "prior", "revision_policy", "producer", "consumer", "meta_operation", "proof_choice", "stage", "step"} {
		if fields[key] == "" {
			return nil, fmt.Errorf("missing source field %q", key)
		}
	}
	return fields, nil
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

func validateObservation(claims []Claim, observation Observation) error {
	if observation.Predicate != "equality" {
		return fmt.Errorf("unsupported predicate %q", observation.Predicate)
	}
	if observation.Producer != ProducerID || observation.Consumer != ConsumerID || observation.MetaOperation != MetaOperation {
		return fmt.Errorf("producer/consumer/meta-operation provenance mismatch")
	}
	if observation.ProofChoice == "" || observation.Coordinate.Stage == "" || observation.Coordinate.Step == "" {
		return fmt.Errorf("proof choice and coordinate are required")
	}
	if observation.PriorState != StatusOpen && observation.PriorState != StatusDischarged && observation.PriorState != StatusRefuted {
		return fmt.Errorf("invalid prior state %q", observation.PriorState)
	}
	if observation.RevisionPolicy == "" || observation.Provenance == "" || len(observation.EvidenceDigest) != len("sha256:")+64 || !strings.HasPrefix(observation.EvidenceDigest, "sha256:") {
		return fmt.Errorf("evidence provenance, digest, and revision policy are required")
	}
	if _, ok := claimIDForKey(claims, strings.TrimPrefix(observation.ClaimID, "gooo://nonmonotonic-refutation/claim/")); !ok {
		return fmt.Errorf("unknown claim %q", observation.ClaimID)
	}
	return nil
}

func knownObservationField(key string) bool {
	for _, known := range []string{"claim", "predicate", "expected", "observed", "provenance", "evidence_digest", "prior", "revision_policy", "producer", "consumer", "meta_operation", "proof_choice", "stage", "step"} {
		if key == known {
			return true
		}
	}
	return false
}

func completeClaims(contract *Contract) error {
	for index := range contract.Claims {
		claim := &contract.Claims[index]
		for _, observation := range contract.Observations {
			if observation.ClaimID != claim.ID {
				continue
			}
			if claim.InitialStatus == "" {
				claim.InitialStatus = observation.PriorState
				claim.Predicate = observation.Predicate
				claim.ExpectedValue = observation.ExpectedValue
				claim.RevisionPolicy = observation.RevisionPolicy
			}
			if claim.Predicate != observation.Predicate || claim.ExpectedValue != observation.ExpectedValue || claim.RevisionPolicy != observation.RevisionPolicy {
				return fmt.Errorf("claim %q changes predicate, expected value, or revision policy", claim.ID)
			}
		}
		if claim.InitialStatus == "" {
			return fmt.Errorf("claim %q has no observations", claim.ID)
		}
	}
	return nil
}
