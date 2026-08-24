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

func validSnapshot(snapshot Snapshot) bool {
	if snapshot.Schema != SnapshotSchema || snapshot.RootDigest != digestEntries(snapshot.Entries) { return false }
	previous := ""
	for index, entry := range snapshot.Entries {
		paths, valid := normalizePaths([]string{entry.Path})
		if !valid || len(paths) != 1 || paths[0] != entry.Path || entry.Digest == "" || (index > 0 && entry.Path <= previous) { return false }
		previous = entry.Path
	}
	return true
}

func changedPaths(before, after Snapshot) []string {
	left, right := make(map[string]Entry, len(before.Entries)), make(map[string]Entry, len(after.Entries))
	for _, entry := range before.Entries { left[entry.Path] = entry }
	for _, entry := range after.Entries { right[entry.Path] = entry }
	candidates := make([]string, 0, len(left)+len(right))
	for candidate := range left { if right[candidate] != left[candidate] { candidates = append(candidates, candidate) } }
	for candidate := range right { if _, exists := left[candidate]; !exists { candidates = append(candidates, candidate) } }
	result, _ := normalizePaths(candidates)
	return result
}
