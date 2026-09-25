package languagesourceexecution

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"
)

func decodeReceipt(raw []byte) (sourceexecution.Receipt, error) {
	var receipt sourceexecution.Receipt
	if err := json.Unmarshal(raw, &receipt); err != nil {
		return receipt, err
	}
	return receipt, sourceexecution.Validate(receipt)
}
