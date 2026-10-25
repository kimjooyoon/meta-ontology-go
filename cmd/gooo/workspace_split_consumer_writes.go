package main

import (
	"bytes"
	"fmt"
)

// The source pair is installed once. Each additional owner contributes exactly
// its manifest edit; every proposed source and pre-state must agree first.
func sourceSplitAllWrites(receipt workspaceSourceSplitReceipt) ([]atomicWrite, error) {
	writes, err := sourceSplitWrites(receipt)
	if err != nil {
		return nil, err
	}
	for _, consumer := range receipt.Consumers {
		candidate, err := sourceSplitWrites(consumer.receipt())
		if err != nil {
			return nil, err
		}
		for i := range 2 {
			if candidate[i].path != writes[i].path || !bytes.Equal(candidate[i].data, writes[i].data) ||
				!sameOutputSnapshot(*candidate[i].before, *writes[i].before) {
				return nil, fmt.Errorf("consumer workspace %q disagrees on the source publication", consumer.ManifestPath)
			}
		}
		for _, write := range writes {
			shared, err := samePublicationEntry(write.path, candidate[2].path)
			if err != nil {
				return nil, err
			}
			if write.path == candidate[2].path || shared {
				return nil, fmt.Errorf("consumer manifest aliases another publication target")
			}
		}
		writes = append(writes, candidate[2])
	}
	return writes, nil
}
