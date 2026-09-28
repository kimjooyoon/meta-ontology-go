package selfimprovementtransport

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
)

func Produce(repository fs.FS, contractPath string, input ProducerInput, observation []byte) (ProducerReceipt, error) {
	contract, err := CompileContract(repository, contractPath)
	if err != nil {
		return ProducerReceipt{}, err
	}
	var source observationHeader
	if err := json.Unmarshal(observation, &source); err != nil {
		return ProducerReceipt{}, fmt.Errorf("decode observation: %w", err)
	}
	receipt := ProducerReceipt{
		Schema: ProducerSchema, Contract: contract,
		RepositoryURI: "https://github.com/" + input.Repository,
		SubjectSHA:    input.SubjectSHA, CheckoutSHA: input.CheckoutSHA,
		WorkflowRef: input.WorkflowRef, WorkflowSHA: input.WorkflowSHA,
		RunID: input.RunID, RunAttempt: input.RunAttempt, Job: input.Job,
		ArtifactName: input.ArtifactName,
		Subject:      LogicalSubject{Name: "first.json", Digest: digestBytes(observation), Bytes: len(observation)},
		Decision:     "BOUND", Resolution: ResolutionExact, Reason: "PRODUCER_SUBJECT_BOUND",
	}
	receipt.Digest = producerDigest(receipt)
	if err := ValidateProducer(receipt, source, observation, contract); err != nil {
		return ProducerReceipt{}, err
	}
	return receipt, nil
}

func ValidateProducer(receipt ProducerReceipt, source observationHeader, observation []byte, contract ContractEvidence) error {
	validRepository := strings.HasPrefix(receipt.RepositoryURI, "https://github.com/") &&
		strings.Count(strings.TrimPrefix(receipt.RepositoryURI, "https://github.com/"), "/") == 1
	if receipt.Schema != ProducerSchema || receipt.Decision != "BOUND" ||
		receipt.Resolution != ResolutionExact || receipt.Reason != "PRODUCER_SUBJECT_BOUND" ||
		!validRepository || !validSHA(receipt.SubjectSHA) || receipt.CheckoutSHA != receipt.SubjectSHA ||
		source.Schema != ObservationSchema || source.SubjectSHA != receipt.SubjectSHA {
		return fmt.Errorf("producer subject binding mismatch")
	}
	if receipt.Contract.ContractID != ContractID || receipt.Contract.CanonicalDigest != contract.CanonicalDigest ||
		receipt.Contract.SemanticDigest != contract.SemanticDigest || receipt.Contract.ObligationTotal != fixedObligationTotal ||
		!validDigest(receipt.Contract.SourceDigest) || contract.ResolutionPolicy.Validate() != nil {
		return fmt.Errorf("producer contract binding mismatch")
	}
	if !validSHA(receipt.WorkflowSHA) || receipt.WorkflowRef == "" || receipt.RunID <= 0 ||
		receipt.RunAttempt <= 0 || receipt.Job == "" || receipt.ArtifactName != ArtifactName {
		return fmt.Errorf("producer run identity mismatch")
	}
	if receipt.Subject.Name != "first.json" || receipt.Subject.Bytes != len(observation) ||
		receipt.Subject.Digest != digestBytes(observation) || !validDigest(receipt.Subject.Digest) {
		return fmt.Errorf("producer logical subject mismatch")
	}
	if receipt.Digest != producerDigest(receipt) {
		return fmt.Errorf("producer receipt digest mismatch")
	}
	return nil
}

func producerDigest(receipt ProducerReceipt) string {
	receipt.Digest = ""
	return digestJSON(receipt)
}
