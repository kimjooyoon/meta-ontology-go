package extractor

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type returnTailProofChain struct {
	contract      []ContractObligationEvidence
	sourceDigest  string
	candidateDigest string
	stages        []ProofStageEvidence
}

type returnTailPredicateResult struct {
	Status  string
	Payload []byte
	Detail  string
}

func newReturnTailProofChain(contract []ContractObligationEvidence, source, candidate []byte) returnTailProofChain {
	return returnTailProofChain{
		contract: append([]ContractObligationEvidence{}, contract...),
		sourceDigest: proofDigest(source),
		candidateDigest: proofDigest(candidate),
		stages: make([]ProofStageEvidence, 0, len(contract)),
	}
}

func (chain *returnTailProofChain) consume(index int, result returnTailPredicateResult) error {
	if index < 0 || index >= len(chain.contract) || index != len(chain.stages) {
		return fail("derive-recipe", "consume-return-tail-proof", "RETURN_TAIL_PROOF_CHAIN_UNPROVEN", "DIRECT_MISSING", "restore-return-tail-proof", nil)
	}
	relation := chain.contract[index]
	inputEvidenceID := proofEvidenceID(relation.InputEntity, chain.sourceDigest, "input")
	if index > 0 {
		inputEvidenceID = chain.stages[index-1].OutputEvidenceID
		if inputEvidenceID == "" || chain.stages[index-1].OutputEntity != relation.InputEntity {
			return failWithDiagnostics("derive-recipe", "consume-return-tail-proof", "RETURN_TAIL_PROOF_CHAIN_UNPROVEN", "DIRECT_MISSING", "restore-return-tail-proof", []string{"obligation=" + relation.Name})
		}
	}
	if result.Status == "" || result.Payload == nil {
		return failWithDiagnostics("derive-recipe", "consume-return-tail-proof", "RETURN_TAIL_PROOF_RESULT_MISSING", "DIRECT_MISSING", "restore-return-tail-proof", []string{"obligation=" + relation.Name})
	}
	payloadDigest := proofDigest(result.Payload)
	outputEvidenceID := proofEvidenceID(relation.OutputEntity, relation.Activity, payloadDigest)
	chain.stages = append(chain.stages, ProofStageEvidence{
		Name: relation.Name, Activity: relation.Activity, InputEntity: relation.InputEntity, OutputEntity: relation.OutputEntity,
		Status: result.Status, SourceDigest: chain.sourceDigest, CandidateDigest: chain.candidateDigest,
		InputEvidenceID: inputEvidenceID, OutputEvidenceID: outputEvidenceID, PayloadDigest: payloadDigest,
		PayloadBytes: len(result.Payload), Detail: result.Detail,
	})
	return nil
}

func proofEvidenceID(entity, activity, payloadDigest string) string {
	return proofDigest([]byte(entity + "\x00" + activity + "\x00" + payloadDigest))
}

func proofDigest(payload []byte) string {
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func proofBindingPayload(bindings []suffixBinding) []byte {
	var payload bytes.Buffer
	for _, binding := range bindings {
		fmt.Fprintf(&payload, "%s\x00%s\n", binding.name, binding.object.Type().String())
	}
	if payload.Len() == 0 {
		return []byte{}
	}
	return payload.Bytes()
}

func obligationsFromProofStages(stages []ProofStageEvidence) []ObligationEvidence {
	result := make([]ObligationEvidence, 0, len(stages))
	for _, stage := range stages {
		result = append(result, ObligationEvidence{Name: stage.Name, Status: stage.Status, Detail: stage.Detail})
	}
	return result
}
