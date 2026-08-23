package transformationeffect

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

type generationReceipt = generation.IndicatorReceipt

func normalizeSplitGoVerdict(verdict string) string {
	switch strings.ToUpper(verdict) {
	case "PASS":
		return "PASS"
	case "FAIL":
		return "FAIL"
	default:
		return "UNKNOWN"
	}
}

func newSplitGoReceipt(indicatorID, verdict, digest string) (generationReceipt, error) {
	receipt := reflect.New(reflect.TypeOf(generation.IndicatorReceipt{})).Elem()
	if err := setSplitGoReceiptField(receipt, []string{"IndicatorID", "ID"}, indicatorID); err != nil {
		return generationReceipt{}, err
	}
	if err := setSplitGoReceiptField(receipt, []string{"Verdict"}, verdict); err != nil {
		return generationReceipt{}, err
	}
	if err := setSplitGoReceiptField(receipt, []string{"EvidenceDigest"}, digest); err != nil {
		return generationReceipt{}, err
	}
	return receipt.Interface().(generation.IndicatorReceipt), nil
}

func setSplitGoReceiptField(receipt reflect.Value, names []string, value string) error {
	for _, name := range names {
		field := receipt.FieldByName(name)
		if !field.IsValid() || !field.CanSet() || field.Kind() != reflect.String {
			continue
		}
		field.SetString(value)
		return nil
	}
	return fmt.Errorf("generation receipt lacks writable string field %s", strings.Join(names, "/"))
}

func splitGoDigestHex(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}
