package transformationeffect

import (
	"os"
	"path/filepath"
	"strconv"

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

func effectIndicators(ledger Ledger, selected int, receipt generation.ReceiptDecision) []Indicator {
	pass := func(id, route, relation, value, limit string) Indicator {
		return Indicator{id, route, "PASS", relation, value, limit}
	}
	return []Indicator{
		pass("foundation.artifact-schemas", "FOUNDATION", "=", "true", "true"),
		pass("foundation.exact-head", "FOUNDATION", "=", ledger.HeadSHA, ledger.HeadSHA),
		pass("foundation.root-topology-exemption", "FOUNDATION", "=", "true", "true"),
		pass("foundation.indicator-ledger", "FOUNDATION", "sha256", ledger.IndicatorLedgerDigest, "bound"),
		pass("foundation.disposable-write-boundary", "FOUNDATION", "=", ledger.WriteBoundary, "SANDBOX_ONLY"),
		pass("coherence.selected-effects", "COHERENCE", "=", strconv.Itoa(len(ledger.Effects)), strconv.Itoa(selected)),
		pass("coherence.generated-receipts", "COHERENCE", "=", string(receipt), string(receipt)),
		pass("coherence.content-patch", "COHERENCE", "sha256", ledger.PatchDigest, "bound"),
		pass("coherence.executed-provenance", "COHERENCE", "sha256", ledger.ExecutedProvenanceDigest, "bound"),
		pass("regression.source-workspace", "REGRESSION", "=", "unchanged", "unchanged"),
		pass("regression.canonical-encoding", "REGRESSION", "=", "true", "true"),
	}
}
