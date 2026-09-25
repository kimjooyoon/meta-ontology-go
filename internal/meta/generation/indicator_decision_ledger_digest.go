package generation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func indicatorDecisionLedgerDigest(ledger IndicatorDecisionLedger) (string, error) {
	ledger.Digest = ""
	payload, err := json.Marshal(ledger)
	if err != nil {
		return "", fmt.Errorf("marshal indicator decision ledger digest material: %w", err)
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func validIndicatorDecisionLedgerDigest(value string) bool {
	return strings.HasPrefix(value, "sha256:") && validDigest(strings.TrimPrefix(value, "sha256:"))
}

func ensureIndicatorLedgerEOF(decoder *json.Decoder) error {
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("unexpected trailing JSON value")
		}
		return err
	}
	return nil
}
