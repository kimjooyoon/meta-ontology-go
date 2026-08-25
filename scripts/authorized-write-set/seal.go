package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func seal(report evidence) evidence {
	c := report.Coordinates
	report.Indicators = []indicator{
		{ID: "source.receipts", Class: "OUTCOME", ProofChoice: "FOUNDATION", MetaOperation: metaOperation,
			Observed: c.SourceReceipts, Expected: c.SourceReceiptsTotal, Satisfied: c.SourceReceipts == c.SourceReceiptsTotal},
		{ID: "write-set.exact", Class: "OUTCOME", ProofChoice: "COHERENCE", MetaOperation: metaOperation,
			Observed: boolInt(report.Exact), Expected: 1, Satisfied: report.Exact},
		{ID: "write-set.created", Class: "DRIVER", ProofChoice: "COHERENCE", MetaOperation: metaOperation,
			Observed: len(report.ObservedCreated), Expected: len(report.ExpectedCreated),
			Satisfied: equalPaths(report.ExpectedCreated, report.ObservedCreated)},
		{ID: "guardrail.unknown", Class: "GUARDRAIL", ProofChoice: "FOUNDATION", MetaOperation: metaOperation,
			Observed: c.Unknowns, Expected: 0, Satisfied: c.Unknowns == 0},
		{ID: "guardrail.unclassified-untracked", Class: "GUARDRAIL", ProofChoice: "REGRESSION", MetaOperation: metaOperation,
			Observed: c.UnclassifiedPaths, Expected: 0, Satisfied: c.UnclassifiedPaths == 0},
	}
	report.Proofs = []proof{
		{Choice: "FOUNDATION", MetaOperation: "validate-rewrite-source-receipts", Passed: c.SourceReceipts == c.SourceReceiptsTotal && c.Unknowns == 0},
		{Choice: "COHERENCE", MetaOperation: metaOperation, Passed: report.Exact},
		{Choice: "REGRESSION", MetaOperation: "reject-unclassified-untracked-rewrites", Passed: c.UnclassifiedPaths == 0},
	}
	unsigned := report
	unsigned.Digest = ""
	data, _ := json.Marshal(unsigned)
	sum := sha256.Sum256(data)
	report.Digest = "sha256:" + hex.EncodeToString(sum[:])
	return report
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
