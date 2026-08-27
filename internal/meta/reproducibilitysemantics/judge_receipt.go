package reproducibilitysemantics

import "strings"

func receiptShape(sourcePath, headSHA string, source []byte, semanticDigest string, receipt Receipt) string {
	if receipt.Schema != ReceiptSchema || receipt.Version != 1 || receipt.ContractID != ContractID {
		return "RECEIPT_SCHEMA_INVALID"
	}
	if receipt.SourcePath != sourcePath || receipt.SourceDigest != digestBytes(source) || receipt.HeadSHA != headSHA {
		return "RECEIPT_SOURCE_BINDING_INVALID"
	}
	if !commitPattern.MatchString(headSHA) || !strings.Contains(string(source), "activity SeparateClaims(ByteArtifact, MeaningClaim) -> WitnessCase") {
		return "SOURCE_OR_HEAD_INVALID"
	}
	for _, declaration := range []string{
		"entity ByteArtifact id", "entity MeaningClaim id", "entity WitnessCase id",
		"entity BothClaimsDischarged id", "entity ReproducibleButWrong id",
		"entity MeaningfulButUnreproduced id", "entity ClaimsOpen id",
	} {
		if !strings.Contains(string(source), declaration) {
			return "GOOO_SOURCE_CONTRACT_INCOMPLETE"
		}
	}
	if receipt.Producer != ProducerID || receipt.Consumer != ConsumerID ||
		receipt.MetaOperation != MetaOperation || receipt.ProofChoice != ProofComposition ||
		receipt.Stage != "receipt" || receipt.Step != "produce" || receipt.Reason != "CLAIM_CHANNELS_SEPARATED" {
		return "RECEIPT_PROVENANCE_INVALID"
	}
	if receipt.SemanticDigest == "" {
		return "DIGEST_ONLY_REFUTED"
	}
	if receipt.SemanticDigest != semanticDigest {
		return "SEMANTIC_DIGEST_BINDING_INVALID"
	}
	if len(receipt.Cases) != CaseCount || len(receipt.Proofs) != 4 || receipt.Authority != (Authority{}) {
		return "RECEIPT_DENOMINATOR_OR_AUTHORITY_INVALID"
	}
	want := receipt.ReceiptDigest
	if want == "" || sealReceipt(receipt).ReceiptDigest != want || !shaPattern.MatchString(receipt.SourceDigest) {
		return "RECEIPT_DIGEST_INVALID"
	}
	return ""
}
