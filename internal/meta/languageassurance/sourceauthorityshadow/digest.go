package sourceauthorityshadow

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/sourceauthorityeval"
)

func seal(receipt Receipt) Receipt {
	receipt.ReceiptDigest = ""
	raw, err := json.Marshal(receipt)
	if err != nil {
		receipt.Observation = "ERROR"
		receipt.Resolution = "INVARIANT_ONLY"
		receipt.Enforcement = "BLOCK"
		receipt.Reason = "SOURCE_AUTHORITY_SHADOW_ENCODING_ERROR"
		return receipt
	}
	receipt.ReceiptDigest = sourceauthorityeval.DigestBytes(raw)
	return receipt
}
