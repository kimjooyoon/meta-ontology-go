package writeset

func Compare(subjectSHA, denominatorDigest string, before, after Snapshot, declared []string) Receipt {
	base := Receipt{Schema: ReceiptSchema, MetricID: MetricID, MetaOperation: MetaOperation,
		ProofChoice: "REGRESSION", ObserverID: "gooo-independent-write-set-observer-v1",
		SubjectSHA: subjectSHA, DenominatorDigest: denominatorDigest}
	declared, declaredValid := normalizePaths(declared)
	if subjectSHA == "" || denominatorDigest == "" || !declaredValid || !validSnapshot(before) || !validSnapshot(after) {
		base.Decision, base.Reason, base.Resolution = "FAIL_CLOSED", "WRITE_SET_EVIDENCE_UNKNOWN", "INVARIANT_ONLY"
		return bindReceipt(base)
	}
	observed := changedPaths(before, after)
	mismatches := symmetricDifference(declared, observed)
	base.BeforeDigest, base.AfterDigest = before.RootDigest, after.RootDigest
	base.DeclaredPaths, base.ObservedPaths, base.MismatchPaths = declared, observed, mismatches
	base.Resolution = "EXACT"
	base.Summary = Summary{DeclaredPaths: len(declared), ObservedPaths: len(observed), MismatchPaths: len(mismatches)}
	if len(mismatches) == 0 {
		base.Decision, base.Reason, base.Summary.ExactnessBPS = "PASS", "WRITE_SET_EXACT", 10000
	} else {
		base.Decision, base.Reason = "BLOCK", "WRITE_SET_MISMATCH"
	}
	return bindReceipt(base)
}
