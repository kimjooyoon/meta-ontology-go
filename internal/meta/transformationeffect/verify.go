package transformationeffect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/transformationeffect/workspace"
)

func VerifyFiles(ledgerPath, receiptPath, provenancePath, patchPath string) error {
	ledgerPayload, err := os.ReadFile(ledgerPath)
	if err != nil {
		return err
	}
	patchPayload, err := os.ReadFile(patchPath)
	if err != nil {
		return err
	}
	receiptPayload, err := os.ReadFile(receiptPath)
	if err != nil {
		return err
	}
	provenancePayload, err := os.ReadFile(provenancePath)
	if err != nil {
		return err
	}
	var ledger Ledger
	var patch workspace.Patch
	var receipts generation.ReceiptReport
	var provenance generation.ArtifactProvenance
	if err := json.Unmarshal(ledgerPayload, &ledger); err != nil {
		return err
	}
	if err := json.Unmarshal(patchPayload, &patch); err != nil {
		return err
	}
	if err := json.Unmarshal(receiptPayload, &receipts); err != nil {
		return err
	}
	if err := json.Unmarshal(provenancePayload, &provenance); err != nil {
		return err
	}
	canonicalLedger, _ := encodeJSON(ledger)
	canonicalPatch, _ := encodeJSON(patch)
	canonicalReceipts, _ := generation.EncodeReceiptReport(receipts)
	canonicalProvenance, _ := generation.EncodeArtifactProvenance(provenance)
	if err := validateLedger(ledger); err != nil {
		return err
	}
	if err := workspace.Validate(patch); err != nil {
		return err
	}
	if !bytes.Equal(ledgerPayload, canonicalLedger) || !bytes.Equal(patchPayload, canonicalPatch) ||
		!bytes.Equal(receiptPayload, canonicalReceipts) || !bytes.Equal(provenancePayload, canonicalProvenance) {
		return fmt.Errorf("effect artifact encoding is not canonical")
	}
	if ledger.PatchDigest != patch.PatchDigest || ledger.GeneratedReceiptReportDigest != receipts.ReportDigest ||
		ledger.ExecutedProvenanceDigest != provenance.EnvelopeDigest || provenance.ReceiptReportDigest != receipts.ReportDigest {
		return fmt.Errorf("effect artifact set is not digest-bound")
	}
	return nil
}
