//go:build partial_reuse_example
// +build partial_reuse_example

package partialreuseexample

import "testing"

func TestOrdersPartition(t *testing.T) {
	CreateReceipt(Order{}, SharedContract{})
}

func TestInventoryPartition(t *testing.T) {
	SnapshotInventory(StockItem{}, SharedContract{})
}
