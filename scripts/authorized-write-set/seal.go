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
		{ID: "guardrail.unknown", Class: "GUARDRAIL", ProofChoice: "FOUNDATION", MetaOperation: metaOperation,
			Observed: c.Unknowns, Expected: 0, Satisfied: c.Unknowns == 0},
		{ID: "guardrail.untracked", Class: "GUARDRAIL", ProofChoice: "REGRESSION", MetaOperation: metaOperation,
			Observed: c.UntrackedPaths, Expected: 0, Satisfied: c.UntrackedPaths == 0},
	}
	report.Proofs = []proof{
		{Choice: "FOUNDATION", MetaOperation: "validate-rewrite-source-receipts", Passed: c.SourceReceipts == 2 && c.Unknowns == 0},
		{Choice: "COHERENCE", MetaOperation: metaOperation, Passed: report.Exact},
		{Choice: "REGRESSION", MetaOperation: "reject-untracked-rewrites", Passed: c.UntrackedPaths == 0},
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
