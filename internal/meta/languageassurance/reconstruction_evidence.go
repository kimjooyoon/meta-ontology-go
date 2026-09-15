package languageassurance

import (
	"reflect"
	"sort"
)

func detectRawReconstructionMismatches(receipts []RawReconstructionReceipt, expected RawReconstructionReceipt) []Finding {
	var findings []Finding
	for _, receipt := range receipts {
		if !reflect.DeepEqual(receipt, expected) {
			findings = append(findings, Finding{
				MetricID: MetricRawReconstruction, PathID: "reconstruction/" + receipt.VerifierID,
				EvidenceID: "raw_transaction", VerifierID: receipt.VerifierID,
				ExpectedDigest: digest(expected), ObservedDigest: digest(receipt),
			})
		}
	}
	return findings
}

func observeRawReconstructions(receipts []RawReconstructionReceipt, mismatches int) (*int, *int) {
	if len(receipts) != 1 {
		return nil, nil
	}
	bps, paths := (1-mismatches)*10000, mismatches
	return &bps, &paths
}

func rawTransactionDigest(transaction Transaction) string {
	copy := transaction
	copy.RawReconstructions = nil
	return digest(copy)
}

func sortFindings(findings []Finding) {
	sort.Slice(findings, func(i, j int) bool {
		return findings[i].MetricID+findings[i].PathID < findings[j].MetricID+findings[j].PathID
	})
}
