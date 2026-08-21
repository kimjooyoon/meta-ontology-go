package transformationeffect

import (
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

func WriteResult(result Result, ledgerPath, receiptPath, provenancePath, patchPath string) error {
	ledger, err := encodeJSON(result.Ledger)
	if err != nil {
		return err
	}
	patch, err := encodeJSON(result.Patch)
	if err != nil {
		return err
	}
	receipts, err := generation.EncodeReceiptReport(result.Receipts)
	if err != nil {
		return err
	}
	provenance, err := generation.EncodeArtifactProvenance(result.Provenance)
	if err != nil {
		return err
	}
	paths := []string{ledgerPath, receiptPath, provenancePath, patchPath}
	payloads := [][]byte{ledger, receipts, provenance, patch}
	for index, path := range paths {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, payloads[index], 0o644); err != nil {
			return err
		}
	}
	return nil
}
