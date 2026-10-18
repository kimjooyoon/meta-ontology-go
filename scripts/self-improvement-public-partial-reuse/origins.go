package main

import (
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicpartialreuse"
)

func sealPartitionReceipt(out, caseID, partition string, binding publicpartialreuse.Binding,
	origin publicpartialreuse.OriginalExecution, observed publicpartialreuse.Metrics, result executionResult,
) (publicpartialreuse.Receipt, string, error) {
	origin.Operation, origin.ResultDigest = publicpartialreuse.ReceiptOperation, result.ResultDigest
	origin.Successful, origin.ExitCode = result.Success, result.ExitCode
	origin.BuildExecutions, origin.TestExecutions = observed.BuildExecutions, observed.TestExecutions
	origin.BuildMS, origin.TestMS, origin.WallMS, origin.PeakRSSKib = observed.BuildMS, observed.TestMS, observed.WallMS, observed.PeakRSSKib
	origin.InvocationID = cache.HashBytes([]byte(binding.GeneratedArtifactDigest + "\x00" + result.ResultDigest + "\x00" + partition)).String()
	receipt := publicpartialreuse.Receipt{Schema: publicpartialreuse.ReceiptSchema, Partition: partition,
		Decision: publicpartialreuse.DecisionClosed, Reason: publicpartialreuse.ReceiptReason, Binding: binding, Original: origin,
		Provenance: publicpartialreuse.Provenance{Operation: publicpartialreuse.Operation, CaseID: caseID, Stage: "v14-authorized"}}
	var err error
	receipt.ReceiptID, err = publicpartialreuse.ReceiptContentDigest(receipt)
	if err != nil {
		return publicpartialreuse.Receipt{}, "", err
	}
	path := filepath.Join(out, partition+".receipt.json")
	if err := publicpartialreuse.WriteReceipt(path, receipt); err != nil {
		return publicpartialreuse.Receipt{}, "", err
	}
	return receipt, path, nil
}
