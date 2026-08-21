package transformationeffect

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

func hashBytes(payload []byte) string {
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}

func hashJSON(value any) string {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return hashBytes(payload)
}

func hashPair(left, right string) string { return hashBytes([]byte(left + "\x00" + right)) }

func validSHA(value string) bool {
	if len(value) != 40 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func sealLedger(ledger Ledger) Ledger {
	ledger.EffectDigest = hashJSON(ledger.Effects)
	ledger.SemanticDigest = hashJSON([]any{ledger.HeadSHA, ledger.Decision, ledger.IndicatorLedgerDigest,
		ledger.SourceTreeBefore, ledger.SandboxTreeAfter, ledger.EffectDigest, ledger.PatchDigest,
		ledger.GeneratedReceiptReportDigest, ledger.ExecutedProvenanceDigest, ledger.Indicators})
	ledger.LedgerDigest, ledger.ReplayDigest = "", ""
	ledger.LedgerDigest = hashJSON(ledger)
	ledger.ReplayDigest = hashPair(ledger.InputDigest, ledger.LedgerDigest)
	return ledger
}

func validateLedger(ledger Ledger) error {
	if ledger.Schema != ledgerSchema || ledger.Status != "BOUND" || !ledger.RootTopologyExempt ||
		ledger.PromotionAuthorized || !validSHA(ledger.BaseSHA) || !validSHA(ledger.HeadSHA) {
		return fmt.Errorf("transformation ledger is not bound")
	}
	for _, indicator := range ledger.Indicators {
		if indicator.Verdict != "PASS" {
			return fmt.Errorf("indicator %s did not pass", indicator.ID)
		}
	}
	expected := ledger
	expected.EffectDigest, expected.SemanticDigest, expected.LedgerDigest, expected.ReplayDigest = "", "", "", ""
	if !reflect.DeepEqual(sealLedger(expected), ledger) {
		return fmt.Errorf("transformation ledger digest diverged")
	}
	return nil
}

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

func encodeJSON(value any) ([]byte, error) {
	payload, err := json.MarshalIndent(value, "", "  ")
	return append(payload, '\n'), err
}
