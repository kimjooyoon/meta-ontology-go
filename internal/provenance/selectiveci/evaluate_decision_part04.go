package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func receiptsFor(input Input) map[semantic.ID]CommandReceipt {
	receipts := make(map[semantic.ID]CommandReceipt, len(input.CommandReceipts))
	for _, receipt := range input.CommandReceipts {
		receipts[receipt.CommandID] = receipt
	}
	return receipts
}
